package main

import (
	"fmt"

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
		if on {
			fmt.Println("turn on led")
		} else {
			fmt.Println("turn off led")
		}
	}

	m.ledState = on
}

func (m *consolePi) SetRelais(on bool) {
	if on != m.relaisState {
		if on {
			fmt.Println("turn on relais")
		} else {
			fmt.Println("turn off relais")
		}
	}

	m.relaisState = on
}
