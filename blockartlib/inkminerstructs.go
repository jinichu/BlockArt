package blockartlib

// Structs for ArtNode -> InkMiner RPC calls

import "crypto/ecdsa"

type Operation struct {
	OpType      int             // Type of operation
	Shape       string          // SVG string of this shape
	ShapeHash   string          // Hash of the SVG string
	OpSig       string          // Signature of the operation, signed by an ArtNode
	PubKey      ecdsa.PublicKey // Public key of the ArtNode that created this operation
	InkCost     uint32          // Cost of ink to do this operation
	ValidateNum uint8           //  Number of blocks that must follow the block with this operation in the blockchain
}

type AddShapeResponse struct {
	BlockHash    string
	InkRemaining uint32
}

type GetShapesResponse struct {
	ShapeHashes []string
}

type GetChildrenResponse struct {
	BlockHashes []string
}
