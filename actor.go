package main

import (
	"database/sql"
	"fmt"
	"log"

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

type Actor struct {
	Id     Id
	Body   *Thing
	Stats  *Stats
	Zone   *Zone
	Player *Player
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

func (a *Actor) ParentID() Id {
	return a.Body.Parent
}

func (a *Actor) GetTitle() string {
	return a.Body.Title
}

func (a *Actor) Take(obj GObject) bool {
	// TODO: check carrying capacity but for now can carry anything
	return true
}

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
