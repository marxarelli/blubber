package config

import (
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func ExpandVariant(config *Config, name string) (*VariantConfig, error) {
	variant, found := config.Variants[name]

	if !found {
		return nil, errors.New("variant does not exist")
	}

	expanded := new(VariantConfig)
	expanded.CommonConfig.Merge(config.CommonConfig)
	expanded.Merge(variant)

	for _, include := range variant.Includes {
		if includedVariant, found := config.Variants[include]; found {
			expanded.Merge(includedVariant)
		}
	}

	return expanded, nil
}

func ReadConfig(data []byte) (*Config, error) {
	var config Config

	err := yaml.Unmarshal(data, &config)

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
