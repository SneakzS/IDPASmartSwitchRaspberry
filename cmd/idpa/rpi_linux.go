//go:build linux

// implement raspberry pi on linux
package main

import (
	"github.com/philip-s/idpa"
	"github.com/stianeikeland/go-rpio"
)

type raspberryPi struct {
	ledPin    rpio.Pin
	relaisPin rpio.Pin
}

var raspberryPiTemplate = raspberryPi{
	ledPin:    rpio.Pin(18),
	relaisPin: rpio.Pin(17),
}

// ensure that raspberryPi implements idpa.PiOutput
var _ idpa.PiOutput = raspberryPi{}

func (rpi raspberryPi) setupGPIO() {
	rpi.ledPin.Output()
	rpi.relaisPin.Output()
	rpi.ledPin.Low()
	rpi.relaisPin.Low()
}

func (rpi raspberryPi) SetLed(on bool) {
	if on {
		rpi.ledPin.High()
	} else {
		rpi.ledPin.Low()
	}
}

func (rpi raspberryPi) SetRelais(on bool) {
	if on {
		rpi.relaisPin.Low() // relais is active low
	} else {
		rpi.relaisPin.High()
	}
}

func setupRPI() (idpa.PiOutput, error) {
	err := rpio.Open()
	if err != nil {
		return nil, err
	}

	rpi := raspberryPiTemplate
	rpi.setupGPIO()

	return rpi, nil
}

func closeRPI() {
	rpio.Close()
}
