package integration

import (
	"fmt"
	"testing"
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
