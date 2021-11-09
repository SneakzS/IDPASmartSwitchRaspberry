package idpa

import "database/sql"

type dbWire struct {
	wireId    int32
	capacityW int32
}

func GetCustomerWires(conn *sql.DB, customerID int32) ([]dbWire, error) {
	res, err := conn.Query(
		`SELECT wireID, capacityW FROM Wire as w 
		LEFT JOIN WireCustomer as wc 
		ON w.wireID = wc.wireID WHERE customerID = %s`, customerID,
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var wires []dbWire
	for res.Next() {
		var wire dbWire
		err = res.Scan(&wire.wireId, &wire.capacityW)
		if err != nil {
			return nil, err
		}
		wires = append(wires, wire)
	}

	return wires, nil
}
