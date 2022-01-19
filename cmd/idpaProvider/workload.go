package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philip-s/idpa/common"
	"github.com/philip-s/idpa/provider"
)

func getOptimalWorkload() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}

	c := provider.Config{}
	err = provider.ReadConfig(&c, ini)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", c.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	startTime, err := time.Parse("2006-01-02 15:04:05", *startTimeStr)
	if err != nil {
		return err
	}

	wires, err := provider.GetCustomerWires(tx, int32(*customerID))
	if err != nil {
		return err
	}

	d := common.WorkloadDefinition{
		WorkloadW:          int32(*workloadW),
		DurationM:          int32(*durationM),
		ToleranceDurationM: int32(*toleranceDurationM),
		IsEnabled:          true,
	}

	offsetM, err := provider.GetOptimalWorkloadOffset(
		tx,                         // tx
		wires,                      // wires
		int32(*durationM),          // durationM
		int32(*toleranceDurationM), // toleranceDurationM
		int32(*workloadW),          // workloadW
		startTime,                  // startTime
	)
	if err != nil {
		return err
	}
	fmt.Printf("Your optimal starts at %s \n", startTime.Add(time.Duration(offsetM)*time.Minute).Format("2006-01-02 15:04:05"))

	if *add {
		for _, wire := range wires {
			err = provider.AddWireWorkload(tx, wire.WireID, startTime, d.DurationM, d.WorkloadW)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func getWireWorkload() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}

	c := provider.Config{}
	err = provider.ReadConfig(&c, ini)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", c.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	startTime, err := time.Parse("2006-01-02 15:04:05", *startTimeStr)
	if err != nil {
		return err
	}

	samples, err := provider.GetWireWorkload(tx, int32(*wireID), startTime, int32(*durationM))
	if err != nil {
		return err
	}

	for _, s := range samples {
		fmt.Printf("%s;%d\n", s.SampleTime.Format("2006-01-02 15:04:05"), s.WorkloadW)
	}

	return nil
}
