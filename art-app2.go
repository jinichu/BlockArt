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
var pubKeyFile = flag.String("pub", "testkeys/test1-public.key", "path to public key file")
var privKeyFile = flag.String("priv", "testkeys/test1-private.key", "path to private key file")

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

	log.Printf("Add yellow shape")

	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 90 90 l 20 30 l 20 10 l 30 10 Z", "yellow", "yellow")
	if err != nil {
		return err
	}

	log.Printf("Add purple trapezoid")

	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 70 40 l 40 0 l 30 15 l 15 30 Z", "purple", "purple")
	if err != nil {
		return err
	}

	log.Printf("Add blue (no fill) square")

	shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 10 50 l 0 20 l 20 0 l 0 -20 Z", "transparent", "blue")
	if err != nil {
		return err
	}

	log.Printf("Add pink shape")

	shapeHash4, blockHash4, ink4, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 20 80 l 0 40 l 20 0 l 0 -20 Z", "pink", "pink")
	if err != nil {
		return err
	}

	// Close the canvas.
	ink5, err = canvas.CloseCanvas()
	if err != nil {
		return err
	}

	// Unused variables
	_ = blockHash
	_ = blockHash2
	_ = blockHash3
	_ = blockHash4
	_ = ink
	_ = ink2
	_ = ink3
	_ = ink4
	_ = ink5
	_ = shapeHash
	_ = shapeHash2
	_ = shapeHash3
	_ = shapeHash4
	_ = settings

	return nil
}
