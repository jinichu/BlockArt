package blockartlib

// Structs for ArtNode -> InkMiner RPC calls

import "crypto/ecdsa"

type Operation struct {
	opType      int             // Type of operation
	shape       string          // SVG string of this shape
	shapeHash   string          // Hash of the SVG string
	opSig       string          // Signature of the operation, signed by an ArtNode
	pubKey      ecdsa.PublicKey // Public key of the ArtNode that created this operation
	inkCost     uint32          // Cost of ink to do this operation
	validateNum uint8           //  Number of blocks that must follow the block with this operation in the blockchain
}

type AddShapeResponse struct {
	blockHash    string
	inkRemaining int
}

type GetShapesResponse struct {
	shapeHashes []string
}

type GetChildrenResponse struct {
	blockHashes []string
}
