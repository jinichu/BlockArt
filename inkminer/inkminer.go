package inkminer

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"sync"

	"../blockartlib"
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
	rs            *rpc.Server

	mu struct {
		sync.Mutex

		l               net.Listener
		currentWIPBlock blockartlib.Block
	}
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

func New(privKey *ecdsa.PrivateKey) (*InkMiner, error) {
	inkMiner := &InkMiner{
		blockchain: make(map[string]*blockartlib.Block),
		shapes:     make(map[string]string),
	}
	inkMiner.privKey = privKey

	inkMiner.rs = rpc.NewServer()
	if err := inkMiner.rs.Register(inkMiner); err != nil {
		return nil, err
	}

	return inkMiner, nil
}

func (i *InkMiner) Listen(serverAddr string) error {
	// TODO: figure out a better IP
	localIP := "127.0.0.1"
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	i.mu.Lock()
	i.mu.l = l
	i.mu.Unlock()

	fmt.Println("InkMiner is up!")

	localAddr := localIP + ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	_ = localAddr

	client, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		return err
	}
	i.client = client
	/*
			  TODO: Do client.Call("Server.Register", args=(localAddr, pubKey),...)
		    to register this InkMiner to the network and get the BlockArt settings
	*/

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Server accept error: %s", err)
		}
		go i.rs.ServeConn(conn)
	}

	return nil
}

func (i *InkMiner) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.mu.l.Close()
}

// Addr returns the listen address or "" if it hasn't started listening yet.
func (i *InkMiner) Addr() string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.mu.l == nil {
		return ""
	}

	return i.mu.l.Addr().String()
}
