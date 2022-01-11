package idpa

import (
	"sync"
	"time"
)

type Pi struct {
	mux     sync.Mutex
	flags   uint32
	output  uint32
	outChan chan<- uint32
	done    chan struct{}
}

func (pi *Pi) SetFlags(flags, mask uint32) {
	pi.mux.Lock()
	defer pi.mux.Unlock()

	newFlags := flags & mask
	newFlags = newFlags | (pi.flags & ^mask)
	pi.flags = newFlags

	if newFlags&FlagIsEnabled > 0 {
		pi.setOutput(OutRelais|OutLed1, OutRelais|OutLed1)
	} else {
		pi.setOutput(0, OutRelais|OutLed1)
	}
}

func (pi *Pi) setOutput(output, mask uint32) {
	newOutput := output & mask
	newOutput = newOutput | (pi.output & ^mask)
	pi.output = newOutput

	pi.outChan <- newOutput
}

func (pi *Pi) SetOutput(output, mask uint32) {
	pi.mux.Lock()
	defer pi.mux.Unlock()
	pi.setOutput(output, mask)
}

func (pi *Pi) Flags() uint32 {
	pi.mux.Lock()
	defer pi.mux.Unlock()

	return pi.flags
}

const (
	OutLed1 = 1 << iota // on / off indicator
	OutLed2             // blinks when UI is not connected
	OutLed3             // blinks when operator client failed
	OutRelais
)

func NewPI(output chan<- uint32) *Pi {
	done := make(chan struct{})
	pi := Pi{
		outChan: output,
		done:    done,
	}

	// blink the leds
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)

		var led2, led3 bool

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				flags := pi.Flags()
				if flags&FlagIsUIConnected == 0 {
					led2 = !led2
				} else {
					led2 = false
				}

				if flags&FlagProviderClientOK == 0 {
					led3 = !led3
				} else {
					led3 = false
				}

				var output uint32
				if led2 {
					output |= OutLed2
				}

				if led3 {
					output |= OutLed3
				}

				pi.SetOutput(output, OutLed2|OutLed3)
			}
		}
	}()

	return &pi
}

func (pi *Pi) Close() {
	close(pi.done)
}
