package client

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/philip-s/idpa/common"
)

type uiClientState struct {
	IsOK          bool
	EnforceOutput bool
	EnableOutput  bool
}

func runUIClient(stateChan chan<- uiClientState, sqlConn *sql.DB, c *Config, done <-chan struct{}) {
	var (
		currentState uiClientState
		messages     = make(chan common.UIMessage, 1)
		isOkChan     = make(chan bool, 1)
	)

	go fetchMessages(isOkChan, messages, c, done)

	for {
		newState := currentState

		select {
		case <-done:
			return

		case isOk := <-isOkChan:
			newState.IsOK = isOk

		case msg := <-messages:
			err := handleUIMessage(&newState, &msg, sqlConn)
			if err != nil {
				log.Println(err)
				newState.IsOK = false
			} else {
				newState.IsOK = true
			}

		}

		if newState != currentState {
			stateChan <- newState
			currentState = newState
		}
	}
}

func fetchMessages(isOkChan chan<- bool, messages chan<- common.UIMessage, c *Config, done <-chan struct{}) {
	var closing bool
	var conn *websocket.Conn
	var err error
	var dialer websocket.Dialer

	go func() {
		<-done
		closing = true

		if conn != nil {
			conn.WriteMessage(websocket.CloseMessage, nil)
		}

	}()

	for {
		conn, _, err = dialer.Dial(c.ServerURL, nil)
		if err != nil {
			goto handleError
		}

		err = sendHeloMessage(conn, c.ClientGUID)
		if err != nil {
			goto handleError
		}

		log.Println("connected to " + c.ServerURL)
		isOkChan <- true

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
				var parsedMessage common.UIMessage
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
		isOkChan <- false
		time.Sleep(5 * time.Second)
	}
}

func sendHeloMessage(c *websocket.Conn, clientGUID string) error {
	msg := common.UIMessage{
		ActionID:   common.ActionHelo,
		ClientGUID: clientGUID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	//log.Println("Hello Message: ", string(data))
	return c.WriteMessage(websocket.TextMessage, data)
}

func handleUIMessage(state *uiClientState, msg *common.UIMessage, conn *sql.DB) error {
	switch msg.ActionID {
	case common.ActionSetFlags:
		if msg.FlagMask&common.FlagIsEnabled > 0 {
			state.EnableOutput = msg.Flags&common.FlagIsEnabled > 0
		}
		if msg.FlagMask&common.FlagEnforce > 0 {
			state.EnforceOutput = msg.Flags&common.FlagEnforce > 0
		}

	case common.ActionSetWorkloadDefinition:
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

	case common.ActionDeleteWorkloadDefinition:
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
