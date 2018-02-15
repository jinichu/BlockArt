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

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	privKey, err := crypto.LoadPrivate(
		"testkeys/test1-public.key", "testkeys/test1-private.key",
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

	log.Printf("add line")

	// Add a square.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 100 100 L 200 100 L 200 110 L 100 110 Z", "blue", "red")
	if err != nil {
		return err
	}

	log.Printf("add line2")

	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 100 120 L 200 120", "transparent", "blue")
	if err != nil {
		return err
	}

	log.Printf("delete shape")

	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if err != nil {
		return err
	}

	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if err != nil {
		return err
	}

	_ = settings
	_ = blockHash
	_ = ink
	_ = shapeHash2
	_ = blockHash2
	_ = ink2
	_ = ink3
	_ = ink4

	return nil
}
