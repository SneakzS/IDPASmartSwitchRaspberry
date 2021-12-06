package serverui

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/philip-s/idpa"
)

type customerView struct {
	baseView
	Customers []idpa.Customer
}

var (
	customersTemplate = compileTemplate("layout.html", "customers.html")
)

func getCustomerRoutes(r *httprouter.Router, conn *sql.DB) {
	r.GET("/customers", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			err        error
			statusCode = 500
			errText    = "Internal Server Error"
		)

		goto begin

	sendError:
		log.Println(err)
		http.Error(w, errText, statusCode)
		return

	begin:

		tx, err := conn.Begin()
		if err != nil {
			goto sendError
		}
		defer tx.Rollback()

		customers, err := idpa.GetCustomers(tx)
		if err != nil {
			goto sendError
		}

		view := customerView{
			Customers: customers,
		}

		err = customersTemplate.Execute(w, view)
		if err != nil {
			log.Println(err)
		}
	})

	r.GET("/customers/:id", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			err        error
			statusCode = 500
			errText    = "Internal Server Error"
		)

		goto begin

	sendError:
		log.Println(err)
		http.Error(rw, errText, statusCode)
		return

	begin:
		id, err := strconv.ParseInt(p.ByName("id"), 10, 32)
		if err != nil {
			goto sendError
		}

		fmt.Println(id)
	})
}
