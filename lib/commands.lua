function look(req)
  req:Write(req.Actor.Room.Desc)
end

function goDirection(req, dir)
   exit = req.Actor.Room:GetExit(dir)
   if exit then
	  newRoom = req.Actor.Zone:GetRoom(exit.Destination)
	  newRoom:Receive(req.Actor)
	  req:Write("You go " .. dir .. "\n")
	  look(req)
   else
	  req:Write("You can't go that way")
   end
end

function quit(req)
   req:Write("Goodbye...\n")
   req.Writer:Close()
end
