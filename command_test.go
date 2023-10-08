package main

import (
	"testing"
)

func DoCommand(text string) *Command {
	a := NewActor("1", NewPlayer())
	zm, _ := GetZoneMgr()
	zone, _ := zm.GetZone("1")
	room := zone.Rooms["R1"]
	cmd := NewCommand(text)
	cmd.ResolveWords(room, a)
	return cmd
}

func TestNewCommand(t *testing.T) {
	cmd := DoCommand(" look ")
	if cmd.Action != "look" {
		t.Errorf("Expected action 'look' but got '%s'", cmd.Action)
	}

	cmd = DoCommand("sw")
	if cmd.Action != "goDirection" {
		t.Errorf("Expected action 'goDirection' but got '%s'", cmd.Action)
	}
	if len(cmd.Params) != 1 {
		t.Fatalf("Expected 1 parameter but got %d", len(cmd.Params))
	}
	if cmd.Params[0] != "southwest" {
		t.Errorf("Expected parameter 'southwest' but got '%s'", cmd.Params[0])
	}

	cmd = DoCommand("take knife")
	if cmd.Action != "take" {
		t.Errorf("Expected action 'take' but got '%s'", cmd.Action)
	}
	if len(cmd.DirectObjs) != 1 {
		t.Fatalf("Expected 1 direct object but got %d", len(cmd.DirectObjs))
	}
	if thing, ok := cmd.DirectObjs[0].Ref.(*Thing); !ok {
		t.Errorf("Expected an entity type 'thing' but that's not what we got")
	} else {
		if thing.ID() != "T1" {
			t.Errorf("Expected T1 but got '%s'", thing.ID())
		}
	}

	cmd = DoCommand("give knife to man")
	if cmd.Action != "give" {
		t.Errorf("Expected action 'give' but got '%s'", cmd.Action)
	}
	if thing, ok := cmd.DirectObjs[0].Ref.(*Thing); !ok {
		t.Errorf("Expected an entity type 'thing' but that's not what we got")
	} else {
		if thing.ID() != "T1" {
			t.Errorf("Expected T1 for direct object but got '%s'", thing.ID())
		}
	}
	if cmd.Preposition != "to" {
		t.Errorf("Expected preposition 'to' but got '%s'", cmd.Preposition)
	}
	if actor, ok := cmd.IndirectObjs[0].Ref.(*Actor); !ok {
		t.Error("Expected an entity type 'actor' but that's not what we got")
	} else {
		if actor.ID() != "T2" {
			t.Errorf("Expected T2 for indirect object but got '%s'", actor.ID())
		}
	}

	cmd = DoCommand("give knife bucket to man")
	if cmd.Action != "give" {
		t.Errorf("Expected action 'give' but got '%s'", cmd.Action)
	}
	if len(cmd.DirectObjs) != 2 {
		t.Fatalf("Expected 2 direct objects but got %d", len(cmd.DirectObjs))
	}
	if cmd.DirectObjs[0].Ref.(*Thing).ID() != "T1" {
		t.Errorf("Expected T1 for direct object but got '%s'", cmd.DirectObjs[0].Ref.(*Thing).ID())
	}
	if cmd.DirectObjs[1].Ref.(*Thing).ID() != "T3" {
		t.Errorf("Expected T3 for direct object but got '%s'", cmd.DirectObjs[1].Ref.(*Thing).ID())
	}
	if cmd.Preposition != "to" {
		t.Errorf("Expected preposition 'to' but got '%s'", cmd.Preposition)
	}
	if cmd.IndirectObjs[0].Ref.(*Actor).ID() != "T2" {
		t.Errorf("Expected T2 for indirect object but got '%s'", cmd.IndirectObjs[0].Ref.(*Actor).ID())
	}

}
