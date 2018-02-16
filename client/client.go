package client

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"../blockartlib"
)

const validateNum = 0

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

	c.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./client/static/"))))

	c.mux.HandleFunc("/api/state", c.handleState)
	c.mux.HandleFunc("/api/add", c.handleAdd)
	c.mux.HandleFunc("/api/delete", c.handleDelete)

	c.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./client/static/index.html")
	})

	return c, nil
}

type Block struct {
	Hash     string
	Shapes   []Shape
	Children []Block
}

type Shape struct {
	Hash string
	Svg  string
}

func wrap(err error, msg string, params ...interface{}) error {
	return fmt.Errorf("%s: %s\n\n%s", fmt.Sprintf(msg, params...), err, debug.Stack())
}

func (c *Client) GetBlock(hash string) (Block, error) {
	shapeHashes, err := c.canvas.GetShapes(hash)
	if err != nil {
		return Block{}, wrap(err, "GetShapes")
	}
	var shapes []Shape
	for _, hash := range shapeHashes {
		svg, err := c.canvas.GetSvgString(hash)
		if err != nil {
			return Block{}, wrap(err, "GetSvgString")
		}
		shapes = append(shapes, Shape{
			Hash: hash,
			Svg:  svg,
		})
	}

	children, err := c.canvas.GetChildren(hash)
	if err != nil {
		return Block{}, wrap(err, "GetChildren")
	}

	var childBlocks []Block
	for _, child := range children {
		block, err := c.GetBlock(child)
		if err != nil {
			return Block{}, wrap(err, "GetBlock [%s]", child)
		}
		childBlocks = append(childBlocks, block)
	}

	return Block{
		Hash:     hash,
		Shapes:   shapes,
		Children: childBlocks,
	}, nil
}

func (c *Client) GetBlockChain() (Block, error) {
	hash, err := c.canvas.GetGenesisBlock()
	if err != nil {
		return Block{}, wrap(err, "GetGenesisBlock")
	}
	return c.GetBlock(hash)
}

type jsonErr struct {
	Error string
	Stack string
}

func handleErr(w http.ResponseWriter, err error) {
	stack := debug.Stack()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	if err := json.NewEncoder(w).Encode(jsonErr{
		Error: err.Error(),
		Stack: string(stack),
	}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (c *Client) handleState(w http.ResponseWriter, r *http.Request) {
	type state struct {
		Settings   blockartlib.CanvasSettings
		Ink        uint32
		BlockChain Block
	}

	ink, err := c.canvas.GetInk()
	if err != nil {
		handleErr(w, err)
		return
	}

	blockChain, err := c.GetBlockChain()
	if err != nil {
		handleErr(w, err)
		return
	}

	resp := state{
		Settings:   c.settings,
		Ink:        ink,
		BlockChain: blockChain,
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleErr(w, err)
		return
	}
}

type shape struct {
	Type   string
	Stroke string
	Fill   string
	Svg    string
}

func (c *Client) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleErr(w, errors.New("must use POST"))
		return
	}

	var body shape
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handleErr(w, err)
		return
	}

	if body.Type != "path" {
		handleErr(w, fmt.Errorf("Type must be path; not %q", body.Type))
		return
	}

	if _, _, _, err := c.canvas.AddShape(validateNum, blockartlib.PATH, body.Svg, body.Fill, body.Stroke); err != nil {
		handleErr(w, err)
		return
	}
}

func (c *Client) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleErr(w, errors.New("must use POST"))
		return
	}

	var hash string
	if err := json.NewDecoder(r.Body).Decode(&hash); err != nil {
		handleErr(w, err)
		return
	}

	if _, err := c.canvas.DeleteShape(validateNum, hash); err != nil {
		handleErr(w, err)
		return
	}
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
