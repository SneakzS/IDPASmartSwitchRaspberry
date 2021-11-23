package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/philip-s/idpa"
	"github.com/stianeikeland/go-rpio"
)

type raspberryPi struct {
	ledPin            rpio.Pin
	relaisPin         rpio.Pin
	lastLedToggleTime time.Time
	flags             uint64
	mux               sync.Mutex
}

func setupGPIO(rpi *raspberryPi) {
	rpi.ledPin.Output()
	rpi.relaisPin.Output()
	rpi.ledPin.Low()
	rpi.relaisPin.Low()
}

func applyGPIO(rpi *raspberryPi, now time.Time) {
	rpi.mux.Lock()
	defer rpi.mux.Unlock()

	switch {
	// We detected an error conditon, flash the led
	case idpa.FlagHasError&rpi.flags > 0:
		if rpi.lastLedToggleTime.Add(500 * time.Millisecond).Before(now) {
			rpi.ledPin.Toggle()
			rpi.lastLedToggleTime = now
			rpi.relaisPin.High() // shut off the power supply

			fmt.Printf("toggle led to %d\n", rpi.ledPin.Read())
		}

	// The output should be enabled
	case idpa.FlagIsEnabled&rpi.flags > 0:
		rpi.relaisPin.Low()
		rpi.ledPin.High()

	// The output should be disabled
	case idpa.FlagIsEnabled&rpi.flags == 0:
		rpi.ledPin.Low()
		rpi.relaisPin.High()
	}

}

func (rpi *raspberryPi) setFlags(flags, mask uint64) {
	// flip the mask so that every bit that we don't care is 1
	mask = ^mask

	rpi.mux.Lock()
	defer rpi.mux.Unlock()

	rpi.flags = rpi.flags & (mask & flags)
}

// setFlagsUnsafe sets the flags without locking the mutex
func (rpi *raspberryPi) setFlagsUnsafe(flags, mask uint64) {
	// flip the mask so that every bit that we don't care is 1
	mask = ^mask
	rpi.flags = rpi.flags & (mask & flags)
}
