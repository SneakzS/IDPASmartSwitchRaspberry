package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/philip-s/idpa"
)

var (
	configFile = flag.String("config", "", "Specify the config file")
	mockPi     = flag.Bool("mock-pi", false, "Set to true to mock the raspberry pi")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Read configuration
	config := idpa.DefaultConfig()
	err := idpa.ReadConfig(&config, *configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Setup GPIO and context
	ctx, cancel := context.WithCancel(context.Background())
	state := idpa.PiState{}

	var output idpa.PiOutput
	if *mockPi {
		output = &consolePi{}
	} else {

		output, err = setupRPI()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	go idpa.RunUIClient(ctx, &state)
	go idpa.RunPi(ctx, output, &state)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	<-done
	cancel()
}
