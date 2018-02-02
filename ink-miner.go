package main

import (
	"log"
	"os"

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

	if err := inkminer.RunInkMiner(serverAddr, pubKeyFile, privKeyFile); err != nil {
		log.Fatal(err)
	}
}
