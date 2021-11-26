package idpa

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type wsProviderHandler struct {
	conn *websocket.Conn
}

func (p wsProviderHandler) Receive(q *ProviderMessage, requestID int32) error {
	msgType, msg, err := p.conn.ReadMessage()
	if err != nil {
		return err
	}
	if msgType != websocket.TextMessage {
		return ErrInvalidMessage
	}

	err = json.Unmarshal(msg, q)
	if err != nil {
		return err
	}

	if q.RequestID == 0 {
		return ErrInvalidMessage
	}

	if requestID != 0 && q.RequestID != requestID {
		return ErrInvalidMessage
	}

	return nil
}

func (p wsProviderHandler) Send(m *ProviderMessage) error {
	msg, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return p.conn.WriteMessage(websocket.TextMessage, msg)
}
