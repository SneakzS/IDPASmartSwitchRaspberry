package provider

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/philip-s/idpa/common"
)

func HandleWorkloadRequest(q *common.WorkloadResponse, req *common.WorkloadRequest, db *sql.DB) error {
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

	q.OffsetM = offsetM

	return tx.Commit()
}

func WorkloadRequestHandler(wr http.ResponseWriter, r *http.Request, db *sql.DB, mux *sync.Mutex) {
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

		req := common.WorkloadRequest{}

		err = json.Unmarshal(body, &req)
		if err != nil {
			log.Println(err)
			errStatus = http.StatusBadRequest
			errMessage = "Bad Request"
			goto sendError
		}

		mux.Lock()
		defer mux.Unlock()

		resp := common.WorkloadResponse{}
		err = HandleWorkloadRequest(&resp, &req, db)
		if err != nil {
			if err == common.ErrWorkloadNotPossible {
				errStatus = http.StatusConflict
				errMessage = "Workload is not possible"
			} else {
				log.Println(err)
			}

			goto sendError
		}

		// Write the response
		data, _ := json.Marshal(&resp)
		wr.WriteHeader(http.StatusOK)
		wr.Write(data)
		return

	default:
		errStatus = http.StatusMethodNotAllowed
		errMessage = "Method not allowed"
		goto sendError
	}

sendError:
	resp := common.ErrorResponse{Message: errMessage}
	data, _ := json.Marshal(&resp)
	wr.WriteHeader(errStatus)
	wr.Write(data)
}
