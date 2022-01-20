package client

import (
	"database/sql"
	"time"
)

func StoreSensorData(db *sql.DB, sampleTime time.Time, power, current, voltage, shunt float64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO SensorSample
			VALUES (datetime(?), ?, ?, ?, ?)`,
		sampleTime, power, current, voltage, shunt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
