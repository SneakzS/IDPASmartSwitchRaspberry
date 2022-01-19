package provider

import (
	"database/sql"
	"time"

	"github.com/philip-s/idpa"
)

type WireWorkload struct {
	WorkloadW  int32
	WireID     int32
	SampleTime time.Time
	dbStored   bool
}

func GetWireWorkload(tx *sql.Tx, wireID int32, startTime time.Time, durationM int32) ([]WireWorkload, error) {

	startTime = startTime.UTC().Truncate(time.Minute)
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

func GetOptimalWorkloadOffset(tx *sql.Tx, wires []Wire, durationM, toleranceDurationM, workloadW int32, startTime time.Time) (int32, error) {
	var err error

	type wireSample struct {
		w       Wire
		samples []WireWorkload
	}
	startTime = startTime.Truncate(time.Minute).UTC()
	wireSamples := make([]wireSample, len(wires))

	for i, wire := range wires {
		wireSamples[i].w = wire
		wireSamples[i].samples, err = GetWireWorkload(tx, wire.WireID, startTime, toleranceDurationM)
		if err != nil {
			return 0, err
		}
	}

	var offsetM int32
search:
	for i := int(offsetM); i < int(offsetM+durationM) && i < len(wireSamples); i++ {
		for i := offsetM; i < offsetM+durationM; i++ {
			for _, ws := range wireSamples {
				if ws.samples[i].WorkloadW+workloadW > ws.w.CapacityW {
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
	return 0, idpa.ErrWorkloadNotPossible

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
