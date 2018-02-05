package integration

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"../blockartlib"
	"../crypto"
	"../inkminer"
	"../server"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func SucceedsSoon(t *testing.T, f func() error) {
	timeout := time.After(time.Second * 2)
	c := make(chan error)
	go func() {
		for {
			select {
			case <-timeout:
				return
			default:
			}
			err := f()
			c <- err
			if err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	var err error
	for {
		select {
		case err = <-c:
			if err == nil {
				return
			}
		case <-timeout:
			t.Fatal(err)
		}
	}
}

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

	c := server.Config{
		RpcIpPort:        "127.0.0.1:0",
		NumMinerToReturn: uint8(nodes),
		MinerSettings: server.MinerNetSettings{
			MinerSettings: server.MinerSettings{
				MinNumMinerConnections: uint8(nodes - 1),
				HeartBeat:              10000,
			},
		},
	}

	s, err := server.New(c)
	if err != nil {
		t.Fatal(err)
	}
	ts.Server = s

	go func() {
		if err := s.Listen(); err != nil {
			t.Fatal(err)
		}
	}()

	SucceedsSoon(t, func() error {
		if s.Addr() == "" {
			return errors.New("missing address")
		}
		return nil
	})

	for i := 0; i < nodes; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			t.Error(err)
		}
		ts.Keys = append(ts.Keys, key)

		m, err := inkminer.New(key)
		if err != nil {
			t.Fatal(err)
		}

		ts.Miners = append(ts.Miners, m)

		go func() {
			if err := m.Listen(s.Addr()); err != nil {
				log.Println(err)
				t.Fatal(err)
			}
		}()
	}
	log.Printf("inkminers up")

	for i, miner := range ts.Miners {
		SucceedsSoon(t, func() error {
			if miner.Addr() == "" {
				return errors.New("missing address")
			}
			return nil
		})

		canvas, _, err := blockartlib.OpenCanvas(miner.Addr(), *ts.Keys[i])
		if err != nil {
			t.Fatal(err)
		}
		ts.ArtNodes = append(ts.ArtNodes, canvas)
	}

	SucceedsSoon(t, func() error {
		n := s.NumMiners()
		if n != nodes {
			return fmt.Errorf("expected %d nodes on server; got %d", nodes, n)
		}
		return nil
	})

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
