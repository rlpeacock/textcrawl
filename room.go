package main

import "strings"

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
	Contents  Inventory
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
	if a.Room == nil || a.Room.Remove(a) {
		r.Occupants = append(r.Occupants, a)
		a.NewRoom(r)
		return true
	}
	return false
}

func (r *Room) Remove(a *Actor) bool {
	for i, o := range r.Occupants {
		if o == a {
			r.Occupants = append(r.Occupants[:i], r.Occupants[i+1:]...)
			return true
		}
	}
	return false
}

func (r *Room) Take(o *Thing) bool {
	if o.Owner == nil || o.Owner.Give(o) {
		r.Contents = append(r.Contents, o)
		o.NewOwner(r)
		return true
	}
	return false
}

func (r *Room) Give(o *Thing) bool {
	for i, c := range r.Contents {
		if c == o {
			r.Contents = append(r.Contents[:i], r.Contents[i+1:]...)
			return true
		}
	}
	return false
}

func (r *Room) ID() Id {
	return r.Id
}

func (r *Room) Match(word string) MatchLevel {
	if r.Title == word {
		return MatchExact
	}
	if strings.HasPrefix(r.Title, word) {
		return MatchPrimary
	}
	if strings.Contains(r.Title, word) {
		return MatchPartial
	}
	return MatchNone
}

func (r *Room) Find(word string) Entity {
	bestMatch := struct {
		match  MatchLevel
		entity Entity
	}{match: MatchNone}
	// An actor match has priority over an object match
	// so search occupants first
	for _, a := range r.Occupants {
		match := a.Body.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.entity = a
		}
	}
	for _, o := range r.Contents {
		match := o.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.entity = o
		}
	}
	return bestMatch.entity
}
