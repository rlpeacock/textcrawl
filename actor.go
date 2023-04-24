package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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

type Attrib struct {
	Real int
	Cur  int
}

func DeserializeAttrib(s string) (Attrib, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Attrib{}, errors.New("Malformed attribute: missing delimiter")
	}
	real, err := strconv.Atoi(parts[0])
	if err != nil {
		return Attrib{}, errors.New("Malformed attribute: real is not a number")
	}
	cur, err := strconv.Atoi(parts[1])
	if err != nil {
		return Attrib{}, errors.New("Malformed attribute: cur is not a number")
	}
	return Attrib{real, cur}, nil
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

func LoadMOBs(db *sql.DB, zone Id) []*Actor {
	rows, err := db.Query(`
SELECT id, obj_id, stats
FROM mob
WHERE zone = %s
ORDER BY room`, zone)
	if err != nil {
		panic("Oh shit, the database is screwed up!")
	}
	mobs := make([]*Actor, 0)
	for rows.Next() {
		var (
			id       Id
			objId    Id
			rawStats string
		)
		rows.Scan(&id, &objId, &rawStats)
		stats, err := DeserializeStats(id, rawStats)
		if err != nil {
			log.Printf("WARN: error loading MOB %s: %s", id, err)
			continue
		}
		mob := Actor{
			Id:    id,
			Stats: stats,
		}
		mobs = append(mobs, &mob)
	}
	return mobs
}
