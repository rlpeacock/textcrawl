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
	Obj    *Obj
	Stats  *Stats
	Room   *Room
	Zone   *Zone
	Player *Player
}

func NewActor(id string, player *Player) *Actor {
	return &Actor{
		Id:     Id(id),
		Player: player,
		Obj: &Obj{
			Title: "yourself",
		},
	}
}

func (a *Actor) NewRoom(r *Room) {
	a.Room = r
	a.Obj.NewOwner(r)
}

func (a *Actor) Give(o *Obj) bool {
	return a.Obj.Give(o)
}

func (a *Actor) Take(o *Obj) bool {
	return a.Obj.Take(o)
}

func (a *Actor) ID() Id {
	// TODO: this is used to tell objects who their owner is,
	// and we want to use the actor's object as the owner to
	// simplify loading. However, the name is confusing for
	// the actor object. Rename?
	return a.Obj.Id
}

func LoadActors(db *sql.DB, objs map[Id]*Obj) map[Id]*Actor {
	rows, err := db.Query(`
SELECT a.id, obj_id, stats
FROM actor a JOIN object o ON a.obj_id = o.id`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	actors := make(map[Id]*Actor, 0)
	for rows.Next() {
		actor := Actor{}
		var (
			rawStats string
			objId    string
		)
		err = rows.Scan(&actor.Id, &objId, &rawStats)
		if err != nil {
			panic(fmt.Sprintf("Error will scanning actor row: %s", err))
		}
		obj := objs[Id(objId)]
		if obj == nil {
			log.Printf("WARN: actor '%s' has invalid object id '%s'", actor.Id, objId)
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
