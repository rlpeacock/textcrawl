package main

import (
	"testing"
)

func TestGetZone(t *testing.T) {
	z, e := GetZoneMgr().GetZone(Id("1"))
	if e != nil {
		t.Errorf("GetZoneMgr().GetZone(Id(1)) returned an error: %s", e)
	}
	loc := z.Rooms[Id("R1")]
	if loc.Title != "a room" {
		t.Errorf(`z.GetRoom(Id(1)) should have returned room 'a room', but got '%s'`, loc.Title)
	}
	if len(loc.Things) == 0 {
		t.Fatalf("Room 1 should have an item in it but is empty")
	}
	// This test is broken because children are no longer in determinate order
	// child := loc.Children[0]
	// if child.Object.Title != "tin knife" {
	// 	t.Errorf("Expected object in room 1 to be 'tin knife', but got '%s'", child.Object.GetTitle())
	// }
}
