package inkminer

import (
	"errors"

	"../blockartlib"
)

type InkMinerRPC struct {
	i *InkMiner
}

func (i *InkMinerRPC) InitConnection(req blockartlib.InitConnectionRequest, resp *blockartlib.CanvasSettings) error {
	// Confirm that this is the right public key for this InkMiner
	if req.PublicKey != i.i.publicKey {
		return errors.New("Public key is incorrect for this InkMiner")
	}
	*resp = i.i.settings.CanvasSettings
	return nil
}

func (i *InkMinerRPC) AddShape(req *blockartlib.Operation, resp *blockartlib.AddShapeResponse) error {
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}

	if i.i.states[blockHash].inkLevels[req.PubKey] < req.InkCost {
		return blockartlib.InsufficientInkError(i.i.states[blockHash].inkLevels[req.PubKey])
	}
	if err := i.i.floodOperation(*req); err != nil {
		return err
	}
	if err := i.i.mineBlock(*req); err != nil {
		return err
	}
	// TODO: InkMiner.currentHead should have the latest block. Compute hash and return this as blockHash
	// TODO: InkMiner.states should be updated to have the current state too

	blockHash, err = i.i.currentHead.Hash()
	if err != nil {
		return err
	}
	addShapeResponse := blockartlib.AddShapeResponse{
		BlockHash:    blockHash,
		InkRemaining: i.i.states[blockHash].inkLevels[req.PubKey],
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
		*resp = i.i.states[blockHash].shapes[*req]
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
	if err := i.i.floodOperation(*req); err != nil {
		return err
	}
	if err := i.i.mineBlock(*req); err != nil {
		return err
	}
	blockHash, err := i.i.currentHead.Hash()
	if err != nil {
		return err
	}
	*resp = i.i.states[blockHash].inkLevels[req.PubKey]
	return nil
}

func (i *InkMinerRPC) GetShapes(req *string, resp *blockartlib.GetShapesResponse) error {
	if block, ok := i.i.GetBlock(*req); ok {
		getShapesResponse := blockartlib.GetShapesResponse{}
		for i := 0; i < len(block.Records); i++ {
			getShapesResponse.ShapeHashes = append(getShapesResponse.ShapeHashes, block.Records[i].ShapeHash)
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

		blockHash, err := i.i.currentHead.Hash()
		if err != nil {
			return err
		}

		for {
			if *req == blockHash {
				break
			}
			getChildrenResponse.BlockHashes = append(getChildrenResponse.BlockHashes, blockHash)
			block, ok := i.i.GetBlock(blockHash)
			if !ok {
				break
			}
			blockHash = block.PrevBlock
		}
		*resp = getChildrenResponse
		return nil
	}
	return blockartlib.InvalidBlockHashError(*req)
}
