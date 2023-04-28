package main

import (
	"bytes"
	"io"
	"testing"
)

type testSession struct {
	req    *Request
	reader io.Reader
	ch     chan string
	e      *Engine
}

func newTestSession(e *Engine) *testSession {
	r, w := io.Pipe()
	a := NewActor("1", NewPlayer())
	c := NewCommand(a, "")
	t := &testSession{
		req:    NewRequest(a, w, c),
		reader: r,
		ch:     make(chan string),
		e:      e,
	}
	go t.readResponses()
	return t
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
	t.req.Cmd = NewCommand(t.req.Actor, text)
	t.e.RequestCh <- t.req
}

func (t *testSession) getResponse() string {
	return <-t.ch
}

func TestLogin(t *testing.T) {
	e := NewEngine()
	go e.Run()
	defer e.TriggerShutdown()
	ts := newTestSession(e)
	ts.sendRequest("a")
	s := ts.getResponse()
	if s != "Please Enter your username: " {
		t.Errorf("Expected username prompt, got '%s'", s)
	}
	ts.sendRequest("b")
	s = ts.getResponse()
	if s != "Please enter your password: " {
		t.Errorf("Expected password prompt, got '%s'", s)
	}
}
