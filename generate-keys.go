package main

import (
	"flag"
	"io/ioutil"
	"log"

	"./crypto"
)

var file = flag.String("file", "key", "file name to generate. will output <file>-public.key <file>-private.key")

func main() {
	flag.Parse()

	log.Printf("HELP: Specify -file=<name> to change the key name.")
	log.Println()
	log.Printf("Generating %s-public.key, %s-private.key...", *file, *file)

	if err := run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Done.")
}
func run() error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	publicKey, err := crypto.MarshalPublic(&key.PublicKey)
	if err != nil {
		return err
	}

	privateKey, err := crypto.MarshalPrivate(key)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(*file+"-public.key", []byte(publicKey), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(*file+"-private.key", []byte(privateKey), 0600); err != nil {
		return err
	}

	return nil
}
