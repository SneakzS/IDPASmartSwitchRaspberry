package main

import (
	"log"
	"time"

	"github.com/philip-s/idpa/client"
)

func runSensor(outChan chan<- client.Input, c *client.Config, done <-chan struct{}) {
	t := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-done:
			return

		case now := <-t.C:
			// only use the sensor if we use the raspberry pi
			if c.Output == client.OutputRpi {
				sampleTime := now.UTC().Truncate(time.Second)
				data := client.Input{
					SensorSampleTime: sampleTime,
				}
				err := readInputRPI(&data)
				if err != nil {
					log.Println(err)
					continue
				}

				outChan <- data
			}
		}
	}
}
