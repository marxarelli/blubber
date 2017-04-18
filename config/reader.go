package config

import (
	"io/ioutil"
	"encoding/json"
)

func ReadConfig(data []byte) (*ConfigType, error) {
	var config ConfigType

	err := json.Unmarshal(data, &config)

	return &config, err
}

func ReadConfigFile(path string) (*ConfigType, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	} else {
		return ReadConfig(data)
	}
}
