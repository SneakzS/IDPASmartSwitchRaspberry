package client

import (
	"database/sql"
	"time"

	"github.com/philip-s/idpa/common"
)

func GetWorkloadDefinitions(tx *sql.Tx) ([]common.WorkloadDefinition, error) {
	res, err := tx.Query(
		`SELECT workloadDefinitionID, workloadW, durationM, 
		toleranceDurationM, isEnabled, description, expiryDate
		FROM WorkloadDefinition`,
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var (
		w           common.WorkloadDefinition
		definitions []common.WorkloadDefinition
	)

	for res.Next() {
		expiryDateStr := ""
		err = res.Scan(
			&w.WorkloadDefinitionID, &w.WorkloadW, &w.DurationM, &w.ToleranceDurationM,
			&w.IsEnabled, &w.Description, &expiryDateStr)
		if err != nil {
			return definitions, err
		}

		w.ExpiryDate, err = time.Parse("2006-01-02 15:04:05", expiryDateStr)
		if err != nil {
			return definitions, err
		}
		definitions = append(definitions, w)
	}

	// get all the time patterns
	for i := range definitions {
		d := &definitions[i]

		res, err = tx.Query(
			`SELECT monthFlags, dayFlags, hourFlags, 
			minuteFlags, weekdayFlags FROM TimePattern 
			WHERE workloadDefinitionID = ?`, d.WorkloadDefinitionID,
		)
		if err != nil {
			return definitions, err
		}

		for res.Next() {
			var rp common.RepeatPattern
			err = res.Scan(&rp.MonthFlags, &rp.DayFlags, &rp.HourFlags, &rp.MinuteFlags, &rp.WeekdayFlags)
			if err != nil {
				return definitions, err
			}
		}
	}

	return definitions, nil
}

func CreateWorkloadDefinition(tx *sql.Tx, d common.WorkloadDefinition) (int32, error) {
	res, err := tx.Exec(
		`INSERT INTO WorkloadDefinition VALUES
		(NULL, ?, ?, ?, ?, ?, datetime(?))`,
		d.WorkloadW, d.DurationM, d.ToleranceDurationM, d.IsEnabled, d.Description, d.ExpiryDate,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, p := range d.RepeatPattern {
		_, err = tx.Exec(`INSERT INTO TimePattern VALUES(NULL, ?, ?, ?, ?, ?, ?)`, id,
			p.MonthFlags, p.DayFlags, p.HourFlags, p.MinuteFlags, p.WeekdayFlags)
		if err != nil {
			return 0, err
		}
	}

	return int32(id), nil
}

func UpdateWorkloadDefinition(tx *sql.Tx, d common.WorkloadDefinition) error {
	_, err := tx.Exec(
		`UPDATE WorkloadDefinition SET 
		workloadW = ?, durationM = ?,
		toleranceDurationM = ?,
		isEnabled = ?, description = ?
		WHERE workloadDefinitionID = ?`,
		d.WorkloadW, d.DurationM,
		d.ToleranceDurationM,
		d.IsEnabled, d.Description,
		d.WorkloadDefinitionID,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM TimePattern WHERE workloadDefinitionID = ?", d.WorkloadDefinitionID)
	if err != nil {
		return err
	}

	for _, p := range d.RepeatPattern {
		_, err = tx.Exec(`INSERT INTO TimePattern VALUES (NULL, ?, ?, ?, ?, ?, ?)`,
			d.WorkloadDefinitionID, p.MonthFlags,
			p.DayFlags, p.HourFlags,
			p.MinuteFlags, p.WeekdayFlags,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteWorkloadDefinition(tx *sql.Tx, workloadDefinitionID int32) error {
	_, err := tx.Exec("DELETE FROM WorkloadDefinition WHERE workloadDefinitionID = ?", workloadDefinitionID)
	return err
}

type WorkloadSample struct {
	SampleTime    time.Time
	WorkloadID    int32
	OutputEnabled bool
	dbStored      bool
}

func CreateWorkload(tx *sql.Tx, d common.WorkloadDefinition, matchTime time.Time, offsetM int32) (int32, error) {
	res, err := tx.Exec(
		`INSERT INTO Workload VALUES (NULL, ?, datetime(?), ?, ?, ?)`,
		d.WorkloadDefinitionID, matchTime, d.WorkloadW, offsetM, d.DurationM,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

func CreateWorkloadSamples(tx *sql.Tx, workloadID int32, startTime time.Time, durationM int32) error {
	startTime = startTime.Truncate(time.Minute)
	samples := make([]WorkloadSample, durationM)
	for i := range samples {
		samples[i] = WorkloadSample{
			SampleTime:    startTime.Add(time.Duration(i) * time.Minute),
			OutputEnabled: true,
			dbStored:      true,
			WorkloadID:    workloadID,
		}
	}

	stmt, err := tx.Prepare("INSERT INTO WorkloadSample VALUES (datetime(?), ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sample := range samples {
		_, err = stmt.Exec(sample.SampleTime, sample.WorkloadID)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateWorkloadAndSamples(tx *sql.Tx, d common.WorkloadDefinition, matchTime time.Time, offsetM int32) error {
	workloadID, err := CreateWorkload(tx, d, matchTime, offsetM)
	if err != nil {
		return err
	}
	err = CreateWorkloadSamples(tx, workloadID, matchTime.Add(time.Duration(offsetM)*time.Minute), d.DurationM)
	return err
}

func GetWorkloadSamples(tx *sql.Tx, startTime time.Time, durationM int32) ([]WorkloadSample, error) {
	startTime = startTime.UTC().Truncate(time.Minute)

	res, err := tx.Query(
		`SELECT sampleTime, workloadID FROM WorkloadSample
		WHERE sampleTime >= datetime(?) AND sampleTime < datetime(?)`,
		startTime, startTime.Add(time.Duration(durationM)*time.Minute),
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	samples := make([]WorkloadSample, durationM)
	for i := range samples {
		samples[i].SampleTime = startTime.Add(time.Duration(i) * time.Minute)
	}

	for res.Next() {
		var (
			sampleTime time.Time
			workloadID int32
		)
		err = res.Scan(&sampleTime, &workloadID)
		if err != nil {
			return nil, err
		}

		samples[int(sampleTime.Sub(startTime).Minutes())] = WorkloadSample{
			SampleTime:    sampleTime,
			WorkloadID:    workloadID,
			OutputEnabled: true,
			dbStored:      true,
		}

	}

	return samples, nil
}

type Workload struct {
	WorkloadID           int32
	WorkloadDefinitionID int32
	MatchTime            time.Time
	WorkloadW            int32
	OffsetM              int32
	DurationM            int32
}

func GetWorkloads(tx *sql.Tx, startTime time.Time, durationM int32) ([]Workload, error) {
	var workloads []Workload
	startTime = startTime.Truncate(time.Minute)

	res, err := tx.Query(
		`SELECT workloadID, workloadDefinitionID, matchTime, workloadW, offsetM, durationM
		FROM Workload WHERE matchTime >= datetime(?) AND matchTime < datetime(?)`,
		startTime, startTime.Add(time.Duration(durationM)*time.Minute),
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		w := Workload{}
		err = res.Scan(&w.WorkloadID, &w.WorkloadDefinitionID, &w.MatchTime, &w.WorkloadW, &w.OffsetM, &w.DurationM)
		if err != nil {
			return nil, err
		}

		workloads = append(workloads, w)
	}

	return workloads, nil
}
