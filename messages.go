package idpa

import (
	"errors"
	"time"
)

const (
	_ = iota
	MsgRequest
	MsgOffer
	MsgSelect
	MsgDiscard
	MsgAck
)

const (
	ActionSetWorkloadDefinition    = 1
	ActionGetWorkloadDefinition    = 2
	ActionDeleteWorkloadDefinition = 3
	ActionSetFlags                 = 4
	ActionGetFlags                 = 5
	ActionHelo                     = 6
)

const (
	FlagEnforce          = 1 << 0
	FlagIsEnabled        = 1 << 1
	FlagIsUIConnected    = 1 << 2
	FlagProviderClientOK = 1 << 3
)

var (
	ErrInvalidMessage      = errors.New("invalid message")
	ErrWorkloadNotPossible = errors.New("workload is not possible")
	ErrNoWires             = errors.New("customer has no wires")
)

type Offer struct {
	OfferID   int32
	OffsetM   int32
	WorkloadW int32
	PriceCNT  int32
}

type ProviderMessage struct {
	MessageTypeID      int32
	RequestID          int32
	CustomerID         int32
	OfferID            int32
	DurationM          int32
	ToleranceDurationM int32
	WorkloadW          int32
	StartTime          time.Time
	Offers             []Offer
}

type WorkloadDefinition struct {
	WorkloadDefinitionID int32           `json:"workloadDefinitionId"`
	Description          string          `json:"description"`
	WorkloadW            int32           `json:"workloadW"`
	DurationM            int32           `json:"durationM"`
	ToleranceDurationM   int32           `json:"toleranceDurationM"`
	RepeatPattern        []RepeatPattern `json:"repeatPattern"`
	IsEnabled            bool            `json:"isEnabled"`
	ExpiryDate           time.Time       `json:"expiryDate"`
}

type ActiveWorkload struct {
	WorkloadDefinitionID int32     `json:"workloadDefinitionId"`
	OffsetM              int32     `json:"offsetM"`
	StartTime            time.Time `json:"startTime"`
}

type UIMessage struct {
	ActionID           int32              `json:"actionId"`
	WorkloadDefinition WorkloadDefinition `json:"workloadDefinition,omitempty"`
	Flags              uint32             `json:"flags,omitempty"`
	FlagMask           uint32             `json:"flagMask,omitempty"`
	ClientGUID         string             `json:"clientGuid"`
	ActiveWorkloads    []ActiveWorkload   `json:"activeWorkloads"`
}
