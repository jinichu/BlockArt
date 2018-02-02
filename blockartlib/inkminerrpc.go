package blockartlib

import (
	"crypto/ecdsa"
	"errors"
)

func (i *InkMiner) InitConnection(req *ecdsa.PublicKey, resp *CanvasSettings) error {
	// Confirm that this is the right public key for this InkMiner
	if *req != i.privKey.Public() {
		return errors.New("Public key is incorrect for this InkMiner")
	}
	*resp = i.settings.canvasSettings
	return nil
}

func (i *InkMiner) AddShape(req *Operation, resp *AddShapeResponse) error {
	// TODO: Check if this operation uses a legal amount of ink and fail if not
	if err := i.floodOperation(req); err != nil {
		return err
	}
	block, err := i.mineBlock(req)
	if err != nil {
		return err
	}
	addShapeResponse := AddShapeResponse{}
	// TODO: Compute blockHash and amount of ink remaining
	_ = block
	*resp = addShapeResponse
	return nil
}

func (i *InkMiner) GetSvgString(req *string, resp *string) error {
	if _, ok := i.shapes[*req]; ok {
		*resp = i.shapes[*req]
		return nil
	}
	return InvalidShapeHashError(*req)
}

func (i *InkMiner) GetInk(req *string, resp *uint32) error {
	*resp = i.inkAmount
	return nil
}

func (i *InkMiner) DeleteShape(req *Operation, resp *uint32) error {

	*resp = i.inkAmount
	return nil
}
