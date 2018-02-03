package blockartlib

import (
	"crypto/ecdsa"

	crypto "../crypto"
)

type Block struct {
	PrevBlock string          // Hash of the previous block
	BlockNum  int             // Block number
	Records   []Operation     // Set of operation records
	PubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	Nonce     uint32
}

func (b Block) Hash() (string, error) {
	return crypto.Hash(b)
}
