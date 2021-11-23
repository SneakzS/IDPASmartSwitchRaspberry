package main

import (
	"fmt"
	"os"
	"time"

	"github.com/philip-s/idpa"
	"github.com/stianeikeland/go-rpio"
)

/*type memoryPin struct {
	rpio.Pin
	v rpio.State
}

func (m memoryPin) Toggle() {
	if m.v == rpio.High {
		m.v = rpio.Low
		m.Low()
	} else {
		m.v = rpio.High
		m.High()
	}
}*/

func main() {
	rpi := raspberryPi{
		ledPin:    rpio.Pin(7),
		relaisPin: rpio.Pin(17),
		flags:     idpa.FlagHasError,
	}

	handleUIConnection(&rpi, "ws://10.158.46.219:5028/ws")
	return

	go handleUIConnection(&rpi, "ws://10.0.0.3/ws")

	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	setupGPIO(&rpi)

	for {
		applyGPIO(&rpi, time.Now())
		time.Sleep(10 * time.Millisecond)
	}
}
