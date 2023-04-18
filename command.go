package main

import "fmt"

func doLook(req *Request) {
	req.Write(req.Actor.Room.Desc)
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
