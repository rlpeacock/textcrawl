package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type LoginState int

// Don't add anything here without updating String() func below!
// TODO: can I use introspection to make this less fragile?
const (
	LoginStateStart LoginState = iota
	LoginStateWantUser
	LoginStateWantPwd
	LoginStateFailed
	LoginStateMaxFailed
	LoginStateLoggedIn
)

type Stats struct {
	Str    Attrib
	Dex    Attrib
	Int    Attrib
	Will   Attrib
	Health Attrib
	Mind   Attrib
}

// An Actor is an entity that can perform actions. Actors may
// be players, or they may be NPCs/MOBs. The Actor structure
// holds attributes of the actor outside of it's actual place
// in the game. That is represented by the 'body' attribute.
type Actor struct {
	Id        Id
	Body      *Thing
	Stats     *Stats
	Zone      *Zone
	Player    *Player
	Inventory []*Thing
}

func NewActor(id string, player *Player) *Actor {
	return &Actor{
		Id:     Id(id),
		Player: player,
		Body: &Thing{
			Title: "yourself",
		},
	}
}

func (a *Actor) ID() Id {
	// TODO: this is used to tell objects who their owner is,
	// and we want to use the actor's body as the owner to
	// simplify loading. However, the name is confusing for
	// the actor object. Rename?
	return a.Body.Id
}

func (a *Actor) SetRoom(r *Room) {
	a.Body.ParentId = r.Id
}

func (a *Actor) GetTitle() string {
	return a.Body.Title
}

// Will this actor accept possession of supplied object?
func (a *Actor) Accept(thing Thing) bool {
	// TODO: check carrying capacity but for now can carry anything
	return true
}

// return object holding position of actor within the object hierarchy
func (a *Actor) Room() *Room {
	return a.Zone.GetRoom(a.Body.ParentId)
}

func (a *Actor) Match(word string) MatchLevel {
	if a.GetTitle() == word {
		return MatchExact
	}
	if strings.HasPrefix(a.GetTitle(), word) {
		return MatchPrimary
	}
	if strings.Contains(a.GetTitle(), word) {
		return MatchPartial
	}
	return MatchNone
}

func (a *Actor) Find(word string) *Thing {
	bestMatch := struct {
		match MatchLevel
		thing *Thing
	}{match: MatchNone}
	for _, item := range a.Inventory {
		match := item.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.thing = item
		}
	}
	return bestMatch.thing
}

func (a *Actor) Insert(child *Thing) {
	a.Inventory = append(a.Inventory, child)
}

func (a *Actor) Remove(thing *Thing) {
	for i, item := range a.Inventory {
		if item == thing {
			a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
		}
	}
}

// Load all actors from SQLite DB
func LoadActors(db *sql.DB, things map[Id]*Thing) map[Id]*Actor {
	rows, err := db.Query(`
SELECT a.id, thing_id, stats
FROM actor a JOIN thing t ON a.thing_id = t.id`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	actors := make(map[Id]*Actor)
	for rows.Next() {
		actor := Actor{}
		var (
			rawStats string
			thingId  string
		)
		err = rows.Scan(&actor.Id, &thingId, &rawStats)
		if err != nil {
			panic(fmt.Sprintf("Error will scanning actor row: %s", err))
		}
		actor.Body = things[Id(thingId)]
		if actor.Body == nil {
			log.Printf("WARN: actor '%s' has invalid object id '%s'", actor.Id, thingId)
			continue
		}
		stats := Stats{}
		err = DeserializeAttribList(rawStats, &stats.Str, &stats.Dex, &stats.Int, &stats.Will, &stats.Health, &stats.Mind)
		if err != nil {
			log.Printf("WARN: error loading actors %s: %s", actor.Id, err)
			continue
		}
		actors[actor.Id] = &actor
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return actors
}
