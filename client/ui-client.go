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
		messagesIN   = make(chan common.UIMessage, 1)
		messagesOUT  = make(chan common.UIMessage, 1)
		isOkChan     = make(chan bool, 1)
	)

	go runUIMessagePump(isOkChan, messagesOUT, messagesIN, c, done)

	for {
		newState := currentState

		select {
		case <-done:
			return

		case isOk := <-isOkChan:
			newState.IsOK = isOk

		case msg := <-messagesIN:
			response := common.UIMessage{}
			hasResponse, err := handleUIMessage(&response, &newState, &msg, sqlConn)
			if err != nil {
				log.Println(err)
				newState.IsOK = false

				if msg.RequestID != 0 && !hasResponse {
					response = common.UIMessage{
						ActionID:     common.ActionNotifyError,
						RequestID:    msg.RequestID,
						ErrorMessage: "internal server error",
					}
					hasResponse = true
				}

			} else {
				newState.IsOK = true

				if msg.RequestID != 0 && !hasResponse {
					response = common.UIMessage{
						ActionID:  common.ActionNotifyNoContent,
						RequestID: msg.RequestID,
					}
					hasResponse = true
				}
			}

			if hasResponse {
				messagesOUT <- response
			}
		}

		if newState != currentState {
			stateChan <- newState
			currentState = newState
		}
	}
}

func runUIMessagePump(isOkChan chan<- bool, messagesOUT <-chan common.UIMessage, messagesIN chan<- common.UIMessage, c *Config, done <-chan struct{}) {
	var closing bool
	var conn *websocket.Conn
	var err error
	var dialer websocket.Dialer

	go func() {

		for {
			select {
			case <-done:
				<-done
				closing = true

				if conn != nil {
					conn.WriteMessage(websocket.CloseMessage, nil)
				}
				return

			case msg := <-messagesOUT:
				data, err := json.Marshal(&msg)
				if err != nil {
					log.Println(err)
					continue
				}
				err = conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Println(err)
					continue
				}
			}
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

				//log.Println(string(msg))
				messagesIN <- parsedMessage
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

func handleUIMessage(response *common.UIMessage, state *uiClientState, msg *common.UIMessage, conn *sql.DB) (hasResponse bool, err error) {
	switch msg.ActionID {
	case common.ActionSetFlags:
		if msg.FlagMask&common.FlagIsEnabled > 0 {
			state.EnableOutput = msg.Flags&common.FlagIsEnabled > 0
		}
		if msg.FlagMask&common.FlagEnforce > 0 {
			state.EnforceOutput = msg.Flags&common.FlagEnforce > 0
		}

	case common.ActionGetWorkloadDefinitions:
		tx, err := conn.Begin()
		if err != nil {
			return false, err
		}
		defer tx.Rollback()
		defs, err := GetWorkloadDefinitions(tx)
		if err != nil {
			return false, err
		}
		*response = common.UIMessage{
			ActionID:          common.ActionNotifyWorkloadDefinitions,
			RequestID:         msg.RequestID,
			CurreontWorkloads: defs,
		}
		return true, nil

	case common.ActionSetWorkloadDefinition:
		def := msg.WorkloadDefinition
		tx, err := conn.Begin()
		if err != nil {
			return false, err
		}
		defer tx.Rollback()

		if def.WorkloadDefinitionID == 0 {
			def.WorkloadDefinitionID, err = CreateWorkloadDefinition(tx, def)
			if err != nil {
				return false, err
			}
		} else {
			err = UpdateWorkloadDefinition(tx, def)
			if err != nil {
				return false, err
			}
		}

		err = tx.Commit()
		if err != nil {
			return false, err
		}
		*response = common.UIMessage{
			ActionID:           common.ActionNotifyWorkloadCreated,
			RequestID:          msg.RequestID,
			WorkloadDefinition: def,
		}
		return true, nil

	case common.ActionDeleteWorkloadDefinition:
		tx, err := conn.Begin()
		if err != nil {
			return false, err
		}
		defer tx.Rollback()

		err = DeleteWorkloadDefinition(tx, msg.WorkloadDefinition.WorkloadDefinitionID)
		if err != nil {
			return false, err
		}
		err = tx.Commit()
		return false, err

	case common.ActionGetWorkloads:
		tx, err := conn.Begin()
		if err != nil {
			return false, err
		}
		defer tx.Rollback()

		workloads, err := GetWorkloads(tx, msg.StartTime, msg.DurationM)
		if err != nil {
			return false, err
		}

		activeWorkloads := make([]common.ActiveWorkload, len(workloads))
		for i, wl := range workloads {
			activeWorkloads[i] = common.ActiveWorkload{
				WorkloadDefinitionID: wl.WorkloadDefinitionID,
				StartTime:            wl.MatchTime,
				OffsetM:              wl.OffsetM,
				DurationM:            wl.DurationM,
				WorkloadW:            wl.WorkloadW,
			}
		}

		*response = common.UIMessage{
			ActionID:        common.ActionNotifyWorkloads,
			RequestID:       msg.RequestID,
			ActiveWorkloads: activeWorkloads,
		}
		return true, nil
	}

	return false, nil
}
