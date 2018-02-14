package crypto

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrivateFile(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	privKey, err := MarshalPrivate(key)
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := MarshalPublic(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "crypto-TestLoadPrivateFile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	publicPath := filepath.Join(dir, "public.key")
	privatePath := filepath.Join(dir, "private.key")

	if err := ioutil.WriteFile(publicPath, []byte(pubKey), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(privatePath, []byte(privKey), 0755); err != nil {
		t.Fatal(err)
	}

	key2, err := LoadPrivate(publicPath, privatePath)
	if err != nil {
		t.Fatal(err)
	}
	privKey2, err := MarshalPrivate(key2)
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, err := MarshalPublic(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	if privKey != privKey2 {
		t.Errorf("%q != %q", privKey, privKey2)
	}
	if pubKey != pubKey2 {
		t.Errorf("%q != %q", pubKey, pubKey2)
	}
}

func TestLoadPrivateDirect(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	privKey, err := MarshalPrivate(key)
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := MarshalPublic(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	key2, err := LoadPrivate(pubKey, privKey)
	if err != nil {
		t.Fatal(err)
	}
	privKey2, err := MarshalPrivate(key2)
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, err := MarshalPublic(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	if privKey != privKey2 {
		t.Errorf("%q != %q", privKey, privKey2)
	}
	if pubKey != pubKey2 {
		t.Errorf("%q != %q", pubKey, pubKey2)
	}
}
