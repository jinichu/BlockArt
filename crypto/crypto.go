package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
)

// GenerateKey generates a ECDSA public-private key pair.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// MarshalPrivate marshals a x509/PEM encoded ECDSA private key.
func MarshalPrivate(key *ecdsa.PrivateKey) (string, error) {
	rawPriv, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}

	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: rawPriv,
	}

	return string(pem.EncodeToMemory(keyBlock)), nil
}

var curves = []elliptic.Curve{elliptic.P224(), elliptic.P256(), elliptic.P384(), elliptic.P521()}

func fixCurve(curve elliptic.Curve) elliptic.Curve {
	for _, c := range curves {
		if c.Params().Name == curve.Params().Name {
			return c
		}
	}
	return curve
}

// MarshalPublic marshals a x509/PEM encoded ECDSA public key.
func MarshalPublic(key *ecdsa.PublicKey) (string, error) {
	key.Curve = fixCurve(key.Curve)

	rawPriv, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	keyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: rawPriv,
	}

	return string(pem.EncodeToMemory(keyBlock)), nil
}

// UnmarshalPrivate unmarshals a x509/PEM encoded ECDSA private key.
func UnmarshalPrivate(key string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("no PEM block found in private key")
	}
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

// UnmarshalPublic unmarshals a x509/PEM encoded ECDSA public key.
func UnmarshalPublic(key string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("no PEM block found in public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}
	return ecdsaPubKey, nil
}

// LoadPrivate loads a public-private key from the specified public and private key
// paths.
func LoadPrivate(publicPath, privatePath string) (*ecdsa.PrivateKey, error) {
	publicBody, err := ioutil.ReadFile(publicPath)
	if err != nil {
		return nil, err
	}
	privateBody, err := ioutil.ReadFile(privatePath)
	if err != nil {
		return nil, err
	}
	publicKey, err := UnmarshalPublic(string(publicBody))
	if err != nil {
		return nil, err
	}
	privateKey, err := UnmarshalPrivate(string(privateBody))
	if err != nil {
		return nil, err
	}
	privateKey.PublicKey = *publicKey
	return privateKey, nil
}

// Compute the Hash of any string
func Hash(a interface{}) (string, error) {
	h := sha1.New()
	if err := json.NewEncoder(h).Encode(a); err != nil {
		return "", nil
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Provides a sig for an operation
func Sign(operation []byte, privKey ecdsa.PrivateKey) (signedR, signedS *big.Int, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, operation)
	if err != nil {
		return big.NewInt(0), big.NewInt(0), err
	}

	signedR = r
	signedS = s
	return
}
