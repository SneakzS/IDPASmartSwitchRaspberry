package serverui

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/philip-s/idpa"
)

type customerController struct {
	conn *sql.DB
}

type customerView struct {
	baseView
	Customers []idpa.Customer
}

var customersTemplate = compileTemplate("layout.html", "customers.html")

func (c customerController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	tx, err := c.conn.Begin()
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
}
