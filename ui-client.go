package idpa

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func RunUIClient(ctx context.Context, s *PiState) {
	dialer := websocket.Dialer{}

mainloop:
	for {
		conn, _, err := dialer.DialContext(ctx, s.c.UIServerAddress, nil)
		if err != nil {
			log.Println(err)
			goto setError
		}

		log.Println("connected to " + s.c.UIServerAddress)

	recevie:
		for {

			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if err == context.Canceled {
					return
				}

				log.Println(err)
				goto setError
			}

			// clear error flag
			s.SetFlags(0, FlagHasConnectionError)

			switch msgType {
			case websocket.BinaryMessage:

			case websocket.TextMessage:
				var parsedMessage UIMessage
				err := json.Unmarshal(msg, &parsedMessage)
				if err != nil {
					log.Println("Invalid message received", err)
					continue recevie
				}

				err = handleUIMessage(s, &parsedMessage)
				if err != nil {
					log.Println(err)
					continue recevie
				}

			case websocket.CloseMessage:
				log.Println("connection closed")
				goto setError
			}
		}
	}

setError:
	s.SetFlags(FlagHasConnectionError, FlagHasConnectionError)
	time.Sleep(5 * time.Second)
	goto mainloop

}

func handleUIMessage(s *PiState, msg *UIMessage) error {
	switch msg.ActionID {
	case ActionSetFlags:
		s.SetFlags(msg.Flags, msg.FlagMask)
	}

	return nil
}
