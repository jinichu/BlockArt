package blockartlib

import "crypto/ecdsa"

// InkMiner methods for mining blocks, verifying blocks, and adding blocks to the blockchain

type Block struct {
	PrevBlock string          // Hash of the previous block
	BlockNum  int             // Block number
	Records   []Operation     // Set of operation records
	PubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	Nonce     uint32
}
