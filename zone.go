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

func NewZone(id Id) (*Zone, error) {
	rooms := make(map[Id]*Room)
	rooms, err := LoadSampleRooms()
	if err != nil {
		return nil, err
	}
	return &Zone{
		Rooms: rooms,
	}, nil
}

type ZoneManager struct {
	zones map[Id]*Zone
}

func GetZoneMgr() *ZoneManager {
	return &ZoneManager{
		zones: make(map[Id]*Zone),
	}
}

func (zm *ZoneManager) GetZone(id Id) (*Zone, error) {
	z := zm.zones[id]
	if z == nil {
		z, err := NewZone(id)
		if err != nil {
			return nil, err
		}
		zm.zones[id] = z
	}
	return z, nil
}
