package main

import (
	"flag"
	"fmt"
	"os"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/philip-s/idpa/client"
	"github.com/philip-s/idpa/common"
)

//go:embed schema-client.sql
var createClientDBScript string

var (
	action         = flag.String("action", "", "Specify the action you want to accomplish")
	config         = flag.String("config", "", "Specify the service config file")
	help           = flag.Bool("help", false, "Print help")
	mockSensorData = flag.Bool("mocksensordata", false, "Create random sensor data if output is console (for testing only)")
)

func main() {
	flag.Parse()

	if *help {
		*action = "help"
	}

	var err error

	switch *action {
	case "help":
		printHelp()
	case "service":
		err = runService()
	case "initdb":
		err = initializeDatabase()
	case "newguid":
		guid := uuid.New()
		fmt.Println(guid)

	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initializeDatabase() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}

	config := client.Config{}
	err = client.ReadConfig(&config, ini)
	if err != nil {
		return err
	}

	return common.InvokeSQLScript(createClientDBScript, config.Database)
}
