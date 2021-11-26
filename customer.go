package idpa

import "database/sql"

type Customer struct {
	CustomerID int32
	FirstName  string
	LastName   string
	Address    string
	Town       string
}

func GetCustomers(tx *sql.Tx) ([]Customer, error) {
	res, err := tx.Query("SELECT customerID, firstName, lastName, address, town FROM Customer")
	if err != nil {
		return nil, err
	}

	var customers []Customer
	for res.Next() {
		c := Customer{}
		err = res.Scan(&c.CustomerID, &c.FirstName, &c.LastName, &c.Address, &c.Town)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	return customers, nil
}
