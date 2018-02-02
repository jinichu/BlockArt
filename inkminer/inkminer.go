package inkminer

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"

	"../blockartlib"
	"../crypto"
)

type InkMiner struct {
	client        *rpc.Client                   // RPC client to connect to the server
	privKey       *ecdsa.PrivateKey             // Pub/priv key pair of this InkMiner
	blockchain    map[string]*blockartlib.Block // Copy of the blockchain
	latest        []*blockartlib.Block          // Latest blocks in the blockchain
	settings      blockartlib.MinerNetSettings  // Settings for this BlockArt network instance
	currentHead   *blockartlib.Block            // Block that InkMiner is mining on (current head)
	mineBlockChan chan blockartlib.Block
	states        map[string]State // States of the canvas at a given block

	mu struct {
		sync.Mutex

		currentWIPBlock blockartlib.Block
	}
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

type State struct {
	shapes      map[string]string          // Map of shape hashes to their SVG string representation
	shapeOwners map[string]ecdsa.PublicKey // Map of shape hashes to their owner (InkMiner PubKey)
	inkLevels   map[ecdsa.PublicKey]uint32 // Current ink levels of every InkMiner
}

func RunInkMiner(serverAddr string, pubKeyFile string, privKeyFile string) error {
	localIP := "127.0.0.1"

	inkMiner := &InkMiner{
		blockchain: make(map[string]*blockartlib.Block),
		states:     make(map[string]State),
	}
	var err error
	inkMiner.privKey, err = crypto.LoadPrivate(pubKeyFile, privKeyFile)
	if err != nil {
		return err
	}

	rpc.Register(inkMiner)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", localIP+":0")
	if err != nil {
		return err
	}
	fmt.Println("InkMiner is up!")
	go http.Serve(l, nil)

	localAddr := localIP + ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	_ = localAddr

	client, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		return err
	}
	inkMiner.client = client
	/*
			  TODO: Do client.Call("Server.Register", args=(localAddr, pubKey),...)
		    to register this InkMiner to the network and get the BlockArt settings
	*/

	select {}

	return nil
}
