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
	if len(cmd.DirectObjs) != 1 {
		t.Fatalf("Expected 1 direct object but got %d", len(cmd.DirectObjs))
	}
	if _, ok := cmd.DirectObjs[0].(*Thing); !ok {
		t.Errorf("Expected and entity type 'thing' but that's not what we got")
	}
	if cmd.DirectObjs[0].ID() != "T1" {
		t.Errorf("Expected T1 but got '%s'", cmd.DirectObjs[0].ID())
	}

	cmd, e = NewCommand(a, "give knife to man")
	if e != nil {
		t.Fatalf("Error parsing command 'give knife to man': %s", e)
	}
	if cmd.Action != "give" {
		t.Errorf("Expected action 'give' but got '%s'", cmd.Action)
	}
	if cmd.DirectObjs[0].ID() != "T1" {
		t.Errorf("Expected T1 for direct object but got '%s'", cmd.DirectObjs[0].ID())
	}
	if cmd.Preposition != "to" {
		t.Errorf("Expected preposition 'to' but got '%s'", cmd.Preposition)
	}
	if cmd.IndirectObjs[0].ID() != "T2" {
		t.Errorf("Expected T2 for indirect object but got '%s'", cmd.IndirectObjs[0].ID())
	}

	cmd, e = NewCommand(a, "give knife bucket to man")
	if e != nil {
		t.Fatalf("Error parsing command 'give knife bucket to man': %s", e)
	}
	if cmd.Action != "give" {
		t.Errorf("Expected action 'give' but got '%s'", cmd.Action)
	}
	if len(cmd.DirectObjs) != 2 {
		t.Fatalf("Expected 2 direct objects but got %d", len(cmd.DirectObjs))
	}
	if cmd.DirectObjs[0].ID() != "T1" {
		t.Errorf("Expected T1 for direct object but got '%s'", cmd.DirectObjs[0].ID())
	}
	if cmd.DirectObjs[1].ID() != "T3" {
		t.Errorf("Expected T3 for direct object but got '%s'", cmd.DirectObjs[1].ID())
	}
	if cmd.Preposition != "to" {
		t.Errorf("Expected preposition 'to' but got '%s'", cmd.Preposition)
	}
	if cmd.IndirectObjs[0].ID() != "T2" {
		t.Errorf("Expected T2 for indirect object but got '%s'", cmd.IndirectObjs[0].ID())
	}

}
