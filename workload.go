package idpa

import (
	"database/sql"
	"time"
)

type WireWorkload struct {
	WorkloadW  int32
	WireID     int32
	SampleTime time.Time
	dbStored   bool
}

func GetWireWorkload(tx *sql.Tx, wireID int32, startTime time.Time, durationM int32) ([]WireWorkload, error) {

	startTime = startTime.UTC().Round(time.Minute)
	samples := make([]WireWorkload, durationM)

	for i := range samples {
		samples[i].SampleTime = startTime.Add(time.Duration(i) * time.Minute)
	}

	res, err := tx.Query(
		`SELECT sampleTime, workloadW 
		FROM WireWorkload WHERE wireID = ? AND sampleTime >= datetime(?) AND sampleTime < datetime(?)`,
		wireID, startTime, startTime.Add(time.Duration(durationM)*time.Minute),
	)

	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		var (
			sampleTime time.Time
			workloadW  int32
		)

		var err = res.Scan(&sampleTime, &workloadW)
		if err != nil {
			return nil, err
		}

		samples[int(sampleTime.Sub(startTime).Minutes())] = WireWorkload{
			WorkloadW:  workloadW,
			WireID:     wireID,
			SampleTime: sampleTime,
			dbStored:   true,
		}
	}

	return samples, nil
}

func GetOptimalWorkloadOffset(tx *sql.Tx, wires []Wire, d WorkloadDefinition, startTime time.Time) (int32, error) {
	var err error

	type wireSample struct {
		w       Wire
		samples []WireWorkload
	}
	startTime = startTime.Round(time.Minute).UTC()
	wireSamples := make([]wireSample, len(wires))

	for i, wire := range wires {
		wireSamples[i].w = wire
		wireSamples[i].samples, err = GetWireWorkload(tx, wire.WireID, startTime, d.ToleranceDurationM)
		if err != nil {
			return 0, err
		}
	}

	var offsetM int32
search:
	for i := int(offsetM); i < int(offsetM+d.DurationM) && i < len(wireSamples); i++ {
		for i := offsetM; i < offsetM+d.DurationM; i++ {
			for _, ws := range wireSamples {
				if ws.samples[i].WorkloadW+d.WorkloadW > ws.w.CapacityW {
					// we detected a wire overload
					offsetM = i + 1
					continue search
				}
			}
		}

		// all whires are fine we can use t as our start time
		goto nooverload
	}
	// the wire will be overloaded at all time, the workload is not possible
	return 0, ErrWorkloadNotPossible

nooverload:
	return offsetM, nil
}

func AddWireWorkload(tx *sql.Tx, wireID int32, startTime time.Time, durationM, workloadW int32) error {

	samples, err := GetWireWorkload(tx, wireID, startTime, durationM)
	if err != nil {
		return err
	}

	for i := range samples {
		samples[i].WorkloadW += workloadW
	}

	for _, sample := range samples {
		if sample.dbStored {
			_, err = tx.Exec(`UPDATE WireWorkload SET workloadW = ? WHERE wireID = ? AND sampleTime = datetime(?)`,
				sample.WorkloadW, wireID, sample.SampleTime,
			)
		} else {
			_, err = tx.Exec(`INSERT INTO WireWorkload VALUES (?, datetime(?), ?)`,
				wireID, sample.SampleTime, sample.WorkloadW)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func GetWorkloadDefinitions(tx *sql.Tx) ([]WorkloadDefinition, error) {
	res, err := tx.Query(
		`SELECT workloadDefinitionID, workloadW, durationM, toleranceDurationM, monthFlags, 
		dayFlags, hourFlags, minuteFlags, weekdayFlags, isEnabled
		FROM WorkloadDefinition`,
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var (
		w           WorkloadDefinition
		definitions []WorkloadDefinition
	)

	for res.Next() {
		err = res.Scan(&w.WorkloadDefinitionID, &w.WorkloadW, &w.DurationM, &w.ToleranceDurationM,
			&w.RepeatPattern.MonthFlags, &w.RepeatPattern.DayFlags, &w.RepeatPattern.HourFlags, &w.RepeatPattern.MinuteFlags, &w.RepeatPattern.WeekdayFlags,
			&w.IsEnabled)
		if err != nil {
			return definitions, err
		}

		definitions = append(definitions, w)
	}

	return definitions, nil
}

type WorkloadSample struct {
	SampleTime       time.Time
	WorkloadID       int32
	MeasuredWorkload int32
	OutputEnabled    bool
	dbStored         bool
}

func CreateWorkload(tx *sql.Tx, d WorkloadDefinition, matchTime time.Time, offsetM int32) (int32, error) {
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
	startTime = startTime.Round(time.Minute)
	samples := make([]WorkloadSample, durationM)
	for i := range samples {
		samples[i] = WorkloadSample{
			SampleTime:    startTime.Add(time.Duration(i) * time.Minute),
			OutputEnabled: true,
			dbStored:      true,
			WorkloadID:    workloadID,
		}
	}

	stmt, err := tx.Prepare("INSERT INTO WorkloadSample VALUES (datetime(?), ?, 0)")
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

func CreateWorkloadAndSamples(tx *sql.Tx, d WorkloadDefinition, matchTime time.Time, offsetM int32) error {
	workloadID, err := CreateWorkload(tx, d, matchTime, offsetM)
	if err != nil {
		return err
	}
	err = CreateWorkloadSamples(tx, workloadID, matchTime.Add(time.Duration(offsetM)*time.Minute), d.DurationM)
	return err
}

func GetWorkloadSamples(tx *sql.Tx, startTime time.Time, durationM int32) ([]WorkloadSample, error) {
	startTime = startTime.Round(time.Minute)

	res, err := tx.Query(
		`SELECT sampleTime, workloadID, measuredWorkloadW FROM WorkloadSample
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
			sampleTime        time.Time
			workloadID        int32
			measuredWorkloadW int32
		)
		err = res.Scan(&sampleTime, &workloadID, &measuredWorkloadW)
		if err != nil {
			return nil, err
		}

		samples[int(sampleTime.Sub(startTime).Minutes())] = WorkloadSample{
			SampleTime:       sampleTime,
			WorkloadID:       workloadID,
			MeasuredWorkload: measuredWorkloadW,
			OutputEnabled:    true,
			dbStored:         true,
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
	startTime = startTime.Round(time.Minute)

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
