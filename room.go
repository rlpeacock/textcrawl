package main

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
	Id     Id
	Title  string
	Desc   string
	Exits  []*Exit
	Actors []*Actor
	Things []*Thing
}

func (r *Room) GetExit(d Direction) *Exit {
	for _, e := range r.Exits {
		if e.Direction == d {
			return e
		}
	}
	return nil
}

func (r *Room) AcceptThing(t *Thing) bool {
	return true
}

func (r *Room) AcceptActor(a *Actor) bool {
	return true
}

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

func (r *Room) InsertActor(actor *Actor) {
	actor.SetRoom(r)
	r.Actors = append(r.Actors, actor)
}

func (r *Room) RemoveActor(actor *Actor) {
	for i, item := range r.Actors {
		if item == actor {
			r.Actors = append(r.Actors[:i], r.Actors[i+1:]...)
		}
	}
}

func (r *Room) Insert(thing *Thing) {
	thing.ParentId = r.Id
	r.Things = append(r.Things, thing)
}

func (r *Room) Remove(thing *Thing) {
	for i, item := range r.Things {
		if item == thing {
			r.Things = append(r.Things[:i], r.Things[i+1:]...)
		}
	}
}
