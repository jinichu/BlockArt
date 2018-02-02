package inkminer

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"

	"../blockartlib"
)

func (i *InkMiner) InitConnection(req *ecdsa.PublicKey, resp *blockartlib.CanvasSettings) error {
	// Confirm that this is the right public key for this InkMiner
	if *req != i.privKey.Public() {
		return errors.New("Public key is incorrect for this InkMiner")
	}
	*resp = i.settings.CanvasSettings
	return nil
}

func (i *InkMiner) AddShape(req *blockartlib.Operation, resp *blockartlib.AddShapeResponse) error {
	// TODO: Check if this operation uses a legal amount of ink and fail if not
	if err := i.floodOperation(req); err != nil {
		return err
	}
	if err := i.mineBlock(*req); err != nil {
		return err
	}
	addShapeResponse := blockartlib.AddShapeResponse{}
	*resp = addShapeResponse
	return nil
}

func (i *InkMiner) GetSvgString(req *string, resp *string) error {
	blockHash := "" // TODO: Compute hash of currentHead
	if _, ok := i.states[blockHash].shapes[*req]; ok {
		*resp = i.states[blockHash].shapes[*req]
		return nil
	}
	return blockartlib.InvalidShapeHashError(*req)
}

func (i *InkMiner) GetInk(req *string, resp *uint32) error {
	*resp = i.inkAmount
	return nil
}

func (i *InkMiner) DeleteShape(req *blockartlib.Operation, resp *uint32) error {
	if err := i.floodOperation(req); err != nil {
		return err
	}
	if err := i.mineBlock(*req); err != nil {
		return err
	}
	*resp = i.inkAmount
	return nil
}

func (i *InkMiner) GetShapes(req *string, resp *blockartlib.GetShapesResponse) error {
	if _, ok := i.blockchain[*req]; ok {
		block := i.blockchain[*req]
		getShapesResponse := blockartlib.GetShapesResponse{}
		for i := 0; i < len(block.Records); i++ {
			getShapesResponse.ShapeHashes = append(getShapesResponse.ShapeHashes, block.Records[i].ShapeHash)
		}
		*resp = getShapesResponse
		return nil
	}
	return blockartlib.InvalidBlockHashError(*req)
}

func (i *InkMiner) GetGenesisBlock(req *string, resp *string) error {
	*resp = i.settings.GenesisBlockHash
	return nil
}

func (i *InkMiner) GetChildrenBlocks(req *string, resp *blockartlib.GetChildrenResponse) error {
	if _, ok := i.blockchain[*req]; ok {
		getChildrenResponse := blockartlib.GetChildrenResponse{}
		bytes, err := json.Marshal(i.currentHead)
		if err != nil {
			return nil
		}
		blockHash := string(bytes) // TODO: Compute block hash properly
		for {
			if *req == blockHash {
				break
			}
			getChildrenResponse.BlockHashes = append(getChildrenResponse.BlockHashes, blockHash)
			blockHash = i.blockchain[blockHash].PrevBlock
		}
		*resp = getChildrenResponse
		return nil
	}
	return blockartlib.InvalidBlockHashError(*req)
}
