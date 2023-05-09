package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Id string

type ObjectFlags int

type Container interface {
	Take(o *Obj) bool
	Give(o *Obj) bool
	ID() Id
}

const (
	MOB = 1
)

type Attrib struct {
	Real int
	Cur  int
}

type Inventory []*Obj

type Obj struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
	Loc        Id
	Owner      Container
	Contents   Inventory
	Flags      ObjectFlags
	dirty      bool
}

func (o *Obj) NewOwner(c Container) {
	o.Owner = c
	o.Loc = c.ID()
	o.dirty = true
}

func (o *Obj) Take(obj *Obj) bool {
	if obj.Owner == nil || obj.Owner.Give(obj) {
		o.Contents = append(o.Contents, obj)
		obj.NewOwner(o)
		return true
	}
	return false
}

func (o *Obj) Give(obj *Obj) bool {
	for i, io := range o.Contents {
		if io == obj {
			o.Contents = append(o.Contents[:i], o.Contents[i+1:]...)
			return true
		}
	}
	return false
}

func (o *Obj) ID() Id {
	return o.Id
}

func DeserializeAttrib(s string) (Attrib, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Attrib{}, errors.New("Malformed attribute: missing delimiter")
	}
	real, err := strconv.Atoi(parts[0])
	if err != nil {
		return Attrib{}, errors.New("Malformed attribute: real is not a number")
	}
	cur, err := strconv.Atoi(parts[1])
	if err != nil {
		return Attrib{}, errors.New("Malformed attribute: cur is not a number")
	}
	return Attrib{real, cur}, nil
}

func DeserializeAttribList(attribStr string, attribs ...*Attrib) error {
	rawAttribs := strings.Split(attribStr, ",")
	if len(rawAttribs) < len(attribs) {
		return errors.New(fmt.Sprintf("Attribute does not contain enough values. Expected %d, got %d", len(attribs), len(rawAttribs)))
	}
	if len(rawAttribs) > len(attribs) {
		return errors.New(fmt.Sprintf("Attribute contains too many values. Expected %d, got %d", len(attribs), len(rawAttribs)))
	}
	for i, a := range rawAttribs {
		attrib, err := DeserializeAttrib(a)
		if err != nil {
			return err
		}
		attribs[i].Cur = attrib.Cur
		attribs[i].Real = attrib.Real
	}
	return nil
}

func LoadObjects(db *sql.DB) map[Id]*Obj {
	rows, err := db.Query(`
SELECT id, attributes, title, description, location, flags
FROM object
ORDER BY location`)
	if err != nil {
		panic(fmt.Sprintf("Oh shit, the database is screwed up! Error: %s", err))
	}
	defer rows.Close()
	objs := make(map[Id]*Obj, 0)
	for rows.Next() {
		obj := Obj{}
		var (
			attribs string
		)
		err = rows.Scan(&obj.Id, &attribs, &obj.Title, &obj.Desc, &obj.Loc, &obj.Flags)
		if err != nil {
			panic(fmt.Sprintf("Error while iterating rows: %s", err))
		}
		DeserializeAttribList(attribs, &obj.Weight, &obj.Size, &obj.Durability)
		objs[obj.Id] = &obj
	}
	err = rows.Err()
	if err != nil {
		panic(fmt.Sprintf("Error while iterating rows: %s", err))
	}
	return objs
}
