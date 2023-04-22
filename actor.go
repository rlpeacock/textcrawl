package main

type LoginFlowState int

const (
	NOT_STARTED = iota
	WAITING_FOR_USERNAME
	WAITING_FOR_PASSWORD
	BAD_LOGIN_INFO
	MAX_FAILURES
	LOGIN_COMPLETE
)

type Player struct {
	LoginState    LoginFlowState
	LoginAttempts int
	Username      string
}

func NewPlayer() *Player {
	return &Player{
		LoginState: NOT_STARTED,
	}
}

type Attrib struct {
	Real int
	Cur  int
}

type Stats struct {
	Str    Attrib
	Dex    Attrib
	Int    Attrib
	Will   Attrib
	Health Attrib
	Mind   Attrib
}

type Inventory struct {
}

type Actor struct {
	Id     Id
	Obj    *Obj
	Stats  *Stats
	Room   *Room
	Zone   *Zone
	Inv    *Inventory
	Player *Player
}

func NewActor(id string, player *Player) *Actor {
	return &Actor{
		Id:     Id(id),
		Player: player,
	}
}
