package inkminer

import (
	"errors"
	"net"
	"net/rpc"
	"sort"
	"time"

	blockartlib "../blockartlib"
	colors "../colors"
)

var ErrUnimplemented = errors.New("unimplemented")

const Timeout = 2 * time.Second
const HeartBeatsPerInterval = 5

func dialRPC(addr string) (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, Timeout)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}

func (i *InkMiner) heartbeatLoop() {
	// send tick n times as fast as required to avoid timeouts
	duration := time.Millisecond * time.Duration(i.settings.HeartBeat) / HeartBeatsPerInterval
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
		case <-i.stopper.ShouldStop():
			return
		}

		var resp bool
		if err := i.client.Call("RServer.HeartBeat", i.privKey.PublicKey, &resp); err != nil {
			i.log.Printf("HeartBeat error: %s", err)
		}
	}
}

func (i *InkMiner) peerDiscover() error {
	if i.NumPeers() >= int(i.settings.MinNumMinerConnections) {
		return nil
	}

	var resp []net.Addr
	if err := i.client.Call("RServer.GetNodes", i.privKey.PublicKey, &resp); err != nil {
		i.log.Printf("GetNodes failed: %s", err)
		return err
	}

	i.log.Printf("got peers: %+v", resp)

	for _, addr := range resp {
		addr := addr
		go func() {
			if _, err := i.addPeer(addr.String()); err != nil {
				i.log.Printf("failed to add peer: %s")
			}
		}()
	}

	return nil
}

func (i *InkMiner) addPeer(address string) (*peer, error) {
	if address == i.addr {
		return nil, nil
	}

	i.mu.Lock()
	p, ok := i.mu.peers[address]
	i.mu.Unlock()

	// don't readd a peer
	if ok {
		return p, nil
	}

	client, err := dialRPC(address)
	if err != nil {
		return nil, err
	}

	p = &peer{
		rpc:     client,
		address: address,
	}

	i.mu.Lock()
	p2, exists := i.mu.peers[address]
	// race condition to add peer, discard this one
	if exists {
		p = p2
	} else {
		i.mu.peers[address] = p
	}
	i.mu.Unlock()

	if !exists {
		if err := p.sendHello(i); err != nil {
			return nil, err
		}
		go i.peerHeartBeat(p)
	}

	return p, nil
}

func (p *peer) sendHello(i *InkMiner) error {
	var resp HelloResponse
	req := HelloRequest{
		Addr:   i.Addr(),
		Blocks: map[string]struct{}{},
	}

	i.mu.Lock()
	for blockHash := range i.mu.blockchain {
		req.Blocks[blockHash] = struct{}{}
	}
	i.mu.Unlock()

	if err := p.rpc.Call("InkMinerRPC.Hello", req, &resp); err != nil {
		return err
	}
	return nil
}

type HelloRequest struct {
	Addr string

	Blocks map[string]struct{}
}

type HelloResponse struct{}

func (i *InkMinerRPC) Hello(req HelloRequest, resp *HelloResponse) error {
	i.i.log.Printf("got Hello: %+v", req)

	p, err := i.i.addPeer(req.Addr)
	if err != nil {
		return err
	}

	if p == nil {
		return nil
	}

	// send all blocks the client doesn't have
	go func() {
		depths := map[string]int{}

		var toSend []string
		i.i.mu.Lock()
		for hash := range i.i.mu.blockchain {
			if _, ok := req.Blocks[hash]; ok {
				continue
			}

			// build up the depths of all the blocks so we can send them in order
			if _, err := i.i.blockDepthLocked(hash, depths); err != nil {
				i.i.log.Printf("blockDepthLocked err: %+v", err)
				// send any blocks without a proper chain last
				depths[hash] = len(i.i.mu.blockchain)
			}

			toSend = append(toSend, hash)
		}
		i.i.mu.Unlock()

		sort.Slice(toSend, func(i, j int) bool {
			return depths[toSend[i]] < depths[toSend[j]]
		})

		for _, hash := range toSend {
			block, ok := i.i.GetBlock(hash)
			if !ok {
				continue
			}

			if err := p.notifyBlock(block); err != nil {
				i.i.log.Printf("failed to send block to new client: %s", err)
			}
		}
	}()

	return nil
}

