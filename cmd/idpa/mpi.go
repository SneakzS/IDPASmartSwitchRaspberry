package main

import (
	"fmt"

	"github.com/philip-s/idpa"
)

type consolePiPin struct {
	mask  uint32
	name  string
	state bool
}

// make it global because gpio pins are global as well
var consolePiPins = []consolePiPin{
	{idpa.OutLed1, "led 1", false},
	{idpa.OutLed2, "led 2", false},
	{idpa.OutLed3, "led 3", false},
	{idpa.OutRelais, "relais", false},
}

func (c *consolePiPin) write(out uint32) {
	if out&c.mask > 0 {
		if !c.state {
			fmt.Println("turn on ", c.name)
			c.state = true
		}
	} else {
		if c.state {
			fmt.Println("turn of ", c.name)
			c.state = false
		}
	}
}

type consolePi struct{}

var _ piOutput = consolePi{}

func (consolePi) write(out uint32) {
	for i := range consolePiPins {
		consolePiPins[i].write(out)
	}
}
