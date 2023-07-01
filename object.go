package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// An Id is a unique identifier.
// Ids are unique across entities.
// They start with a prefix which indicates entity type.
// (e.g. rooms all start with R)
type Id string

// IdType Specifies what type of object is referred to by an Id
type IdType int

const (
	IdTypeRoom IdType = iota
	IdTypeContainer
	IdTypeInventory
	IdTypeUnknown
)

// IdTypeForId Returns what type of object a given Id refers to.
// TODO: this is a workaround for lack of polymorphism.
// What's a more idiomatic way of doing stuff like this?
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

// ThingFlags Bit flags indicating state or features on a thing.
type ThingFlags int

// MatchLevel A type for specifying how close a match has matched.
// We use this for looking up what objects words refer to.
type MatchLevel int

// to indicate how closely a word matches the object's title
const (
	MatchNone MatchLevel = iota
	MatchPartial
	MatchPrimary
	MatchExact
)

// An Attrib holds a trait for an Actor or Thing.
// It has 2 components:
//   - the maximum, or healthy, value
//   - what the value currently is
type Attrib struct {
	Real int
	Cur  int
}

// --------------------------------

// A Thing is a physical object within the game.
// Things can contain other things
// and are themselves contained,
// either by another thing or by a room.
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

// NewThing Create a generic thing.
func NewThing() *Thing {
	return &Thing{
		Contents: make([]*Thing, 0),
	}
}

// ID Return the thing's ID.
// This is a method because we'll be calling it from Lua where polymorphism actually works.
func (t *Thing) ID() Id {
	return t.Id
}

// Accept Determine whether the supplied thing can be contained within this thing.
func (t *Thing) Accept(_ *Thing) bool {
	// TODO: later this will actually check capacity
	// but for now, anything can go in anything
	return true
}

// Match How closely does this word match the title of this thing?
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

// Find Search the contents of this thing for something that matches the supplied word.
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
	t.dirty = true
}

func (t *Thing) Remove(thing *Thing) bool {
	for i, item := range t.Contents {
		if item == thing {
			t.Contents = append(t.Contents[:i], t.Contents[i+1:]...)
			t.dirty = true
			return true
		}
	}
	return false
}

func SerializeAttrib(attrib Attrib) string {
	return fmt.Sprintf("%d:%d", attrib.Real, attrib.Cur)
}
func DeserializeAttrib(s string) (Attrib, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Attrib{}, errors.New("malformed attribute: missing delimiter")
	}
	realVal, err := strconv.Atoi(parts[0])
	if err != nil {
		return Attrib{}, errors.New("malformed attribute: realVal is not a number")
	}
	curVal, err := strconv.Atoi(parts[1])
	if err != nil {
		return Attrib{}, errors.New("malformed attribute: curVal is not a number")
	}
	return Attrib{realVal, curVal}, nil
}

func SerializeAttribList(attribs ...Attrib) string {
	serialized := ""
	for _, attrib := range attribs {
		serialized = serialized + "," + SerializeAttrib(attrib)
	}
	return serialized
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
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	things := make(map[Id]*Thing)
	for rows.Next() {
		thing := NewThing()
		var (
			attribs string
		)
		err = rows.Scan(&thing.Id, &attribs, &thing.Title, &thing.Desc, &thing.ParentId, &thing.Flags)
		if err != nil {
			panic(fmt.Sprintf("Error while iterating rows: %s", err))
		}
		err = DeserializeAttribList(attribs, &thing.Weight, &thing.Size, &thing.Durability)
		if err != nil {
			log.Printf("Failed to deserialize attributes for object %s", thing.Id)
			continue
		}
		things[thing.Id] = thing
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return things
}

func (t *Thing) Save(db *sql.DB) error {
	for _, child := range t.Contents {
		err := child.Save(db)
		if err != nil {
			return err
		}
	}
	if !t.dirty {
		return nil
	}
	attribs := SerializeAttribList(t.Weight, t.Size, t.Durability)
	res, err := db.Exec(`UPDATE thing SET attributes = ?, location = ?, flags =? WHERE id = ?`, attribs, t.ParentId, t.Id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows != 1 {
		log.Printf("Something went wrong with update to %s. %d rows were updated rather than 1.", t.Id, rows)
		return errors.New(fmt.Sprintf("unexpected update result when saving t %s: %d rows affected", t.Id, rows))
	}
	t.dirty = false
	return nil
}
