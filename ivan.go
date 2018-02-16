package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"os"

	"./blockartlib"
	"./crypto"
)

var (
	minerAddr = flag.String("miner", "127.0.0.1:8080", "the address of the miner to connect to")
	public    = flag.String("public", "testkeys/test1-public.key", "public key file")
	private   = flag.String("private", "testkeys/test1-private.key", "private key file")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Flags() | log.Lshortfile)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func webColor(c color.Color) {
	r, g, b, a := c.RGBA()
	const max = 0xffff
	return fmt.Sprintf("rgba(%f, %f, %f, %f)", float64(r)/max*255, float64(g)/max*255, float64(b)/max*255, float64(a)/max)
}

func run() error {
	privKey, err := crypto.LoadPrivate(*public, *private)
	if err != nil {
		return err
	}

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(*minerAddr, *privKey)
	if err != nil {
		return err
	}

	reader, err := os.Open("testdata/ivanb.jpeg")
	if err != nil {
		return err
	}
	defer reader.Close()

	img, err := image.Decode(reader)
	if err != nil {
		return err
	}

	bounds := img.Bounds()

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; x < bounds.Max.Y; y++ {
			color := img.At(x, y)
			//"M 402 300 v 1 h 1 v -1 Z",
			_ = color
		}
	}
}
