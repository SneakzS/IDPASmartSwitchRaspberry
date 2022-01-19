package client

import (
	"fmt"

	"github.com/drbig/simpleini"
)

const (
	OutputConsole = 0
	OutputRpi     = 1
)

type Config struct {
	CustomerID  int
	Database    string
	Output      int
	ProviderURL string
	ClientGUID  string
	ServerURL   string
}

const ConfigSection = "Client"

func ReadConfig(c *Config, ini *simpleini.INI) (err error) {
	c.CustomerID, err = ini.GetInt(ConfigSection, "customerid")
	if err != nil {
		return
	}
	c.Database, err = ini.GetString(ConfigSection, "database")
	if err != nil {
		return
	}
	output, err := ini.GetString(ConfigSection, "output")
	if err != nil {
		return
	}
	switch output {
	case "console":
		c.Output = OutputConsole
	case "rpi":
		c.Output = OutputRpi

	default:
		err = fmt.Errorf("output must be console or rpi not %s", output)
		return
	}
	c.ProviderURL, err = ini.GetString(ConfigSection, "providerurl")
	if err != nil {
		return
	}
	c.ClientGUID, err = ini.GetString(ConfigSection, "clientguid")
	if err != nil {
		return
	}
	c.ServerURL, err = ini.GetString(ConfigSection, "serverurl")
	return

}
