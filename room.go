package main

import (
	"gopkg.in/yaml.v3"
)

type Direction string

type MoveType string

type Exit struct {
	Direction Direction
	Destination Id
	MoveType MoveType
	MoveDifficulty Attrib
	MoveSpeed Attrib
}

type Room struct {
	Id        Id
	Title     string
	Desc      string
	Occupants []*Actor
	Contents  []*Obj
	Exits     []*Exit
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
`

