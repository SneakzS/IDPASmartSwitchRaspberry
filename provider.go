package idpa

import (
	"database/sql"
	"time"
)

func handleProviderServer(p wsProviderHandler, conn *sql.DB) error {
	var msg ProviderMessage

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = p.Receive(&msg, 0)
	if err != nil {
		return err
	}

	if msg.MessageTypeID != MsgRequest {
		return ErrInvalidMessage
	}

	wires, err := GetCustomerWires(tx, msg.CustomerID)
	if err != nil {
		return err
	}

	workloadPossible := false

	d := WorkloadDefinition{
		WorkloadW:          msg.WorkloadW,
		DurationM:          msg.DurationM,
		ToleranceDurationM: msg.ToleranceDurationM,
		IsEnabled:          true,
	}

	startTime := msg.StartTime
	offsetM, err := GetOptimalWorkloadOffset(tx, wires, d, startTime)
	if err == ErrWorkloadNotPossible {
		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgOffer,
			Offers:        nil,
		}
	} else if err != nil {
		return err
	} else {
		workloadPossible = true
		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgOffer,
			Offers: []Offer{{
				OfferID:   1,
				OffsetM:   offsetM,
				WorkloadW: d.WorkloadW,
				PriceCNT:  0,
			}},
		}
	}

	err = p.Send(&msg)
	if err != nil {
		return err
	}

	err = p.Receive(&msg, msg.RequestID)
	if err != nil {
		return err
	}

	switch msg.MessageTypeID {
	case MsgDiscard:
		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgAck,
		}

	case MsgSelect:
		if msg.OfferID != 1 || !workloadPossible {
			return ErrInvalidMessage
		}

		for _, wire := range wires {
			err = AddWireWorkload(tx, wire.WireID, startTime.Add(time.Duration(offsetM)*time.Minute), d.DurationM, d.WorkloadW)
			if err != nil {
				return err
			}
		}

		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgAck,
			OfferID:       msg.OfferID,
		}

	default:
		return ErrInvalidMessage
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = p.Send(&msg)
	return err
}

func handleProviderClient(q *Offer, p wsProviderHandler, d WorkloadDefinition, matchTime time.Time, customerID int32) error {
	var (
		msg           ProviderMessage
		offerReceived bool
	)

	startTime := matchTime

	msg = ProviderMessage{
		RequestID:          1,
		MessageTypeID:      MsgRequest,
		CustomerID:         customerID,
		DurationM:          d.DurationM,
		ToleranceDurationM: d.ToleranceDurationM,
		WorkloadW:          d.WorkloadW,
		StartTime:          startTime,
	}

	err := p.Send(&msg)
	if err != nil {
		return err
	}

	err = p.Receive(&msg, msg.RequestID)
	if err != nil {
		return err
	}
	if msg.MessageTypeID != MsgOffer {
		return ErrInvalidMessage
	}

	if len(msg.Offers) == 0 {
		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgDiscard,
		}
	} else {
		*q = msg.Offers[0]
		offerReceived = true

		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgSelect,
			OfferID:       msg.Offers[0].OfferID,
		}
	}

	err = p.Send(&msg)
	if err != nil {
		return err
	}

	err = p.Receive(&msg, msg.RequestID)
	if err != nil {
		return err
	}
	if msg.MessageTypeID != MsgAck {
		return ErrInvalidMessage
	}

	if !offerReceived {
		return ErrWorkloadNotPossible
	}
	return nil
}
