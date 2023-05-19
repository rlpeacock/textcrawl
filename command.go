package main

import (
	"errors"
	"fmt"
	"strings"
)

type Command struct {
	Text         string
	Action       string
	Params       []string
	Preposition  string
	DirectObjs   []*Locus
	IndirectObjs []*Locus
}

var translations = map[string][]string{
	"s":         {"goDirection", "south"},
	"south":     {"goDirection", "south"},
	"n":         {"goDirection", "north"},
	"north":     {"goDirection", "north"},
	"e":         {"goDirection", "east"},
	"east":      {"goDirection", "east"},
	"w":         {"goDirection", "west"},
	"west":      {"goDirection", "west"},
	"nw":        {"goDirection", "northwest"},
	"northwest": {"goDirection", "northwest"},
	"ne":        {"goDirection", "northeast"},
	"northeast": {"goDirection", "northeast"},
	"sw":        {"goDirection", "southwest"},
	"southwest": {"goDirection", "southwest"},
	"se":        {"goDirection", "southeast"},
	"southeast": {"goDirection", "southeast"},
	"l":         {"look"},
	"i":         {"inventory"},
	"inv":       {"inventory"},
}

var prepositions = []string{
	"above",
	"across",
	"after",
	"against",
	"around",
	"before",
	"behind",
	"below",
	"down",
	"from",
	"in",
	"inside",
	"into",
	"off",
	"on",
	"over",
	"through",
	"to",
	"under",
}

func NewCommand(text string) *Command {
	return &Command{
		Text:         strings.TrimSpace(text),
		Preposition:  "",
		DirectObjs:   make([]*Locus, 0),
		IndirectObjs: make([]*Locus, 0),
	}
}

// look for abbreviations and other mappings
func TranslateAction(text string) (string, []string) {
	t := translations[text]
	if t == nil {
		t = []string{text}
	}
	return t[0], t[1:]
}

func (c *Command) ResolveWords(room *Locus, actor *Locus) error {
	words := strings.Split(c.Text, " ")
	c.Action, c.Params = TranslateAction(words[0])
	for _, w := range words[1:] {
		entity := room.Find(w)
		// if not in room, check actor's inventory
		if entity == nil {
			entity = actor.Find(w)
		}
		if entity != nil {
			if c.Preposition == "" {
				c.DirectObjs = append(c.DirectObjs, entity)
			} else {
				c.IndirectObjs = append(c.IndirectObjs, entity)
			}
		} else if c.Preposition == "" {
			for _, p := range prepositions {
				if w == p {
					c.Preposition = p
					break
				}
			}
			if c.Preposition == "" {
				return errors.New(fmt.Sprintf("What is '%s'?", w))
			}
		} else {
			return errors.New(fmt.Sprintf("What is '%s'?", w))
		}
	}
	return nil
}
