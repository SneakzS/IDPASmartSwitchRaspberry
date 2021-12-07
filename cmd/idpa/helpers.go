package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/drbig/simpleini"
)

func loadConfig() (*simpleini.INI, error) {
	if *config == "" {
		return nil, fmt.Errorf("-config is required")
	}

	cfp, err := os.Open(*config)
	if err != nil {
		return nil, err
	}
	defer cfp.Close()

	return simpleini.Parse(cfp)
}

func saveConfig(cfg *simpleini.INI) error {
	cfp, err := os.Create(*config)
	if err != nil {
		return err
	}
	defer cfp.Close()

	return cfg.Write(cfp, true)
}

func getServiceMode(cfg *simpleini.INI) (string, error) {
	mode, err := cfg.GetString("Service", "mode")
	if err != nil {
		return "", err
	}

	// Override the service mode based on the flag
	if *serviceMode != "" {
		mode = *serviceMode
	}

	return mode, nil
}

func invokeScriptFile(scriptFile, databaseFile string) error {
	scriptBody, err := os.ReadFile(scriptFile)
	if err != nil {
		return err
	}
	return invokeScript(string(scriptBody), databaseFile)
}

func invokeScript(script, databaseFile string) error {
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
