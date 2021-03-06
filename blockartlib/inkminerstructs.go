package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
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
	Type   ShapeType
	Svg    string // SVG string of this shape
	Fill   string
	Stroke string
}

func (s Shape) SvgString() string {
	return fmt.Sprintf(`<%s d="%s" stroke="%s" fill="%s"/>`, s.Type, s.Svg, s.Stroke, s.Fill)
}

func (s Shape) Valid() error {
	if s.Svg == "" || s.Fill == "" || s.Stroke == "" {
		return fmt.Errorf("one of Svg, Fill, Stroke is empty: %+v", s)
	}
	if s.Type != PATH {
		return fmt.Errorf("unknown shape type: %+v", s.Type)
	}
	if err := svgStringValidityCheck(s.Svg); err != nil {
		return err
	}
	if err := svgShapeValidityCheck(s.Svg, s.Fill, s.Stroke); err != nil {
		return err
	}
	return nil
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
	key, err := crypto.MarshalPublic(&o.PubKey)
	if err != nil {
		return "", fmt.Errorf("PubKeyString error for operation %+v: %+v", o, err)
	}
	return key, nil
}

// TestShape returns a test shape with a specific offset and cost
func TestShape(cost, offset int) Shape {
	return Shape{
		Type:   PATH,
		Svg:    fmt.Sprintf("M %d 0 L %d %d", offset, offset, cost),
		Fill:   "transparent",
		Stroke: "red",
	}
}
