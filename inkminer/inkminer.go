package inkminer

import (
	"crypto/ecdsa"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"sync"

	"../blockartlib"
	"../crypto"
)

type InkMiner struct {
	client        *rpc.Client       // RPC client to connect to the server
	privKey       *ecdsa.PrivateKey // Pub/priv key pair of this InkMiner
	publicKey     string
	blockchain    map[string]*blockartlib.Block // Copy of the blockchain
	latest        []*blockartlib.Block          // Latest blocks in the blockchain
	settings      blockartlib.MinerNetSettings  // Settings for this BlockArt network instance
	currentHead   *blockartlib.Block            // Block that InkMiner is mining on (current head)
	mineBlockChan chan blockartlib.Block
	rs            *rpc.Server
	states        map[string]State // States of the canvas at a given block

	mu struct {
		sync.Mutex

		l               net.Listener
		currentWIPBlock blockartlib.Block
	}
	// TODO: Keep track of shapes on the canvas and the owners (ArtNode) of every shape
}

type State struct {
	shapes      map[string]string          // Map of shape hashes to their SVG string representation
	shapeOwners map[string]ecdsa.PublicKey // Map of shape hashes to their owner (InkMiner PubKey)
	inkLevels   map[ecdsa.PublicKey]uint32 // Current ink levels of every InkMiner
}

func New(privKey *ecdsa.PrivateKey) (*InkMiner, error) {
	inkMiner := &InkMiner{
		blockchain: make(map[string]*blockartlib.Block),
		states:     make(map[string]State),
	}
	inkMiner.privKey = privKey
	var err error
	inkMiner.publicKey, err = crypto.MarshalPublic(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}

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

	log.Printf("InkMiner is up! %s", l.Addr())

	localAddr := localIP + ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	_ = localAddr

	client, err := rpc.Dial("tcp", serverAddr)
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
		log.Printf("New connection from: %s", conn.RemoteAddr())
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
