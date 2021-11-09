package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/philip-s/idpa"
)

var (
	action       = flag.String("action", "", "Specify the action you want to accomplish")
	configFile   = flag.String("config", "", "Specify a config file")
	script       = flag.String("script", "", "Specify a script file")
	wireID       = flag.Int("wire-id", 0, "Specify the ID of the wire")
	startTimeStr = flag.String("start-time", "", "Specify the start time")
	endTimeStr   = flag.String("end-time", "", "Specify the end time")
	customerID   = flag.Int("customer-id", 0, "Specify the customer ID")
	durationM    = flag.Int("duration-m", 0, "Duration in minutes")
	workloadW    = flag.Int("workload-w", 0, "Workload in Watts")
	add          = flag.Bool("add", false, "add the to the db")
)

func getStartAndEndTime() (startTime, endTime time.Time, err error) {
	startTime, err = time.Parse("2006-01-02 15:04", *startTimeStr)
	if err != nil {
		return
	}
	endTime, err = time.Parse("2006-01-02 15:04", *endTimeStr)
	if err != nil {
		return
	}

	return
}

func executeScript(config idpa.Config) error {
	if *script == "" {
		return fmt.Errorf("-script is required")
	}

	db, err := sql.Open("sqlite3", config.DatabaseFileName)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	script, err := os.ReadFile(*script)
	if err != nil {
		return err
	}

	statements := strings.Split(string(script), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if len(stmt) > 0 {
			_, err = db.Exec(stmt)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func getWireWorkload(config idpa.Config) error {
	conn, err := sql.Open("sqlite3", config.DatabaseFileName)
	if err != nil {
		return err
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	startTime, endTime, err := getStartAndEndTime()
	if err != nil {
		return err
	}

	samples, err := idpa.GetWireWorkload(tx, int32(*wireID), startTime, endTime)
	if err != nil {
		return err
	}

	for _, s := range samples {
		fmt.Printf("%s;%d\n", s.Time.Format("2006-01-02 15:04:05"), s.WorkloadW)
	}

	return nil
}

func getOptimalWorkload(config idpa.Config) error {
	db, err := sql.Open("sqlite3", config.DatabaseFileName)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	startTime, endTime, err := getStartAndEndTime()
	if err != nil {
		return err
	}

	wires, err := idpa.GetCustomerWires(tx, int32(*customerID))
	if err != nil {
		return err
	}

	workload, err := idpa.GetOptimalWorkload(tx, wires, int32(*durationM), int32(*workloadW), startTime, endTime)
	if err != nil {
		return err
	}
	fmt.Printf("Your optimal start time is %s until %s\n", workload.StartTime, workload.EndTime)

	if *add {
		for _, wire := range wires {
			err = idpa.AddWireWorkload(tx, wire.WireID, workload)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()

}

func main() {
	flag.Parse()

	var err error

	config := idpa.DefaultConfig()
	if *configFile != "" {
		err = idpa.ReadConfig(&config, *configFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	switch *action {
	case "execute-script":
		err = executeScript(config)
	case "get-wire-workload":
		err = getWireWorkload(config)
	case "get-optimal-workload":
		err = getOptimalWorkload(config)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
