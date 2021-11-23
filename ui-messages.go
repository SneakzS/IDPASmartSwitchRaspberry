package idpa

const (
	ActionSetWorkload    = 1
	ActionGetWorkload    = 2
	ActionDeleteWorkload = 3
	ActionSetFlags       = 4
	ActionGetFlags       = 5
)

const (
	FlagEnforce            = 1 << 0
	FlagIsEnabled          = 1 << 1
	FlagHasConnectionError = 1 << 2
)

type WorkloadDefinition struct {
	WorkloadPlanID     int32         `json:"workloadPlanId"`
	WorkloadW          int32         `json:"workloadW"`
	DurationM          int32         `json:"duratonM"`
	ToleranceDurationM int32         `json:"toleranceDurationM"`
	RepeatPattern      RepeatPattern `json:"repeatPattern"`
	IsEnabled          bool          `json:"isEnabled"`
}

type UIMessage struct {
	ActionID           int32 `json:"actionId"`
	WorkloadDefinition `json:"workloadDefinition,omitempty"`
	Flags              uint64 `json:"flags,omitempty"`
	FlagMask           uint64 `json:"flagMask,omitempty"`
}
