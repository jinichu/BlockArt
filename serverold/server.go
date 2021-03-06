package server

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	blockartlib "../blockartlib"
	colors "../colors"
	stopper "../stopper"
)

type Server struct {
	rs      *rpc.Server
	stopper *stopper.Stopper
	log     *log.Logger

	mu struct {
		sync.Mutex

		l      net.Listener
		miners map[string]string
	}
}

func New() (*Server, error) {
	s := &Server{
		rs:      rpc.NewServer(),
		stopper: stopper.New(),
		log:     log.New(os.Stderr, colors.Orange("server "), log.Flags()|log.Lshortfile),
	}

	s.mu.miners = map[string]string{}

	if err := s.rs.Register(&ServerRPC{s}); err != nil {
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
		select {
		case <-s.stopper.ShouldStop():
			return nil
		default:
		}

		conn, err := ln.Accept()
		if err != nil {
			s.log.Printf("Server accept error: %s", err)
			continue
		}
		s.log.Printf("New connection from: %s", conn.RemoteAddr())
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

	s.stopper.Stop()

	return s.mu.l.Close()
}

type RegisterRequest struct {
	PublicKey, Address string
}

func (s *Server) NumMiners() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.mu.miners)
}

type ServerRPC struct {
	s *Server
}

func (s *ServerRPC) Register(req RegisterRequest, resp *blockartlib.MinerNetSettings) error {
	s.s.mu.Lock()
	defer s.s.mu.Unlock()

	s.s.mu.miners[req.PublicKey] = req.Address

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

func (s *ServerRPC) GetNodes(req GetNodesRequest, resp *GetNodesResponse) error {
	s.s.mu.Lock()
	defer s.s.mu.Unlock()

	for _, addr := range s.s.mu.miners {
		resp.Addrs = append(resp.Addrs, addr)
	}

	return nil
}

type HeartBeatRequest struct {
	PublicKey string
}

type HeartBeatResponse struct{}

func (s *ServerRPC) HeartBeat(req HeartBeatRequest, resp *HeartBeatResponse) error {
	return nil
}
