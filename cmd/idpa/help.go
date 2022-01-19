package main

import (
	"fmt"
	"os"
)

const helpText = `IDPA Client Software Help

`

func printHelp() {
	fmt.Fprintln(os.Stderr, helpText)
}
