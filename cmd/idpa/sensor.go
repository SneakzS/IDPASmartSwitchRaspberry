package main

import (
	"log"
	"time"

	"github.com/philip-s/idpa/client"
)

func runSensor(outChan chan<- client.Input, c *client.Config, done <-chan struct{}) {

	if c.Output != client.OutputRpi && !*mockSensorData {
		log.Println("warning: Sensor is not available")
		return
	}

	t := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-done:
			return

		case now := <-t.C:
			sampleTime := now.UTC().Truncate(time.Second)
			data := client.Input{
				SensorSampleTime: sampleTime,
			}
			err := error(nil)

			switch c.Output {
			case client.OutputRpi:
				err = readInputRPI(&data)
			case client.OutputConsole:
				err = readInputConsole(&data)

			}

			if err != nil {
				log.Println(err)
				continue
			}
			outChan <- data
		}
	}
}
