package integration

import (
	"fmt"
	"testing"

	"../blockartlib"
	"../inkminer"
)

func TestSimpleCluster(t *testing.T) {
	ts := NewTestCluster(t, 1)
	defer ts.Close()
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
		Operation: blockartlib.Operation{
			InkCost: 1,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[0].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: blockartlib.Operation{
			InkCost: 1,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[1].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: blockartlib.Operation{
			InkCost: 2,
		},
	}, &resp); err != nil {
		t.Fatal(err)
	}

	if err := ts.Miners[2].RPC().NotifyOperation(inkminer.NotifyOperationRequest{
		Operation: blockartlib.Operation{
			InkCost: 3,
		},
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
