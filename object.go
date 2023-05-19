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

type ThingFlags int

type MatchLevel int

// to indicate how closely a word matches the object's title
const (
	MatchNone MatchLevel = iota
	MatchPartial
	MatchPrimary
	MatchExact
)

type GObject interface {
	ID() Id
	ParentID() Id
	GetTitle() string
	Take(obj GObject) bool
}

type Locus struct {
	Parent   *Locus
	Children []*Locus
	Object   GObject
}

func NewLocus(obj GObject) *Locus {
	return &Locus{
		Children: make([]*Locus, 0),
		Object:   obj,
	}
}

func (n *Locus) Insert(child *Locus) bool {
	if n.Object.Take(child.Object) {
		if child.Parent != nil {
			for i, o := range child.Parent.Children {
				if o == child {
					child.Parent.Children = append(child.Parent.Children[:i], child.Parent.Children[i+1:]...)
				}
			}
		}
		child.Parent = n
		n.Children = append(n.Children, child)
		return true
	}
	return false
}

func (l *Locus) ID() Id {
	return l.Object.ID()
}

func (l *Locus) ParentID() Id {
	return l.Object.ParentID()
}

type Attrib struct {
	Real int
	Cur  int
}

func (n *Locus) Match(word string) MatchLevel {
	if n.Object.GetTitle() == word {
		return MatchExact
	}
	if strings.HasPrefix(n.Object.GetTitle(), word) {
		return MatchPrimary
	}
	if strings.Contains(n.Object.GetTitle(), word) {
		return MatchPartial
	}
	return MatchNone
}

func (n *Locus) Find(word string) *Locus {
	bestMatch := struct {
		match MatchLevel
		node  *Locus
	}{match: MatchNone}
	for _, child := range n.Children {
		match := child.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.node = child
		}
	}
	return bestMatch.node
}

// --------------------------------

type Thing struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
	Parent     Id
	Flags      ThingFlags
	dirty      bool
}

func (t *Thing) GetTitle() string {
	return t.Title
}

func (t *Thing) ID() Id {
	return t.Id
}

func (t *Thing) ParentID() Id {
	return t.Parent
}

func (t *Thing) Take(obj GObject) bool {
	// TODO: later this will actually check capacity
	// but for now, anything can go in anything
	return true
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
		thing := Thing{}
		var (
			attribs string
		)
		err = rows.Scan(&thing.Id, &attribs, &thing.Title, &thing.Desc, &thing.Parent, &thing.Flags)
		if err != nil {
			panic(fmt.Sprintf("Error while iterating rows: %s", err))
		}
		DeserializeAttribList(attribs, &thing.Weight, &thing.Size, &thing.Durability)
		things[thing.Id] = &thing
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return things
}
