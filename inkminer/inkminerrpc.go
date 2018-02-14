package inkminer

import (
	"fmt"

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
	block := i.i.currentHead()
	blockHash, err := block.Hash()
	if err != nil {
		return err
	}

	pubKey, err := req.PubKeyString()
	if err != nil {
		return err
	}

	state, err := i.i.CalculateState(block)
	if err != nil {
		return err
	}

	inkLevel := state.inkLevels[pubKey]
	if inkLevel < req.InkCost {
		return blockartlib.InsufficientInkError(state.inkLevels[pubKey])
	}
	if err := i.i.addOperation(*req); err != nil {
		return fmt.Errorf("add operation error: %+v", err)
	}
	// TODO: InkMiner.currentHead should have the latest block. Compute hash and return this as blockHash
	// TODO: InkMiner.states should be updated to have the current state too

	block = i.i.currentHead()

	blockHash, err = block.Hash()
	if err != nil {
		return err
	}

	state, err = i.i.CalculateState(block)
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
	state, err := i.i.CalculateState(i.i.currentHead())
	if err != nil {
		return err
	}

	if shape, ok := state.shapes[*req]; ok {
		*resp = shape.SvgString()
		return nil
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
	if err := i.i.addOperation(*req); err != nil {
		return err
	}

	// TODO: wait for ValidateNum

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
