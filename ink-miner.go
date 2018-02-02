package main

import (
	"log"
	"os"

	"./blockartlib"
)

func main() {
	args := os.Args[1:]
	serverAddr := args[0]
	pubKeyFile := args[1]
	privKeyFile := args[2]

	if err := blockartlib.RunInkMiner(serverAddr, pubKeyFile, privKeyFile); err != nil {
		log.Fatal(err)
	}
}
