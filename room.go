package main

import (
	"log"

	"gopkg.in/yaml.v3"
)

type Direction string

type MoveType string

type Exit struct {
	Direction      Direction
	Destination    Id
	MoveType       MoveType
	MoveDifficulty Attrib
	MoveSpeed      Attrib
}

type Room struct {
	Id        Id
	Title     string
	Desc      string
	Occupants []*Actor
	Contents  []*Obj
	Exits     []*Exit
}

func (r *Room) GetExit(d Direction) *Exit {
	for _, e := range r.Exits {
		if e.Direction == d {
			return e
		}
	}
	return nil
}

func (r *Room) Receive(a *Actor) bool {
	if a.Room.Remove(a) {
		r.Occupants = append(r.Occupants, a)
		a.Room = r
		return true
	}
	log.Printf("Actor is now in %s", a.Room.Id)
	return false
}

func (r *Room) Remove(a *Actor) bool {
	for i, o := range r.Occupants {
		if o == a {
			r.Occupants = append(r.Occupants[:i], r.Occupants[i+1:]...)
			return true
		}
	}
	// TODO: we don't actually properly maintain occupant lists yet
	return true
}

func LoadSampleRooms() (map[Id]*Room, error) {
	rooms := make(map[Id]*Room, 0)
	err := yaml.Unmarshal([]byte(sample_rooms), &rooms)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

var sample_rooms = `
1:
  title: a room
  desc: This is an empty room. It only exists as a sample.
  exits:
    - direction: north
      destination: 2
2:
  title: another room
  desc: This is an even emptier room!
  exits:
    - direction: south
      destination: 1
`
