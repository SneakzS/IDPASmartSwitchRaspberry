//go:build linux

// implement raspberry pi on linux
package main

import (
	"github.com/philip-s/idpa"
	"github.com/stianeikeland/go-rpio"
)

var (
	led1Pin   = rpio.Pin(18)
	led2Pin   = rpio.Pin(22)
	led3Pin   = rpio.Pin(23)
	relaisPin = rpio.Pin(17)
)

type raspberryPi struct{}

func applyFlagToPin(pin rpio.Pin, mask, out uint) {
	if out&mask > 0 {
		pin.High()
	} else {
		pin.Low()
	}
}

func (raspberryPi) Set(out uint) {
	applyFlagToPin(led1Pin, idpa.OutLed1, out)
	applyFlagToPin(led2Pin, idpa.OutLed2, out)
	applyFlagToPin(led3Pin, idpa.OutLed3, out)
	applyFlagToPin(relaisPin, idpa.OutRelais, ^out) // relais is active low
}

// ensure that raspberryPi implements idpa.PiOutput
var _ idpa.PiOutput = raspberryPi{}

func setupGPIO() {
	led1Pin.Output()
	led2Pin.Output()
	led3Pin.Output()
	relaisPin.Output()

	led1Pin.Low()
	led2Pin.Low()
	led3Pin.Low()
	relaisPin.High() // relais is active low
}

func setupRPI() (idpa.PiOutput, error) {
	err := rpio.Open()
	if err != nil {
		return nil, err
	}

	setupGPIO()

	return raspberryPi{}, nil
}

func closeRPI() error {
	return rpio.Close()
}
