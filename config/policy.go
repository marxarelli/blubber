package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/utahta/go-openuri"
	"gopkg.in/yaml.v2"
)

// Policy validates a number of rules against a given configuration.
//
type Policy struct {
	Enforcements []Enforcement `yaml:"enforcements"`
}

// Validate checks the given config against all policy enforcements.
//
func (pol Policy) Validate(config Config) error {
	validate := NewValidator()

	for _, enforcement := range pol.Enforcements {
		cfg, err := ResolveYAMLPath(enforcement.Path, config)

		if err != nil {
			// If the path resolved nothing, there's nothing to enforce
			continue
		}

		// Flags are a special case in which the True field should be compared
		// against the validator, not the struct itself.
		if flag, ok := cfg.(Flag); ok {
			cfg = flag.True
		}

		err = validate.Var(cfg, enforcement.Rule)

		if err != nil {
			return fmt.Errorf(
				`value for "%s" violates policy rule "%s"`,
				enforcement.Path, enforcement.Rule,
			)
		}
	}

	return nil
}

// Enforcement represents a policy rule and config path on which to apply it.
//
type Enforcement struct {
	Path string `yaml:"path"`
	Rule string `yaml:"rule"`
}

// ReadPolicy unmarshals the given YAML bytes into a new Policy struct.
//
func ReadPolicy(data []byte) (*Policy, error) {
	var policy Policy

	err := yaml.Unmarshal(data, &policy)

	if err != nil {
		return nil, err
	}

	return &policy, err
}

// ReadPolicyFromURI fetches the policy file from the given URL or file path
// and loads its contents with ReadPolicy.
//
func ReadPolicyFromURI(uri string) (*Policy, error) {
	io, err := openuri.Open(uri)

	if err != nil {
		return nil, err
	}

	defer io.Close()

	data, err := ioutil.ReadAll(io)

	if err != nil {
		return nil, err
	}

	return ReadPolicy(data)
}

// ResolveYAMLPath returns the config value found at the given YAML-ish
// namespace/path (e.g. "variants.production.runs.as").
//
func ResolveYAMLPath(path string, cfg interface{}) (interface{}, error) {
	parts := strings.SplitN(path, ".", 2)
	name := parts[0]

	v := reflect.ValueOf(cfg)
	t := v.Type()

	var subcfg interface{}

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Anonymous {
				if subsubcfg, err := ResolveYAMLPath(path, v.Field(i).Interface()); err == nil {
					return subsubcfg, nil
				}
			}

			if name == resolveYAMLTagName(t.Field(i)) {
				subcfg = v.Field(i).Interface()
				break
			}
		}

	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			for _, key := range v.MapKeys() {
				if key.Interface().(string) == name {
					subcfg = v.MapIndex(key).Interface()
					break
				}
			}
		}
	}

	if subcfg == nil {
		return nil, errors.New("invalid path")
	}

	if len(parts) > 1 {
		return ResolveYAMLPath(parts[1], subcfg)
	}

	return subcfg, nil
}
