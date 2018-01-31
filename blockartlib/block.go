package blockartlib

import "crypto/ecdsa"

type Block struct {
	blockNum int              // Block number
	records  []Operation      // Set of operation records
	privKey  ecdsa.PrivateKey // Pub/priv key pair of the InkMiner that mined this block
	nonce    uint32
}
