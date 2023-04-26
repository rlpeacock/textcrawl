package main

import (
	"log"
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
