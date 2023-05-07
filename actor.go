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

type Inventory struct {
}

type Actor struct {
	Id     Id
	Obj    *Obj
	Stats  *Stats
	Room   *Room
	Zone   *Zone
	Inv    *Inventory
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

func LoadActors(db *sql.DB, objs map[Id]*Obj) []*Actor {
	rows, err := db.Query(`
SELECT a.id, obj_id, stats
FROM actor a JOIN object o ON a.obj_id = o.id
ORDER BY o.room`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	actors := make([]*Actor, 0)
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
		actors = append(actors, &actor)
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return actors
}
