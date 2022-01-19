package idpa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type UIConfig struct {
	ServerURL  string
	ClientGUID string
}

func RunUIClient(ctx context.Context, pi *Pi, sqlConn *sql.DB, c UIConfig) {
	messages := make(chan UIMessage, 1)
	go fetchMessages(ctx, pi, c, messages)

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-messages:
			err := handleUIMessage(pi, &msg, sqlConn)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func fetchMessages(ctx context.Context, pi *Pi, c UIConfig, messages chan<- UIMessage) {
	var closing bool
	var conn *websocket.Conn
	var err error
	var dialer websocket.Dialer

	go func() {
		<-ctx.Done()
		closing = true

		if conn != nil {
			conn.WriteMessage(websocket.CloseMessage, nil)
		}

	}()

	for {
		conn, _, err = dialer.DialContext(ctx, c.ServerURL, nil)
		if err != nil {
			goto handleError
		}

		err = sendHeloMessage(conn, c.ClientGUID)
		if err != nil {
			goto handleError
		}

		log.Println("connected to " + c.ServerURL)
		pi.SetFlags(FlagIsUIConnected, FlagIsUIConnected)

	receive:
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
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
					continue receive
				}

				log.Println(string(msg))
				messages <- parsedMessage
			}
		}

	handleError:
		var errClose *websocket.CloseError
		if closing && errors.As(err, &errClose) {
			return
		}

		if conn != nil {
			conn.Close()
		}
		pi.SetFlags(0, FlagIsUIConnected)
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
	//log.Println("Hello Message: ", string(data))
	return c.WriteMessage(websocket.TextMessage, data)
}

func handleUIMessage(pi *Pi, msg *UIMessage, conn *sql.DB) error {
	switch msg.ActionID {
	case ActionSetFlags:
		pi.SetFlags(msg.Flags, uint32(msg.FlagMask))
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
