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
	"log"

	"./blockartlib"
	"./crypto"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	minerAddr := "127.0.0.1:8080"
	privKey, err := crypto.LoadPrivate(
		"testkeys/test1-public.key", "testkeys/test2-private.key",
	)
	if err != nil {
		return err
	}

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	if err != nil {
		return err
	}

	validateNum := uint8(2)

	// Add a line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	if err != nil {
		return err
	}

	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if err != nil {
		return err
	}

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
