package blockartlib

// Structs for ArtNode -> InkMiner RPC calls

type Operation struct {
	OpType      OpType // Type of operation
	Shape       string // SVG string of this shape
	ShapeHash   string // Hash of the SVG string
	OpSig       OpSig  // Signature of the operation, signed by an ArtNode
	PubKey      string // Public key of the ArtNode that created this operation
	InkCost     uint32 // Cost of ink to do this operation
	ValidateNum uint8  //  Number of blocks that must follow the block with this operation in the blockchain
}

type OpSig struct {
	r string
	s string
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
