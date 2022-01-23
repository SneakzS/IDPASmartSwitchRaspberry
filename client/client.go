package client

import (
	"database/sql"
	"log"
	"sort"
	"time"
)

type Output struct {
	Relais bool
	Led1   bool
	Led2   bool
	Led3   bool
}

type Input struct {
	SensorSampleTime time.Time
	Power            float64
	Current          float64
	Voltage          float64
}

func Run(outChan chan<- Output, inChan <-chan Input, c *Config, done <-chan struct{}) error {

	var (
		t               = time.NewTicker(200 * time.Millisecond) // create a blinking signal to blink leds in case of error
		currentOutput   Output
		db              *sql.DB
		isOutputEnabled bool

		uiState     uiClientState
		uiStateChan = make(chan uiClientState, 1)

		providerState     providerClientState
		providerStateChan = make(chan providerClientState, 1)

		sensorSamples []Input
	)

	db, err := sql.Open("sqlite3", c.Database)
	if err != nil {
		return err
	}

	go runProviderClient(providerStateChan, db, c, done)
	go runUIClient(uiStateChan, db, c, done)

	for {
		// update the output based on our flags
		newOutput := currentOutput

		select {
		case <-done:
			return nil

		case <-t.C:
			// blink led2 if the ui has an error
			if !uiState.IsOK {
				newOutput.Led2 = !newOutput.Led2
			} else {
				newOutput.Led2 = false
			}

			// blink led3 if the provider failed
			if !providerState.IsOK || providerState.HasWarning {
				newOutput.Led3 = !newOutput.Led3
			} else {
				newOutput.Led3 = false
			}

		case uiState = <-uiStateChan:
		case providerState = <-providerStateChan:

		case input := <-inChan:
			if len(sensorSamples) > 0 && input.SensorSampleTime.Sub(sensorSamples[0].SensorSampleTime) >= time.Minute {
				s := make([]float64, len(sensorSamples))
				for i, inp := range sensorSamples {
					s[i] = inp.Power
				}
				power := getMedianFloat(s)

				for i, inp := range sensorSamples {
					s[i] = inp.Current
				}
				current := getMedianFloat(s)

				for i, inp := range sensorSamples {
					s[i] = inp.Voltage
				}
				voltage := getMedianFloat(s)

				tx, err := db.Begin()
				if err != nil {
					return err
				}
				defer tx.Rollback()

				err = StoreSensorData(tx, input.SensorSampleTime, power, current, voltage)
				if err != nil {
					log.Println(err)
				} else {
					sensorSamples = nil
				}

			}
			sensorSamples = append(sensorSamples, input)

		}

		if uiState.IsOK && uiState.EnforceOutput {
			isOutputEnabled = uiState.EnableOutput
		} else if providerState.IsOK {
			isOutputEnabled = providerState.EnableOutput
		} else {
			isOutputEnabled = false
		}

		if isOutputEnabled {
			newOutput.Led1 = true
			newOutput.Relais = true
		} else {
			newOutput.Led1 = false
			newOutput.Relais = false
		}

		if newOutput != currentOutput {
			outChan <- newOutput
			currentOutput = newOutput
		}

	}

}

func getMedianFloat(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}

	sort.Float64s(s)

	if len(s)%2 == 0 {
		high := len(s) / 2
		low := high - 1
		return (s[low] + s[high]) / 2
	}

	return s[len(s)/2]
}
