package main

import (
	"log"
	"os"

	"./crypto"
	"./inkminer"
)

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		log.Fatal("inkminer <server addr> <public key file> <private key file>")
	}

	serverAddr := args[0]
	pubKeyFile := args[1]
	privKeyFile := args[2]

	privKey, err := crypto.LoadPrivate(pubKeyFile, privKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	m, err := inkminer.New(privKey)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Listen(serverAddr); err != nil {
		log.Fatal(err)
	}
}
