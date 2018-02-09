package inkminer

import (
	"crypto/ecdsa"
	"encoding/json"

	"../blockartlib"

	crypto "../crypto"
)

func (i *InkMiner) isOperationValid(operation blockartlib.Operation) bool {
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
