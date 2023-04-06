package main

import (
	"fmt"
	"io"
	"log"
	"time"
)

// each connection, when it receives a message, will put it on a channel
// the server will select on the channels and queue up the actiosn
// when it gets a tick it will order them by precedence, then execute them
// one by one. After all events are processed, it will generate responses
// and send them to a channel that the reader thread is selecting on

type Actor struct {
	Id string
}

func NewActor(id string) *Actor {
	return &Actor{
		Id: id,
	}
}

type Request struct {
	Writer io.Writer
	Text string
	Actor *Actor
}

func NewRequest(actor *Actor, writer io.Writer, text string) *Request {
	return &Request{
		Actor: actor,
		Writer: writer,
		Text: text,
	}
}

type Heartbeat struct {
	tick int
	cmd string
}

func newHeartbeat(tick int, cmd string) *Heartbeat {
	return &Heartbeat{
		tick: tick,
		cmd: cmd,
	}
}

type Engine struct {
	RequestCh   chan *Request
	HeartbeatCh chan *Heartbeat
	reqsByActor map[string][]*Request
}

func NewEngine() *Engine {
	return &Engine{
		RequestCh:   make(chan *Request, 0),
		HeartbeatCh: make(chan *Heartbeat, 0),
		reqsByActor: make(map[string][]*Request),
	}
}

func (e *Engine) Run() {
	for {
		select {
		case req := <-e.RequestCh:
			q := e.reqsByActor[req.Actor.Id]
			if q == nil {
				q = []*Request{req }
			} else {
				q = append(q, req)
			}
			e.reqsByActor[req.Actor.Id] = q
		case hb := <-e.HeartbeatCh:
			if hb.cmd == "quit" {
				return
			}
			e.processRequests(hb)
		}
	}
}

// Take the first unprocessed request we have from each actor.
func (e *Engine) processRequests(hb *Heartbeat) {
	log.Printf("tick %d", hb.tick)
	todo := make([]*Request, 0)
	for id, q := range e.reqsByActor {
		todo = append(todo, q[0])
		q = q[1:]
		if len(q) == 0 {
			delete(e.reqsByActor, id)
		} else {
			e.reqsByActor[id] = q
		}
	}
	for _, req := range todo {
		t := fmt.Sprintf("processing: %s (%d)\r\n", req.Text, hb.tick)
		req.Writer.Write([]byte(t))
		log.Print(t)
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
	s := NewServer(e.RequestCh)
	go s.Serve()
	go heartbeat(e.HeartbeatCh)
	e.Run()
	fmt.Println("Stopping")
}