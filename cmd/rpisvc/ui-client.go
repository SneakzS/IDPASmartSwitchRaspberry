package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/philip-s/idpa"
)

func handleUIConnection(rpi *raspberryPi, urlStr string) {
	dialer := websocket.Dialer{}

mainloop:
	for {
		conn, _, err := dialer.Dial(urlStr, nil)
		if err != nil {
			log.Println(err)
			goto setError
		}

		log.Println("connected to " + urlStr)

	recevie:
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				goto setError
			}

			// clear error flag
			rpi.setFlags(0, idpa.FlagHasError)

			switch msgType {
			case websocket.BinaryMessage:

			case websocket.TextMessage:
				var parsedMessage idpa.UIMessage
				err := json.Unmarshal(msg, &parsedMessage)
				if err != nil {
					log.Println("Invalid message received", err)
					continue recevie
				}

				err = handleUIMessage(rpi, &parsedMessage)
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
	rpi.setFlags(idpa.FlagHasError, idpa.FlagHasError)
	time.Sleep(5 * time.Second)
	goto mainloop

}

func handleUIMessage(rpi *raspberryPi, msg *idpa.UIMessage) error {
	rpi.mux.Lock()
	defer rpi.mux.Unlock()

	switch msg.ActionID {
	case idpa.ActionSetFlags:
		rpi.setFlagsUnsafe(msg.Flags, msg.FlagMask)
		fmt.Println("processed message with action " + fmt.Sprintf("%d", msg.ActionID))
		fmt.Printf("flags are 0b%b\n", msg.Flags)
	}

	return nil
}
