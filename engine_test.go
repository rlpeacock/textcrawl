package main

import (
	"bytes"
	"io"
	"testing"
	entity "rob.co/textcrawl/entity"
	cmd "rob.co/textcrawl/command"
)

type testSession struct {
	req    Request
	reader io.Reader
	ch     chan string
	e      *Engine
	t      *testing.T
}

func newTestSession(e *Engine, t *testing.T) *testSession {
	r, w := io.Pipe()
	a := entity.NewActor("1", entity.NewPlayer())
	c := cmd.NewCommand("")
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
	cmd := cmd.NewCommand(text)
	t.req.Cmd = cmd
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
	if ts.req.Actor.Player.LoginState != entity.LoginStateStart {
		t.Errorf("Should have started in %s, but actually was %s", entity.LoginStateStart, ts.req.Actor.Player.LoginState)
	}
	// Because of how we're testing, we need to send an initial request to prime the session.
	// Normally this would happen when server received a connection.
	ts.sendRequest("something")
	s := ts.getResponse()
	if s != "Please enter your username: " {
		t.Errorf("Expected username prompt, got '%s'", s)
	}
	if ts.req.Actor.Player.LoginState != entity.LoginStateWantUser {
		t.Errorf("Should have been in %s, but actually was %s", entity.LoginStateWantUser, ts.req.Actor.Player.LoginState)
	}

	ts.sendRequest("username")
	s = ts.getResponse()
	if s != "Please enter your password: " {
		t.Errorf("Expected password prompt, got '%s'", s)
	}
	if ts.req.Actor.Player.LoginState != entity.LoginStateWantPwd {
		t.Errorf("Should have been in %s, but actually was %s", entity.LoginStateWantPwd, ts.req.Actor.Player.LoginState)
	}

	// TODO: This test predated password validation. Need to put in some fixtures to make this work again.
	// ts.sendRequest("pwd")
	// s = ts.getResponse()
	// if ts.req.Actor.Player.LoginState != entity.LoginStateLoggedIn {
	// 	t.Errorf("Should have been in %s, but actually was %s", entity.LoginStateLoggedIn, ts.req.Actor.Player.LoginState)
	// }
}
