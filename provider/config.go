package provider

import "github.com/drbig/simpleini"

type Config struct {
	Database string
	Listen   string
}

const ConfigSection = "Provider"

func ReadConfig(c *Config, ini *simpleini.INI) (err error) {
	c.Database, err = ini.GetString(ConfigSection, "database")
	if err != nil {
		return
	}
	c.Listen, err = ini.GetString(ConfigSection, "listen")
	return
}
