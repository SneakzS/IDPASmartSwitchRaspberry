package idpa

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type PiState struct {
	mux               sync.Mutex
	flags             uint64
	lastLedToggleTime time.Time
	ledState          bool
	c                 Config
}

type PiOutput interface {
	SetLed(on bool)
	SetRelais(on bool)
}

func (s *PiState) Set(o PiOutput, now time.Time) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch {
	// We detected an error conditon, flash the led
	case FlagHasConnectionError&s.flags > 0:

		if s.lastLedToggleTime.Add(500 * time.Millisecond).Before(now) {
			fmt.Println("toggle led")
			s.ledState = !s.ledState
			o.SetRelais(false)
			o.SetLed(s.ledState)
			s.lastLedToggleTime = now
		}

	// The output should be enabled
	case FlagIsEnabled&s.flags > 0:
		o.SetRelais(true)
		o.SetLed(true)

	// The output should be disabled
	case FlagIsEnabled&s.flags == 0:
		o.SetRelais(false)
		o.SetLed(false)
	}
}

func (s *PiState) SetFlags(flags, mask uint64) {
	// flip the mask so that every bit that we don't care is 1
	mask = ^mask

	s.mux.Lock()
	defer s.mux.Unlock()

	s.flags = (s.flags & mask) | flags
}

func RunPi(ctx context.Context, o PiOutput, s *PiState) {
	t := time.NewTimer(10 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case now := <-t.C:
			t.Reset(10 * time.Millisecond)
			s.Set(o, now)
		}
	}
}
