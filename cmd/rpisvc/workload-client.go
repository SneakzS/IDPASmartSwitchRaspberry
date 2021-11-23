package main

type configuration struct {
	serverAddress string
	customerID    int32
}

/*
func planWorkload(c configuration, d workloadDefinition, startTime time.Time, requestID int32) error {
	endTime := startTime.Add(time.Duration(d.toleranceDurationM) * time.Minute)

	request := idpa.MsgRequest{
		RequestID:  requestID,
		CustomerID: c.customerID,
		DurationM:  d.durationM,
		AmountW:    d.workloadW,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	conn, r, err := websocket.DefaultDialer.Dial(c.serverAddress, nil)
	if err != nil {
		return err
	}
}

func planNextWorkload(tx *sql.Tx, d workloadDefinition)
*/
