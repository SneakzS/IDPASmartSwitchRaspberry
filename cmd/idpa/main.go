package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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
	config            = flag.String("config", "", "Specify the service config file")
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
	case "new-uiclientguid":
		err = newUIClientGuid()
	case "init-db":
		err = initializeDatabase()
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

	return invokeScriptFile(*script, *databaseFile)
}

func newUIClientGuid() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	newUUID := uuid.New()
	cfg.SetString("Client", "uiclientguid", newUUID.String())

	return saveConfig(cfg)
}

func initializeDatabase() error {
	if *serviceMode == "" {
		return fmt.Errorf("-service-mode is required")
	}

	if *databaseFile == "" {
		return fmt.Errorf("-db is required")
	}

	switch *serviceMode {
	case "server":
		return invokeScript(createServerDBScript, *databaseFile)
	case "client":
		return invokeScript(createClientDBScript, *databaseFile)
	default:
		return fmt.Errorf("invalid service mode %s", *serviceMode)
	}
}
