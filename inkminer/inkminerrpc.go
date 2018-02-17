package inkminer

import (
	"errors"
	"fmt"
	"log"

	"../blockartlib"
	server "../server"
)

type InkMinerRPC struct {
	i *InkMiner
}

func (i *InkMinerRPC) TestConnection(req *string, resp *bool) error {
	*resp = true
	return nil
}

func (i *InkMinerRPC) InitConnection(req blockartlib.InitConnectionRequest, resp *server.CanvasSettings) error {
	// Confirm that this is the right public key for this InkMiner
	if req.PublicKey != i.i.publicKey {
		return blockartlib.DisconnectedError(i.i.Addr())
	}
	*resp = i.i.settings.CanvasSettings
	return nil
}

func (i *InkMiner) testOperation(op blockartlib.Operation) error {
	if err := i.validateOp(op); err != nil {
		return err
	}

	block := i.currentHead()
	blockHash, err := block.Hash()
	if err != nil {
		return err
	}

	state, err := i.CalculateState(block)
	if err != nil {
		return err
	}

	log.Printf("CalculateState Passed! %+v", op)

	testBlock := blockartlib.Block{
		PrevBlock: blockHash,
		BlockNum:  block.BlockNum + 1,
		PubKey:    i.privKey.PublicKey,
		Records:   []blockartlib.Operation{op},
	}
	if _, err := i.TransformState(state, testBlock); err != nil {
		log.Printf("TransformState failed!")
		return err
	}

	log.Printf("TransformState Passed! %+v", op)

	return nil
}

func (i *InkMinerRPC) AddShape(req *blockartlib.Operation, resp *blockartlib.AddShapeResponse) error {

	if err := i.i.testOperation(*req); err != nil {
		return err
	}

	if err := i.i.addOperation(*req); err != nil {
		return fmt.Errorf("add operation error: %+v", err)
	}

	opHash, err := req.Hash()
	if err != nil {
		return err
	}
	blockHash, err := i.i.waitForValidateNum(opHash, req.ValidateNum)
	if err != nil {
		return err
	}

	state, err := i.i.CalculateState(i.i.currentHead())
	if err != nil {
		return err
	}

	pubKey, err := req.PubKeyString()
	if err != nil {
		return err
	}

	*resp = blockartlib.AddShapeResponse{
		BlockHash:    blockHash,
		InkRemaining: state.inkLevels[pubKey],
	}
	return nil
}

func (i *InkMinerRPC) GetSvgString(req *string, resp *string) error {
	tryBlocks := []blockartlib.Block{i.i.currentHead()}
	i.i.mu.Lock()
	for _, block := range i.i.mu.blockchain {
		tryBlocks = append(tryBlocks, block)
	}
	i.i.mu.Unlock()

	for _, block := range tryBlocks {
		state, err := i.i.CalculateState(block)
		if err != nil {
			i.i.log.Printf("CalculateState error: %+v", err)
			continue
		}

		if shape, ok := state.shapes[*req]; ok {
			*resp = shape.SvgString()
			return nil
		}
	}
	return blockartlib.InvalidShapeHashError(*req)
}

func (i *InkMinerRPC) GetInk(req *string, resp *uint32) error {
	state, err := i.i.CalculateState(i.i.currentHead())
	if err != nil {
		return err
	}

	*resp = state.inkLevels[i.i.publicKey]
	return nil
}

func (i *InkMinerRPC) DeleteShape(req *blockartlib.Operation, resp *uint32) error {
	if err := i.i.testOperation(*req); err != nil {
		return err
	}
	if err := i.i.addOperation(*req); err != nil {
		return err
	}

	opHash, err := req.Hash()
	if err != nil {
		return err
	}

	if _, err := i.i.waitForValidateNum(opHash, req.ValidateNum); err != nil {
		return err
	}

	state, err := i.i.CalculateState(i.i.currentHead())
	if err != nil {
		return err
	}

	pubKey, err := req.PubKeyString()
	if err != nil {
		return err
	}

	*resp = state.inkLevels[pubKey]
	return nil
}

func (i *InkMinerRPC) GetShapes(req *string, resp *blockartlib.GetShapesResponse) error {

	if *req == i.i.settings.GenesisBlockHash {
		return nil
	}

	block, ok := i.i.GetBlock(*req)
	if !ok {
		return blockartlib.InvalidBlockHashError(*req)
	}

	getShapesResponse := blockartlib.GetShapesResponse{}
	for i := 0; i < len(block.Records); i++ {
		op := block.Records[i]
		hash, err := op.Hash()
		if err != nil {
			return err
		}
		getShapesResponse.ShapeHashes = append(getShapesResponse.ShapeHashes, hash)
	}
	*resp = getShapesResponse
	return nil
}

func (i *InkMinerRPC) GetGenesisBlock(req *string, resp *string) error {
	*resp = i.i.settings.GenesisBlockHash
	return nil
}

func (i *InkMinerRPC) GetChildrenBlocks(req *string, resp *blockartlib.GetChildrenResponse) error {
	_, ok := i.i.GetBlock(*req)
	if ok || *req == i.i.settings.GenesisBlockHash {
		getChildrenResponse := blockartlib.GetChildrenResponse{}

		i.i.mu.Lock()
		defer i.i.mu.Unlock()

		for hash, block := range i.i.mu.blockchain {
			if block.PrevBlock == *req {
				getChildrenResponse.BlockHashes = append(getChildrenResponse.BlockHashes, hash)
			}
		}

		*resp = getChildrenResponse
		return nil
	}
	return blockartlib.InvalidBlockHashError(*req)
}

func (i *InkMiner) waitForValidateNum(opHash string, validateNum uint8) (string, error) {
	validateNumWaiter := ValidateNumWaiter{
		done:        make(chan string),
		err:         make(chan error),
		validateNum: validateNum,
	}

	defer func() {
		close(validateNumWaiter.done)
		close(validateNumWaiter.err)
	}()

	i.mu.Lock()
	i.mu.validateNumMap[opHash] = validateNumWaiter
	i.mu.Unlock()

	select {
	case <-i.stopper.ShouldStop():
		return "", errors.New("stopping")
	case blockHash := <-validateNumWaiter.done:
		return blockHash, nil
	case err := <-validateNumWaiter.err:
		return "", err
	}
}
