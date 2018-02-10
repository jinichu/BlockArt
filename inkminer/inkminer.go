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

	latest        []*blockartlib.Block    // Latest blocks in the blockchain
	settings      server.MinerNetSettings // Settings for this BlockArt network instance
	currentHead   blockartlib.Block       // Block that InkMiner is mining on (current head)
	mineBlockChan chan blockartlib.Block  // Channel used to distribute blocks
	rs            *rpc.Server             // RPC Server
	states        map[string]State        // States of the canvas at a given block
	stopper       *stopper.Stopper        // TODO: [Jonathan] Figure what this is
	log           *log.Logger             // TODO: [Jonathan] Figure what this is

	mu struct {
		sync.Mutex

		l               net.Listener
		currentWIPBlock blockartlib.Block
		peers           map[string]*peer

		// blockchain is a map between blockhash and the block
		blockchain map[string]blockartlib.Block
		// all operations that haven't been added to the current block chain
		mempool map[string]blockartlib.Operation

		// closed is whether the miner is closed, mostly used for tests
		closed bool
	}
}

type State struct {
	shapes      map[string]blockartlib.Shape // Map of shape hashes to their SVG string representation
	shapeOwners map[string]string            // Map of shape hashes to their owner (InkMiner PubKey)
	inkLevels   map[string]uint32            // Current ink levels of every InkMiner
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

/*
addr      string							// IP Address of the InkMiner
client    *rpc.Client       				// RPC client to connect to the server

latest        []*blockartlib.Block         	// Latest blocks in the blockchain
settings      server.MinerNetSettings 		// Settings for this BlockArt network instance
stopper       *stopper.Stopper

*/

func New(privKey *ecdsa.PrivateKey) (*InkMiner, error) {
	i := &InkMiner{
		states:  make(map[string]State),
		stopper: stopper.New(),
	}

	i.mu.blockchain = make(map[string]blockartlib.Block)
	i.mu.mempool = make(map[string]blockartlib.Operation)
	i.mu.peers = make(map[string]*peer)
	i.privKey = privKey

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

	// Initialize the map and channel variables
	i.states = make(map[string]State)
	i.mineBlockChan = make(chan blockartlib.Block)

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
	i.settings = resp

	go i.peerDiscoveryLoop()
	go i.heartbeatLoop()

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
