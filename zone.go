package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

// a zone is a collection of rooms that are managed together
// it acts as a boundary for interactivity - that is, everything
// in a zone will be handled by a single process and can interact
// without distributed transaction semantics. Cross zone interaction
// should be tightly restricted and needs to be hardened against
// IPC failure.

type Zone struct {
	Id    Id
	Rooms map[Id]*Room
	db    *sql.DB
}

func (z *Zone) GetRoom(id Id) *Room {
	return z.Rooms[id]
}

func loadRooms(id Id) map[Id]*Room {
	rooms := make(map[Id]*Room, 0)
	filename := filepath.Join("world", fmt.Sprintf("%s.yaml", id))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Unable to read zone file for zone %s: %s", id, err))
	}
	err = yaml.Unmarshal(content, &rooms)
	if err != nil {
		panic(fmt.Sprintf("Zone %s YAML is not valid: %s:", id, err))
	}
	return rooms
}

func ensureZoneStateDB(id Id) *sql.DB {
	f := filepath.Join("world", fmt.Sprintf("%s.dat", id))
	// TODO: create the DB if it does not exist. Should not only
	// create the DB but also apply the DDL.
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		panic(fmt.Sprintf("Could not find database %s", f))
	}
	return db
}

func (z *Zone) loadZoneState() {
	// z.db = ensureZoneStateDB(z.Id)
	// objs := LoadObjects(z.db, z.Id)
	// mobs := LoadMOBs(z.db, z.Id)
	// TODO: process zone state objs into room members
	// we also need to hook up the actors with the objs
}

func NewZone(id Id) (*Zone, error) {
	log.Printf("Loading zone %s", id)
	// db, objs, mobs := loadZoneState(id)
	zone := &Zone{
		Id:    id,
		Rooms: loadRooms(id),
		// db: db,
	}
	zone.loadZoneState()
	return zone, nil
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
	// TODO: lazy loading kind of sucks since we can't handle panics well.
	// If we have corrupt world data we probably shouldn't even start up.
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
