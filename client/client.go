package client

import (
	"crypto/ecdsa"
	"log"
	"net/http"

	"../blockartlib"
)

type Client struct {
	mux      *http.ServeMux
	privKey  *ecdsa.PrivateKey
	canvas   blockartlib.Canvas
	settings blockartlib.CanvasSettings
}

func New(privKey *ecdsa.PrivateKey) (*Client, error) {
	c := &Client{
		mux:     http.NewServeMux(),
		privKey: privKey,
	}

	return c, nil
}

func (c *Client) Listen(bind, minerAddr string) error {
	var err error
	c.canvas, c.settings, err = blockartlib.OpenCanvas(minerAddr, *c.privKey)
	if err != nil {
		return err
	}

	log.Printf("Listening... %s", bind)

	if err := http.ListenAndServe(bind, c.mux); err != nil {
		return err
	}
	return nil
}
