package main

import (
	"database/sql"
	"fmt"
	"log"
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

type Player struct {
	LoginState    LoginState
	LoginAttempts int
	Username      string
}

func NewPlayer() *Player {
	return &Player{
		LoginState: LoginStateStart,
	}
}

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
	}
}

func LoadActors(db *sql.DB, zone Id) []*Actor {
	rows, err := db.Query(`
SELECT a.id, obj_id, stats
FROM actor a JOIN object o ON a.obj_id = o.id
WHERE o.zone = $1
ORDER BY o.room`, zone)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	actors := make([]*Actor, 0)
	for rows.Next() {
		actor := Actor{}
		var rawStats string
		rows.Scan(&actor.Id, &actor.Obj, &rawStats)
		stats := Stats{}
		err = DeserializeAttribList(rawStats, &stats.Str, &stats.Dex, &stats.Int, &stats.Will, &stats.Health, &stats.Mind)
		if err != nil {
			log.Printf("WARN: error loading actors %s: %s", actor.Id, err)
			continue
		}
		actors = append(actors, &actor)
	}
	return actors
}
