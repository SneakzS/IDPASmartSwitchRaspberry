//go:build !linux

// this file is required to build on platforms where rpio is not supported
// it just provides setupRPI and closeRPI but does not do anything
package main

import (
	"os"

	"github.com/philip-s/idpa/client"
)

func writeOutputRPI(o client.Output) {

}

func readInputRPI(inp *client.Input) error {
	return os.ErrNotExist
}

func setupRPI() error {
	return os.ErrNotExist
}

func closeRPI() {
	// nop
}
