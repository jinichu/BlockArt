package blockartlib

import "crypto/ecdsa"

type Block struct {
	prevBlock string          // Hash of the previous block
	blockNum  int             // Block number
	records   []Operation     // Set of operation records
	pubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	nonce     uint32
}
