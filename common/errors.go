package common

import "errors"

var (
	ErrWorkloadNotPossible = errors.New("workload is not possible")
	ErrNoWires             = errors.New("customer has no wires")
)
