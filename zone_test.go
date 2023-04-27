package main

import (
	"testing"
)

func TestGetZone(t *testing.T) {
	z, e := GetZoneMgr().GetZone(Id("1"))
	if e != nil {
		t.Errorf("GetZoneMgr().GetZone(Id(1)) returned an error: %s", e)
	}
	r := z.GetRoom(Id("1"))
	if r.Title != "a room" {
		t.Errorf(`z.GetRoom(Id(1)) should have returned room 'a room', but got '%s'`, r.Title)
	}
}
