package main

import (
	"fmt"
	"time"

	"github.com/philip-s/idpa"
)

type consolePi struct {
	ledState    bool
	relaisState bool
}

// consolePi must implement idpa.PiOutput
var _ idpa.PiOutput = &consolePi{}

func (m *consolePi) SetLed(on bool) {
	if on != m.ledState {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		if on {
			fmt.Println(timestamp, ": turn on led")
		} else {
			fmt.Println(timestamp, ": turn off led")
		}
	}

	m.ledState = on
}

func (m *consolePi) SetRelais(on bool) {
	if on != m.relaisState {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		if on {
			fmt.Println(timestamp, ": turn on relais")
		} else {
			fmt.Println(timestamp, ": turn off relais")
		}
	}

	m.relaisState = on
}
