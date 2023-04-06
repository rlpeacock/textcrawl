package main

type Id string

type Obj struct {
	Id         Id
	Weight     Attrib
	Size       Attrib
	Title      string
	Desc       string
	Durability Attrib
}
