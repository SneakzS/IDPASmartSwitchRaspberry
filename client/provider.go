package client

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/philip-s/idpa/common"
)

type providerClientState struct {
	IsOK         bool
	EnableOutput bool
}

func runProviderClient(stateChan chan<- providerClientState, conn *sql.DB, c *Config, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	var (
		now              time.Time
		lastSampleUpdate time.Time
		sampleMap        map[int64]WorkloadSample
		currentState     providerClientState
	)

	for {
		newState := currentState

		select {
		case <-done:
			return

		case now = <-ticker.C:
			if lastSampleUpdate.Add(time.Minute).Before(now) {
				wls, err := updateProviderClient(now, conn, c.ProviderURL, int32(c.CustomerID))
				if err != nil {
					log.Println(err)
					newState.IsOK = false
					continue
				}

				newState.IsOK = true

				sampleMap = make(map[int64]WorkloadSample)
				for _, sample := range wls {
					sampleMap[sample.SampleTime.Unix()] = sample
				}
			}

			// update our output based on the loaded samples
			nowTC := now.Truncate(time.Minute)
			sample := sampleMap[nowTC.Unix()]
			newState.EnableOutput = sample.OutputEnabled
		}

		if newState != currentState {
			stateChan <- newState
			currentState = newState
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
			wl := Workload{}
			err = RequestWorkload(&wl, pw, serverURL, customerID)
			if err != nil {
				log.Println(err)
				continue
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
	Definition common.WorkloadDefinition
	MatchTime  time.Time
}

func PlanWorloads(definitions []common.WorkloadDefinition, startTime time.Time, durationM int32) []PlannedWorkload {
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

func RequestWorkload(w *Workload, pw PlannedWorkload, url string, customerID int32) error {

	req := common.WorkloadRequest{
		CustomerID:         customerID,
		DurationM:          pw.Definition.DurationM,
		ToleranceDurationM: pw.Definition.ToleranceDurationM,
		WorkloadW:          pw.Definition.WorkloadW,
		StartTime:          pw.MatchTime,
	}

	data, _ := json.Marshal(&req)

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errResp := common.ErrorResponse{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return fmt.Errorf("invalid server response: %w", err)
		}

		err := ErrWorkloadPlan{
			WorkloadDefinitionID: pw.Definition.WorkloadDefinitionID,
			MatchTime:            pw.MatchTime,
			Reason:               errResp.Message,
		}
		return err
	}

	workloadResp := common.WorkloadResponse{}
	err = json.Unmarshal(body, &workloadResp)
	if err != nil {
		return err // this is very bad
	}

	*w = Workload{
		WorkloadDefinitionID: pw.Definition.WorkloadDefinitionID,
		MatchTime:            pw.MatchTime,
		WorkloadW:            pw.Definition.WorkloadW,
		OffsetM:              workloadResp.OffsetM,
		DurationM:            pw.Definition.DurationM,
	}
	return nil
}
