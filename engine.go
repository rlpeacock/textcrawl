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

type Request struct {
	Writer io.Writer
	Text string
}

func NewRequest(writer io.Writer, text string) *Request {
	return &Request{
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
	reqQueue    []*Request
}

func NewEngine() *Engine {
	return &Engine{
		RequestCh:   make(chan *Request, 0),
		HeartbeatCh: make(chan *Heartbeat, 0),
		reqQueue:    make([]*Request, 0),
	}
}

func (e *Engine) Run() {
	for {
		select {
		case req := <-e.RequestCh:
			e.reqQueue = append(e.reqQueue, req)
		case hb := <-e.HeartbeatCh:
			if hb.cmd == "quit" {
				return
			}
			e.processRequests(hb)
		}
	}
}

func (e *Engine) processRequests(hb *Heartbeat) {
	log.Printf("tick %d", hb.tick)
	for _, r := range e.reqQueue {
		log.Printf("processing: %s", r.Text)
		t := fmt.Sprintf("tick %d\r\n", hb.tick)
		r.Writer.Write([]byte(t))
	}
	e.reqQueue = nil
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