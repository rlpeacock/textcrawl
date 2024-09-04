package entity

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

func loadRooms(worldDir string, id Id) map[Id]*Room {
	rooms := make(map[Id]*Room)
	filename := filepath.Join(worldDir, fmt.Sprintf("%s.yaml", id))
	content, err := os.ReadFile(filename)
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

func ensureZoneDB(worldDir, destFile string) {
	if info, err := os.Stat(destFile); err != nil {
		newFile, err := os.Create(destFile)
		if err != nil {
			panic(fmt.Sprintf("Unable to create db %s: %s", destFile, err))
		}
		defer func(newFile *os.File) {
			err := newFile.Close()
			if err != nil {
				panic(fmt.Sprintf("Failed to close the db file %s: %s", destFile, err))
			}
		}(newFile)
		template, err := os.Open(filepath.Join(worldDir, "0.dat"))
		if err != nil {
			panic(fmt.Sprintf("Could not find template file for creating new zone: %s", err))
		}
		defer func(template *os.File) {
			err := template.Close()
			if err != nil {
				panic(fmt.Sprintf("Failed to close the template file: %s", err))
			}
		}(template)
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

func (z *Zone) loadZoneState(worldDir string) {
	f := filepath.Join(worldDir, fmt.Sprintf("%s.dat", z.Id))
	ensureZoneDB(worldDir, f)
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
		actor.Zone = z
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

func LoadZone(worldDir string, id Id) (*Zone, error) {
	log.Printf("Loading zone %s", id)
	zone := &Zone{
		Id:     id,
		Rooms:  loadRooms(worldDir, id),
		Actors: make(map[Id]*Actor),
	}
	zone.loadZoneState(worldDir)
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

func (z *Zone) TakeThing(thing *Thing, actor *Actor) bool {
	idType := IdTypeForId(thing.ParentId)
	if idType == IdTypeRoom {
		room := z.Rooms[thing.ParentId]
		room.Remove(thing)
	} else if idType == IdTypeContainer {
		// oh shit, I don't know how to find the container!
	} else if idType == IdTypeInventory {
		actor := z.Actors[thing.ParentId]
		actor.Remove(thing)
	}
	thing.ParentId = actor.ID()
	actor.Insert(thing)
	// TODO: we'll eventually check various things, like capacity, etc.
	return true
}

func (z *Zone) Save() {
	trans, err := z.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Unable to create a transaction in which to save state for zone %s: %s", z.Id, err))
	}
	defer func(trans *sql.Tx) {
		err := trans.Commit()
		if err != nil {
			panic(fmt.Sprintf("Unable to commit transaction for zone %s: %s", z.Id, err))
		}
	}(trans)
	for _, room := range z.Rooms {
		err := room.Save(z.db)
		if err != nil {
			panic(fmt.Sprintf("Unable to save state for zone %s: %s", z.Id, err))
		}
	}
}

type ZoneManager struct {
	zones map[Id]*Zone
}

func GetZoneMgr() (ZoneManager, error) {
	zones, err := loadZones()
	if err != nil {
		return ZoneManager{
			zones: map[Id]*Zone{},
		}, err
	}
	return ZoneManager{
		zones: zones,
	}, nil
}

func (zm *ZoneManager) GetZone(id Id) (*Zone, error) {
	z := zm.zones[id]
	if z == nil {
		return nil, fmt.Errorf("Unable to find zone %s", id)
	}
	return z, nil
}

func loadZones() (map[Id]*Zone, error) {
	zones := make(map[Id]*Zone, 0)
	worldDir := os.Getenv("TEXTCRAWL_WORLD")
	entries, err := os.ReadDir(worldDir)
	if err != nil {
		return nil, fmt.Errorf("Could not access world directory: %s", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		// any file in the world directory with a number for a name and .yaml suffix is a zone
		isZone, _ := regexp.MatchString(`\d+\.yaml`, name)
		if isZone {
			id, _ := strings.CutSuffix(name, ".yaml")
			z, err := LoadZone(worldDir, Id(id))
			if err != nil {
				return nil, fmt.Errorf("Error loading zones: %s", err)
			}
			zones[z.Id] = z
		}
	}
	return zones, nil
}

func (zm *ZoneManager) FindActor(actorId Id) (*Actor, error) {
	for _, zone := range zm.zones {
		// TODO: create a global actor list in zonemgr when zones are loaded
		for _, actor := range zone.Actors {
			if actor.Id == actorId {
				return actor, nil
			}
		}
	}
	return nil, fmt.Errorf("Actor '%s' cannot be found!", actorId)
}
