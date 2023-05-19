function look(req)
   req:Write(req.Actor.Room.Desc .. "\n")
   if req.Actor.Room.Contents and #req.Actor.Room.Contents > 0 then
	  req:Write("You see here:\n")
	  for i = 1, #req.Actor.Room.Contents do
		 req:Write("  " .. req.Actor.Room.Contents[i].Title .. "\n")
	  end
   end
   if req.Actor.Room.Occupants and #req.Actor.Room.Occupants > 0 then
	  req:Write("There are in this room:\n")
	  for i = 1, #req.Actor.Room.Occupants do
		 req:Write("  " .. req.Actor.Room.Occupants[i].Body.Title .. "\n")
	  end
   end

   if req.Actor.Room.Exits and #req.Actor.Room.Exits > 0 then
	  req:Write("Exits: ")
	  for i = 1, #req.Actor.Room.Exits do
		 req:Write(req.Actor.Room.Exits[i].Direction .. " ")
	  end
   end
end

function goDirection(req, dir)
   exit = req.Actor.Room:GetExit(dir)
   if exit then
	  newRoom = req.Actor.Zone:GetRoom(exit.Destination)
	  if newRoom:Insert(req.Actor) then
		 req:Write("You go " .. dir .. "\n")
		 look(req)
	  else
		 req:Write("You tried but it didn't work")
	  end
   else
	  req:Write("You can't go that way")
   end
end

function take(req)
   if #req.Cmd.DirectObjs == 0 then
	  req:Write("Take what?")
   else
	  for i = 1, #req.Cmd.DirectObjs do
		 obj = req.Cmd.DirectObjs[i]
		 if req.Actor:Insert(obj) then
			req:Write("You get the " .. obj.Title .. "\n")
		 else
			req:Write("You failed to take " .. obj.Title .. "\n")
		 end
	  end
   end
end

function drop(req)
   if #req.Cmd.DirectObjs == 0 then
	  req:Write("Drop what?")
   else
	  for i = 1, #req.Cmd.DirectObjs do
		 obj = req.Cmd.DirectObjs[i]
		 if req.Actor.Room:Insert(obj) then
			req:Write("You drop the " .. obj.Title .. "\n")
		 else
			req:Write("You failed to drop the " .. obj.Title .. "\n")
		 end
	  end
   end

end

function inventory(req)
   req:Write("You have:\n")
   for i = 1, #req.Actor.Body.Contents do
	  req:Write(req.Actor.Body.Contents[i].Title .. "\n")
   end
end


function quit(req)
   req:Write("Goodbye...\n")
   req.Writer:Close()
end
