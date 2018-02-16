package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"os"
	"strings"
	"time"

	"./blockartlib"
	"./crypto"
)

var (
	minerAddr = flag.String("miner", "127.0.0.1:8080", "the address of the miner to connect to")
	public    = flag.String("public", "testkeys/test1-public.key", "public key file")
	private   = flag.String("private", "testkeys/test1-private.key", "private key file")
	img       = flag.String("f", "testdata/ivanb.jpg", "the image to render")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Flags() | log.Lshortfile)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func webColor(c color.Color) string {
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
	canvas, _, err := blockartlib.OpenCanvas(*minerAddr, *privKey)
	if err != nil {
		return err
	}

	reader, err := os.Open(*img)
	if err != nil {
		return err
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	bounds := img.Bounds()

	margin := 100

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; x < bounds.Max.Y; y++ {
			x := x
			y := y
			go func() {
				for {
					color := webColor(img.At(x, y))
					path := fmt.Sprintf("M %d %d v 1 h 1 v -1 Z", margin+x*2, margin+y*2)
					if _, _, _, err := canvas.AddShape(0, blockartlib.PATH, path, color, color); err != nil {
						if strings.HasPrefix(err.Error(), "BlockArt: Not enough ink to addShape") {
							log.Printf("%d, %d: sleeping... %s", x, y, err)
							time.Sleep(1 * time.Second)
							continue
						} else {
							log.Fatal(err)
						}
					}
					break
				}
			}()
		}
	}

	return nil
}
