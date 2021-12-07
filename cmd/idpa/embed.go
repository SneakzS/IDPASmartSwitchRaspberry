package main

import _ "embed"

var (
	//go:embed sql/schema-client.sql
	createClientDBScript string

	//go:embed sql/schema-server.sql
	createServerDBScript string
)
