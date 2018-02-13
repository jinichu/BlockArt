package blockartlib

import (
	"crypto/ecdsa"
	"math/big"

	crypto "../crypto"
)

// Structs for ArtNode -> InkMiner RPC calls

type Operation struct {
	OpType      OpType          // Type of operation
	OpSig       OpSig           // Signature of the operation, signed by an ArtNode
	PubKey      ecdsa.PublicKey // Public key of the ArtNode that created this operation
	ValidateNum uint8           //  Number of blocks that must follow the block with this operation in the blockchain
	Id          int64           // Unique ID for this Operation (to prevent replay attacks), given by a timestamp

	// TODO: get rid of InkCost and recompute when validating
	InkCost uint32 // Cost of ink to do this operation

	// These fields are only used for specific operations.

	DELETE struct {
		// ShapeHash is used to notify which shape to delete.
		ShapeHash string
	}

	ADD struct {
		// The shape object
		Shape Shape
	}
}

type OpSig struct {
	R *big.Int
	S *big.Int
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
	o.OpSig = OpSig{}
	return crypto.Hash(o)
}

func (o Operation) Sign(key ecdsa.PrivateKey) (Operation, error) {
	hash, err := o.Hash()
	if err != nil {
		return Operation{}, err
	}

	r, s, err := crypto.Sign([]byte(hash), key)
	if err != nil {
		return Operation{}, err
	}

	o.OpSig = OpSig{r, s}

	return o, nil
}

func (o Operation) PubKeyString() (string, error) {
	return crypto.MarshalPublic(&o.PubKey)
}
