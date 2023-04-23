package main

import (
	"fmt"
	"io"
	"log"
	"time"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// each connection, when it receives a message, will put it on a channel
// the server will select on the channels and queue up the actiosn
// when it gets a tick it will order them by precedence, then execute them
// one by one. After all events are processed, it will generate responses
// and send them to a channel that the reader thread is selecting on

// Message types the engine can handle
type MessageType int

const (
	CONNECT MessageType = iota
	DISCONNECT
)

type Message struct {
	mType MessageType
	Actor *Actor
}

func NewMessage(t MessageType, a *Actor) Message {
	return Message{
		mType: t,
		Actor: a,
	}
}

type Request struct {
	Writer io.Writer
	Cmd    *Command
	Actor  *Actor
}

func (r *Request) Write(msg string) {
	r.Writer.Write([]byte(msg))
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
	zoneMgr     *ZoneManager
	luaState    *lua.LState
}

func NewEngine() *Engine {
	ls := lua.NewState()
	if err := ls.DoFile("lib/commands.lua"); err != nil {
		panic(fmt.Sprintf("Script execution failed: %s", err))
	}

	return &Engine{
		RequestCh:   make(chan *Request, 0),
		HeartbeatCh: make(chan *Heartbeat, 0),
		MessageCh:   make(chan Message, 0),
		reqsByActor: make(map[Id][]*Request),
		zoneMgr:     GetZoneMgr(),
		luaState:    ls,
	}
}

func (e *Engine) ensureLoggedIn(req *Request) bool {
	switch req.Actor.Player.LoginState {
	case LOGIN_COMPLETE:
		return true
	case NOT_STARTED:
		req.Write("Please Enter your username: ")
		req.Actor.Player.LoginState = WAITING_FOR_USERNAME
	case WAITING_FOR_USERNAME:
		if req.Cmd.Action != "" {
			req.Actor.Player.Username = req.Cmd.Action
			req.Write("Please enter your password: ")
			req.Actor.Player.LoginState = WAITING_FOR_PASSWORD
		}
	case WAITING_FOR_PASSWORD:
		if req.Cmd.Action != "" {
			// TODO: for now, we don't actually have passwords!
			req.Write("Login successful\n")
			req.Actor.Player.LoginState = LOGIN_COMPLETE
			// TODO: we also don't have persistent sessions so give an arbitrary location
			z, err := e.zoneMgr.GetZone(Id("1"))
			if err != nil {
				log.Printf("Zone get failed: %s", err)
				req.Write("WTF")
				return false
			}
			req.Actor.Room = z.Rooms[Id("1")]
			req.Actor.Zone = z
			e.sendPrompt(req)
		}
	}
	return false
}

func (e *Engine) sendPrompt(req *Request) {
	// this will eventually have status in it
	req.Write("\n> ")
}

func (e *Engine) dispatch(req *Request) {
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

// We queue up requests for each actor. When we receive a
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
			case CONNECT:
				log.Printf("INFO: %s has connected", msg.Actor.Id)
				e.reqsByActor[msg.Actor.Id] = []*Request{}
			case DISCONNECT:
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
	// Go through and handle each request. TODO: we should order these
	// by init value and account for multi-tick actions.
	for _, req := range todo {
		log.Print(fmt.Sprintf("processing: %s (%d)\r\n", req.Cmd.Action, hb.tick))
		e.dispatch(req)
	}
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
