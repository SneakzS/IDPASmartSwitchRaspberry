package main

import (
	"fmt"
	"math/rand"

	"github.com/philip-s/idpa/client"
)

type consolePiPin struct {
	name  string
	state bool
}

// make it global because gpio pins are global as well
var consolePiPins = [...]consolePiPin{
	{"led 1", false},
	{"led 2", false},
	{"led 3", false},
	{"relais", false},
}

func writeOutputConsole(o client.Output) {
	writeBoolToConsole(&consolePiPins[0], o.Led1)
	writeBoolToConsole(&consolePiPins[1], o.Led2)
	writeBoolToConsole(&consolePiPins[2], o.Led3)
	writeBoolToConsole(&consolePiPins[3], o.Relais)
}

func readInputConsole(inp *client.Input) error {
	if *mockSensorData {
		inp.Power = rand.Float64() * 10   // between 0 and 10
		inp.Current = rand.Float64() * 15 // between 0 and 15
		inp.Voltage = rand.Float64() * 36 // between 0 and 36
		return nil
	}
	return fmt.Errorf("sensor not available")
}

func writeBoolToConsole(pin *consolePiPin, b bool) {
	if pin.state != b {
		pin.state = b

		if b {
			fmt.Println("turn on ", pin.name)
		} else {
			fmt.Println("turn of ", pin.name)
		}
	}
}
