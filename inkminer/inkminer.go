package inkminer

import (
	"crypto/ecdsa"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	blockartlib "../blockartlib"
	colors "../colors"
	crypto "../crypto"
	server "../server"
	stopper "../stopper"
)

type InkMiner struct {
	addr      string            // IP Address of the InkMiner
	client    *rpc.Client       // RPC client to connect to the server
	privKey   *ecdsa.PrivateKey // Pub/priv key pair of this InkMiner
	publicKey string            // Public key of the Miner (Note: is this needed?)

	settings server.MinerNetSettings // Settings for this BlockArt network instance
	rs       *rpc.Server             // RPC Server
	stopper  *stopper.Stopper
	log      *log.Logger

	// newOpChan should be used to notify the mining loop about new operations
	newOpChan chan blockartlib.Operation
	// newBlockChan should be used to notify the mining loop about new blocks
	// received
	newBlockChan chan blockartlib.Block

	mu struct {
		sync.Mutex

		l     net.Listener
		peers map[string]*peer

		// blockchain is a map between blockhash and the block
		blockchain map[string]blockartlib.Block
		// all operations that haven't been added to the current block chain
		mempool map[string]blockartlib.Operation
		// currentHead is the block that InkMiner is mining on
		currentHead blockartlib.Block
		// states of the canvas at a given block
		states map[string]State

		validateNumMap map[string]ValidateNumWaiter

		// closed is whether the miner is closed, mostly used for tests
		closed bool
	}
}

func (i *InkMiner) currentHead() blockartlib.Block {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.mu.currentHead
}

// getOutboundIP sets up a UDP connection (but doesn't send anything) and uses
// the local IP addressed assigned.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func New(privKey *ecdsa.PrivateKey) (*InkMiner, error) {
	i := &InkMiner{
		stopper:      stopper.New(),
		newOpChan:    make(chan blockartlib.Operation, 1),
		newBlockChan: make(chan blockartlib.Block, 1),
	}

	i.mu.states = make(map[string]State)
	i.mu.blockchain = make(map[string]blockartlib.Block)
	i.mu.mempool = make(map[string]blockartlib.Operation)
	i.mu.peers = make(map[string]*peer)
	i.privKey = privKey
	i.mu.validateNumMap = make(map[string]ValidateNumWaiter)

	i.log = log.New(os.Stderr, "", log.Flags()|log.Lshortfile)

	var err error
	i.publicKey, err = crypto.MarshalPublic(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}

	i.rs = rpc.NewServer()
	if err := i.rs.Register(i.RPC()); err != nil {
		return nil, err
	}

	return i, nil
}

func (i *InkMiner) RPC() *InkMinerRPC {
	return &InkMinerRPC{i}
}

func (i *InkMiner) BlockPoolSize() int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return len(i.mu.blockchain)
}

func (i *InkMiner) MemPoolSize() int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return len(i.mu.mempool)
}

func (i *InkMiner) Listen(serverAddr string) error {
	localIP := getOutboundIP()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	i.mu.Lock()
	i.mu.l = l
	i.mu.Unlock()

	localAddr := localIP + ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	i.addr = localAddr

	i.log.SetPrefix(colors.Green(i.addr) + " ")
	i.log.Printf("InkMiner running...")

	client, err := dialRPC(serverAddr)
	if err != nil {
		return err
	}
	i.client = client

	tcpAddr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return err
	}

	req := server.MinerInfo{
		Key:     i.privKey.PublicKey,
		Address: tcpAddr,
	}
	var resp server.MinerNetSettings
	if err := client.Call("RServer.Register", req, &resp); err != nil {
		return err
	}

	i.mu.Lock()
	i.settings = resp
	// Set currentHead to a dummy block initially so we can return saneish
	// results. This might be a terrible idea.
	i.mu.currentHead = blockartlib.Block{
		PrevBlock: i.settings.GenesisBlockHash,
		BlockNum:  1,
		PubKey:    i.privKey.PublicKey,
	}
	i.mu.Unlock()

	go i.peerDiscoveryLoop()
	go i.heartbeatLoop()

	if err := i.startMining(); err != nil {
		return err
	}

	for {
		select {
		case <-i.stopper.ShouldStop():
			return nil
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			i.log.Printf("Server accept error: %s", err)
			continue
		}
		i.log.Printf("New connection from: %s", conn.RemoteAddr())
		go i.rs.ServeConn(conn)
	}

	return nil
}

func (i *InkMiner) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.mu.closed {
		return nil
	}
	i.mu.closed = true

	i.log.Printf("closing...")

	if err := i.client.Close(); err != nil {
		return err
	}

	i.stopper.Stop()

	return i.mu.l.Close()
}

// Addr returns the listen address or "" if it hasn't started listening yet.
func (i *InkMiner) Addr() string {
	return i.addr
}
