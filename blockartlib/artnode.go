package blockartlib

import (
	"crypto/ecdsa"
	"errors"
	"net/rpc"
)

type ArtNode struct {
	client  *rpc.Client      // RPC client to connect to the InkMiner
	privKey ecdsa.PrivateKey // Pub/priv key pair of this ArtNode
}

// Adds a new shape to the canvas.
// Can return the following errors:
// - DisconnectedError
// - InsufficientInkError
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
func (a *ArtNode) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	// TODO: client.Call("InkMiner.AddShape", args, &resp)
	return "", "", 0, errors.New("Not implemented")
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (a *ArtNode) GetSvgString(shapeHash string) (svgString string, err error) {
	// TODO: client.Call("InkMiner.GetSvgString", args, &resp)
	return "", errors.New("Not implemented")
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetInk() (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMiner.GetInk", args, &resp)
	return 0, errors.New("Not implemented")
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutOfBoundsError
// - ShapeOverlapError
func (a *ArtNode) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMiner.DeleteShape", args, &resp)
	return 0, errors.New("Not implemented")
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetShapes(blockHash string) (shapeHashes []string, err error) {
	// TODO: client.Call("InkMiner.GetShapes", args, &resp)
	return nil, errors.New("Not implemented")
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetGenesisBlock() (blockHash string, err error) {
	// TODO: client.Call("InkMiner.GetGenesisBlock", args, &resp)
	return "", errors.New("Not implemented")
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetChildren(blockHash string) (blockHashes []string, err error) {
	// TODO: client.Call("InkMiner.GetChildrenBlocks", args, &resp)
	return nil, errors.New("Not implemented")
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (a *ArtNode) CloseCanvas() (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMiner.GetInk", args, &resp)
	return 0, errors.New("Not implemented")
}
