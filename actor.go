package main

type Attrib struct {
	Real int
	Cur  int
}

type Stats struct {
	Str    Attrib
	Dex    Attrib
	Int    Attrib
	Will   Attrib
	Health Attrib
	Mind   Attrib
}

type Inventory struct {
}

type Actor struct {
	Id    Id
	Obj   *Obj
	Stats *Stats
	Room  *Room
	Inv   *Inventory
}

func NewActor(id string) *Actor {
	return &Actor{
		Id: Id(id),
	}
}
