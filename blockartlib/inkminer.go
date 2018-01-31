package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type InkMiner struct {
	client     *rpc.Client       // RPC client to connect to the server
	inkAmount  int               // Amount of ink this InkMiner hash
	privKey    ecdsa.PrivateKey  // Pub/priv key pair of this InkMiner
	blockchain map[string]*Block // Copy of the blockchain
	latest     *Block            // Latest block in the blockchain
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

func RunInkMiner(serverAddr string, pubKeyFile string, privKeyFile string) {
	localIP := "127.0.0.1"

	inkMiner := &InkMiner{}
	rpc.Register(inkMiner)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", localIP+":0")
	if e != nil {
		log.Fatal("Listen Error:", e)
	}
	fmt.Println("InkMiner is up!")
	go http.Serve(l, nil)

	localAddr := localIP + ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)

	client, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	inkMiner.client = client
	// TODO: Do client.Call("Server.Register", args=(localAddr, pubKey),...)

	select {}
}
