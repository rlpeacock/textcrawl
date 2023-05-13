package main

import (
	"errors"
	"fmt"
	"strings"
)

type Command struct {
	Text        string
	Action      string
	Params      []string
	Preposition string
	Obj         []Entity
	DObj        []Entity
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

func NewCommand(actor *Actor, text string) (*Command, error) {
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")
	action, params := TranslateAction(words[0])
	cmd := &Command{
		Text:        text,
		Action:      action,
		Params:      params,
		Preposition: "",
		Obj:         make([]Entity, 0),
		DObj:        make([]Entity, 0),
	}
	for _, w := range words[1:] {
		entity := actor.Room.Find(w)
		if entity != nil {
			if cmd.Preposition == "" {
				cmd.Obj = append(cmd.Obj, entity)
			} else {
				cmd.DObj = append(cmd.DObj, entity)
			}
		} else if cmd.Preposition == "" {
			for _, p := range prepositions {
				if w == p {
					cmd.Preposition = p
					break
				}
			}
			if cmd.Preposition == "" {
				return nil, errors.New(fmt.Sprintf("bWhat is '%s'?", w))
			}
		} else {
			return nil, errors.New(fmt.Sprintf("bWhat is '%s'?", w))
		}
	}
	return cmd, nil
}

// look for abbreviations and other mappings
func TranslateAction(text string) (string, []string) {
	t := translations[text]
	if t == nil {
		t = []string{text}
	}
	return t[0], t[1:]
}
