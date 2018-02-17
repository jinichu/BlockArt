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
var pubKeyFile = flag.String("pub", "testkeys/test4-public.key", "path to public key file")
var privKeyFile = flag.String("priv", "testkeys/test4-private.key", "path to private key file")

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

	log.Printf("Add orange rectangle")

	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 200 30 l 0 50 l 30 0 l 0 -50 l -30 0", "orange", "orange")
	if err != nil {
		log.Println(err)
	}

	log.Printf("Add pink rectangle with red outline")

	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 250 50 l 0 10 l 30 0 l 0 -10 l -30 0", "pink", "red")
	if err != nil {
		return err
	}

	log.Printf("Add fancy blue line")

	shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 250 80 L 260 90 l 30 0 l 30 -10", "transparent", "blue")
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
	_ = blockHash3
	_ = ink
	_ = ink2
	_ = ink3
	_ = ink4
	_ = shapeHash
	_ = shapeHash2
	_ = shapeHash3
	_ = settings

	return nil
}
