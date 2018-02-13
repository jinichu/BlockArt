package integration

import (
	"fmt"
	"log"
	"testing"
	"time"

	"../blockartlib"
	"../inkminer"
)

func TestSimpleCluster(t *testing.T) {
	ts := NewTestCluster(t, 1)
	defer ts.Close()

	log.Printf("cluster up")
}

func TestClusterHeartBeat(t *testing.T) {
	heartbeat := uint32(10)
	defer SetHeartBeat(heartbeat)()

	nodes := 2
	ts := NewTestCluster(t, nodes)
	defer ts.Close()

	SucceedsSoon(t, func() error {
		n := ts.Server.NumMiners()
		if n != nodes {
			return fmt.Errorf("expected %d miners, got %d", nodes, n)
		}
		return nil
	})

	ts.Miners[1].Close()

	time.Sleep(time.Duration(heartbeat) * time.Millisecond * 5)

	SucceedsSoon(t, func() error {
		n := ts.Server.NumMiners()
		if n != 1 {
			return fmt.Errorf("expected %d miners, got %d", 1, n)
		}
		return nil
	})
}

func TestClusterP2P(t *testing.T) {
	ts := NewTestCluster(t, 5)
	defer ts.Close()

	SucceedsSoon(t, func() error {
		for i, im := range ts.Miners {
			n := im.NumPeers()
			if n != 4 {
				return fmt.Errorf("%d. expected 4 peers, only have %d", i, n)
			}
		}
		return nil
	})
}

func TestClusterBlockPropagation(t *testing.T) {
	ts := NewTestCluster(t, 5)
	defer ts.Close()

	for i, im := range ts.Miners {
		if n := im.BlockPoolSize(); n != 0 {
			t.Fatalf("%d. expected empty blockpool, found %d", i, n)
		}
	}

	var resp inkminer.NotifyBlockResponse
	if err := ts.Miners[0].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 1,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[0].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 1,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[1].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 2,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[2].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 3,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	SucceedsSoon(t, func() error {
		for i, im := range ts.Miners {
			n := im.BlockPoolSize()
			if n != 3 {
				return fmt.Errorf("%d. expected 3 blocks, have %d", i, n)
			}
		}
		return nil
	})
}

func TestClusterOperationPropagation(t *testing.T) {
	ts := NewTestCluster(t, 5)
	defer ts.Close()

	for i, im := range ts.Miners {
		if n := im.MemPoolSize(); n != 0 {
			t.Fatalf("%d. expected empty mempool, found %d", i, n)
		}
	}

	var resp inkminer.NotifyOperationResponse
	if err := ts.Miners[0].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: ts.NewAddOp(blockartlib.Operation{
			InkCost: 1,
		}),
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[0].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: ts.NewAddOp(blockartlib.Operation{
			InkCost: 1,
		}),
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[1].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: ts.NewAddOp(blockartlib.Operation{
			InkCost: 2,
		}),
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[2].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: ts.NewAddOp(blockartlib.Operation{
			InkCost: 3,
		}),
	}, &resp); err != nil {
		t.Fatal(err)
	}

	SucceedsSoon(t, func() error {
		for i, im := range ts.Miners {
			n := im.MemPoolSize()
			if n != 3 {
				return fmt.Errorf("%d. expected 3 blocks, have %d", i, n)
			}
		}
		return nil
	})
}

func TestClusterMinerHeartBeat(t *testing.T) {
	heartbeat := uint32(10)
	defer SetHeartBeat(heartbeat)()

	nodes := 2
	ts := NewTestCluster(t, nodes)
	defer ts.Close()

	SucceedsSoon(t, func() error {
		n := ts.Miners[0].NumPeers()
		want := nodes - 1
		if n != want {
			return fmt.Errorf("expected %d peers, got %d", want, n)
		}
		return nil
	})

	ts.Miners[1].Close()

	time.Sleep(time.Duration(heartbeat) * time.Millisecond * 5)

	SucceedsSoon(t, func() error {
		n := ts.Miners[0].NumPeers()
		want := 0
		if n != want {
			return fmt.Errorf("expected %d peers, got %d", want, n)
		}
		return nil
	})
}

func TestSimpleClusterBlockPropagateOnJoin(t *testing.T) {
	ts := NewTestCluster(t, 1)
	defer ts.Close()

	var resp inkminer.NotifyBlockResponse
	if err := ts.Miners[0].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 1,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[0].RPC().NotifyBlock(inkminer.NotifyBlockRequest{
		Block: blockartlib.Block{
			BlockNum: 2,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	SucceedsSoon(t, func() error {
		n := ts.Miners[0].BlockPoolSize()
		if n != 2 {
			return fmt.Errorf("0. expected 2 blocks, have %d", n)
		}
		return nil
	})

	ts.AddNode()

	SucceedsSoon(t, func() error {
		n := ts.Miners[1].BlockPoolSize()
		if n != 2 {
			return fmt.Errorf("1. expected 2 blocks, have %d", n)
		}
		return nil
	})
}
