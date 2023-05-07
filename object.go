package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Id string

type ObjectFlags int

const (
	MOB = 1
)

type Attrib struct {
	Real int
	Cur  int
}

type Obj struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
	Room       *Room
	Flags      ObjectFlags
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

func DeserializeAttribList(attribStr string, attribs ...*Attrib) error {
	rawAttribs := strings.Split(attribStr, ",")
	if len(rawAttribs) < len(attribs) {
		return errors.New(fmt.Sprintf("Attribute does not contain enough values. Expected %d, got %d", len(attribs), len(rawAttribs)))
	}
	if len(rawAttribs) > len(attribs) {
		return errors.New(fmt.Sprintf("Attribute contains too many values. Expected %d, got %d", len(attribs), len(rawAttribs)))
	}
	for i, a := range rawAttribs {
		attrib, err := DeserializeAttrib(a)
		if err != nil {
			return err
		}
		attribs[i].Cur = attrib.Cur
		attribs[i].Real = attrib.Real
	}
	return nil
}

func LoadObjects(db *sql.DB, rooms map[Id]*Room) []*Obj {
	rows, err := db.Query(`
SELECT id, attributes, title, description, room, flags
FROM object
ORDER BY room`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	objs := make([]*Obj, 0)
	for rows.Next() {
		obj := Obj{}
		var (
			attribs string
			roomId  string
		)
		err = rows.Scan(&obj.Id, &attribs, &obj.Title, &obj.Desc, &roomId, &obj.Flags)
		if err != nil {
			panic(fmt.Sprintf("Error while iterating rows: %s", err))
		}
		room := rooms[Id(roomId)]
		if room == nil {
			print(fmt.Sprintf("WARN: object '%s' is in invalid room '%s'", obj.Id, roomId))
		}
		room.Take(&obj)
		DeserializeAttribList(attribs, &obj.Weight, &obj.Size, &obj.Durability)
		objs = append(objs, &obj)
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return objs
}
