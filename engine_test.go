package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type testSession struct {
	req    *Request
	reader io.Reader
	ch     chan string
	e      *Engine
	t      *testing.T
}

func newTestSession(e *Engine, t *testing.T) *testSession {
	r, w := io.Pipe()
	a := NewActor("1", NewPlayer())
	c, err := NewCommand(a, "")
	if err != nil {
		panic(fmt.Sprintf("Session creation failed because command creation got an error: %s", err))
	}
	ts := &testSession{
		req:    NewRequest(a, w, c),
		reader: r,
		ch:     make(chan string),
		e:      e,
		t:      t,
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
	cmd, e := NewCommand(t.req.Actor, text)
	t.req.Cmd = cmd
	if e != nil {
		t.t.Fatalf("Could not send request because parse failed: %s", e)
	}
	t.e.RequestCh <- t.req
}

func (t *testSession) getResponse() string {
	return <-t.ch
}

func TestLogin(t *testing.T) {
	e := NewEngine()
	go e.Run()
	defer e.TriggerShutdown()
	ts := newTestSession(e, t)
	if ts.req.Actor.Player.LoginState != LoginStateStart {
		t.Errorf("Should have started in %s, but actually was %s", LoginStateStart, ts.req.Actor.Player.LoginState)
	}
	// Because of how we're testing, we need to send an initial request to prime the session.
	// Normally this would happen when server received a connection.
	ts.sendRequest("something")
	s := ts.getResponse()
	if s != "Please enter your username: " {
		t.Errorf("Expected username prompt, got '%s'", s)
	}
	if ts.req.Actor.Player.LoginState != LoginStateWantUser {
		t.Errorf("Should have been in %s, but actually was %s", LoginStateWantUser, ts.req.Actor.Player.LoginState)
	}

	ts.sendRequest("username")
	s = ts.getResponse()
	if s != "Please enter your password: " {
		t.Errorf("Expected password prompt, got '%s'", s)
	}
	if ts.req.Actor.Player.LoginState != LoginStateWantPwd {
		t.Errorf("Should have been in %s, but actually was %s", LoginStateWantPwd, ts.req.Actor.Player.LoginState)
	}

	ts.sendRequest("pwd")
	s = ts.getResponse()
	if ts.req.Actor.Player.LoginState != LoginStateLoggedIn {
		t.Errorf("Should have been in %s, but actually was %s", LoginStateLoggedIn, ts.req.Actor.Player.LoginState)
	}
}
