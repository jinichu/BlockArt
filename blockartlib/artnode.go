package blockartlib

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"net/rpc"

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
	r, s, err := sign([]byte(string(ADD)), a.privKey)
	inkCost, err := calculateInkCost(shapeSvgString, fill, stroke)
	shapeHash, err = crypto.Hash(shapeSvgString)

	if err != nil {
		return "", "", 0, err
	}

	args := Operation{
		OpType:      ADD,
		Shape:       shapeSvgString,
		OpSig:       OpSig{r, s},
		PubKey:      a.privKey.PublicKey,
		InkCost:     inkCost,
		ShapeHash:   shapeHash,
		ValidateNum: validateNum,
	}

	var resp AddShapeResponse
	err = a.client.Call("InkMinerRPC.AddShape", args, &resp)
	//TODO: retrieve blockHash, inkRemaining from call to ink miner to add shape

	if err != nil {
		return "", "", 0, err
	}

	return shapeHash, resp.BlockHash, resp.InkRemaining, nil
}

func sign(operation []byte, privKey ecdsa.PrivateKey) (signedR, signedS string, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, operation)
	if err != nil {
		return "", "", err
	}

	signedR = fmt.Sprint(r)
	signedS = fmt.Sprint(s)
	return
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (a *ArtNode) GetSvgString(shapeHash string) (svgString string, err error) {
	// TODO: client.Call("InkMinerRPC.GetSvgString", args, &resp)
	return "", errors.New("Not implemented")
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetInk() (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMinerRPC.GetInk", args, &resp)

	return 0, errors.New("Not implemented")
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutOfBoundsError
// - ShapeOverlapError
func (a *ArtNode) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMinerRPC.DeleteShape", args, &resp)
	return 0, errors.New("Not implemented")
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetShapes(blockHash string) (shapeHashes []string, err error) {
	// TODO: client.Call("InkMinerRPC.GetShapes", args, &resp)
	return nil, errors.New("Not implemented")
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (a *ArtNode) GetGenesisBlock() (blockHash string, err error) {
	// TODO: client.Call("InkMinerRPC.GetGenesisBlock", args, &resp)
	return "", errors.New("Not implemented")
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (a *ArtNode) GetChildren(blockHash string) (blockHashes []string, err error) {
	// TODO: client.Call("InkMinerRPC.GetChildrenBlocks", args, &resp)
	return nil, errors.New("Not implemented")
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (a *ArtNode) CloseCanvas() (inkRemaining uint32, err error) {
	// TODO: client.Call("InkMinerRPC.GetInk", args, &resp)
	return 0, errors.New("Not implemented")
}

// Gets the ink cost of a particular operation
func calculateInkCost(shapeSvgString string, fill string, stroke string) (cost uint32, err error) {
	return 0, errors.New("Not implemented")
}
