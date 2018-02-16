/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"flag"
	"log"

	"./blockartlib"
	"./crypto"
)

var minerAddr = flag.String("miner", "127.0.0.1:8080", "the address of the miner to connect to")
var pubKeyFile = flag.String("pub", "testkeys/test2-public.key", "path to public key file")
var privKeyFile = flag.String("priv", "testkeys/test2-private.key", "path to private key file")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	privKey, err := crypto.LoadPrivate(
		*pubKeyFile, *privKeyFile,
	)
	if err != nil {
		return err
	}

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(*minerAddr, *privKey)
	if err != nil {
		return err
	}

	validateNum := uint8(2)

	log.Printf("Add yellow line that conflicts")

	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 60 10 l 0 20 L 60 20 Z", "yellow", "yellow")
	if err != nil {
		log.Println(err)
	}

	log.Printf("Delete a shape that doesn't exist")

	// Delete shape that doesn't exist
	ink2, err := canvas.DeleteShape(validateNum, "foobar")
	if err != nil {
		log.Println(err)
	}

	log.Printf("Add black shape")

	shapeHash2, blockHash2, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 160 100 l 0 20 l 20 40 l 40 0 Z", "black", "black")
	if err != nil {
		return err
	}

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if err != nil {
		return err
	}

	// Unused variables
	_ = blockHash
	_ = blockHash2
	_ = ink
	_ = ink2
	_ = ink3
	_ = ink4
	_ = shapeHash
	_ = shapeHash2
	_ = settings

	return nil
}
