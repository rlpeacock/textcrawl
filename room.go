package main

import "database/sql"

// Direction A cardinal direction
type Direction string

// MoveType Kind of movement
// (e.g. walk, fly, swim)
type MoveType string

// An Exit connects one room to another.
type Exit struct {
	Direction      Direction
	Destination    Id
	MoveType       MoveType
	MoveDifficulty Attrib
	MoveSpeed      Attrib
}

// A Room describes a location within the game.
// It is not necessarily an actual room.
type Room struct {
	Id     Id
	Title  string
	Desc   string
	Exits  []*Exit
	Actors []*Actor
	Things []*Thing
	dirty  bool
}

// GetExit Gets the Exit associated with a particular direction, if any.
func (r *Room) GetExit(d Direction) *Exit {
	for _, e := range r.Exits {
		if e.Direction == d {
			return e
		}
	}
	return nil
}

// AcceptThing Determines whether the supplied thing can be placed within this room.
// Eventual reasons for not might include:
//   - the object is too big
//   - the room is full
//   - magic
func (r *Room) AcceptThing(_ *Thing) bool {
	return true
}

// AcceptActor Determines whether the supplied actor can be placed within the room.
// See [AcceptThing] for reasons why it might not.
func (r *Room) AcceptActor(_ *Actor) bool {
	return true
}

// Find Search the room for objects which match the supplied word.
func (r *Room) Find(word string) interface{} {
	bestMatch := struct {
		match   MatchLevel
		matched interface{}
	}{match: MatchNone}
	for _, item := range r.Actors {
		match := item.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.matched = item
		}
	}
	for _, item := range r.Things {
		match := item.Match(word)
		if match > bestMatch.match {
			bestMatch.match = match
			bestMatch.matched = item
		}
	}
	return bestMatch.matched
}

// InsertActor Unconditionally add an actor to the room
func (r *Room) InsertActor(actor *Actor) {
	actor.SetRoom(r)
	r.Actors = append(r.Actors, actor)
}

// RemoveActor Unconditionally remove an actor from the room.
func (r *Room) RemoveActor(actor *Actor) {
	for i, item := range r.Actors {
		if item == actor {
			r.Actors = append(r.Actors[:i], r.Actors[i+1:]...)
		}
	}
}

// Insert Unconditionally add a thing to the room.
func (r *Room) Insert(thing *Thing) {
	thing.ParentId = r.Id
	r.Things = append(r.Things, thing)
	r.dirty = true
}

// Remove Unconditionally remove a thing from the room.
func (r *Room) Remove(thing *Thing) {
	for i, item := range r.Things {
		if item == thing {
			r.Things = append(r.Things[:i], r.Things[i+1:]...)
			r.dirty = true
		}
	}
}

func (r *Room) Save(db *sql.DB) error {
	// TODO: the player's actor is currently not in the databases, which causes
	// saves to fail!
	for _, actor := range r.Actors {
		err := actor.Save(db)
		if err != nil {
			return err
		}
	}
	return nil
}
