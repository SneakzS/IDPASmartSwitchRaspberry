package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/drbig/simpleini"
	"github.com/philip-s/idpa"
	"github.com/philip-s/idpa/serverui"
)

func runService() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	mode, err := getServiceMode(cfg)
	if err != nil {
		return err
	}

	switch mode {
	case "client":
		return runClient(cfg)
	case "server":
		return runServer(cfg)
	default:
		return fmt.Errorf("invalid service mode %s", mode)
	}
}

type piOutput interface {
	write(uint32)
}

func runClient(cfg *simpleini.INI) error {
	outputStr, err := cfg.GetString("Client", "output")
	if err != nil {
		return err
	}

	var output piOutput

	switch outputStr {
	case "rpi":
		output, err = setupRPI()
		if err != nil {
			return err
		}
		defer closeRPI()

	case "console":
		output = &consolePi{}

	default:
		return fmt.Errorf("invalid output %s", outputStr)
	}

	uiServerURL, err := cfg.GetString("Client", "uiserverurl")
	if err != nil {
		return err
	}
	uiClientGUID, err := cfg.GetString("Client", "uiclientguid")
	if err != nil {
		return err
	}
	if uiClientGUID == "" {
		return fmt.Errorf("uiclientguid is required")
	}

	providerServerURL, err := cfg.GetString("Client", "providerserverurl")
	if err != nil {
		return err
	}
	database, err := cfg.GetString("Client", "database")
	if err != nil {
		return err
	}
	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return err
	}
	defer conn.Close()

	customerID, err := cfg.GetInt("Client", "customerid")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outChan := make(chan uint32)

	pi := idpa.NewPI(outChan)

	go idpa.RunUIClient(ctx, pi, conn, idpa.UIConfig{
		ServerURL:  uiServerURL,
		ClientGUID: uiClientGUID,
	})
	go idpa.RunProviderClient(ctx, pi, conn, providerServerURL, int32(customerID))

	// Wait for SIGINT to quit
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	for {
		select {
		case q := <-outChan:
			output.write(q)
		case <-done:
			return nil
		}
	}
}

func runServer(cfg *simpleini.INI) error {
	database, err := cfg.GetString("Server", "database")
	if err != nil {
		return err
	}

	listen, err := cfg.GetString("Server", "listen")
	if err != nil {
		return err
	}

	conn, err := sql.Open("sqlite3", database)
	if err != nil {
		return err
	}
	defer conn.Close()

	handler := serverui.GetHTTPHandler(conn)
	http.Handle("/", handler)
	return http.ListenAndServe(listen, nil)
}
