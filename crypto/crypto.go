package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
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

// MarshalPublic marshals a x509/PEM encoded ECDSA public key.
func MarshalPublic(key *ecdsa.PublicKey) (string, error) {
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
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

// UnmarshalPublic unmarshals a x509/PEM encoded ECDSA public key.
func UnmarshalPublic(key string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
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
