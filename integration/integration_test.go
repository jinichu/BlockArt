package integration

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"../blockartlib"
	"../crypto"
	"../inkminer"
	"../server"
)

type TestCluster struct {
	Server *server.Server

	Keys     []*ecdsa.PrivateKey
	Miners   []*inkminer.InkMiner
	ArtNodes []blockartlib.Canvas

	t *testing.T
}

func NewTestCluster(t *testing.T, nodes int) *TestCluster {
	ts := &TestCluster{
		t: t,
	}

	s, err := server.New()
	if err != nil {
		t.Fatal(err)
	}
	ts.Server = s

	go func() {
		if err := s.Listen("127.0.0.1:0"); err != nil {
			t.Error(err)
		}
	}()

	for s.Addr() == "" {
		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < nodes; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			t.Error(err)
		}
		ts.Keys = append(ts.Keys, key)

		m, err := inkminer.New(key)
		if err != nil {
			t.Error(err)
		}

		ts.Miners = append(ts.Miners, m)

		go func() {
			if err := m.Listen(s.Addr()); err != nil {
				t.Error(err)
			}
		}()
	}

	for i, miner := range ts.Miners {
		for miner.Addr() == "" {
			time.Sleep(10 * time.Millisecond)
		}

		canvas, _, err := blockartlib.OpenCanvas(miner.Addr(), *ts.Keys[i])
		if err != nil {
			t.Error(err)
		}
		ts.ArtNodes = append(ts.ArtNodes, canvas)
	}

	return ts
}

func (ts *TestCluster) Close() {
	for _, an := range ts.ArtNodes {
		if _, err := an.CloseCanvas(); err != nil {
			ts.t.Error(err)
		}
	}
	for _, i := range ts.Miners {
		if err := i.Close(); err != nil {
			ts.t.Error(err)
		}
	}
	if err := ts.Server.Close(); err != nil {
		ts.t.Error(err)
	}
}
