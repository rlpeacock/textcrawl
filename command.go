package main

import "strings"

type Command struct {
	Text   string
	Action string
	Params []string
	Obj    []*Obj
	DObj   []*Obj
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
}

func NewCommand(actor *Actor, text string) *Command {
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")
	action, params := TranslateAction(words[0])
	// TODO: match remaining words against object in room
	return &Command{
		Text:   text,
		Action: action,
		Params: params,
		Obj:    make([]*Obj, 0),
		DObj:   make([]*Obj, 0),
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
