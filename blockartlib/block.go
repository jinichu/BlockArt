package blockartlib

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"
)

type Block struct {
	PrevBlock string          // Hash of the previous block
	BlockNum  int             // Block number
	Records   []Operation     // Set of operation records
	PubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
	Nonce     uint32
}

func (b Block) HashNoNonce() ([]byte, error) {
	b.Nonce = 0
	hash := sha1.New()
	if err := json.NewEncoder(hash).Encode(b); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func (b Block) HashApplyNonce(noNonceHash []byte) (string, error) {
	hash := sha1.New()
	hash.Write(noNonceHash)
	hash.Write([]byte(strconv.Itoa(int(b.Nonce))))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (b Block) Hash() (string, error) {
	noNonceHash, err := b.HashNoNonce()
	if err != nil {
		return "", err
	}
	return b.HashApplyNonce(noNonceHash)
}
