package entity

import (
	"path/filepath"
	"fmt"
	"os"
	"testing"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func openDB(zone string) (*sql.DB, error) {
	worldDir, ok := os.LookupEnv("TEXTCRAWL_WORLD")
	if !ok {
		worldDir = "./world"
	}
	f := filepath.Join(worldDir, fmt.Sprintf("%s.dat", zone))
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		return nil, fmt.Errorf("Could not open database %s", f)
	}
	return db,nil
}

func TestSaveLoadActor(t *testing.T) {
	db, err := openDB("1")
	if err != nil {
		t.Errorf("Unable to open DB: %v", err)
	}
	things := map[Id]*Thing{
		"T2": {},
	}
	actors := LoadActors(db, things)
	actor := actors["A1"]
	if actor == nil {
		t.Error(`Actor "A1" not found`)
	}

	actor.dirty = true
	actor.Save(db)
	actors = LoadActors(db, things)
	actor = actors["A1"]
	if actor == nil {
		t.Error(`Actor could not be loaded after save`)
	}
	
}
