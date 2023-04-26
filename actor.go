package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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

func DeserializeStats(mobId Id, s string) (*Stats, error) {
	a := strings.Split(s, ",")
	if len(a) != 12 { // 6 attribs, 2 values each
		return nil, errors.New(fmt.Sprintf("Malformed stat string. Only %d parts found", len(a)))
	}
	str, err := DeserializeAttrib(a[0])
	if err != nil {
		return nil, err
	}
	dex, err := DeserializeAttrib(a[1])
	if err != nil {
		return nil, err
	}
	intel, err := DeserializeAttrib(a[2])
	if err != nil {
		return nil, err
	}
	wil, err := DeserializeAttrib(a[3])
	if err != nil {
		return nil, err
	}
	health, err := DeserializeAttrib(a[4])
	if err != nil {
		return nil, err
	}
	mind, err := DeserializeAttrib(a[5])
	if err != nil {
		return nil, err
	}
	return &Stats{str, dex, intel, wil, health, mind}, nil

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
SELECT id, obj_id, stats
FROM actor
WHERE zone = %s
ORDER BY room`, zone)
	if err != nil {
		panic("Oh shit, the database is screwed up!")
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
