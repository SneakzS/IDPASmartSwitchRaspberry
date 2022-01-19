package provider

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/philip-s/idpa"
)

func HandleWorkloadRequest(q *WorkloadResponse, req *WorkloadRequest, db *sql.DB) error {
	//TODO: Verify signature

	// begin a db transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// get all wires for the customer
	wires, err := GetCustomerWires(tx, req.CustomerID)
	if err != nil {
		return err
	}

	// check if the workload is possible and at which offset
	offsetM, err := GetOptimalWorkloadOffset(tx, wires, req.DurationM, req.ToleranceDurationM, req.WorkloadW, req.StartTime)
	if err != nil {
		return err
	}

	// insert the workload into the database
	for _, wire := range wires {
		err = AddWireWorkload(tx, wire.WireID, req.StartTime.Add(time.Duration(offsetM)*time.Minute), req.DurationM, req.WorkloadW)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunServer(m *http.ServeMux, db *sql.DB) error {
	mux := sync.Mutex{}

	m.HandleFunc("/request", func(wr http.ResponseWriter, r *http.Request) {
		errStatus := 500
		errMessage := "Internal Server Error"

		wr.Header().Add("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				goto sendError
			}

			req := WorkloadRequest{}

			err = json.Unmarshal(body, &req)
			if err != nil {
				log.Println(err)
				errStatus = http.StatusBadRequest
				errMessage = "Bad Request"
				goto sendError
			}

			mux.Lock()
			defer mux.Unlock()

			resp := WorkloadResponse{}
			err = HandleWorkloadRequest(&resp, &req, db)
			if err != nil {
				if err == idpa.ErrWorkloadNotPossible {
					errStatus = http.StatusConflict
				}
			}

		default:
			errStatus = http.StatusMethodNotAllowed
			errMessage = "Method not allowed"
			goto sendError
		}

	sendError:
		resp := ErrorResponse{errMessage}
		data, _ := json.Marshal(&resp)
		wr.WriteHeader(errStatus)
		wr.Write(data)
	})
}

func handleProviderServer(c *websocket.Conn, conn *sql.DB) error {

	var (
		state            int
		msg              idpa.ProviderMessage
		requestID        int32
		customerID       int32
		wires            []Wire
		workloadPossible bool
	)

	for {
		err := idpa.ReceiveProviderMessage(c, &msg)
		if err != nil {
			return err
		}

		if state > 0 && msg.RequestID != requestID {
			return idpa.ErrInvalidMessage
		}

		switch state {
		case 0:
			// we expect the initial request
			if msg.MessageTypeID != idpa.MsgRequest {
				return idpa.ErrInvalidMessage
			}

			requestID = msg.RequestID
			customerID = msg.CustomerID

			d := idpa.WorkloadDefinition{
				WorkloadW:          msg.WorkloadW,
				DurationM:          msg.DurationM,
				ToleranceDurationM: msg.ToleranceDurationM,
				IsEnabled:          true,
			}

			startTime := msg.StartTime

			if err == idpa.ErrWorkloadNotPossible {
				msg = idpa.ProviderMessage{
					RequestID:     requestID,
					MessageTypeID: idpa.MsgOffer,
					Offers:        nil,
				}
			} else if err != nil {
				return err
			} else {
				workloadPossible = true
				msg = ProviderMessage{
					RequestID:     msg.RequestID,
					MessageTypeID: MsgOffer,
					Offers: []Offer{{
						OfferID:   1,
						OffsetM:   offsetM,
						WorkloadW: d.WorkloadW,
						PriceCNT:  0,
					}},
				}
			}

		}

	}

	err = p.Send(&msg)
	if err != nil {
		return err
	}

	err = p.Receive(&msg, msg.RequestID)
	if err != nil {
		return err
	}

	switch msg.MessageTypeID {
	case MsgDiscard:
		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgAck,
		}

	case MsgSelect:
		if msg.OfferID != 1 || !workloadPossible {
			return ErrInvalidMessage
		}

		msg = ProviderMessage{
			RequestID:     msg.RequestID,
			MessageTypeID: MsgAck,
			OfferID:       msg.OfferID,
		}

	default:
		return ErrInvalidMessage
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = p.Send(&msg)
	return err
}
