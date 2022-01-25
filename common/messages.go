package common

import "time"

type WorkloadRequest struct {
	CustomerID         int32     `json:"customerId"`
	DurationM          int32     `json:"durationM"`
	ToleranceDurationM int32     `json:"toleranceDurationM"`
	WorkloadW          int32     `json:"workloadW"`
	StartTime          time.Time `json:"startTime"`
}

type WorkloadResponse struct {
	OffsetM  int32 `json:"offsetM"`
	PriceCNT int32 `json:"priceCnt"`
}

type ErrorResponse struct {
	Message string
}

const (
	ActionSetWorkloadDefinition    = 1
	ActionGetWorkloadDefinitions   = 2
	ActionDeleteWorkloadDefinition = 3
	ActionSetFlags                 = 4
	ActionGetFlags                 = 5
	ActionHelo                     = 6
	ActionGetWorkloads             = 7
	ActionGetSensorSamples         = 8
	ActionGetLogEntries            = 9

	// Response Actions
	ActionNotifyError               = 101
	ActionNotifyWorkloadCreated     = 102
	ActionNotifyNoContent           = 103
	ActionNotifyWorkloads           = 104
	ActionNotifyWorkloadDefinitions = 105
	ActionNotifySensorSamples       = 106
	ActionNotifyLogEntries          = 107
)

const (
	FlagEnforce          = 1 << 0
	FlagIsEnabled        = 1 << 1
	FlagIsUIConnected    = 1 << 2
	FlagProviderClientOK = 1 << 3
)

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
	DurationM            int32     `json:"durationM"`
	WorkloadW            int32     `json:"workloadW"`
}

type SensorSample struct {
	SampleTime time.Time `json:"sampleTime"`
	Power      float64   `json:"power"`
	Current    float64   `json:"current"`
	Voltage    float64   `json:"voltage"`
}

type LogEntry struct {
	LogID    int32     `json:"logId"`
	LogTime  time.Time `json:"logTime"`
	Severity int32     `json:"severity"`
	Source   string    `json:"source"`
	Message  string    `json:"message"`
}

type UIMessage struct {
	ActionID                   int32                `json:"actionId"`
	RequestID                  int32                `json:"requestId,omitempty"`
	WorkloadDefinition         WorkloadDefinition   `json:"workloadDefinition,omitempty"`
	Flags                      uint32               `json:"flags,omitempty"`
	FlagMask                   uint32               `json:"flagMask,omitempty"`
	ClientGUID                 string               `json:"clientGuid,omitempty"`
	ActiveWorkloads            []ActiveWorkload     `json:"activeWorkloads,omitempty"`
	CurrentWorkloadDefinitions []WorkloadDefinition `json:"currentWorkloadDefinitions,omitempty"`
	ErrorMessage               string               `json:"errorMessage,omitempty"`
	StartTime                  time.Time            `json:"startTime,omitempty"`
	DurationM                  int32                `json:"durationM,omitempty"`
	SensorSamples              []SensorSample       `json:"sensorSamples,omitempty"`
	LogEntries                 []LogEntry           `json:"logEntries,omitempty"`
}
