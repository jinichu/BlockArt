package inkminer

import (
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

func (i *InkMinerRPC) AddShape(req *blockartlib.Operation, resp *blockartlib.AddShapeResponse) error {
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}

	pubKey, err := req.PubKeyString()
	if err != nil {
		return err
	}

	if i.i.states[blockHash].inkLevels[pubKey] < req.InkCost {
		return blockartlib.InsufficientInkError(i.i.states[blockHash].inkLevels[pubKey])
	}
	if err := i.i.addOperation(*req); err != nil {
		return err
	}
	// TODO: InkMiner.currentHead should have the latest block. Compute hash and return this as blockHash
	// TODO: InkMiner.states should be updated to have the current state too
	opHash, err := req.Hash()
	if err != nil {
		return err
	}
	validateNum := req.ValidateNum
	blockHash = i.i.waitForValidateNum(opHash, validateNum)

	addShapeResponse := blockartlib.AddShapeResponse{
		BlockHash:    blockHash,
		InkRemaining: i.i.states[blockHash].inkLevels[pubKey],
	}
	*resp = addShapeResponse
	return nil
}

func (i *InkMinerRPC) GetSvgString(req *string, resp *string) error {
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}

	if _, ok := i.i.states[blockHash].shapes[*req]; ok {
		shape := i.i.states[blockHash].shapes[*req]
		*resp = shape.SvgString()
		return nil
	}
	return blockartlib.InvalidShapeHashError(*req)
}

func (i *InkMinerRPC) GetInk(req *string, resp *uint32) error {
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}
	*resp = i.i.states[blockHash].inkLevels[i.i.publicKey]
	return nil
}

func (i *InkMinerRPC) DeleteShape(req *blockartlib.Operation, resp *uint32) error {
	if err := i.i.addOperation(*req); err != nil {
		return err
	}
	// TODO: wait for ValidateNum
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}

	pubKey, err := req.PubKeyString()
	if err != nil {
		return err
	}

	*resp = i.i.states[blockHash].inkLevels[pubKey]
	return nil
}

func (i *InkMinerRPC) GetShapes(req *string, resp *blockartlib.GetShapesResponse) error {
	if block, ok := i.i.GetBlock(*req); ok {
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
	return blockartlib.InvalidBlockHashError(*req)
}

func (i *InkMinerRPC) GetGenesisBlock(req *string, resp *string) error {
	*resp = i.i.settings.GenesisBlockHash
	return nil
}

func (i *InkMinerRPC) GetChildrenBlocks(req *string, resp *blockartlib.GetChildrenResponse) error {
	if _, ok := i.i.GetBlock(*req); ok {
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

func (i *InkMiner) waitForValidateNum(opHash string, validateNum uint8) string {
	i.mu.Lock()
	validateNumWaiter := ValidateNumWaiter{
		done:        make(chan string),
		validateNum: validateNum,
	}
	i.mu.validateNumMap[opHash] = validateNumWaiter
	i.mu.Unlock()

	blockHash := <-validateNumWaiter.done

	return blockHash
}
