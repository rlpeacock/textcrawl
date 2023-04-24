package main

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// a zone is a collection of rooms that are managed together
// it acts as a boundary for interactivity - that is, everything
// in a zone will be handled by a single process and can interact
// without distributed transaction semantics. Cross zone interaction
// should be tightly restricted and needs to be hardened against
// IPC failure.

type Zone struct {
	Rooms map[Id]*Room
}

func (z *Zone) GetRoom(id Id) *Room {
	return z.Rooms[id]
}

func loadZoneState(id Id) ([]*Obj, []*Actor) {
	f := filepath.Join("world", fmt.Sprintf("%s.dat", id))
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		panic(fmt.Sprintf("Could not find database %s", f))
	}
	objs := LoadObjects(db, id)
	mobs := LoadMOBs(db, id)
	return objs, mobs
}

func NewZone(id Id) (*Zone, error) {
	rooms := make(map[Id]*Room)
	// rooms, err := LoadSampleRooms()
	// TODO: process zone state objs into room members
	// we also need to hook up the actors with the objs
	return &Zone{
		Rooms: rooms,
	}, nil
}

type ZoneManager struct {
	zones map[Id]*Zone
}

func GetZoneMgr() *ZoneManager {
	return &ZoneManager{
		zones: make(map[Id]*Zone),
	}
}

func (zm *ZoneManager) GetZone(id Id) (*Zone, error) {
	z := zm.zones[id]
	if z == nil {
		var err error
		z, err = NewZone(id)
		if err != nil {
			return nil, err
		}
		zm.zones[id] = z
	}
	return z, nil
}
