package integration

import (
	"fmt"
	"testing"
	"time"
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