func (i *InkMinerRPC) HeartBeat(req struct{}, resp *struct{}) error {
	select {
	case <-i.i.stopper.ShouldStop():
		return errors.New("server is closing...")
	default:
	}

	return nil
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
	return i.i.addOperation(req.Operation)
}

func (i *InkMiner) addOperation(op blockartlib.Operation) error {
	if err := i.validateOp(op); err != nil {
		return err
	}

	hash, err := op.Hash()
	if err != nil {
		return err
	}

	i.mu.Lock()
	_, ok := i.mu.mempool[hash]
	if !ok {
		i.mu.mempool[hash] = op
	}
	i.mu.Unlock()

	if ok {
		return nil
	}

	select {
	case i.newOpChan <- op:
	default:
	}

	// if it's a new operation, announce it to all peers
	return i.floodOperation(op)
}

type NotifyBlockRequest struct {
	Block blockartlib.Block
}

type NotifyBlockResponse struct{}

func (i *InkMiner) announceBlock(block blockartlib.Block) error {
	return i.asyncSend(func(p *peer) error {
		return p.notifyBlock(block)
	})
}

func (p *peer) notifyBlock(block blockartlib.Block) error {
	req := NotifyBlockRequest{
		Block: block,
	}
	var resp NotifyBlockResponse
	return p.rpc.Call("InkMinerRPC.NotifyBlock", req, &resp)
}

func (i *InkMinerRPC) NotifyBlock(req NotifyBlockRequest, resp *NotifyBlockResponse) error {
	if _, err := i.i.AddBlock(req.Block); err != nil {
		return err
	}
	return nil
}

func (i *InkMiner) removePeer(p *peer) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.mu.peers, p.address)
	return p.rpc.Close()
}

func (i *InkMiner) peerHeartBeat(p *peer) {
	timeout := time.Millisecond * time.Duration(i.settings.HeartBeat)
	duration := timeout / HeartBeatsPerInterval
	ticker := time.NewTicker(duration)

	remove := func() {
		i.log.Printf("peer timed out: %s", p)
		if err := i.removePeer(p); err != nil {
			i.log.Printf("failed to remove peer: %s", err)
		}
	}

	for {
		select {
		case <-i.stopper.ShouldStop():
			return
		case <-ticker.C:
		}

		var resp struct{}
		call := p.rpc.Go("InkMinerRPC.HeartBeat", struct{}{}, &resp, nil)

		select {
		case <-i.stopper.ShouldStop():
			return
		case reply := <-call.Done:
			if reply.Error != nil {
				i.log.Printf("got heartbeat error: %s", reply.Error)
				remove()
				return
			}
		case <-time.After(timeout):
			remove()
			return
		}
	}
}

// Helper Function: Adds block to the InkMiner
func (i *InkMiner) AddBlock(block blockartlib.Block) (success bool, err error) {
	if err := i.isBlockNonceValid(block); err != nil {
		return false, err
	}

	hash, err := block.Hash()
	if err != nil {
		return false, err
	}

	i.mu.Lock()
	if _, ok := i.mu.blockchain[hash]; ok {
		i.mu.Unlock()
		return false, nil
	}
	i.mu.blockchain[hash] = block
	i.mu.Unlock()

	select {
	case i.newBlockChan <- block:
	default:
	}

	state, err := i.CalculateState(block)
	if err != nil {
		i.log.Printf("got possibly invalid block, orphaned/out of order?: %+v: %+v", block, err)
	} else {
		i.mu.Lock()
		for opHash, validateNumWaiter := range i.mu.validateNumMap {
			if state.commitedOperations[opHash] >= int(validateNumWaiter.validateNum) {
				var blockHash string
				currentBlock := block
				for j := 0; j < int(validateNumWaiter.validateNum); j++ {
					blockHash = currentBlock.PrevBlock
					currentBlock = i.mu.blockchain[currentBlock.PrevBlock]
				}
				delete(i.mu.validateNumMap, opHash)
				validateNumWaiter.done <- blockHash
			}
		}
		i.mu.Unlock()
	}

	return true, i.announceBlock(block)
}

type peer struct {
	address string
	rpc     *rpc.Client
}

func (p peer) String() string {
	return colors.Green(p.address)
}
