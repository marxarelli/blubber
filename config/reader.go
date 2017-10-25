package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func expandIncludes(config *Config, name string, included map[string]bool) ([]string, error) {
	variant, found := config.Variants[name]

	if !found {
		return nil, fmt.Errorf("variant '%s' does not exist", name)
	}

	if included[name] == true {
		return nil, errors.New("variant expansion detected loop")
	}

	for _, include := range variant.Includes {
		included[name] = true
		inc, err := expandIncludes(config, include, included)

		if err != nil {
			return nil, err
		}

		return append(inc, include), nil
	}

	return []string{}, nil
}

// ExpandVariant merges a named variant with a config. It also attempts to
// recursively expand any included variants in the expanded variant.
//
func ExpandVariant(config *Config, name string) (*VariantConfig, error) {
	expanded := new(VariantConfig)
	expanded.CommonConfig.Merge(config.CommonConfig)

	includes, err := expandIncludes(config, name, map[string]bool{})
	includes = append(includes, name)

	if err != nil {
		return nil, err
	}

	for _, include := range includes {
		expanded.Merge(config.Variants[include])
	}

	return expanded, nil
}

// ReadConfig unmarshals the given YAML bytes into a Config struct.
//
func ReadConfig(data []byte) (*Config, error) {
	var config Config

	err := yaml.Unmarshal(data, &config)

	return &config, err
}

// ReadConfigFile unmarshals the given YAML file contents into a Config
// struct.
//
func ReadConfigFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	} else {
		return ReadConfig(data)
	}
}
