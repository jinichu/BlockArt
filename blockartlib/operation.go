package blockartlib

import "crypto/ecdsa"

type Operation struct {
	opType  int              // Type of operation
	params  []string         // Parameters for this operation
	opSig   string           // Signature of the operation, signed by an ArtNode
	privKey ecdsa.PrivateKey // Pub/priv key pair of the ArtNode that created this operation
}
