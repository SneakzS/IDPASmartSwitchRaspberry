package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philip-s/idpa"
)

func getOptimalWorkload() error {
	if *databaseFile == "" {
		return fmt.Errorf("-db is required")
	}

	db, err := sql.Open("sqlite3", *databaseFile)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	startTime, err := time.Parse("2006-01-02 15:04:05", *startTimeStr)
	if err != nil {
		return err
	}

	wires, err := idpa.GetCustomerWires(tx, int32(*customerID))
	if err != nil {
		return err
	}

	d := idpa.WorkloadDefinition{
		WorkloadW:          int32(*workloadW),
		DurationM:          int32(*durationM),
		ToleranceDurationM: int32(*toleranceDuration),
		IsEnabled:          true,
	}

	offsetM, err := idpa.GetOptimalWorkloadOffset(tx, wires, d, startTime)
	if err != nil {
		return err
	}
	fmt.Printf("Your optimal starts at %s \n", startTime.Add(time.Duration(offsetM)*time.Minute).Format("2006-01-02 15:04:05"))

	if *add {
		for _, wire := range wires {
			err = idpa.AddWireWorkload(tx, wire.WireID, startTime, d.DurationM, d.WorkloadW)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func getWireWorkload() error {
	if *databaseFile == "" {
		return fmt.Errorf("-db is required")
	}

	conn, err := sql.Open("sqlite3", *databaseFile)
	if err != nil {
		return err
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	startTime, err := time.Parse("2006-01-02 15:04:05", *startTimeStr)
	if err != nil {
		return err
	}

	samples, err := idpa.GetWireWorkload(tx, int32(*wireID), startTime, int32(*durationM))
	if err != nil {
		return err
	}

	for _, s := range samples {
		fmt.Printf("%s;%d\n", s.SampleTime.Format("2006-01-02 15:04:05"), s.WorkloadW)
	}

	return nil
}
