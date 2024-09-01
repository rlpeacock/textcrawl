package main

import (
	"fmt"
	"io"
	"log"
	cmd "rob.co/textcrawl/command"
	entity "rob.co/textcrawl/entity"
	"time"
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
	Player  entity.Player
}

func NewMessage(t MessageType, p entity.Player, w io.Writer) Message {
	return Message{
		mType:  t,
		Player:  p,
		Writer: w,
	}
}

type Request struct {
	Writer io.Writer
	Text   string
	Player  entity.Player
}

func (r Request) Write(msg string) {
	// Ignore errors for now. Not clear what we can do. Possibly add a counter to track eventually.
	_, _ = r.Writer.Write([]byte(msg))
}

func NewRequest(player entity.Player, writer io.Writer, rawCmd string) Request {
	return Request{
		Player:  player,
		Writer: writer,
		Text:    rawCmd,
	}
}

type Heartbeat struct {
	tick int
	cmd  string
}

func newHeartbeat(tick int, cmd string) Heartbeat {
	return Heartbeat{
		tick: tick,
		cmd:  cmd,
	}
}

type Engine struct {
	RequestCh   chan Request
	HeartbeatCh chan Heartbeat
	MessageCh   chan Message
	reqsByActor map[entity.Id][]Request
	playerMgr   entity.PlayerMgr
	zoneMgr     entity.ZoneManager
	loadTime    time.Time
}

func NewEngine() *Engine {
	zm, err := entity.GetZoneMgr()
	if err != nil {
		log.Fatalf("Unable to start engine: %s", err)
	}

	return &Engine{
		RequestCh:   make(chan Request),
		HeartbeatCh: make(chan Heartbeat),
		MessageCh:   make(chan Message),
		reqsByActor: make(map[entity.Id][]Request),
		playerMgr:   entity.NewPlayerMgr(),
		zoneMgr:     zm,
		loadTime:    time.Now(),
	}
}



func (e *Engine) sendPrompt(req Request) {
	// this will eventually have status in it
	req.Write("\n> ")
}

func (e *Engine) dispatch(req Request, cmd cmd.Command, actor *entity.Actor) {
	// TODO: this should be done earlier
	//cmd.ResolveWords(actor.Room(), actor)
	if cmd.Action == "" {
		return
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
			q := e.reqsByActor[req.Player.ActorId]
			if q == nil {
				// Should have been created connect message, but just to be safe...
				log.Printf("WARN: Request queue missing for actor %s", req.Player.ActorId)
				q = []Request{req}
			} else {
				q = append(q, req)
			}
			e.reqsByActor[req.Player.ActorId] = q
		case hb := <-e.HeartbeatCh:
			if hb.cmd == "quit" {
				return
			}
			e.processRequests(hb)
		case msg := <-e.MessageCh:
			switch msg.mType {
			case Connect:
				log.Printf("INFO: %s has connected", msg.Player.ActorId)
				e.reqsByActor[msg.Player.ActorId] = []Request{}
			case Disconnect:
				log.Printf("INFO: %s has disconnected", msg.Player.ActorId)
				delete(e.reqsByActor, msg.Player.ActorId)
			}
		}
	}
}

func (e *Engine) processRequests(hb Heartbeat) {
	log.Printf("tick %d", hb.tick)
	todo := make([]Request, 0)
	// Take the first unprocessed request we have from each actor.
	for id, q := range e.reqsByActor {
		if len(q) > 0 {
			todo = append(todo, q[0])
			e.reqsByActor[id] = q[1:]
		}
	}
	// Go through and handle each request. TODO: we should order these
	// by init value and account for multi-tick actions.


	zoneIds := make(map[entity.Id]bool)
	for _, req := range todo {
		a, err := e.zoneMgr.FindActor(req.Player.ActorId)
		if err != nil {
			log.Printf("We are receiving commmands from unknown actor '%s'. Command was '%s'", 
				req.Player.ActorId, req.Text)
			continue
		}
		c := cmd.NewCommand(req.Text, a, a.Room())
		log.Print(fmt.Sprintf("processing: %s (%d)\r\n", c.Action, hb.tick))
		cmd.Perform(c, req.Writer)
		e.sendPrompt(req)
		zoneIds[a.Zone.Id] = true
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
	e.HeartbeatCh <- Heartbeat{cmd: "quit"}
}

func heartbeat(c chan Heartbeat) {
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
	s := NewServer(e.MessageCh, e.RequestCh, e.playerMgr)
	go s.Serve()
	go heartbeat(e.HeartbeatCh)
	e.Run()
	fmt.Println("Stopping")
}
