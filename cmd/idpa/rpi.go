//go:build !linux

// this file is required to build on platforms where rpio is not supported
// it just provides setupRPI and closeRPI but does not do anything
package main

import (
	"os"

	"github.com/philip-s/idpa"
)

func setupRPI() (idpa.PiOutput, error) {
	return nil, os.ErrNotExist
}

func closeRPI() {
	// nop
}
