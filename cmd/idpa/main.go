package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/drbig/simpleini"
	_ "github.com/mattn/go-sqlite3"
	"github.com/philip-s/idpa"
)

var (
	action            = flag.String("action", "", "Specify the action you want to accomplish")
	script            = flag.String("script", "", "Specify a script file")
	wireID            = flag.Int("wire-id", 0, "Specify the ID of the wire")
	startTimeStr      = flag.String("start-time", "", "Specify the start time")
	customerID        = flag.Int("customer-id", 0, "Specify the customer ID")
	durationM         = flag.Int("duration-m", 0, "Duration in minutes")
	toleranceDuration = flag.Int("tolerance-duration-m", 0, "Specify the tolerance duration")
	workloadW         = flag.Int("workload-w", 0, "Workload in Watts")
	add               = flag.Bool("add", false, "add the to the db")
	databaseFile      = flag.String("db", "", "Specify the database")
	serviceConfig     = flag.String("service", "", "Specify the service config file")
	serviceMode       = flag.String("service-mode", "", "Override the service mode specified in the config")
)

func main() {
	flag.Parse()

	var err error

	switch *action {
	case "execute-script":
		err = executeScript()
	case "get-wire-workload":
		err = getWireWorkload()
	case "get-optimal-workload":
		err = getOptimalWorkload()
	case "service":
		err = runService()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func executeScript() error {
	if *script == "" {
		return fmt.Errorf("-script is required")
	}
	if *databaseFile == "" {
		return fmt.Errorf("-db is required")
	}

	db, err := sql.Open("sqlite3", *databaseFile)
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

func loadServiceConfig() (*simpleini.INI, error) {
	if *serviceConfig == "" {
		return nil, fmt.Errorf("-service is required")
	}

	cfp, err := os.Open(*serviceConfig)
	if err != nil {
		return nil, err
	}
	defer cfp.Close()

	return simpleini.Parse(cfp)
}

func runService() error {
	cfg, err := loadServiceConfig()
	if err != nil {
		return err
	}

	mode, err := cfg.GetString("Service", "mode")
	if err != nil {
		return err
	}

	// Override the service mode based on the flag
	if *serviceMode != "" {
		mode = *serviceMode
	}

	switch mode {
	case "client":
		return runClient(cfg)
	case "server":
		return runServer(cfg)
	default:
		return fmt.Errorf("invalid service mode %s", mode)
	}
}

func runClient(cfg *simpleini.INI) error {
	var output idpa.PiOutput
	outputStr, err := cfg.GetString("Client", "output")
	if err != nil {
		return err
	}
	switch outputStr {
	case "rpi":
		output, err = setupRPI()
		if err != nil {
			return err
		}
		defer closeRPI()

	case "console":
		output = &consolePi{}

	default:
		return fmt.Errorf("invalid output %s", outputStr)
	}

	/*uiServerURL, err := cfg.GetString("Client", "uiserverurl")
	if err != nil {
		return err
	}*/
	providerServerURL, err := cfg.GetString("Client", "providerserverurl")
	if err != nil {
		return err
	}
	database, err := cfg.GetString("Client", "database")
	if err != nil {
		return err
	}
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return err
	}
	defer conn.Close()

	customerID, err := cfg.GetInt("Client", "customerid")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan idpa.PiEvent, 64)

	go idpa.RunPI(ctx, events, output)
	//go idpa.RunUIClient(ctx, events, uiServerURL)
	go idpa.RunProviderClient(ctx, events, conn, providerServerURL, int32(customerID))

	// Wait for SIGINT to quit
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done
	return nil
}

func runServer(cfg *simpleini.INI) error {
	database, err := cfg.GetString("Server", "database")
	if err != nil {
		return err
	}

	listen, err := cfg.GetString("Server", "listen")
	if err != nil {
		return err
	}

	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return err
	}
	defer conn.Close()

	return idpa.RunServer(listen, conn)
}
