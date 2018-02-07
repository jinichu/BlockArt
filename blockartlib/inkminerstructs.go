package blockartlib

import crypto "../crypto"

// Structs for ArtNode -> InkMiner RPC calls

type Operation struct {
	OpType      OpType // Type of operation
	Shape       Shape  // Shape object
	ShapeHash   string // Hash of the SVG string
	OpSig       OpSig  // Signature of the operation, signed by an ArtNode
	PubKey      string // Public key of the ArtNode that created this operation
	InkCost     uint32 // Cost of ink to do this operation
	ValidateNum uint8  //  Number of blocks that must follow the block with this operation in the blockchain
}

type OpSig struct {
	R string
	S string
}

type Shape struct {
	Svg    string // SVG string of this shape
	Fill   string
	Stroke string
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

type InitConnectionRequest struct {
	PublicKey string
}

func (o Operation) Hash() (string, error) {
	return crypto.Hash(o)
}
