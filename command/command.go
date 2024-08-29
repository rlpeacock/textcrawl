package command

import (
	"io"
	"log"
	entity "rob.co/textcrawl/entity"
	"strings"
)

type Noun struct {
	Text string
	Ref  any
}

func NewNoun(text string, ref any) Noun {
	return Noun{text, ref}
}

type Command struct {
	Text         string
	Action       string
	Params       []string
	Preposition  string
	DirectObjs   []Noun
	IndirectObjs []Noun
}

type Action func(cmd *Command) (bool, error)

var dispatchTable = map[string][]Action{
	"goDirection": {goDirection},
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
		DirectObjs:   make([]Noun, 0),
		IndirectObjs: make([]Noun, 0),
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

func (c *Command) ResolveWords(room *entity.Room, actor *entity.Actor) {
	words := strings.Split(c.Text, " ")
	c.Action, c.Params = TranslateAction(words[0])
words:
	for _, w := range words[1:] {
		entity := room.Find(w)
		// if not in room, check actor's inventory
		if entity == nil {
			entity = actor.Find(w)
		}
		// if not a noun, maybe a preposition?
		if entity == nil && c.Preposition == "" && len(c.DirectObjs) > 0 {
			for _, p := range prepositions {
				if w == p {
					c.Preposition = p
					continue words
				}
			}
		}
		// TODO: we don't know what this thing is...add special 'unknown' object?
		// if entity == nil {
		// }
		if c.Preposition == "" {
			c.DirectObjs = append(c.DirectObjs, NewNoun(w, entity))
		} else {
			c.IndirectObjs = append(c.IndirectObjs, NewNoun(w, entity))
		}
	}
}

func Perform(cmd *Command, actor *entity.Actor, writer io.Writer) {
	cmd.ResolveWords(actor.Room(), actor)
	// basically means blank line
	if cmd.Action == "" {
		return
	}
	handlers := dispatchTable[cmd.Action]
	for _, h := range handlers {
		done, err := h(cmd)
		if done {
			break
		}
		if err != nil {
			log.Printf("%s", err)
			// TODO: break?
		}
	}

}
