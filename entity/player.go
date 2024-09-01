package entity

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
	ActorId	  Id
}

func NewPlayer() Player {
	return Player{
		LoginState: LoginStateStart,
	}
}

type PlayerMgr interface {
	LookupPlayer(username string, pwd string) (Id, error)
}

type DBPlayerMgr struct {
	db *sql.DB
}

func NewPlayerMgr() PlayerMgr {
	f := filepath.Join("world", "player.dat")
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		panic(fmt.Sprintf("Could not open database %s", f))
	}
	return DBPlayerMgr{
		db: db,
	}
}

// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

func (pm DBPlayerMgr) LookupPlayer(username string, pwd string) (Id, error) {
	rows, err := pm.db.Query(`
SELECT password, actor_id, active FROM player WHERE username = ?`, username)
	if err != nil {
		return "", err
	}
	if rows.Next() {
		var (
			storedPwd string
			active    bool
			actorId   string
		)
		rows.Scan(&storedPwd, &actorId, &active)
		// TODO: don't check passwords...we can't write them to DB yet!
		if pwd != "" {
			err = bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(pwd))
			if err != nil {
				return "", errors.New("Invalid username or password")
			}
		}
		if active {
			return "", errors.New("User is already logged in")
		}
		return Id(actorId), nil
	}
	return "", errors.New("Invalid username or password")
}
