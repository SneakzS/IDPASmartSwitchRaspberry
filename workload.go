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

func GetWireWorkload(conn *sql.DB, wireID int32, startTime, endTime time.Time) ([]WireWorkloadSample, error) {

	/*qry := fmt.Sprintf(`SELECT workloadW, startTime, endTime
	FROM WireWorkload WHERE wireID = %d AND startTime >= strftime('%%s', '%s') AND endTime <= strftime('%%s', '%s')`,
		wireID,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
	)*/
	res, err := conn.Query(
		`SELECT workloadW, startTime, endTime
		FROM WireWorkload WHERE wireID = ? AND startTime >= datetime(?) AND endTime <= datetime(?)`,
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
	EndTime   time.Time
}

func getOptimalWorkload(conn *sql.DB, wires []dbWire, durationM, workloadW int32, startTime, endTime time.Time) (Workload, error) {
	var err error

	type wireSample struct {
		w       dbWire
		samples []WireWorkloadSample
	}

	wireSamples := make([]wireSample, len(wires))

	for i, wire := range wires {
		wireSamples[i].w = wire
		wireSamples[i].samples, err = GetWireWorkload(conn, wire.wireId, startTime, endTime)
		if err != nil {
			return Workload{}, err
		}
	}

	var offsetM int32
search:
	for startTime.Add(time.Duration(offsetM+durationM) * time.Minute).Before(endTime) {
		for i := offsetM; i < offsetM+durationM; i++ {
			for _, ws := range wireSamples {
				if ws.samples[i].WorkloadW+workloadW > ws.w.capacityW {
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
	return Workload{}, ErrWorkloadNotPossible

nooverload:
	return Workload{
		StartTime: startTime.Add(time.Duration(offsetM) * time.Minute),
		EndTime:   startTime.Add(time.Duration(offsetM+durationM) * time.Minute),
	}, nil
}

func GetOptimalWorkload(conn *sql.DB, customerID, durationM, workloadW int32, startTime, endTime time.Time) (Workload, error) {
	wires, err := GetCustomerWires(conn, customerID)
	if err != nil {
		return Workload{}, err
	}
	if len(wires) == 0 {
		return Workload{}, ErrCustomerUnknown
	}
	return getOptimalWorkload(conn, wires, durationM, workloadW, startTime, endTime)
}
