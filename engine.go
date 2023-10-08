package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// each connection, when it receives a message, will put it on a channel
// the server will select on the channels and queue up the actiosn
// when it gets a tick it will order them by precedence, then execute them
// one by one. After all events are processed, it will generate responses
// and send them to a channel that the reader thread is selecting on

// MessageType message types the engine can handle
type MessageType int

const (
	Connect MessageType = iota
	Disconnect
)

const LuaEntrypoint = "lib/commands.lua"

type Message struct {
	mType  MessageType
	Writer io.Writer
	Actor  *Actor
}

func NewMessage(t MessageType, a *Actor, w io.Writer) Message {
	return Message{
		mType:  t,
		Actor:  a,
		Writer: w,
	}
}

type Request struct {
	Writer io.Writer
	Cmd    *Command
	Actor  *Actor
}

func (r *Request) Write(msg string) {
	// Ignore errors for now. Not clear what we can do. Possibly add a counter to track eventually.
	_, _ = r.Writer.Write([]byte(msg))
}

func NewRequest(actor *Actor, writer io.Writer, cmd *Command) *Request {
	return &Request{
		Actor:  actor,
		Writer: writer,
		Cmd:    cmd,
	}
}

type Heartbeat struct {
	tick int
	cmd  string
}

func newHeartbeat(tick int, cmd string) *Heartbeat {
	return &Heartbeat{
		tick: tick,
		cmd:  cmd,
	}
}

type Engine struct {
	RequestCh   chan *Request
	HeartbeatCh chan *Heartbeat
	MessageCh   chan Message
	reqsByActor map[Id][]*Request
	playerMgr   *PlayerMgr
	zoneMgr     *ZoneManager
	luaState    *lua.LState
	loadTime    time.Time
}

func NewEngine() *Engine {
	ls := lua.NewState()
	if err := ls.DoFile(LuaEntrypoint); err != nil {
		panic(fmt.Sprintf("Script execution failed: %s", err))
	}
	zm, err := GetZoneMgr()
	if err != nil {
		log.Fatalf("Unable to start engine: %s", err)
	}

	return &Engine{
		RequestCh:   make(chan *Request),
		HeartbeatCh: make(chan *Heartbeat),
		MessageCh:   make(chan Message),
		reqsByActor: make(map[Id][]*Request),
		playerMgr:   NewPlayerMgr(),
		zoneMgr:     zm,
		luaState:    ls,
		loadTime:    time.Now(),
	}
}

func (e *Engine) ensureLoggedIn(req *Request) bool {
	switch req.Actor.Player.LoginState {
	case LoginStateLoggedIn:
		return true
	case LoginStateStart:
		req.Actor.Player.LoginState = LoginStateWantUser
		req.Write("Please enter your username: ")
	case LoginStateWantUser:
		if req.Cmd.Text != "" {
			req.Actor.Player.Username = req.Cmd.Text
			req.Actor.Player.LoginState = LoginStateWantPwd
			req.Write("Please enter your password: ")
		}
	case LoginStateWantPwd:
		if req.Cmd.Text != "" {
			// TODO: for now, we don't actually have passwords!
			player, actorId, err := e.playerMgr.LookupPlayer(req.Actor.Player.Username, "")
			if err != nil {
				log.Printf("Player lookup failed: %s", err)
				req.Write("I'm sorry...who?")
				e.sendPrompt(req)
				return false
			}
			actor, err := e.zoneMgr.FindActor(actorId)
			if err != nil {
				log.Printf("Player actor lookup failed: %s", err)
				req.Write("I know who you are but I don't know WHO you are!")
				e.sendPrompt(req)
				return false
			}
			actor.Player = player
			req.Actor = actor

			req.Write("Login successful\n")
			req.Actor.Player.LoginState = LoginStateLoggedIn
			e.sendPrompt(req)
			return true
		}
	}
	return false
}

func (e *Engine) sendPrompt(req *Request) {
	// this will eventually have status in it
	req.Write("\n> ")
}

