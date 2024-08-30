package main

import (
	"log"
	"net"
	cmd "rob.co/textcrawl/command"
	entity "rob.co/textcrawl/entity"
)

type Server struct {
	msgChan chan Message
	reqChan chan Request
	conns   []net.Conn
}

func NewServer(msgChan chan Message, reqChan chan Request) *Server {
	return &Server{
		msgChan: msgChan,
		reqChan: reqChan,
		conns:   make([]net.Conn, 0),
	}
}

func (s *Server) Serve() {
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
		s.conns = append(s.conns, conn)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("Got a connection from %svr", conn.RemoteAddr())
	// TODO: for now using IP address, not sure what should really be done
	// Note that the new actor will not be placed anywhere until completing
	// login flow.
	actor := entity.NewActor(conn.RemoteAddr().String(), entity.NewPlayer())
	// Tell the engine we've got somebody joining
	s.msgChan <- NewMessage(Connect, actor, conn)
	// Loop forever, processing input from the user. Break if the
	// connection drops.
	b := make([]byte, 100)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Printf("Got connection read error: %s", err)
			s.msgChan <- NewMessage(Disconnect, actor, nil)
			break
		}
		text := string(b[:n])
		cmd := cmd.NewCommand(text)
		req := NewRequest(actor, conn, cmd)
		s.reqChan <- req
	}
}
