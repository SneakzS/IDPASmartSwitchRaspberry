package idpa

import (
	"encoding/json"
	"os"
)

type Config struct {
	DatabaseFileName string `json:"databaseFileName"`
	UIServerAddress  string `json:"uiServerAddress"`
}

var defaultConfig = Config{
	DatabaseFileName: "idpa.sqlite3",
}

func DefaultConfig() Config {
	return defaultConfig
}

func ReadConfig(c *Config, fileName string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}
