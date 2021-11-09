package idpa

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type Server struct {
	db *sql.DB
}

func (s *Server) HandleWS(conn *websocket.Conn) error {
	var (
		request MsgRequest
	)

	_, msgData, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	err = json.Unmarshal(msgData, &request)
	if err != nil {
		return err
	}

	wires, err := GetCustomerWires(s.db, request.CustomerID)
	if err != nil {
		return err
	}

	wl, err := getOptimalWorkload(s.db, wires, request.DurationM, request.AmountW, request.StartTime, request.EndTime)
	if err != nil {
		return err
	}

	// use wl
	fmt.Println(wl)

	return nil
}
