package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

// a zone is a collection of rooms that are managed together
// it acts as a boundary for interactivity - that is, everything
// in a zone will be handled by a single process and can interact
// without distributed transaction semantics. Cross zone interaction
// should be tightly restricted and needs to be hardened against
// IPC failure. Zones are stored as a yaml file and a SQLite database.
// The yaml file defines the layout and the DB holds current state.

type Zone struct {
	Id     Id
	Rooms  map[Id]*Room
	Actors map[Id]*Actor
	db     *sql.DB
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
	nodes := make(map[Id]*Room)
	for rid, room := range rooms {
		// Need to add id to room struct because it's a key rather than a field
		// in YAML peristence. I'm going to regret this...
		room.Id = "R" + rid
		for _, exit := range room.Exits {
			exit.Destination = "R" + exit.Destination
		}
		nodes[room.Id] = room
	}
	return nodes
}

func ensureZoneDB(destFile string) {
	if info, err := os.Stat(destFile); err != nil {
		newFile, err := os.Create(destFile)
		if err != nil {
			panic(fmt.Sprintf("Unable to create db %s: %s", destFile, err))
		}
		defer newFile.Close()
		template, err := os.Open("world/0.dat")
		if err != nil {
			panic(fmt.Sprintf("Could not find template file for creating new zone: %s", err))
		}
		defer template.Close()
		nBytes, err := io.Copy(newFile, template)
		if err != nil || nBytes == 0 {
			panic(fmt.Sprintf("Failed to make a new version of zone db: %s", err))
		}
	} else {
		if info.IsDir() {
			panic(fmt.Sprintf("DB creation failed. %s is a directory!", destFile))
		}
	}
}

func (z *Zone) loadZoneState() {
	f := filepath.Join("world", fmt.Sprintf("%s.dat", z.Id))
	ensureZoneDB(f)
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		panic(fmt.Sprintf("Could not open database %s", f))
	}
	z.db = db
	thingsById := LoadThings(z.db)
	// actors contain a thing reference for their physical form
	actorsById := LoadActors(z.db, thingsById)
	// add actors to rooms
	z.Actors = make(map[Id]*Actor)
	for _, actor := range actorsById {
		z.Actors[actor.ID()] = actor
		z.Rooms[actor.Body.ParentId].InsertActor(actor)
	}
	// add things to inventory, containers, or rooms
	for _, thing := range thingsById {
		switch IdTypeForId(thing.ParentId) {
		case IdTypeRoom:
			z.Rooms[thing.ParentId].Insert(thing)
		case IdTypeInventory:
			z.Actors[thing.ParentId].Insert(thing)
		case IdTypeContainer:
			thingsById[thing.ParentId].Insert(thing)
		default:
			log.Printf("Unable to find parent for %s with id %s", thing.Id, thing.ParentId)
		}
	}
}

func NewZone(id Id) (*Zone, error) {
	log.Printf("Loading zone %s", id)
	zone := &Zone{
		Id:     id,
		Rooms:  loadRooms(id),
		Actors: make(map[Id]*Actor),
	}
	zone.loadZoneState()
	return zone, nil
}

func (z *Zone) MoveActor(actor *Actor, room *Room) bool {
	curRoom := actor.Room()
	if curRoom != nil {
		curRoom.RemoveActor(actor)
	}
	room.InsertActor(actor)
	// In the future we might check any number of things,
	// but for now always succeed.
	return true
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
