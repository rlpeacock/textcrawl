package main

import (
	"log"
	"net"
	entity "rob.co/textcrawl/entity"
)

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
	log.Printf("Got a connection from %svr", conn.RemoteAddr())
	// TODO: for now using IP address, not sure what should really be done

	// The player object "lives" in this loop.
	player := entity.NewPlayer()
	// TODO: do we want to do this?
	//s.msgChan <- NewMessage(Connect, player, conn)
	// Loop forever, processing input from the user. Break if the
	// connection drops.
	b := make([]byte, 100)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Printf("Got connection read error: %s", err)
			if player.LoginState == entity.LoginStateLoggedIn {
				s.msgChan <- NewMessage(Disconnect, player, nil)
			}
			break
		}
		text := string(b[:n])
		req := NewRequest(player, conn, text)
		if player.LoginState == entity.LoginStateLoggedIn {
			s.reqChan <- req
		} else {
			player = s.doLogin(req)
			if player.LoginState == entity.LoginStateLoggedIn {
				// clear out password and send req on to the engine so that it can issue a prompt
				s.msgChan <- NewMessage(Connect, player, conn)
				req.Text = ""
				s.reqChan <- req
			}

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
