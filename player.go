package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Player struct {
	LoginState    LoginState
	LoginAttempts int
	Username      string
}

func NewPlayer() *Player {
	return &Player{
		LoginState: LoginStateStart,
	}
}

type PlayerMgr struct {
	db *sql.DB
}

func NewPlayerMgr() *PlayerMgr {
	f := filepath.Join("world", "player.dat")
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		panic(fmt.Sprintf("Could not open database %s", f))
	}
	return &PlayerMgr{
		db: db,
	}
}

// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

func (pm *PlayerMgr) LookupPlayer(username string, pwd string) (*Player, int, error) {
	rows, err := pm.db.Query(`
SELECT password, actor_id, active FROM player WHERE username = ?`, username)
	if err != nil {
		return nil, 0, err
	}
	if rows.Next() {
		var (
			storedPwd string
			active    bool
			actorId   int
		)
		rows.Scan(&storedPwd, &actorId, &active)
		err = bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(pwd))
		if err != nil {
			return nil, 0, errors.New("Invalid username or password")
		}
		if active {
			return nil, 0, errors.New("User is already logged in")
		}
		return &Player{Username: username, LoginState: LoginStateStart}, actorId, nil
	}
	return nil, 0, errors.New("Invalid username or password")
}
