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

type IdType int

const (
	IdTypeRoom IdType = iota
	IdTypeContainer
	IdTypeInventory
	IdTypeUnknown
)

func IdTypeForId(id Id) IdType {
	if strings.HasPrefix(string(id), "R") {
		return IdTypeRoom
	} else if strings.HasPrefix(string(id), "C") {
		return IdTypeContainer
	} else if strings.HasPrefix(string(id), "A") {
		return IdTypeInventory
	}
	return IdTypeUnknown
}

type ThingFlags int

type MatchLevel int

// to indicate how closely a word matches the object's title
const (
	MatchNone MatchLevel = iota
	MatchPartial
	MatchPrimary
	MatchExact
)

type Attrib struct {
	Real int
	Cur  int
}

// --------------------------------

type Thing struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
	Contents   []*Thing
	ParentId   Id
	Flags      ThingFlags
	dirty      bool
}

func NewThing() *Thing {
	return &Thing{
		Contents: make([]*Thing, 0),
	}
}

func (t *Thing) ID() Id {
	return t.Id
}

func (t *Thing) Accept(child *Thing) bool {
	// TODO: later this will actually check capacity
	// but for now, anything can go in anything
	return true
}

func (t *Thing) Match(word string) MatchLevel {
	if t.Title == word {
		return MatchExact
	}
	if strings.HasPrefix(t.Title, word) {
		return MatchPrimary
	}
	if strings.Contains(t.Title, word) {
		return MatchPartial
	}
	return MatchNone
}

func (t *Thing) Find(word string) *Thing {
	bestMatch := struct {
		match MatchLevel
		thing *Thing
	}{match: MatchNone}
	for _, item := range t.Contents {
		match := item.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.thing = item
		}
	}
	return bestMatch.thing
}

func (t *Thing) Insert(child *Thing) {
	t.Contents = append(t.Contents, child)
}

func (t *Thing) Remove(thing *Thing) bool {
	for i, item := range t.Contents {
		if item == thing {
			t.Contents = append(t.Contents[:i], t.Contents[i+1:]...)
			return true
		}
	}
	return false
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

func LoadThings(db *sql.DB) map[Id]*Thing {
	rows, err := db.Query(`
SELECT id, attributes, title, description, location, flags
FROM thing
ORDER BY location`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	things := make(map[Id]*Thing, 0)
	for rows.Next() {
		thing := NewThing()
		var (
			attribs string
		)
		err = rows.Scan(&thing.Id, &attribs, &thing.Title, &thing.Desc, &thing.ParentId, &thing.Flags)
		if err != nil {
			panic(fmt.Sprintf("Error while iterating rows: %s", err))
		}
		DeserializeAttribList(attribs, &thing.Weight, &thing.Size, &thing.Durability)
		things[thing.Id] = thing
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return things
}
