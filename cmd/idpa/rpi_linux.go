//go:build linux

// implement raspberry pi on linux
package main

import (
	"github.com/d2r2/go-logger"
	"github.com/philip-s/idpa/client"
	ina260 "github.com/phipus/goina260"
	"github.com/stianeikeland/go-rpio"
)

var (
	led1Pin   = rpio.Pin(16)
	led2Pin   = rpio.Pin(20)
	led3Pin   = rpio.Pin(21)
	relaisPin = rpio.Pin(17)

	sensor1 ina260.S
)

func writeOutputRPI(o client.Output) {
	writeBooltoPin(led1Pin, o.Led1)
	writeBooltoPin(led2Pin, o.Led2)
	writeBooltoPin(led3Pin, o.Led3)
	writeBooltoPin(relaisPin, o.Relais)
}

func readInputRPI(inp *client.Input) (err error) {
	inp.Voltage, inp.Current, inp.Power, err = sensor1.ReadData(true, true, true)
	return
}

func writeBooltoPin(p rpio.Pin, b bool) {
	if b {
		p.Low() // all pins are active low
	} else {
		p.High()
	}
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

func setupRPI() error {
	err := rpio.Open()
	if err != nil {
		return err
	}

	setupGPIO()
	err = setupSensor1()
	if err != nil {
		rpio.Close()
		return err
	}

	return nil
}

func closeRPI() error {
	return rpio.Close()
}

func setupSensor1() (err error) {
	logger.ChangePackageLogLevel("i2c", logger.WarnLevel)
	sensor1, err = ina260.New(
		0x40, // ina260 address
		0x01, // i2c bus
	)
	return
}
