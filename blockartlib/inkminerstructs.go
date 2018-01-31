package blockartlib

// Structs for ArtNode -> InkMiner RPC calls

import "crypto/ecdsa"

type Operation struct {
	opType  int             // Type of operation
	shape   string          // SVG string of this shape
	opSig   string          // Signature of the operation, signed by an ArtNode
	pubKey  ecdsa.PublicKey // Public key of the ArtNode that created this operation
	inkCost int             // Cost of ink to do this operation
}
