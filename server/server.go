/*

Implements an example server for the BlockArt project, to be used in
project 1 of UBC CS 416 2017W2.

This server takes in settings from an input json files and implements
a simple strategy for GetNodes: return a fixed number of random miners
("num-miner-to-return" in the json config file).

Usage:

$ go run server.go
  -c string
    	Path to the JSON config

*/

package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sort"
	"sync"
	"time"

	colors "../colors"
	stopper "../stopper"
)

// Errors that the server could return.
type UnknownKeyError error

type KeyAlreadyRegisteredError string

func (e KeyAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("BlockArt server: key already registered [%s]", string(e))
}

type AddressAlreadyRegisteredError string

func (e AddressAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("BlockArt server: address already registered [%s]", string(e))
}

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32 `json:"canvas-x-max"`
	CanvasYMax uint32 `json:"canvas-y-max"`
}

type MinerSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`

	// Canvas settings
	CanvasSettings CanvasSettings `json:"canvas-settings"`
}

type RServer struct {
	s *Server
}

type Miner struct {
	Address         net.Addr
	RecentHeartbeat int64
}

type Config struct {
	MinerSettings    MinerNetSettings `json:"miner-settings"`
	RpcIpPort        string           `json:"rpc-ip-port"`
	NumMinerToReturn uint8            `json:"num-miner-to-return"`
}

type AllMiners struct {
	sync.RWMutex
	all map[string]*Miner
}

var (
	unknownKeyError UnknownKeyError = errors.New("BlockArt server: unknown key")
	errLog          *log.Logger     = log.New(os.Stderr, colors.Orange("server "), log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog          *log.Logger     = log.New(os.Stderr, colors.Orange("server "), log.Lshortfile|log.LUTC|log.Lmicroseconds)
)

// Parses args, setups up RPC server.
func init() {
	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})
	gob.Register(elliptic.P256())
}

type Server struct {
	rpc     *rpc.Server
	l       net.Listener
	stopper *stopper.Stopper

	allMiners AllMiners
	config    Config
}

func New(config Config) (*Server, error) {
	server := &Server{
		rpc:       rpc.NewServer(),
		stopper:   stopper.New(),
		allMiners: AllMiners{all: make(map[string]*Miner)},
		config:    config,
	}
	server.rpc.Register(&RServer{server})

	return server, nil
}

func (s *Server) Listen() error {
	l, err := net.Listen("tcp", s.config.RpcIpPort)
	if err != nil {
		return err
	}
	s.l = l

	outLog.Printf("Server started. Receiving on %s\n", s.config.RpcIpPort)

	for {
		select {
		case <-s.stopper.ShouldStop():
			return nil
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			errLog.Printf("accept error: %s", err)
			continue
		}
		go s.rpc.ServeConn(conn)
	}
}

func (s *Server) Addr() string {
	if s.l == nil {
		return ""
	}

	return s.l.Addr().String()
}

func (s *Server) NumMiners() int {
	s.allMiners.Lock()
	defer s.allMiners.Unlock()

	return len(s.allMiners.all)
}

func (s *Server) Close() error {
	s.stopper.Stop()
	return s.l.Close()
}

type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}

// Function to delete dead miners (no recent heartbeat)
func (s *Server) monitor(k string, heartBeatInterval time.Duration) {
	for {
		select {
		case <-s.stopper.ShouldStop():
			return
		default:
		}

		s.allMiners.Lock()
		if time.Now().UnixNano()-s.allMiners.all[k].RecentHeartbeat > int64(heartBeatInterval) {
			outLog.Printf("%s timed out\n", s.allMiners.all[k].Address.String())
			delete(s.allMiners.all, k)
			s.allMiners.Unlock()
			return
		}
		outLog.Printf("%s is alive\n", s.allMiners.all[k].Address.String())
		s.allMiners.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}

// Registers a new miner with an address for other miner to use to
// connect to it (returned in GetNodes call below), and a
// public-key for this miner. Returns error, or if error is not set,
// then setting for this canvas instance.
//
// Returns:
// - AddressAlreadyRegisteredError if the server has already registered this address.
// - KeyAlreadyRegisteredError if the server already has a registration record for publicKey.
func (s *RServer) Register(m MinerInfo, r *MinerNetSettings) error {
	s.s.allMiners.Lock()
	defer s.s.allMiners.Unlock()

	k := pubKeyToString(m.Key)
	if miner, exists := s.s.allMiners.all[k]; exists {
		return KeyAlreadyRegisteredError(miner.Address.String())
	}

	for _, miner := range s.s.allMiners.all {
		if miner.Address.Network() == m.Address.Network() && miner.Address.String() == m.Address.String() {
			return AddressAlreadyRegisteredError(m.Address.String())
		}
	}

	s.s.allMiners.all[k] = &Miner{
		m.Address,
		time.Now().UnixNano(),
	}

	go s.s.monitor(k, time.Duration(s.s.config.MinerSettings.HeartBeat)*time.Millisecond)

	*r = s.s.config.MinerSettings

	outLog.Printf("Got Register from %s\n", m.Address.String())

	return nil
}

type Addresses []net.Addr

func (a Addresses) Len() int           { return len(a) }
func (a Addresses) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Addresses) Less(i, j int) bool { return a[i].String() < a[j].String() }

// Returns addresses for a subset of miners in the system.
//
// Returns:
// - UnknownKeyError if the server does not know a miner with this publicKey.
func (s *RServer) GetNodes(key ecdsa.PublicKey, addrSet *[]net.Addr) error {

	// TODO: validate miner's GetNodes protocol? (could monitor state
	// of network graph/connectivity and validate protocol FSM)

	s.s.allMiners.RLock()
	defer s.s.allMiners.RUnlock()

	k := pubKeyToString(key)

	if _, ok := s.s.allMiners.all[k]; !ok {
		return unknownKeyError
	}

	minerAddresses := make([]net.Addr, 0, len(s.s.allMiners.all)-1)

	for pubKey, miner := range s.s.allMiners.all {
		if pubKey == k {
			continue
		}
		minerAddresses = append(minerAddresses, miner.Address)
	}

	sort.Sort(Addresses(minerAddresses))

	deterministicRandomNumber := key.X.Int64() % 32
	r := rand.New(rand.NewSource(deterministicRandomNumber))
	for n := len(minerAddresses); n > 0; n-- {
		randIndex := r.Intn(n)
		minerAddresses[n-1], minerAddresses[randIndex] = minerAddresses[randIndex], minerAddresses[n-1]
	}

	n := int(s.s.config.NumMinerToReturn)
	if n > len(minerAddresses) {
		n = len(minerAddresses)
	}
	*addrSet = minerAddresses[:n]

	return nil
}

// The server also listens for heartbeats from known miners. A miner must
// send a heartbeat to the server every HeartBeat milliseconds
// (specified in settings from server) after calling Register, otherwise
// the server will stop returning this miner's address/key to other
// miners.
//
// Returns:
// - UnknownKeyError if the server does not know a miner with this publicKey.
func (s *RServer) HeartBeat(key ecdsa.PublicKey, _ignored *bool) error {
	s.s.allMiners.Lock()
	defer s.s.allMiners.Unlock()

	k := pubKeyToString(key)
	if _, ok := s.s.allMiners.all[k]; !ok {
		return unknownKeyError
	}

	s.s.allMiners.all[k].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

func handleErrorFatal(msg string, e error) {
	if e != nil {
		errLog.Fatalf("%s, err = %s\n", msg, e.Error())
	}
}
