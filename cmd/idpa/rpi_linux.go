//go:build linux

// implement raspberry pi on linux
package main

import (
	ina219 "github.com/JeffAlyanak/goina219"
	"github.com/philip-s/idpa/client"
	"github.com/stianeikeland/go-rpio"
)

var (
	led1Pin   = rpio.Pin(16)
	led2Pin   = rpio.Pin(20)
	led3Pin   = rpio.Pin(21)
	relaisPin = rpio.Pin(17)

	sensor1 *ina219.INA219
)

func writeOutputRPI(o client.Output) {
	writeBooltoPin(led1Pin, o.Led1)
	writeBooltoPin(led2Pin, o.Led2)
	writeBooltoPin(led3Pin, o.Led3)
	writeBooltoPin(relaisPin, o.Relais)
}

func readInputRPI(inp *client.Input) error {
	err := ina219.Read(sensor1)
	if err != nil {
		return err
	}

	inp.Power = sensor1.Power
	inp.Current = sensor1.Current
	inp.Voltage = sensor1.Bus
	inp.Shunt = sensor1.Shunt

	return nil
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

func setupSensor1() error {
	config := ina219.Config(
		ina219.Range32V,
		ina219.Gain320MV,
		ina219.Adc12Bit,
		ina219.Adc12Bit,
		ina219.ModeContinuous,
	)

	ina, err := ina219.New(
		0x40, // ina219 address
		0x01, // i2c bus
		0.01, // Shunt resistance in ohms
		config,
		ina219.Gain320MV,
	)
	if err != nil {
		return err
	}

	sensor1 = ina
	return nil
}
