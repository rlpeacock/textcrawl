package main

import (
	"log"
	"net"
	"time"

	entity "rob.co/textcrawl/entity"
)

type Sender struct {
	ch   chan []byte
	conn net.Conn
}

func NewSender(conn net.Conn) Sender {
	ch := make(chan []byte, 30)
	return Sender{
		ch:   ch,
		conn: conn,
	}
}

func (s Sender) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	s.ch <- pCopy
	return len(pCopy), nil
}

func (s Sender) doSend() {
	toSend := []byte{}
	for {
		p, ok := <-s.ch
		if !ok {
			s.Close()
			break
		}
		toSend = append(toSend, p...)
		err := s.conn.SetWriteDeadline(time.Now().Add(time.Second))
		if err != nil {
			log.Printf("Attempt to set write deadling failed: %s", err)
			s.Close()
			break
		}
		n, err := s.conn.Write(toSend)
		if err != nil {
			log.Printf("Error writing to connection: %s", err)
			s.Close()
			break
		}
		toSend = toSend[n:]
		if len(toSend) > 5000 {
			log.Printf("Backing up too much on connection. Dropping connection")
			s.Close()
			break
		}
	}
}

func (s Sender) Close() {
	err := s.conn.Close()
	if err != nil {
		print("Could not close the socket either!")
	}
	close(s.ch)
}

type Server struct {
	msgChan   chan Message
	reqChan   chan Request
	conns     []net.Conn
	playerMgr entity.PlayerMgr
}

func NewServer(msgChan chan Message, reqChan chan Request, playerMgr entity.PlayerMgr) *Server {
	return &Server{
		msgChan:   msgChan,
		reqChan:   reqChan,
		conns:     make([]net.Conn, 0),
		playerMgr: playerMgr,
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
	defer conn.Close()
	log.Printf("Got a connection from %svr", conn.RemoteAddr())
	// TODO: for now using IP address, not sure what should really be done

	// The player object "lives" in this loop.
	player := entity.NewPlayer()
	// TODO: do we want to do this?
	//s.msgChan <- NewMessage(Connect, player, conn)

	// Loop forever, processing input from the user. Break if the
	// connection drops.
	received := []byte{}
	buf := make([]byte, 500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Got connection read error: %s", err)
			if player.LoginState == entity.LoginStateLoggedIn {
				s.msgChan <- NewMessage(Disconnect, player, nil)
			}
			break
		}
		received = append(received, buf[:n]...)
		if len(received) > 1000 {
			log.Printf("Request too large. Killing connection")
			if player.LoginState == entity.LoginStateLoggedIn {
				s.msgChan <- NewMessage(Disconnect, player, nil)
			}
			break
		}
		if len(received) >= 2 && received[len(received)-2] == '\r' && received[len(received)-1] == '\n' {
			text := string(received[:len(received)-2])
			req := NewRequest(player, conn, text)
			if player.LoginState == entity.LoginStateLoggedIn {
				s.reqChan <- req
			} else {
				player = s.doLogin(req)
				if player.LoginState == entity.LoginStateLoggedIn {
					// clear out password and send req on to the engine so that it can issue a prompt
					s.msgChan <- NewMessage(Connect, player, conn)
					req.Text = ""
					req.Player = player
					s.reqChan <- req
				}

			}
			received = []byte{}
		}
	}
}

func (e *Server) doLogin(req Request) entity.Player {
	switch req.Player.LoginState {
	case entity.LoginStateStart:
		req.Write("Please enter your username: ")
		req.Player.LoginState = entity.LoginStateWantUser
	case entity.LoginStateWantUser:
		if req.Text != "" {
			req.Player.Username = req.Text
			req.Write("Please enter your password: ")
			req.Player.LoginState = entity.LoginStateWantPwd
		}
	case entity.LoginStateWantPwd:
		if req.Text != "" {
			// TODO: for now, we don't actually have passwords!
			actorId, err := e.playerMgr.LookupPlayer(req.Player.Username, "")
			if err != nil {
				log.Printf("Player lookup failed: %s", err)
				req.Write("I'm sorry...who?")
				req.Write("\n> ")
				req.Player.LoginAttempts++
			} else {
				req.Write("Login successful\n")
				req.Player.ActorId = actorId
				req.Player.LoginState = entity.LoginStateLoggedIn
			}
		}
	}
	return req.Player
}
