package main

type Room struct {
	Id        Id
	Title     string
	Desc      string
	Occupants []*Actor
	Contents  []*Obj
}
