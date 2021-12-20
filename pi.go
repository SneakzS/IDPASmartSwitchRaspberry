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

const (
	OutLed1 = 1 << iota
	OutLed2
	OutLed3
	OutRelais
)

type PiOutput interface {
	Set(state uint)
}

// RunPI listens for events on the events chan and processes them accordingly
func RunPI(ctx context.Context, events <-chan PiEvent, o PiOutput) {
	var (
		flags             uint64
		lastLedToggleTime time.Time
		now               time.Time
		sampleMap         map[int64]WorkloadSample
		blink             bool
		out               uint
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

	// if output is enabled, enable the relais and led 1
	if flags&FlagIsEnabled > 0 {
		out = setOut(out, OutLed1|OutRelais)
	} else {
		out = clearOut(out, OutLed1|OutRelais)
	}

	// if FlagIsUIConnected or FlagProviderClientOK ist cleared
	// toggle blink every 200 ms
	if flags&FlagIsUIConnected == 0 || flags&FlagProviderClientOK == 0 {
		if lastLedToggleTime.Add(200 * time.Millisecond).Before(now) {
			blink = !blink
			lastLedToggleTime = now
		}
	}

	// blink led 2 if the ui is not connected
	if flags&FlagIsUIConnected == 0 {
		if blink {
			out = setOut(out, OutLed2)
		} else {
			out = clearOut(out, OutLed2)
		}
	} else {
		out = clearOut(out, OutLed2)
	}

	// blink led 3 if the provider returned an error
	if flags&FlagProviderClientOK == 0 {
		if blink {
			out = setOut(out, OutLed3)
		} else {
			out = clearOut(out, OutLed3)
		}
	} else {
		out = clearOut(out, OutLed3)
	}

	o.Set(out)

	goto loop
}

func setFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: flag, FlagMask: flag}
}

func clearFlag(flag uint64) PiEvent {
	return PiEvent{EventID: EventSetFlags, Flags: 0, FlagMask: flag}
}

func setOut(out, f uint) uint {
	return out | f
}

func clearOut(out, f uint) uint {
	return out & (^f)
}
