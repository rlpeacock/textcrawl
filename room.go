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
	Id    Id
	Title string
	Desc  string
	Exits []*Exit
}

func (r *Room) GetTitle() string {
	return r.Title
}

func (r *Room) GetExit(d Direction) *Exit {
	for _, e := range r.Exits {
		if e.Direction == d {
			return e
		}
	}
	return nil
}

func (r *Room) ID() Id {
	return r.Id
}

func (r *Room) ParentID() Id {
	return ""
}

func (r *Room) Take(obj GObject) bool {
	return true
}
