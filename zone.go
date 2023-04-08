package main

// a zone is a collection of rooms that are managed together
// it acts as a boundary for interactivity - that is, everything
// in a zone will be handled by a single process and can interact
// without distributed transaction semantics. Cross zone interaction
// should be tightly restricted and needs to be hardened against
// IPC failure.

type Zone struct {
	Rooms map[Id]*Room	
}

func NewZone(id Id) *Zone {
	rooms := make(map[Id]*Room)
	samples := SampleRooms()
	for _, s := range(samples) {
		rooms[s.Id] = s
	}
	return &Zone{
		Rooms: rooms,	
	}
}