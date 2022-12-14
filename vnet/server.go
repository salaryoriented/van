package vnet

import (
	"net"
	"sync/atomic"
	"van/core/log"
)

type Server struct {
	// 为每个连接分配的 id
	connId *int64

	*Config
	*log.Log
	ConnectionMgr *ConnectionMgr
	DataPack      *DataPack
	MsgHandle     *MsgHandle
}

func NewServer(config *Config, opts ...Option) (*Server, error) {
	s := &Server{
		connId:        new(int64),
		Config:        config,
		ConnectionMgr: NewConnectionMgr(),
		DataPack:      NewDataPack(),
		MsgHandle:     NewMsgHandle(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Server) setUp() error {
	if err := s.check(); err != nil {
		return err
	}

	return nil
}

func (s *Server) start() error {
	tcpAddr, err := net.ResolveTCPAddr(s.Network, s.Address())
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP(s.Network, tcpAddr)
	if err != nil {
		return err
	}

	s.LogInfo("listen tcp on: %s", s.Address())

	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				s.LogErr("lister accept tcp err: %v", err)
				return
			}
			s.LogInfo("receive a tcp conn from: %s", conn.RemoteAddr())
			_ = conn.SetReadBuffer(s.ReadBuffer)
			_ = conn.SetWriteBuffer(s.WriteBuffer)
			workConn := NewConnection(s.autoIncrConnId(), conn, s)
			go workConn.start()
		}
	}()

	return nil
}

func (s *Server) autoIncrConnId() int64 {
	return atomic.AddInt64(s.connId, 1)
}

func (s *Server) Server() {
	if err := s.start(); err != nil {
		s.LogErr("server start err: %v", err)
		return
	}

	select {}
}

func (s *Server) Stop() {
	s.LogInfo("stop Server")
}

func (s *Server) GetConnectionMgr() *ConnectionMgr {
	return s.ConnectionMgr
}

func (s *Server) GetDataPack() *DataPack {
	return s.DataPack
}

func (s *Server) AddRouter(router Router) {
	s.MsgHandle.Add(router)
}
