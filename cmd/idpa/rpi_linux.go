//go:build linux

// implement raspberry pi on linux
package main

import (
	ina219 "github.com/JeffAlyanak/goina219"
	"github.com/philip-s/idpa"
	"github.com/stianeikeland/go-rpio"
)

var (
	led1Pin   = rpio.Pin(16)
	led2Pin   = rpio.Pin(20)
	led3Pin   = rpio.Pin(21)
	relaisPin = rpio.Pin(17)
)

type raspberryPi struct{}

func applyFlagToPin(pin rpio.Pin, mask, out uint32) {
	if out&mask > 0 {
		pin.High()
	} else {
		pin.Low()
	}
}

func (raspberryPi) write(out uint32) {
	// invert output because all outputs are active low
	out = ^out
	applyFlagToPin(led1Pin, idpa.OutLed1, out)
	applyFlagToPin(led2Pin, idpa.OutLed2, out)
	applyFlagToPin(led3Pin, idpa.OutLed3, out)
	applyFlagToPin(relaisPin, idpa.OutRelais, out)
}

func setupGPIO() {
	led1Pin.Output()
	led2Pin.Output()
	led3Pin.Output()
	relaisPin.Output()

	// All outputs are active low
	led1Pin.High()
	led2Pin.High()
	led3Pin.High()
	relaisPin.High()
}

func setupRPI() (raspberryPi, error) {
	err := rpio.Open()
	if err != nil {
		return raspberryPi{}, err
	}

	setupGPIO()

	setupi2c()

	return raspberryPi{}, nil
}

func closeRPI() error {
	return rpio.Close()
}

func setupi2c() error {
	config := ina219.Config(
		ina219.Range32V,
		ina219.Gain320MV,
		ina219.Adc12Bit,
		ina219.Adc12Bit,
		ina219.ModeContinuous,
	)

	myINA219, err := ina219.New(
		0x40, // ina219 address
		0x00, // i2c bus
		0.01, // Shunt resistance in ohms
		config,
		ina219.Gain320MV,
	)
	if err != nil {
		return err
	}
}
