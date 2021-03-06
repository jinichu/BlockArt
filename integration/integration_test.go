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

	inkminer.TestBlockDelay = 10 * time.Second
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
			time.Sleep(5 * time.Millisecond)
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

var heartbeatTime uint32 = 10000

// SetHeartBeat sets the amount of time between heartbeats in MS.
func SetHeartBeat(time uint32) func() {
	old := heartbeatTime
	heartbeatTime = time
	return func() {
		heartbeatTime = old
	}
}

// SetBlockDelay sets the amount of time between blocks mined.
func SetBlockDelay(time time.Duration) func() {
	old := inkminer.TestBlockDelay
	inkminer.TestBlockDelay = time
	return func() {
		inkminer.TestBlockDelay = old
	}
}

func NewTestCluster(t *testing.T, nodes int) *TestCluster {
	ts := &TestCluster{
		t: t,
	}

	min := nodes - 1
	if min <= 0 {
		min = 1
	}

	c := server.Config{
		RpcIpPort:        "127.0.0.1:0",
		NumMinerToReturn: uint8(nodes),
		MinerSettings: server.MinerNetSettings{
			MinNumMinerConnections: uint8(min),
			HeartBeat:              heartbeatTime,
			GenesisBlockHash:       "genesis!",
			InkPerOpBlock:          200,
			InkPerNoOpBlock:        50,
			CanvasSettings: server.CanvasSettings{
				CanvasXMax: 1000000000,
				CanvasYMax: 1000000000,
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
		ts.AddNode()
	}
	log.Printf("inkminers up")

	SucceedsSoon(t, func() error {
		n := s.NumMiners()
		if n != nodes {
			return fmt.Errorf("expected %d nodes on server; got %d", nodes, n)
		}
		return nil
	})

	return ts
}

func (ts *TestCluster) AddNode() {
	t := ts.t

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	ts.Keys = append(ts.Keys, key)

	m, err := inkminer.New(key)
	if err != nil {
		t.Fatal(err)
	}

	ts.Miners = append(ts.Miners, m)

	go func() {
		if err := m.Listen(ts.Server.Addr()); err != nil {
			log.Println(err)
			t.Fatal(err)
		}
	}()

	SucceedsSoon(t, func() error {
		if m.Addr() == "" {
			return errors.New("missing address")
		}
		return nil
	})

	canvas, _, err := blockartlib.OpenCanvas(m.Addr(), *key)
	if err != nil {
		t.Fatal(err)
	}
	ts.ArtNodes = append(ts.ArtNodes, canvas)
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

func (ts *TestCluster) NewAddOp(op blockartlib.Operation) blockartlib.Operation {
	op.PubKey = ts.Keys[0].PublicKey
	op.OpType = blockartlib.ADD
	op.ADD.Shape = blockartlib.Shape{
		Type:   blockartlib.PATH,
		Svg:    "M 0 0 H 20 V 20 h -20 Z",
		Fill:   "transparent",
		Stroke: "red",
	}

	op2, err := op.Sign(*ts.Keys[0])
	if err != nil {
		ts.t.Fatal(err)
	}
	return op2
}