func (e *Engine) dispatch(req *Request) {
	req.Cmd.ResolveWords(req.Actor.Room(), req.Actor)
	if req.Cmd.Action == "" {
		return
	}
	args := []lua.LValue{luar.New(e.luaState, req)}
	for _, a := range req.Cmd.Params {
		args = append(args, luar.New(e.luaState, a))
	}
	err := e.luaState.CallByParam(lua.P{
		Fn:      e.luaState.GetGlobal(req.Cmd.Action),
		NRet:    1,
		Protect: true,
	}, args...)
	if err != nil {
		log.Printf("Script did not succed: %s", err)
		req.Write(fmt.Sprintf("We failed to %s\n", req.Cmd.Action))
	}
	e.sendPrompt(req)
}

// Run We queue up requests for each actor. When we receive a
// heartbeat message, we process the events we've received.
// Generally this means taking the first message from each
// actor.
func (e *Engine) Run() {
	for {
		select {
		case req := <-e.RequestCh:
			if !e.ensureLoggedIn(req) {
				break
			}
			q := e.reqsByActor[req.Actor.Id]
			if q == nil {
				// Should have been created connect message, but just to be safe...
				log.Printf("WARN: Request queue missing for actor %s", req.Actor.Id)
				q = []*Request{req}
			} else {
				q = append(q, req)
			}
			e.reqsByActor[req.Actor.Id] = q
		case hb := <-e.HeartbeatCh:
			if hb.cmd == "quit" {
				return
			}
			e.processRequests(hb)
		case msg := <-e.MessageCh:
			switch msg.mType {
			case Connect:
				log.Printf("INFO: %s has connected", msg.Actor.Id)
				e.reqsByActor[msg.Actor.Id] = []*Request{}
				// For a new connection, kick the login flow so the user gets a prompt
				e.ensureLoggedIn(NewRequest(msg.Actor, msg.Writer, NewCommand("")))
			case Disconnect:
				log.Printf("INFO: %s has disconnected", msg.Actor.Id)
				delete(e.reqsByActor, msg.Actor.Id)
			}
		}
	}
}

func (e *Engine) processRequests(hb *Heartbeat) {
	log.Printf("tick %d", hb.tick)
	todo := make([]*Request, 0)
	// Take the first unprocessed request we have from each actor.
	for id, q := range e.reqsByActor {
		if len(q) > 0 {
			todo = append(todo, q[0])
			e.reqsByActor[id] = q[1:]
		}
	}
	// check whether we need to reload the lua engine
	info, err := os.Stat(LuaEntrypoint)
	if err != nil {
		panic("What happened to our lua library!")
	}
	if info.ModTime().After(e.loadTime) {
		ls := lua.NewState()
		err := ls.DoFile(LuaEntrypoint)
		if err == nil {
			e.luaState.Close()
			e.luaState = ls
		} else {
			log.Printf("Unable to reload lua engine: %s", err)
		}
	}
	// Go through and handle each request. TODO: we should order these
	// by init value and account for multi-tick actions.
	zoneIds := make(map[Id]bool)
	for _, req := range todo {
		log.Print(fmt.Sprintf("processing: %s (%d)\r\n", req.Cmd.Action, hb.tick))
		e.dispatch(req)
		zoneIds[req.Actor.Zone.Id] = true
	}
	// Now save any zones in which actions have occurred
	for zid := range zoneIds {
		zone, err := e.zoneMgr.GetZone(zid)
		if err != nil {
			panic(fmt.Sprintf("WTF, no zone %s", zid))
		}
		zone.Save()
	}
}

func (e *Engine) TriggerShutdown() {
	e.HeartbeatCh <- &Heartbeat{cmd: "quit"}
}

func heartbeat(c chan *Heartbeat) {
	tick := 0
	for {
		tick += 1
		msg := ""
		if tick == 10000 {
			msg = "quit"
		}
		hb := newHeartbeat(tick, msg)
		c <- hb
		time.Sleep(1 * time.Second)
	}

}

func main() {
	fmt.Println("Starting engine")
	e := NewEngine()
	s := NewServer(e.MessageCh, e.RequestCh)
	go s.Serve()
	go heartbeat(e.HeartbeatCh)
	e.Run()
	fmt.Println("Stopping")
}
