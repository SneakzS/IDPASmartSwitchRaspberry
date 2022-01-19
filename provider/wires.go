package provider

import "database/sql"

type Wire struct {
	WireID    int32
	CapacityW int32
}

func GetCustomerWires(tx *sql.Tx, customerID int32) ([]Wire, error) {
	res, err := tx.Query(
		`SELECT w.wireID, w.capacityW FROM Wire as w 
		LEFT JOIN WireCustomer as wc 
		ON w.wireID = wc.wireID WHERE customerID = ?`, customerID,
	)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var wires []Wire
	for res.Next() {
		var wire Wire
		err = res.Scan(&wire.WireID, &wire.CapacityW)
		if err != nil {
			return nil, err
		}
		wires = append(wires, wire)
	}

	return wires, nil
}
