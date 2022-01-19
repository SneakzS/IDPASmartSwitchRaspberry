package client

import (
	"database/sql"
	"time"
)

func StoreSensorData(db *sql.DB, inp *Input, sampleTime time.Time) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO SensorSample
			VALUES (datetime(?), ?, ?, ?, ?)`,
		sampleTime, inp.Power, inp.Current, inp.Voltage, inp.Shunt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
