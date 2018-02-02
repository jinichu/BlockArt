package integration

import (
	"testing"
)

func TestSimpleCluster(t *testing.T) {
	ts := NewTestCluster(t, 1)
	defer ts.Close()
}

func TestClusterP2P(t *testing.T) {
	ts := NewTestCluster(t, 5)
	defer ts.Close()
}
