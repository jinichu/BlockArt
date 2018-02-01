package blockartlib

import "crypto/ecdsa"

// InkMiner methods for mining blocks, verifying blocks, and adding blocks to the blockchain

type Block struct {
	prevBlock string          // Hash of the previous block
	blockNum  int             // Block number
	records   []Operation     // Set of operation records
	pubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	nonce     uint32
}
