package integration

import (
	"fmt"
	"testing"
	"time"

	"../blockartlib"
)

func TestMiningCluster(t *testing.T) {
	defer SetBlockDelay(10 * time.Millisecond)()

	ts := NewTestCluster(t, 1)
	defer ts.Close()

	SucceedsSoon(t, func() error {
		n := ts.Miners[0].BlockPoolSize()
		want := 5
		if n < want {
			return fmt.Errorf("0. expected %d blocks, have %d", want, n)
		}
		_, depth, err := ts.Miners[0].BlockWithLongestChain()
		if err != nil {
			return err
		}
		if depth != n {
			return fmt.Errorf("0. BlockWithLongestChain() depth = %d; wanted %d", depth, n)
		}
		return nil
	})
}

func TestMiningClusterOrphanBlock(t *testing.T) {
	defer SetBlockDelay(10 * time.Millisecond)()

	ts := NewTestCluster(t, 1)
	defer ts.Close()

	if _, err := ts.Miners[0].AddBlock(ts.Miners[0].TestMine(t, blockartlib.Block{
		PrevBlock: "doesn't exist",
		BlockNum:  1,
		PubKey:    ts.Keys[0].PublicKey,
	})); err != nil {
		t.Fatal(err)
	}

	SucceedsSoon(t, func() error {
		_, depth, err := ts.Miners[0].BlockWithLongestChain()
		if err != nil {
			return err
		}
		want := 5
		if depth < want {
			return fmt.Errorf("0. BlockWithLongestChain() depth = %d; wanted %d", depth, want)
		}
		return nil
	})
}
