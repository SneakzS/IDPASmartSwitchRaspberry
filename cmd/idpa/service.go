package main

import (
	"os"
	"os/signal"

	"github.com/philip-s/idpa/client"
	"github.com/philip-s/idpa/common"
)

func runService() error {
	ini, err := common.ParseINIFile(*config)
	if err != nil {
		return err
	}

	c := client.Config{}
	err = client.ReadConfig(&c, ini)
	if err != nil {
		return err
	}

	var (
		clientInputChan  = make(chan client.Input, 1)
		clientOutputChan = make(chan client.Output, 1)
		doneChan         = make(chan struct{})
		doneSignalChan   = make(chan os.Signal, 1)
	)

	go runSensor(clientInputChan, &c, doneChan)
	go client.Run(clientOutputChan, clientInputChan, &c, doneChan)

	// Wait for SIGINT to quit
	signal.Notify(doneSignalChan, os.Interrupt)

	for {
		select {
		case q := <-clientOutputChan:
			switch c.Output {
			case client.OutputConsole:
				writeOutputConsole(q)
			case client.OutputRpi:
				writeOutputRPI(q)
			}
		case <-doneSignalChan:
			return nil
		}
	}
}
