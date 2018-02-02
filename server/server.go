package server

import (
	"log"
	"net"
	"net/rpc"
	"sync"

	"../blockartlib"
)

type Server struct {
	rs *rpc.Server

	mu struct {
		sync.Mutex

		l      net.Listener
		miners map[string]string
	}
}

func New() (*Server, error) {
	s := &Server{
		rs: rpc.NewServer(),
	}

	s.mu.miners = map[string]string{}

	if err := s.rs.Register(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.mu.l = ln
	s.mu.Unlock()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		go s.rs.ServeConn(conn)
	}
	return nil
}

func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mu.l == nil {
		return ""
	}

	return s.mu.l.Addr().String()
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.mu.l.Close()
}

type RegisterRequest struct {
	PublicKey, Address string
}

func (s *Server) Register(req RegisterRequest, resp *blockartlib.MinerNetSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.miners[req.PublicKey] = req.Address

	*resp = blockartlib.MinerNetSettings{
		GenesisBlockHash:       "genesis",
		MinNumMinerConnections: 3,
		InkPerNoOpBlock:        25,
		InkPerOpBlock:          50,
		HeartBeat:              100,
		PoWDifficultyNoOpBlock: 5,
		PoWDifficultyOpBlock:   5,
		CanvasSettings: blockartlib.CanvasSettings{
			CanvasXMax: 1024,
			CanvasYMax: 1024,
		},
	}
	return nil
}

type GetNodesRequest struct {
	PublicKey string
}
type GetNodesResponse struct {
	Addrs []string
}

func (s *Server) GetNodes(req GetNodesRequest, resp *GetNodesResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, addr := range s.mu.miners {
		resp.Addrs = append(resp.Addrs, addr)
	}

	return nil
}

type HeartBeatRequest struct {
	PublicKey string
}

type HeartBeatResponse struct{}

func (s *Server) HeartBeat(req HeartBeatRequest, resp *HeartBeatResponse) error {
	return nil
}
