package idpa

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type UIConfig struct {
	ServerURL  string
	ClientGUID string
}

func RunUIClient(ctx context.Context, events chan<- PiEvent, c UIConfig) {
	dialer := websocket.Dialer{}

	for {
		conn, _, err := dialer.DialContext(ctx, c.ServerURL, nil)
		if err != nil {
			log.Println(err)
			goto handleError
		}

		err = sendHeloMessage(conn, c.ClientGUID)
		if err != nil {
			goto handleError
		}

		log.Println("connected to " + c.ServerURL)
		events <- setFlag(FlagIsUIConnected)

	recevie:
		for {

			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if err == context.Canceled {
					return
				}

				log.Println(err)
				goto handleError
			}

			switch msgType {
			case websocket.BinaryMessage:

			case websocket.TextMessage:
				var parsedMessage UIMessage
				err := json.Unmarshal(msg, &parsedMessage)
				if err != nil {
					log.Println("Invalid message received", err)
					continue recevie
				}

				err = handleUIMessage(events, &parsedMessage)
				if err != nil {
					log.Println(err)
					continue recevie
				}

			case websocket.CloseMessage:
				log.Println("connection closed")
				goto handleError
			}
		}

	handleError:
		if conn != nil {
			conn.Close()
		}
		events <- clearFlag(FlagIsUIConnected)
		time.Sleep(5 * time.Second)
	}

}

func sendHeloMessage(c *websocket.Conn, clientGUID string) error {
	msg := UIMessage{
		ActionID:   ActionHelo,
		ClientGUID: clientGUID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, data)
}

func handleUIMessage(events chan<- PiEvent, msg *UIMessage) error {
	switch msg.ActionID {
	case ActionSetFlags:
		events <- PiEvent{EventID: EventSetFlags, Flags: msg.Flags, FlagMask: msg.FlagMask}
	}

	return nil
}
