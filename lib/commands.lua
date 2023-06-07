
function look(req)
   room = req.Actor:Room()
   req:Write(room.Desc .. "\n")
   if room.Things and #room.Things > 0 then
	  req:Write("You see here:\n")
	  for i = 1, #room.Things do
		 req:Write("  " .. room.Things[i].Title .. "\n")
	  end
   end
   if room.Actors and #room.Actors > 0 then
	  req:Write("There are in this room:\n")
	  for i = 1, #room.Actors do
		 req:Write("  " .. room.Actors[i].Body.Title .. "\n")
	  end
   end

   if room.Exits and #room.Exits > 0 then
	  req:Write("Exits: ")
	  for i = 1, #room.Exits do
		 req:Write(room.Exits[i].Direction .. " ")
	  end
   end
end

function goDirection(req, dir)
   exit = req.Actor:Room():GetExit(dir)
   if exit then
	  newRoom = req.Actor.Zone:GetRoom(exit.Destination)
	  
	  if req.Actor.Zone:MoveActor(req.Actor, newRoom) then
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
		 if req.Actor:Take(obj) then
			req:Write("You get the " .. obj.Title .. "\n")
		 else
			req:Write("You failed to take " .. obj.Title .. "\n")
		 end
	  end
   end
end

function drop( req )
	if #req.Cmd.DirectObjs == 0 then
		req:Write("Drop what?")
	else
	  for i = 1, #req.Cmd.DirectObjs do
		 obj = req.Cmd.DirectObjs[i]
		 if req.Actor:Drop(obj) then
			req:Write("You dropped the " .. obj.Title .. "\n")
		 else
			req:Write("You failed to drop the " .. obj.Title .. "\n")
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
		 if req.Actor:Drop(obj) then
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
