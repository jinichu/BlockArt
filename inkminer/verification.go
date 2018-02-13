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

func validateShape(shape blockartlib.Shape) error {
	// TODO: validate the shape data
	if shape.Svg == "" || shape.Fill == "" || shape.Stroke == "" {
		return fmt.Errorf("one of Svg, Fill, Stroke is empty: %+v", shape)
	}
	return nil
}

func validateOp(operation blockartlib.Operation) error {
	if err := isOpSigValid(operation); err != nil {
		return err
	}

	if operation.OpType != blockartlib.ADD && operation.OpType != blockartlib.DELETE {
		return fmt.Errorf("invalid operation type: %+v", operation.OpType)
	}

	if err := validateShape(operation.ADD.Shape); err != nil {
		return err
	}

	return nil
}

// Returns true if this block has the correct nonce
func (i *InkMiner) isBlockNonceValid(block blockartlib.Block) bool {
	blockHash, err := block.Hash()
	if err != nil {
		return false
	}
	if uint8(numZeros(blockHash)) == i.settings.PoWDifficultyOpBlock {
		return true
	}
	return false
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
