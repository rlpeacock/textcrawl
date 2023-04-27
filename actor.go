package main

import (
	"database/sql"
	"fmt"
	"log"
)

type LoginFlowState int

const (
	NOT_STARTED = iota
	WAITING_FOR_USERNAME
	WAITING_FOR_PASSWORD
	BAD_LOGIN_INFO
	MAX_FAILURES
	LOGIN_COMPLETE
)

type Player struct {
	LoginState    LoginFlowState
	LoginAttempts int
	Username      string
}

func NewPlayer() *Player {
	return &Player{
		LoginState: NOT_STARTED,
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
