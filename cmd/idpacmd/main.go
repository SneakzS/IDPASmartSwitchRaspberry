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
)

func main() {
	flag.Parse()

	config := idpa.DefaultConfig()
	if *configFile != "" {
		err := idpa.ReadConfig(&config, *configFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	switch *action {
	case "execute-script":
		if *script == "" {
			fmt.Fprintln(os.Stderr, "-script is required")
			goto printUseage
		}

		db, err := sql.Open("sqlite3", config.DatabaseFileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		script, err := os.ReadFile(*script)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var hasError bool
		statements := strings.Split(string(script), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if len(stmt) > 0 {
				_, err = db.Exec(stmt)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					hasError = true
				}
			}
		}

		if hasError {
			os.Exit(1)
		}

	case "get-workload":
		db, err := sql.Open("sqlite3", config.DatabaseFileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		startTime, err := time.Parse("2006-01-02 15:04", *startTimeStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		endTime, err := time.Parse("2006-01-02 15:04", *endTimeStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		samples, err := idpa.GetWireWorkload(db, int32(*wireID), startTime, endTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, s := range samples {
			fmt.Printf("%s;%d\n", s.Time.Format("2006-01-02 15:04:05"), s.WorkloadW)
		}
	}

	os.Exit(0)

printUseage:
	flag.PrintDefaults()
	os.Exit(1)
}
