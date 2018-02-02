package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"

	"../crypto"
)

type InkMiner struct {
	client     *rpc.Client       // RPC client to connect to the server
	inkAmount  int               // Amount of ink this InkMiner hash
	privKey    *ecdsa.PrivateKey // Pub/priv key pair of this InkMiner
	blockchain map[string]*Block // Copy of the blockchain
	latest     []*Block          // Latest blocks in the blockchain
	settings   MinerNetSettings  // Settings for this BlockArt network instance
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

func RunInkMiner(serverAddr string, pubKeyFile string, privKeyFile string) error {
	localIP := "127.0.0.1"

	inkMiner := &InkMiner{}
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
