package idpa

import (
	"time"
)

func handleProviderClient(q *Offer, p ProviderConnection, d WorkloadDefinition, matchTime time.Time, customerID int32) error {
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
