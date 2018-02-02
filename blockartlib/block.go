package blockartlib

import "crypto/ecdsa"

type Block struct {
	PrevBlock string          // Hash of the previous block
	BlockNum  int             // Block number
	Records   []Operation     // Set of operation records
	PubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	Nonce     uint32
}
