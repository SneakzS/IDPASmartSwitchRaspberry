package idpa

import (
	"database/sql"
	"time"
)

type dbWorkload struct {
	startTime time.Time
	endTime   time.Time
	workload  int32
}

type WireWorkloadSample struct {
	Time      time.Time
	WorkloadW int32
}

func GetWireWorkload(db *sql.DB, wireID int32, startTime, endTime time.Time) ([]WireWorkloadSample, error) {

	/*qry := fmt.Sprintf(`SELECT workloadW, startTime, endTime
	FROM WireWorkload WHERE wireID = %d AND startTime >= strftime('%%s', '%s') AND endTime <= strftime('%%s', '%s')`,
		wireID,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
	)*/
	res, err := db.Query(
		`SELECT workloadW, startTime, endTime
		FROM WireWorkload WHERE wireID = ? AND startTime >= strftime('%s', ?) AND endTime <= strftime('%s', ?)`,
		wireID,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
	)
	//res, err := db.Query(qry)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var workloads []dbWorkload

	for res.Next() {
		var wl dbWorkload

		var err = res.Scan(&wl.workload, &wl.startTime, &wl.endTime)
		if err != nil {
			return nil, err
		}

		workloads = append(workloads, wl)
	}

	samples := make([]WireWorkloadSample, 0, int(endTime.Sub(startTime).Minutes())+1)

	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(1 * time.Minute) {
		s := WireWorkloadSample{Time: t}

		for _, wl := range workloads {
			if wl.startTime.Equal(t) || wl.endTime.Equal(t) || (wl.startTime.Before(t) && wl.endTime.After(t)) {
				s.WorkloadW += wl.workload
			}
		}

		samples = append(samples, s)
	}

	return samples, nil
}

type Workload struct {
	StartTime time.Time
	DurationM int32
}

func getOptimalWorkload(db *sql.DB, wires []dbWire, durationM, workloadW int32, startTime, endTime time.Time) (Workload, error) {
	var err error

	type wireSample struct {
		w       dbWire
		samples []WireWorkloadSample
	}

	wireSamples := make([]wireSample, len(wires))

	for i, wire := range wires {
		wireSamples[i].w = wire
		wireSamples[i].samples, err = GetWireWorkload(db, wire.wireId, startTime, endTime)
		if err != nil {
			return Workload{}, err
		}
	}

	t := startTime
search:
	for t.Add(time.Duration(durationM) * time.Minute).Before(endTime) {
		for i := 0; i < int(durationM); i++ {
			for _, ws := range wireSamples {
				if ws.samples[i].WorkloadW+workloadW > ws.w.capacityW {
					// we detected a wire overload
					t = t.Add(time.Duration(i) * time.Minute)
					continue search
				}
			}
		}

		// all whires are fine we can use t as our start time
		goto nooverload
	}
	// the wire will be overloaded at all time, the workload is not possible
	return Workload{}, ErrWorkloadNotPossible

nooverload:
	return Workload{StartTime: t, DurationM: durationM}, nil
}
