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

func TestLoadingActor(t *testing.T) {
	db, err := openDB("1")
	if err != nil {
		t.Errorf("Unable to open DB: %v", err)
	}
	things := map[Id]*Thing{
		"T2": &Thing{},
	}
	actors := LoadActors(db, things)
	if actors["A1"] == nil {
		t.Error(`Actor "foo" not found`)
	}
}
