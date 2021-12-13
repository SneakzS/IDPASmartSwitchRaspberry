package idpa

import (
	"context"
	"fmt"
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

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

loop:
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

				fmt.Printf("new flags received: 0b%b\n", flags)

				now = time.Now()
				goto applyOutput

			case EventSetWorkloads:
				sampleMap = make(map[int64]WorkloadSample)
				for _, sample := range ev.Samples {
					sampleMap[sample.SampleTime.Unix()] = sample
				}

				now = time.Now()
				goto applyOutput
			}

		case now = <-ticker.C:
			// The time was updated

			// if no specific output is enforced, check if we have an active workflow
			// and act acordingly
			if flags&FlagEnforce == 0 {
				nowTC := now.Truncate(time.Minute)
				sample := sampleMap[nowTC.Unix()]

				if sample.OutputEnabled {
					flags = flags | FlagIsEnabled
				} else {
					flags = flags & (^uint64(FlagIsEnabled))
				}
			}
			goto applyOutput
		}

	}

applyOutput:

	hasError := FlagIsUIConnected&flags == 0

	if hasError {
		if lastLedToggleTime.Add(200 * time.Millisecond).Before(now) {
			ledState = !ledState
			lastLedToggleTime = now
		}
	} else {
		ledState = false
	}

	ledEnabled := FlagIsEnabled&flags > 0 || ledState
	relaisEnabled := FlagIsEnabled&flags > 0

	o.SetLed(ledEnabled)
	o.SetRelais(relaisEnabled)

	goto loop
}

func setFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: flag, FlagMask: flag}
}

func clearFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: 0, FlagMask: flag}
}
