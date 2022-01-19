package idpa

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

func ReceiveProviderMessage(c *websocket.Conn, q *ProviderMessage) error {
	msgType, msg, err := c.ReadMessage()
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

	return nil
}

func SendProviderMessage(c *websocket.Conn, m *ProviderMessage) error {
	msg, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, msg)
}
