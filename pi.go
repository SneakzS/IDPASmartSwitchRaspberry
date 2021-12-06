package idpa

import (
	"context"
	"time"
)

const (
	_ = iota
	EventSetFlags
	EventSetWorkloads
)

type PiEvent struct {
	EventID  int
	Flags    uint64
	FlagMask uint64
	Samples  []WorkloadSample
}

type PiOutput interface {
	SetLed(on bool)
	SetRelais(on bool)
}

// RunPI listens for events on the events chan and processes them accordingly
func RunPI(ctx context.Context, events <-chan PiEvent, o PiOutput) {
	var (
		flags             uint64
		ledState          bool
		lastLedToggleTime time.Time
		now               time.Time
		sampleMap         map[int64]WorkloadSample
	)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// the context is done
			return
		case ev := <-events:
			// An event occured
			switch ev.EventID {
			case EventSetFlags:
				// flip the mask so that every bit that we don't care is 1
				mask := ^ev.FlagMask
				flags = (flags & mask) | ev.Flags

			case EventSetWorkloads:
				sampleMap = make(map[int64]WorkloadSample)
				for _, sample := range ev.Samples {
					sampleMap[sample.SampleTime.Unix()] = sample
				}
			}
			continue
		case now = <-ticker.C:
			// The time was updated
			now = now.UTC().Truncate(time.Minute)
		}

		// Update our output based on the event

		if flags&FlagEnforce == 0 {
			// the output is not enforced. Check if we have an active workload and act acordingly
			sample, ok := sampleMap[now.Unix()]
			if !ok {
				continue
			}
			if sample.OutputEnabled {
				flags = flags | FlagIsEnabled
			} else {
				flags = flags & (^uint64(FlagIsEnabled))
			}

		}

		switch {
		// We detected an error conditon, flash the led
		case FlagIsUIConnected&flags == 0:
			// ui is not connected, flash the led
			if false && lastLedToggleTime.Add(500*time.Millisecond).Before(now) {
				ledState = !ledState
				o.SetRelais(false)
				o.SetLed(ledState)
				lastLedToggleTime = now
			}
		}

		switch {
		// The output should be enabled
		case FlagIsEnabled&flags > 0:
			o.SetRelais(true)
			o.SetLed(true)

		// The output should be disabled
		case FlagIsEnabled&flags == 0:
			o.SetRelais(false)
			o.SetLed(false)
		}

	}
}

func setFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: flag, FlagMask: flag}
}

func clearFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: 0, FlagMask: flag}
}
