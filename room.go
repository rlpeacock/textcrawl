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
	if a.Room == nil || a.Room.Remove(a) {
		r.Occupants = append(r.Occupants, a)
		a.Room = r
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

func (r *Room) Take(o *Obj) bool {
	if o.Room == nil || o.Room.Give(o) {
		r.Contents = append(r.Contents, o)
		o.Room = r
		return true
	}
	return false
}

func (r *Room) Give(o *Obj) bool {
	for i, c := range r.Contents {
		if c == o {
			r.Contents = append(r.Contents[:i], r.Contents[i+1:]...)
			return true
		}
	}
	return false
}
