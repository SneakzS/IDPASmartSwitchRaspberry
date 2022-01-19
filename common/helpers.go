package common

import (
	"database/sql"
	"os"
	"strings"

	"github.com/drbig/simpleini"
)

func InvokeSQLScript(script, databaseFile string) error {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := strings.Split(string(script), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if len(stmt) > 0 {
			_, err = db.Exec(stmt)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func ParseINIFile(fileName string) (ini *simpleini.INI, err error) {
	configFile, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer configFile.Close()

	ini, err = simpleini.Parse(configFile)
	return
}
