package main

import "os"

func main() {
	args := os.Args[1:]
	serverAddr := args[0]
	pubKey := args[1]
	privKey := args[2]
}
