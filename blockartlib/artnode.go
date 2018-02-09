package blockartlib

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"net/rpc"
	"time"

	crypto "../crypto"
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
	inkCost, err := calculateInkCost(shapeSvgString, fill, stroke)
	if err != nil {
		return "", "", 0, err
	}

	publicKey, err := crypto.MarshalPublic(&a.privKey.PublicKey)
	if err != nil {
		return "", "", 0, err
	}

	shape := Shape{
		Svg:    shapeSvgString,
		Fill:   fill,
		Stroke: stroke,
	}

	shapeHash, err = crypto.Hash(shape)
	if err != nil {
		return "", "", 0, err
	}

	id := time.Now().String()

	args := Operation{
		OpType:      ADD,
		Shape:       shape,
		OpSig:       OpSig{},
		PubKey:      publicKey,
		InkCost:     inkCost,
		ShapeHash:   shapeHash,
		ValidateNum: validateNum,
		Id:          id,
	}

	bytes, err := json.Marshal(args)
	if err != nil {
		return "", "", 0, err
	}

	r, s, err := crypto.Sign(bytes, a.privKey)
	if err != nil {
		return "", "", 0, err
	}

	args.OpSig = OpSig{r, s}

	var resp AddShapeResponse
	err = a.client.Call("InkMinerRPC.AddShape", args, &resp)
	//TODO: retrieve blockHash, inkRemaining from call to ink miner to add shape

	if err != nil {
		return "", "", 0, err
	}

	return shapeHash, resp.BlockHash, resp.InkRemaining, nil
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (a *ArtNode) GetSvgString(shapeHash string) (svgString string, err error) {
	var resp string

	err = a.client.Call("InkMinerRPC.GetSvgString", shapeHash, &resp)
	if err != nil {
		return "", err
	}

	svgString = resp
	return
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetInk() (inkRemaining uint32, err error) {
	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutOfBoundsError
// - ShapeOverlapError
func (a *ArtNode) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	publicKey, err := crypto.MarshalPublic(&a.privKey.PublicKey)
	if err != nil {
		return 0, err
	}

	id := time.Now().String()

	args := Operation{
		OpType:      DELETE,
		OpSig:       OpSig{},
		PubKey:      publicKey,
		ShapeHash:   shapeHash,
		ValidateNum: validateNum,
		Id:          id,
	}

	bytes, err := json.Marshal(args)
	if err != nil {
		return 0, err
	}

	r, s, err := crypto.Sign(bytes, a.privKey)
	if err != nil {
		return 0, err
	}

	args.OpSig = OpSig{r, s}

	var resp uint32

	err = a.client.Call("InkMinerRPC.DeleteShape", args, &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetShapes(blockHash string) (shapeHashes []string, err error) {
	var resp GetShapesResponse

	err = a.client.Call("InkMinerRPC.GetShapes", blockHash, &resp)
	if err != nil {
		return nil, err
	}

	shapeHashes = resp.ShapeHashes
	return
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetGenesisBlock() (blockHash string, err error) {
	var resp string

	err = a.client.Call("InkMinerRPC.GetGenesisBlock", "", &resp)
	if err != nil {
		return "", err
	}

	blockHash = resp
	return
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetChildren(blockHash string) (blockHashes []string, err error) {
	var resp GetChildrenResponse

	err = a.client.Call("InkMinerRPC.GetChildrenBlocks", blockHash, &resp)
	if err != nil {
		return nil, err
	}

	blockHashes = resp.BlockHashes
	return
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (a *ArtNode) CloseCanvas() (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMinerRPC.GetInk", args, &resp)
	var resp uint32

	err = a.client.Call("InkMinerRPC.GetInk", "", &resp)
	if err != nil {
		return 0, err
	}

	inkRemaining = resp
	return
}

// HELPERS
// Gets the ink cost of a particular operation
func calculateInkCost(shapeSvgString string, fill string, stroke string) (cost uint32, err error) {
	return 0, errors.New("Not implemented")
}

// Checks if valid svg string
// - InvalidShapeSvgString Error
// - ShapeSvgStringTooLong Error
func svgStringValidityCheck(svgString string) (err error) {
	return errors.New("Not implemented")
}
