package integration

import (
	"log"
	"testing"
)

func TestIndividualCluster(t *testing.T) {
	log.Println("cluster test")
	ts := NewTestCluster(t, 1)
	defer ts.Close()
}
