package client

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetSensorData(t *testing.T) {
	db, err := sql.Open("sqlite3", `C:\Users\Phipu\client.sqlite3`)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	samples, err := GetSensorData(tx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(samples)
}
