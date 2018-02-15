package inkminer

import (
	"crypto/ecdsa"
	"fmt"

	"../blockartlib"
)

// Returns true if op-sig is valid
func isOpSigValid(operation blockartlib.Operation) error {
	hash, err := operation.Hash()
	if err != nil {
		return err
	}

	if operation.PubKey.X == nil || operation.PubKey.Y == nil {
		return fmt.Errorf("invalid public key for operation: %+v", operation)
	}

	opSig := operation.OpSig
	if opSig.S == nil || opSig.R == nil {
		return fmt.Errorf("invalid operation signature for operation: %+v", operation)
	}
	if !ecdsa.Verify(&operation.PubKey, []byte(hash), opSig.R, opSig.S) {
		return fmt.Errorf("invalid operation signature for operation: %+v", operation)
	}

	return nil
}

func (i *InkMiner) validateShape(shape blockartlib.Shape) error {
	// TODO: validate the shape data
	if shape.Svg == "" || shape.Fill == "" || shape.Stroke == "" {
		return fmt.Errorf("one of Svg, Fill, Stroke is empty: %+v", shape)
	}
	if shape.Type != blockartlib.PATH {
		return fmt.Errorf("unknown shape type: %+v", shape.Type)
	}
	points := blockartlib.ComputeVertices(shape.Svg)
	for j := 0; j < len(points); j++ {
		if points[j].GetX() < 0 || points[j].GetX() > float64(i.settings.CanvasSettings.CanvasXMax) ||
			points[j].GetY() < 0 || points[j].GetY() > float64(i.settings.CanvasSettings.CanvasYMax) {
			return fmt.Errorf("svg is out of canvas bounds")
		}
	}
	return nil
}

func (i *InkMiner) validateOp(operation blockartlib.Operation) error {
	if err := isOpSigValid(operation); err != nil {
		return err
	}

	switch operation.OpType {
	case blockartlib.ADD:
		if err := i.validateShape(operation.ADD.Shape); err != nil {
			return err
		}
	case blockartlib.DELETE:
		if operation.DELETE.ShapeHash == "" {
			return fmt.Errorf("missing ShapeHash")
		}
	default:
		return fmt.Errorf("invalid operation type: %+v", operation.OpType)
	}

	return nil
}

// Returns true if this block has the correct nonce
func (i *InkMiner) isBlockNonceValid(block blockartlib.Block) error {
	blockHash, err := block.Hash()
	if err != nil {
		return err
	}

	want := i.settings.PoWDifficultyOpBlock
	if len(block.Records) == 0 {
		want = i.settings.PoWDifficultyNoOpBlock
	}
	zeros := uint8(numZeros(blockHash))
	if zeros != want {
		return fmt.Errorf("invalid block nonce: got %d zeros, wanted %d in %q, %+v", zeros, want, blockHash, block)
	}
	return nil
}

// Returns true if the inkCost is valid in the given state
// !!!
func (i *InkMiner) isInkCostValid(inkCost int, pubKey string, state State) bool {
	return state.inkLevels[pubKey] > uint32(inkCost)
}

// Returns true if the given shape is valid to add in the state (does not conflict with something else)
// TODO: Complete this
// !!!
func (i *InkMiner) canAddShape(shape blockartlib.Shape, pubKey string) bool {
	return false // stub
}