package idpa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type UIConfig struct {
	ServerURL  string
	ClientGUID string
}

func RunUIClient(ctx context.Context, events chan<- PiEvent, sqlConn *sql.DB, c UIConfig) {
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
				fmt.Println("ReadMessage returned, error is ", err)
				var errClose *websocket.CloseError
				if errors.As(err, &errClose) {
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

				err = handleUIMessage(events, &parsedMessage, sqlConn)
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

func handleUIMessage(events chan<- PiEvent, msg *UIMessage, conn *sql.DB) error {
	switch msg.ActionID {
	case ActionSetFlags:
		events <- PiEvent{EventID: EventSetFlags, Flags: msg.Flags, FlagMask: msg.FlagMask}
	case ActionSetWorkloadDefinition:
		def := msg.WorkloadDefinition
		tx, err := conn.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if def.WorkloadDefinitionID == 0 {
			_, err := CreateWorkloadDefinition(tx, def)
			if err != nil {
				return err
			}
		} else {
			err := UpdateWorkloadDefinition(tx, def)
			if err != nil {
				return err
			}
		}

		return tx.Commit()

	case ActionDeleteWorkloadDefinition:
		tx, err := conn.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		err = DeleteWorkloadDefinition(tx, msg.WorkloadDefinition.WorkloadDefinitionID)
		if err != nil {
			return err
		}
		return tx.Commit()
	}

	return nil
}
