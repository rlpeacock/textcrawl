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
	// Need to add id to room struct
	for id, room := range rooms {
		room.Id = id
	}
	return rooms
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
	_ = LoadActors(z.db, thingsById)
	// insert objects into their owner's inventory
	for _, thing := range thingsById {
		switch thing.Loc[0] {
		case 'C':
			container := thingsById[thing.Loc]
			if container != nil {
				container.Take(thing)
			} else {
				log.Printf(fmt.Sprintf("Object '%s' belongs to unknown container '%s'", thing.Id, thing.Loc))
			}
		case 'R':
			// Room IDs don't start with R to keep things simple for YAML
			room := z.Rooms[thing.Loc[1:]]
			if room != nil {
				room.Take(thing)
			} else {
				log.Printf(fmt.Sprintf("Object '%s' belongs to unknown room '%s'", thing.Id, thing.Loc))
			}

		}
	}
}

func NewZone(id Id) (*Zone, error) {
	log.Printf("Loading zone %s", id)
	// db, things, mobs := loadZoneState(id)
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
