package config

import (
	"io/ioutil"
	"encoding/json"
)

//func ExpandConfig(config *Config) {
//	for i := range config.Variants {
//		if config.Variants
//		config.Variants[i].
//	}
//}

func ReadConfig(data []byte) (*Config, error) {
	var config Config

	err := json.Unmarshal(data, &config)

	return &config, err
}

func ReadConfigFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	} else {
		return ReadConfig(data)
	}
}
