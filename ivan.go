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
	"sync"
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
	return fmt.Sprintf("rgba(%d, %d, %d, %f)", int(float64(r)/max*255), int(float64(g)/max*255), int(float64(b)/max*255), float64(a)/max)
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

	const workers = 32

	type work struct {
		path, fill, stroke string
	}

	workChan := make(chan work, workers)
	var wg sync.WaitGroup

	log.Printf("spinning up %d workers", workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for w := range workChan {
				for {
					if _, _, _, err := canvas.AddShape(0, blockartlib.PATH, w.path, w.fill, w.stroke); err != nil {
						if strings.HasPrefix(err.Error(), "BlockArt: Not enough ink to addShape") {
							log.Printf("%q: sleeping... %s", w.path, err)
							time.Sleep(1 * time.Second)
							continue
						} else {
							log.Fatal(err)
						}
					}
					break
				}
			}
		}()
	}

	const stride = 8

	for x := bounds.Min.X; x < bounds.Max.X; x += stride {
		for y := bounds.Min.Y; y < bounds.Max.Y; y += stride {
			log.Printf("drawing %d x %d", x, y)
			color := webColor(img.At(x, y))
			path := fmt.Sprintf("M %d %d v %d h %d v -%d Z", margin+x, margin+y, stride, stride, stride)
			workChan <- work{
				path:   path,
				fill:   color,
				stroke: color,
			}
		}
	}

	close(workChan)
	wg.Wait()

	return nil
}
