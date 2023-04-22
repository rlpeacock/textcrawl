package main

import (
	"fmt"
	"log"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)	

var ls *lua.LState = lua.NewState()

func doLook(req *Request) {

	look := `
function look(f, a)
  f:Write(f.Actor.Room.Desc)
  return "\nLooked in lua!"
end
`
	if err := ls.DoString(look); err != nil {
		req.Write("We failed\n")
		log.Printf("Script execution failed: %s", err)
		return
	}
	err := ls.CallByParam(lua.P{
		Fn: ls.GetGlobal("look"),
		NRet: 1,
		Protect: true,
	}, luar.New(ls, req))
	if err != nil {
		log.Printf("Script did not succed: %s", err)
		req.Write("We failed the look\n")
		return
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


