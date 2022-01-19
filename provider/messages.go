package provider

import "time"

type WorkloadRequest struct {
	CustomerID         int32     `json:"customerId"`
	DurationM          int32     `json:"durationM"`
	ToleranceDurationM int32     `json:"toleranceDurationM"`
	WorkloadW          int32     `json:"workloadW"`
	StartTime          time.Time `json:"startTime"`
	Signature          string    `json:"signature"`
}

type WorkloadResponse struct {
	OffsetM  int32 `json:"offsetM"`
	PriceCNT int32 `json:"priceCnt"`
}

type ErrorResponse struct {
	Message string
}
