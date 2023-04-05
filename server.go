package main

import (
	"log"
	"net"
)

type Server struct {
	reqChan chan *Request
	conns []net.Conn
}

func NewServer(reqChan chan *Request) *Server {
	return &Server {
		reqChan: reqChan,
		conns: make([]net.Conn, 0),
	}
}

func (s *Server)  Serve() {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		s.conns = append(s.conns, conn)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("Got a connection from %s", conn.RemoteAddr())
	b := make([]byte, 100)
	for {
		_, err := conn.Read(b)
		if err != nil {
			log.Printf("Got a read error: %s", err)
		}
		r := NewRequest(conn, string(b))
		s.reqChan <- r
	}
}