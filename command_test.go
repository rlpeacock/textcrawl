package main

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	z, e := GetZoneMgr().GetZone(Id("1"))
	if e != nil {
		t.Fatalf("Couldn't even get the zone! %s", e)
	}
	a := NewActor("", nil)
	a.Room = z.GetRoom(Id("1"))
	cmd, e := NewCommand(a, " look ")
	if e != nil {
		t.Fatalf("Error parsing command ' look ': %s", e)
	}
	if cmd.Action != "look" {
		t.Errorf("Expected action 'look' but got '%s'", cmd.Action)
	}

	cmd, e = NewCommand(a, "sw")
	if e != nil {
		t.Fatalf("Error parsing command 'sw': %s", e)
	}
	if cmd.Action != "goDirection" {
		t.Errorf("Expected action 'goDirection' but got '%s'", cmd.Action)
	}
	if len(cmd.Params) != 1 {
		t.Fatalf("Expected 1 parameter but got %d", len(cmd.Params))
	}
	if cmd.Params[0] != "southwest" {
		t.Errorf("Expected parameter 'southwest' but got '%s'", cmd.Params[0])
	}

	cmd, e = NewCommand(a, "take knife")
	if e != nil {
		t.Fatalf("Error parsing command 'take knife': %s", e)
	}
	if cmd.Action != "take" {
		t.Errorf("Expected action 'take' but got '%s'", cmd.Action)
	}
	if len(cmd.Obj) != 1 {
		t.Fatalf("Expected 1 direct object but got %d", len(cmd.Obj))
	}
	if _, ok := cmd.Obj[0].(*Thing); !ok {
		t.Errorf("Expected and entity type 'thing' but that's not what we got")
	}
	if cmd.Obj[0].ID() != "T1" {
		t.Errorf("Expected T1 but got '%s'", cmd.Obj[0].ID())
	}
}
