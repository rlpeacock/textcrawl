package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)	

var ls *lua.LState = lua.NewState()

func doLook(req *Request) {
	look := `
function look(f) return f .. " looked in lua!" end
`
	if err := ls.DoString(look); err != nil {
		req.Write("We failed")
	}
	err := ls.CallByParam(lua.P{
		Fn: ls.GetGlobal("look"),
		NRet: 1,
		Protect: true,
	}, luar.New(ls, req.Actor.Player.Username))
	if err != nil {
		req.Write("We failed the look")
	}	
	ret := ls.Get(-1)
	req.Write(ret.String() + "\n")
}

func doUnknown(req *Request) {
	req.Write(fmt.Sprintf("Unknown action '%s'", req.Text))
}

func ProcessRequest(req *Request) {
	switch req.Text {
	case "look": doLook(req)
	default: doUnknown(req)
	}
}


