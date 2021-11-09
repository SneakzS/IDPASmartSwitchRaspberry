package idpa

import (
	"encoding/json"
	"errors"
	"time"
)

type MessageType int32

const (
	MsgTypeRequest MessageType = iota
	MsgTypeOffer
	MsgTypeOfferSelectAck
	MsgTypeServerAck
)

type MsgRequest struct {
	RequestID  int32     `json:"requestID"`
	CustomerID int32     `json:"customerID"`
	DurationM  int32     `json:"durationM"`
	AmountW    int32     `json:"amountW"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

type MsgOffer struct {
	RequestID    int32     `json:"requestID"`
	OfferID      int32     `json:"offerID"`
	OffersAmount int32     `json:"offersAmount"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	PriceCNT     int32     `json:"priceCNT"`
}

type MsgOfferSelectAck struct {
	RequestID int32 `json:"requestID"`
	OfferID   int32 `json:"offerID"`
	PriceCNT  int32 `json:"priceCNT"`
}

type MsgServerAck struct {
	RequestID int32 `json:"requestID"`
	OfferID   int32 `json:"offerID"`
}

type Message struct {
	Type MessageType
	Data json.RawMessage
}

var (
	ErrInvalidMessage      = errors.New("invalid message")
	ErrWorkloadNotPossible = errors.New("workload is not possible")
)

func ParseMessage(buf []byte) (typ MessageType, msg []byte, err error) {
	var m Message
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return
	}

	typ = m.Type
	msg = []byte(m.Data)
	return
}
