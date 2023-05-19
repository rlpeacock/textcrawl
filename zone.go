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
	Rooms  map[Id]*Locus
	Actors map[Id]*Locus
	db     *sql.DB
}

func (z *Zone) GetRoom(id Id) *Room {
	return z.Rooms[id].Object.(*Room)
}

func loadRooms(id Id) map[Id]*Locus {
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
	nodes := make(map[Id]*Locus)
	for rid, room := range rooms {
		// Need to add id to room struct because it's a key rather than a field
		// in YAML peristence.
		room.Id = rid
		nodes[id] = NewLocus(room)
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
	actorsById := LoadActors(z.db, thingsById)
	z.Actors = make(map[Id]*Locus)
	// Wrap everything in Locus objects to represent their placement
	// within the game.
	lociById := make(map[Id]*Locus)
	for _, thing := range thingsById {
		lociById[thing.ID()] = NewLocus(thing)
	}
	for _, actor := range actorsById {
		actorLoc := NewLocus(actor)
		lociById[actor.ID()] = actorLoc
		// Also add actors to actor lookup map
		z.Actors[actor.ID()] = actorLoc
	}
	// Rooms are already wrapped, just add them to the big map
	for _, room := range z.Rooms {
		lociById[room.ID()] = room
	}
	// Create a tree of loci by adding every Locus to it's parent
	for _, locus := range lociById {
		pId := locus.Object.ParentID()
		parent := lociById[pId]
		if parent == nil {
			log.Printf(fmt.Sprintf("WARN: object '%s' has an invalid parent '%s'", locus.ID(), pId))
			continue
		}
		parent.Insert(locus)
	}
}

func NewZone(id Id) (*Zone, error) {
	log.Printf("Loading zone %s", id)
	zone := &Zone{
		Id:     id,
		Rooms:  loadRooms(id),
		Actors: make(map[Id]*Locus),
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
