package main

import "database/sql"

type Id string

type ObjectFlags int

const (
	MOB = 1
)

type Obj struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
	Room       Id
	Flags      ObjectFlags
}

func LoadObjects(db *sql.DB, zone Id) []*Obj {
	rows, err := db.Query(`
SELECT id, weight, size, title, description, durability, room, flags
FROM object
WHERE zone = %s
ORDER BY room`, zone)
	if err != nil {
		panic("Oh shit, the database is screwed up!")
	}
	objs := make([]*Obj, 0)
	for rows.Next() {
		obj := Obj{}
		rows.Scan(&obj.Id, &obj.Weight, &obj.Size, &obj.Title, &obj.Desc, &obj.Durability, &obj.Room, &obj.Flags)
		objs = append(objs, &obj)
	}
	return objs
}
