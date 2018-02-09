package inkminer

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"time"

	blockartlib "../blockartlib"
	colors "../colors"
)

var ErrUnimplemented = errors.New("unimplemented")

const Timeout = 2 * time.Second

func dialRPC(addr string) (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, Timeout)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}

func (i *InkMiner) peerDiscover() error {
	if i.NumPeers() >= int(i.settings.MinNumMinerConnections) {
		return nil
	}

	var resp []net.Addr
	if err := i.client.Call("RServer.GetNodes", i.privKey.PublicKey, &resp); err != nil {
		log.Printf("GetNodes failed: %s", err)
		return err
	}

	log.Printf("got peers: %+v", resp)

	for _, addr := range resp {
		addr := addr
		go func() {
			if err := i.addPeer(addr.String()); err != nil {
				i.log.Printf("failed to add peer: %s")
			}
		}()
	}

	return nil
}

func (i *InkMiner) addPeer(address string) error {
	if address == i.addr {
		return nil
	}

	i.mu.Lock()
	_, ok := i.mu.peers[address]
	i.mu.Unlock()

	// don't readd a peer
	if ok {
		return nil
	}

	client, err := dialRPC(address)
	if err != nil {
		return err
	}

	p := &peer{
		rpc:     client,
		address: address,
	}

	i.mu.Lock()
	p2, old := i.mu.peers[address]
	// race condition to add peer, discard this one
	if old {
		p = p2
	} else {
		i.mu.peers[address] = p
	}
	i.mu.Unlock()

	if !old {
		var resp HelloResponse
		if err := p.rpc.Call("InkMinerRPC.Hello", HelloRequest{Addr: i.addr}, &resp); err != nil {
			return err
		}
	}

	return nil
}

type HelloRequest struct {
	Addr string
}

type HelloResponse struct{}

func (i *InkMinerRPC) Hello(req HelloRequest, resp *HelloResponse) error {
	return i.i.addPeer(req.Addr)
}

func (i *InkMiner) NumPeers() int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return len(i.mu.peers)
}

func (i *InkMiner) peerDiscoveryLoop() {
	t := time.NewTicker(5 * time.Second)
	for {
		if err := i.peerDiscover(); err != nil {
			i.log.Printf("Peer discovery error: %s", err)
		}

		select {
		case <-i.stopper.ShouldStop():
			return
		case <-t.C:
		}
	}
}

func (i *InkMiner) asyncSend(f func(p *peer) error) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, p := range i.mu.peers {
		p := p
		go func() {
			if err := f(p); err != nil {
				i.log.Printf("asyncSend error (to %s): %s", p, err)
			}
		}()
	}

	return nil
}

type NotifyOperationRequest struct {
	Operation blockartlib.Operation
}

type NotifyOperationResponse struct{}

func (i *InkMiner) floodOperation(operation blockartlib.Operation) error {
	req := NotifyOperationRequest{
		Operation: operation,
	}
	return i.asyncSend(func(p *peer) error {
		var resp NotifyOperationResponse
		return p.rpc.Call("InkMinerRPC.NotifyOperation", req, &resp)
	})
}

func (i *InkMinerRPC) NotifyOperation(req NotifyOperationRequest, resp *NotifyOperationResponse) error {
	// TODO: validate operation

	hash, err := req.Operation.Hash()
	if err != nil {
		return err
	}

	i.i.mu.Lock()
	_, ok := i.i.mu.mempool[hash]
	if !ok {
		i.i.mu.mempool[hash] = req.Operation
	}
	i.i.mu.Unlock()

	if ok {
		return nil
	}

	// if it's a new operation, announce it to all peers
	return i.i.floodOperation(req.Operation)
}

type NotifyBlockRequest struct {
	Block blockartlib.Block
}

type NotifyBlockResponse struct{}

func (i *InkMiner) announceBlock(block blockartlib.Block) error {
	req := NotifyBlockRequest{
		Block: block,
	}
	return i.asyncSend(func(p *peer) error {
		var resp NotifyBlockResponse
		return p.rpc.Call("InkMinerRPC.NotifyBlock", req, &resp)
	})
}

func (i *InkMinerRPC) NotifyBlock(req NotifyBlockRequest, resp *NotifyBlockResponse) error {
	// TODO: validate block

	success, err := i.i.AddBlock(req.Block)
	if err != nil {
		return err
	}

	// If we didn't end up adding the block at all
	if !success {
		return nil
	}

	// if it's a new block, announce it to all peers
	return i.i.announceBlock(req.Block)
}

// Helper Function: Adds block to the InkMiner
func (i *InkMiner) AddBlock(block blockartlib.Block) (success bool, err error){
	i.mu.Lock()
	defer i.mu.Unlock()

	hash, err := block.Hash()
	if err != nil {
		return false, err
	}

	_, ok := i.mu.blockchain[hash]
	if !ok {
		// Check every block in the latest queue
		newLatest := false
		for _, blk := range i.latest {
			//
			if block.BlockNum > blk.BlockNum {
				i.latest = append([]*blockartlib.Block{}, &block)
				newLatest = true
				break
			}
		}

		// If we did not clean out the latest list at all
		if !newLatest {

		}
		i.mu.blockchain[hash] = block
	} else {
		return false, nil
	}

	return true, nil
}

type peer struct {
	address string
	rpc     *rpc.Client
}

func (p peer) String() string {
	return colors.Green(p.address)
}
