package integration

import (
	"log"
	"testing"
)

func TestSimpleCluster(t *testing.T) {
	log.Println("cluster test")
	ts := NewTestCluster(t, 1)
	defer ts.Close()
}

func TestClusterP2P(t *testing.T) {
	log.Println("cluster test")
	ts := NewTestCluster(t, 5)
	defer ts.Close()
}
