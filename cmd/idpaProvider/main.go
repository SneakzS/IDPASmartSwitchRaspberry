package main

import (
	_ "embed"
	"flag"
	"log"
	"os"

	"github.com/philip-s/idpa/common"
	"github.com/philip-s/idpa/provider"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema-provider.sql
var createProviderDBScript string

var (
	action             = flag.String("action", "", "Specify the action you want to accomplish")
	config             = flag.String("config", "", "Specify config file")
	startTimeStr       = flag.String("starttime", "", "Specify the start time")
	customerID         = flag.Int("customerid", 0, "Specify the customer id")
	workloadW          = flag.Int("workload", 0, "Specify the workload in watts")
	durationM          = flag.Int("duration", 0, "Specify the duration in minutes")
	toleranceDurationM = flag.Int("toleranceduration", 0, "Specify the tolerance duration in minutes")
	wireID             = flag.Int("wireid", 0, "Specify the Wire ID")
	add                = flag.Bool("add", false, "Add to the database")
)

func main() {
	var err error

	flag.Parse()

	switch *action {
	case "initdb":
		err = initializeDatabase()

	case "wireworkload":
		err = getWireWorkload()

	case "optimalworkload":
		err = getOptimalWorkload()

	case "service":
		err = runService()
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func initializeDatabase() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}
	c := provider.Config{}
	err = provider.ReadConfig(&c, ini)
	if err != nil {
		return err
	}

	err = common.InvokeSQLScript(createProviderDBScript, c.Database)
	return err
}
