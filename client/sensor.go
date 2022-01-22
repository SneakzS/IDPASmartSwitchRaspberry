package client

import (
	"database/sql"
	"time"

	"github.com/philip-s/idpa/common"
)

func StoreSensorData(tx *sql.Tx, sampleTime time.Time, power, current, voltage float64) error {

	_, err := tx.Exec(`INSERT INTO SensorSample
			VALUES (datetime(?), ?, ?, ?)`,
		sampleTime, power, current, voltage,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetSensorData(tx *sql.Tx) (samples []common.SensorSample, err error) {
	rows, err := tx.Query("SELECT sampleTime, power, current, voltage FROM SensorSample")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		s := common.SensorSample{}
		sampleTimeStr := ""
		err = rows.Scan(&sampleTimeStr, &s.Power, &s.Current, &s.Voltage)
		if err != nil {
			return
		}

		s.SampleTime, _ = time.Parse("2006-01-02 15:04:05", sampleTimeStr)
		samples = append(samples, s)
	}

	return
}
