package main

import (
	"bytes"
	"io"
	"testing"
	entity "rob.co/textcrawl/entity"
)

type DummyPlayerMgr struct {
}

func (pm DummyPlayerMgr) LookupPlayer(username string, pwd string) (entity.Id, error) {
	return "A1", nil
}

func newTestEngine() *Engine {
	e := NewEngine()
	e.playerMgr = DummyPlayerMgr{}
	return e
}

type testSession struct {
	player entity.Player
	writer io.Writer
	reader io.Reader
	ch     chan string
	engine      *Engine
	t      *testing.T
	actor *entity.Actor
}

func newTestSession(e *Engine, t *testing.T) *testSession {
	r, w := io.Pipe()
	p := entity.NewPlayer()
	a := entity.NewActor("1", p)
	ts := &testSession{
		player: p,
		writer: w,
		reader: r,
		ch:     make(chan string),
		engine:      e,
		t:      t,
		actor: a,
	}
	go ts.readResponses()
	return ts
}

func (t *testSession) readResponses() {
	for {
		b := make([]byte, 1000)
		_, e := t.reader.Read(b)
		if e != nil {
			break
		}
		s := string(bytes.Trim(b, "\x00"))
		t.ch <- s
	}
}

func (t *testSession) sendRequest(text string) {
	req := NewRequest(t.player, t.writer, text)
	t.engine.RequestCh <- req
}

func (t *testSession) getResponse() string {
	return <-t.ch
}


