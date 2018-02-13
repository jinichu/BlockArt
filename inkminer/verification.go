package inkminer

import (
	"crypto/ecdsa"
	"encoding/json"

	"../blockartlib"

	crypto "../crypto"
)

// Returns true if op-sig is valid
func (i *InkMiner) isOpSigValid(operation blockartlib.Operation) bool {
	opSig := operation.OpSig
	operation.OpSig = blockartlib.OpSig{}
	pubKey, err := crypto.UnmarshalPublic(operation.PubKey)
	if err != nil {
		return false
	}

	bytes, err := json.Marshal(operation)
	if err != nil {
		return false
	}

	return ecdsa.Verify(pubKey, bytes, opSig.R, opSig.S)
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
