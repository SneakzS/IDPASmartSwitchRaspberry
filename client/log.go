package client

import (
	"database/sql"
	"time"

	"github.com/philip-s/idpa/common"
)

const (
	LogInfo = iota
	LogWarn
	LogError
)

func WritePersistendLog(db *sql.DB, severity int32, source, message string) error {
	_, err := db.Exec(
		`INSERT INTO Log VALUES (NULL, datetime(?), ?, ?, ?)`,
		time.Now().UTC(), severity, source, message)
	return err
}

func GetLogEntries(tx *sql.Tx) ([]common.LogEntry, error) {
	rows, err := tx.Query("SELECT logID, datetime(logTime), severity, source, message FROM Log")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []common.LogEntry{}
	for rows.Next() {
		entry := common.LogEntry{}
		logTimeStr := ""
		err = rows.Scan(&entry.LogID, &logTimeStr, &entry.Severity, &entry.Source, &entry.Message)
		if err != nil {
			return entries, err
		}

		entry.LogTime, err = time.Parse(sqliteDatetimeFormat, logTimeStr)
		if err != nil {
			return entries, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
