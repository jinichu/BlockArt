package main

import (
	"flag"
	"log"

	"./client"
	"./crypto"
)

var (
	bind      = flag.String("bind", ":8081", "the address to listen on")
	minerAddr = flag.String("miner", "127.0.0.1:8080", "the address of the miner to connect to")
	public    = flag.String("public", "testkeys/test1-public.key", "public key file")
	private   = flag.String("private", "testkeys/test1-private.key", "private key file")
)

func run() error {
	flag.Parse()

	privKey, err := crypto.LoadPrivate(*public, *private)
	if err != nil {
		return err
	}

	c, err := client.New(privKey)
	if err != nil {
		return err
	}

	return c.Listen(*bind, *minerAddr)
}
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
