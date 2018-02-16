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
	"fmt"
	"io/ioutil"
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

	log.Printf("Add blue triangle")

	// Add a blue line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 40 30 L 20 10 L 10 10 L 10 10 Z", "blue", "blue")
	if err != nil {
		return err
	}

	log.Printf("Add red line")

	// Add a red line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 70 20 L 40 20 L 60 20 Z", "red", "red")
	if err != nil {
		return err
	}

	log.Printf("Delete blue triangle")

	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if err != nil {
		return err
	}

	log.Printf("Add green triangle")

	// Add a rectangle.
	shapeHash3, blockHash3, ink4, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 20 30 L 40 10 L 40 10 L 20 10 Z", "green", "green")
	if err != nil {
		return err
	}

	log.Printf("Get genesis block hash")
	// Get genesis block
	genesisBlockHash, err := canvas.GetGenesisBlock()
	if err != nil {
		return err
	}

	canvasShapes, _, err := GetCanvas(genesisBlockHash, canvas, 0)

	htmlString := fmt.Sprintf(`<html><svg viewbox="0 0 %s %s">`, settings.CanvasXMax, settings.CanvasYMax)
	for i := 0; i < len(canvasShapes); i++ {
		htmlString += canvasShapes[i]
	}
	htmlString += "</svg></html>"

	// Close the canvas.
	ink5, err := canvas.CloseCanvas()
	if err != nil {
		return err
	}

	log.Printf("Writing to html file")
	err = ioutil.WriteFile("canvas.html", []byte(htmlString), 0755)
	if err != nil {
		return err
	}

	// Unused variables
	_ = blockHash
	_ = ink
	_ = shapeHash2
	_ = shapeHash3
	_ = blockHash2
	_ = blockHash3
	_ = ink2
	_ = ink3
	_ = ink4
	_ = ink5

	return nil
}

func GetCanvas(hash string, canvas blockartlib.Canvas, depth int) ([]string, int, error) {
	shapeHashes, err := canvas.GetShapes(hash)
	if err != nil {
		return nil, 0, err
	}
	var shapes []string
	for _, hash := range shapeHashes {
		svg, err := canvas.GetSvgString(hash)
		if err != nil {
			return nil, 0, err
		}
		shapes = append(shapes, svg)
	}

	children, err := canvas.GetChildren(hash)
	if err != nil {
		return nil, 0, err
	}

	var accumulatorShapes []string
	max := 0
	for _, child := range children {
		childShapes, d, err := GetCanvas(child, canvas, depth+1)
		if err != nil {
			return nil, 0, err
		}
		if d > max {
			accumulatorShapes = childShapes
			max = d
		}
	}
	shapes = append(shapes, accumulatorShapes...)

	return shapes, depth, nil
}
