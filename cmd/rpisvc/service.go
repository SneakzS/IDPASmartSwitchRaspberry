package main

import (
	"database/sql"
	"time"

	"github.com/philip-s/idpa"
)

type serviceState struct {
	conn      *sql.Conn
	rpi       raspberryPi
	workloads []idpa.WorkloadDefinition
}

func runService(s *serviceState, t time.Time) error {
	t = t.UTC()
	return nil
}
