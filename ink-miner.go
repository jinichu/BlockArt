package main

import (
	"os"

	"./blockartlib"
)

func main() {
	args := os.Args[1:]
	serverAddr := args[0]
	pubKeyFile := args[1]
	privKeyFile := args[2]

	blockartlib.RunInkMiner(serverAddr, pubKeyFile, privKeyFile)
}
