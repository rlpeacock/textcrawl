package main

import "strings"

type Direction string

type MoveType string

type Exit struct {
	Direction Direction
	Destination Id
	MoveType MoveType
	MoveDifficulty Attrib
	MoveSpeed Attrib
}

type Room struct {
	Id        Id
	Title     string
	Desc      string
	Occupants []*Actor
	Contents  []*Obj
	Exits     []*Exit
}

// simpleConnect create symetric connections between
// two rooms. The direction will be the one used for
// the first to the seccond. The second will get the
// inverse.
func simpleConnect(r1 *Room, r2 *Room, dir string) {
	e1 := &Exit{
		Direction: Direction(dir),
		Destination: r2.Id,
	}
	r1.Exits = append(r1.Exits, e1)
	if strings.Contains(dir, "north") {
		dir = strings.ReplaceAll(dir, "north", "south")
	} else {
		dir = strings.ReplaceAll(dir, "south", "north")
	}
	if strings.Contains(dir, "east") {
		dir = strings.ReplaceAll(dir, "east", "west")
	} else {
		dir = strings.ReplaceAll(dir, "west", "east")
	}
	if strings.Contains(dir, "up") {
		dir = strings.ReplaceAll(dir, "up", "down")
	} else {
		dir = strings.ReplaceAll(dir, "down", "up")
	}
	e2 := &Exit{
		Direction: Direction(dir),
		Destination: r1.Id,
		}
	r2.Exits = append(r2.Exits, e2)
}

func SampleRooms() []*Room {
	r1 := &Room{
		Id: Id(1),
		Title: "Featureless room",
		Desc: "A strangely featureless room with a single door to the east.",
		Exits: make([]*Exit, 0),
	}
	r2 := &Room{
		Id: Id(2),
		Title: "Featureless room",
		Desc: "A strangely featureless room with a single door to the west.",
		Exits: make([]*Exit, 0),
	}
	simpleConnect(r1, r2, "east")
	return []*Room{
		r1, r2,
	}
}
