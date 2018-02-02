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
	inkAmount     uint32                        // Amount of ink this InkMiner hash
	privKey       *ecdsa.PrivateKey             // Pub/priv key pair of this InkMiner
	blockchain    map[string]*blockartlib.Block // Copy of the blockchain
	latest        []*blockartlib.Block          // Latest blocks in the blockchain
	settings      blockartlib.MinerNetSettings  // Settings for this BlockArt network instance
	shapes        map[string]string             // Map of shape hashes to their SVG string representation
	currentHead   *blockartlib.Block            // Block that InkMiner is mining on (current head)
	mineBlockChan chan blockartlib.Block

	mu struct {
		sync.Mutex

		currentWIPBlock blockartlib.Block
	}
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

func RunInkMiner(serverAddr string, pubKeyFile string, privKeyFile string) error {
	localIP := "127.0.0.1"

	inkMiner := &InkMiner{
		blockchain: make(map[string]*blockartlib.Block),
		shapes:     make(map[string]string),
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
