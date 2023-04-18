package main

import (
	"log"
	"net"
	"strings"
)

type Server struct {
	reqChan chan *Request
	conns   []net.Conn
}

func NewServer(reqChan chan *Request) *Server {
	return &Server{
		reqChan: reqChan,
		conns:   make([]net.Conn, 0),
	}
}

func (svr *Server) Serve() {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %svr", err)
			continue
		}
		svr.conns = append(svr.conns, conn)
		go svr.handleConnection(conn)
	}
}

func (svr *Server) handleConnection(conn net.Conn) {
	log.Printf("Got a connection from %svr", conn.RemoteAddr())
	// TODO: for now using IP address, not sure what should really be done here
	actor := NewActor(conn.RemoteAddr().String(), NewPlayer())
	b := make([]byte, 100)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Printf("Got actor read error: %svr", err)
			continue
		}
		txt := strings.TrimSpace(string(b[:n]))
		req := NewRequest(actor, conn, txt)
		svr.reqChan <- req
	}
}
