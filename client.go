package idpa

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func RunProviderClient(ctx context.Context, pi *Pi, conn *sql.DB, serverURL string, customerID int32) {
	ticker := time.NewTicker(time.Second)
	var (
		now              time.Time
		lastSampleUpdate time.Time
		sampleMap        map[int64]WorkloadSample
	)

	for {
		select {
		case <-ctx.Done():
			return

		case now = <-ticker.C:
			if lastSampleUpdate.Add(time.Minute).Before(now) {
				wls, err := updateProviderClient(now, conn, serverURL, customerID)
				if err != nil {
					log.Println(err)
					pi.SetFlags(0, FlagProviderClientOK)
					continue
				}

				pi.SetFlags(FlagProviderClientOK, FlagProviderClientOK)

				sampleMap = make(map[int64]WorkloadSample)
				for _, sample := range wls {
					sampleMap[sample.SampleTime.Unix()] = sample
				}
			}

			// if no specific output is enforced, check if we have an active workflow
			// and act acordingly

			flags := pi.Flags()

			if flags&FlagEnforce == 0 {
				nowTC := now.Truncate(time.Minute)
				sample := sampleMap[nowTC.Unix()]

				if sample.OutputEnabled {
					pi.SetFlags(FlagIsEnabled, FlagIsEnabled)
				} else {
					pi.SetFlags(0, FlagIsEnabled)
				}
			}
		}
	}
}

type ErrWorkloadPlan struct {
	WorkloadDefinitionID int32
	MatchTime            time.Time
	Reason               string
}

func (e ErrWorkloadPlan) Error() string {
	return fmt.Sprintf("cannot create workload for definition %d matching %s: %s",
		e.WorkloadDefinitionID, e.MatchTime.Format("2006-01-02 15:04:05"), e.Reason,
	)
}

func updateProviderClient(now time.Time, conn *sql.DB, serverURL string, customerID int32) ([]WorkloadSample, error) {
	type int32Time struct {
		i  int32
		ts int64
	}

	tx, err := conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	definitions, err := GetWorkloadDefinitions(tx)
	if err != nil {
		return nil, err
	}

	workloads, err := GetWorkloads(tx, now, 12*60)
	if err != nil {
		return nil, err
	}

	samples, err := GetWorkloadSamples(tx, now, 12*60)
	if err != nil {
		return nil, err
	}

	set := make(map[int32Time]struct{})
	for _, w := range workloads {
		set[int32Time{w.WorkloadDefinitionID, w.MatchTime.Unix()}] = struct{}{}
	}

	planned := PlanWorloads(definitions, now, 12*60)

	for _, pw := range planned {
		_, exists := set[int32Time{pw.Definition.WorkloadDefinitionID, pw.MatchTime.Unix()}]
		overlaps := false
		s := WorkloadSample{}
		if !exists {
			s, overlaps = CheckWorkloadOverlaps(samples, pw.MatchTime, pw.Definition.ToleranceDurationM)

		}

		if overlaps {
			log.Println(ErrWorkloadPlan{
				WorkloadDefinitionID: pw.Definition.WorkloadDefinitionID,
				MatchTime:            pw.MatchTime,
				Reason:               fmt.Sprintf("it overlaps with workload %d", s.WorkloadID),
			})
			continue
		}

		if !exists {
			wl, err := RequestWorkload(pw, serverURL, customerID)
			if err != nil {
				if err == ErrWorkloadNotPossible {
					log.Println(ErrWorkloadPlan{
						WorkloadDefinitionID: pw.Definition.WorkloadDefinitionID,
						MatchTime:            pw.MatchTime,
						Reason:               "workload is not possible",
					})
					continue
				} else {
					return nil, err
				}
			}

			err = CreateWorkloadAndSamples(tx, pw.Definition, pw.MatchTime, wl.OffsetM)
			if err != nil {
				return nil, err
			}

			samples, err = GetWorkloadSamples(tx, now, 12*60)
			if err != nil {
				return nil, err
			}
		}
	}

	samples, err = GetWorkloadSamples(tx, now, 12*60)
	if err != nil {
		return nil, err
	}

	return samples, tx.Commit()
}

type PlannedWorkload struct {
	Definition WorkloadDefinition
	MatchTime  time.Time
}

func PlanWorloads(definitions []WorkloadDefinition, startTime time.Time, durationM int32) []PlannedWorkload {
	var planned []PlannedWorkload
	startTime = startTime.Truncate(time.Minute)

	for i := int32(0); i < durationM; i++ {
		t := startTime.Add(time.Duration(i) * time.Minute)
		for _, d := range definitions {
			if d.IsEnabled {
				// one of the repeat patterns must match
				for _, rp := range d.RepeatPattern {
					if rp.Matches(t) {
						planned = append(planned, PlannedWorkload{
							Definition: d,
							MatchTime:  t,
						})
						break
					}
				}
			}
		}
	}

	return planned
}

func CheckWorkloadOverlaps(samples []WorkloadSample, startTime time.Time, durationM int32) (WorkloadSample, bool) {
	set := make(map[int64]WorkloadSample)
	startTime = startTime.Truncate(time.Minute)

	for _, s := range samples {
		set[s.SampleTime.Unix()] = s
	}

	for i := int32(0); i < durationM; i++ {
		if s := set[startTime.Add(time.Duration(i)*time.Minute).Unix()]; s.OutputEnabled {
			return s, true
		}
	}

	return WorkloadSample{}, false
}

func RequestWorkload(pw PlannedWorkload, serverURL string, customerID int32) (Workload, error) {
	dialer := websocket.Dialer{}

	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		return Workload{}, err
	}
	defer conn.Close()

	o := Offer{}
	p := ProviderConnection{conn}
	err = handleProviderClient(&o, p, pw.Definition, pw.MatchTime, customerID)
	if err != nil {
		return Workload{}, err
	}

	if o.WorkloadW != pw.Definition.WorkloadW {
		return Workload{}, ErrInvalidMessage
	}

	return Workload{
		WorkloadDefinitionID: pw.Definition.WorkloadDefinitionID,
		MatchTime:            pw.MatchTime,
		WorkloadW:            o.WorkloadW,
		OffsetM:              o.OffsetM,
		DurationM:            pw.Definition.DurationM,
	}, nil

}
