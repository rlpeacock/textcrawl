package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// A LoginState holds the state of the player w.r.t login flow
type LoginState int

const (
	LoginStateStart     LoginState = iota // initial state
	LoginStateWantUser                    // waiting for username
	LoginStateWantPwd                     // waiting for password
	LoginStateFailed                      // previous creds were invalid, we're starting over
	LoginStateMaxFailed                   // player has failed login too many times
	LoginStateLoggedIn                    // player has successfully logged in
)

// A Stats is a structure for holding the attributes of an actor.
type Stats struct {
	Str    Attrib
	Dex    Attrib
	Int    Attrib
	Will   Attrib
	Health Attrib
	Mind   Attrib
}

// An Actor is an entity that can perform actions.
// Actors may be players, or they may be NPCs/MOBs.
// The Actor structure contains information about the actor's identity
// (e.g. stats, name, species)
// Information about the embodyment of the actor within the game
// (e.g. location, inventory)
// are stored in the Body member.
// In other words, if the actor dies, the Body is what's left.
type Actor struct {
	Id     Id
	Body   *Thing  // The physical presence of the actor within the game
	Stats  *Stats  //
	Zone   *Zone   // The current zone for this actor, if any
	Player *Player // The player, if any, associated with this actor
	dirty  bool    // whether the actor has been modified from initial state
}

// Returns a generic actor
func NewActor(id string, player *Player) *Actor {
	return &Actor{
		Id:     Id(id),
		Player: player,
		Body: &Thing{
			Title:    "yourself", // TODO: something real
			Contents: make([]*Thing, 0),
		},
	}
}

// Returns the ID associated with this actor.
// Note that while the actor has an ID, this is not it.
// Why? Because when we load from persistence,
// We need to be able to find the body to contain it's inventory.
// TODO: fix this abomination!
func (a *Actor) ID() Id {
	return a.Body.Id
}

// Modifies the actor's body's location
func (a *Actor) SetRoom(r *Room) {
	a.Body.ParentId = r.Id
	a.dirty = true
}

// Gets a title for the actor.
// Note that the title actually comes from the body.
// TODO: I don't remember why this is this way.
// Can we stop doing this?
func (a *Actor) GetTitle() string {
	return a.Body.Title
}

// Will this actor accept possession of supplied object?
func (a *Actor) Accept(thing Thing) bool {
	// TODO: check carrying capacity but for now can carry anything
	return true
}

// Return object holding position of actor within the object hierarchy.
// This is currently a room, but conceivably someday it could be another
// Thing.
func (a *Actor) Room() *Room {
	return a.Zone.GetRoom(a.Body.ParentId)
}

// Determine how closely a word matches the title of this actor.
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

// Search actor's inventory to see if it contains an object that matches this word.
// If multiple objects match, it will return the closest match. In cases of tie,
// the first match is returned.
func (a *Actor) Find(word string) interface{} {
	bestMatch := struct {
		match MatchLevel
		thing interface{}
	}{match: MatchNone}
	for _, item := range a.Body.Contents {
		match := item.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.thing = item
		}
	}
	return bestMatch.thing
}

// Attempt to place a thing in the actor's inventory.
// Returns whether the attempt succeeded.
func (a *Actor) Take(thing *Thing) bool {
	return a.Zone.TakeThing(thing, a)
}

// Attempt to drop something from the actor's inventory into the current room.
// Returns whether the attempt succeeded.
func (a *Actor) Drop(thing *Thing) bool {
	room := a.Room()
	thing.ParentId = room.Id
	if a.Body.Remove(thing) {
		room.Insert(thing)
		a.dirty = true
		return true
	}
	return false
}

// Unconditionally add something to the actor's inventory.
func (a *Actor) Insert(child *Thing) {
	a.Body.Insert(child)
}

// Unconditionally remove something from the actor's inventory.
func (a *Actor) Remove(thing *Thing) {
	a.Body.Remove(thing)
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
