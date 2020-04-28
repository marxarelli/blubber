package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// DefaultConfig contains YAML that is applied before the user's
// configuration.
//
const DefaultConfig = `{
"lives": {
  "in": "/srv/app",
  "as": "somebody",
  "uid": 65533,
  "gid": 65533
},
"runs": {
  "as": "runuser",
  "uid": 900,
  "gid": 900}}`

// ResolveIncludes iterates over and recurses through a given variant's
// includes to build a flat slice of variant names in the correct order by
// which they should be expanded/merged. It checks for both the existence of
// included variants and maintains a recursion stack to protect against
// infinite loops.
//
// Variant names found at a greater depth of recursion are first and siblings
// last, the order in which config should be merged.
//
func ResolveIncludes(config *Config, name string) ([]string, error) {
	stack := map[string]bool{}
	includes := []string{}

	var resolve func(string) error

	resolve = func(name string) error {
		if instack, found := stack[name]; found && instack {
			return errors.New("variant expansion detected loop")
		}

		stack[name] = true
		defer func() { stack[name] = false }()

		variant, found := config.Variants[name]

		if !found {
			return fmt.Errorf("variant '%s' does not exist", name)
		}

		for _, include := range variant.Includes {
			if err := resolve(include); err != nil {
				return err
			}
		}

		// Appending _after_ recursion ensures the correct ordering
		includes = append(includes, name)

		return nil
	}

	err := resolve(name)

	return includes, err
}

// ExpandVariant merges a named variant with a config. It also attempts to
// recursively expand any included variants in the expanded variant.
//
func ExpandVariant(config *Config, name string) (*VariantConfig, error) {
	expanded := new(VariantConfig)
	expanded.CommonConfig.Merge(config.CommonConfig)

	includes, err := ResolveIncludes(config, name)

	if err != nil {
		return nil, err
	}

	for _, include := range includes {
		expanded.Merge(config.Variants[include])
	}

	return expanded, nil
}

// ExpandIncludesAndCopies resolves 'includes' and 'copies' for the
// specified variant.  This should be run before policy verfication
// so that the policy enforcement is applied to the final blubber spec
//
func ExpandIncludesAndCopies(config *Config, name string) error {
	vcfg, err := ExpandVariant(config, name)

	if err != nil {
		return fmt.Errorf("expanding variant '%s': %s", name, err)
	}

	config.Variants[name] = *vcfg

	for _, stage := range vcfg.Copies.Variants() {
		dependency, err := ExpandVariant(config, stage)

		if err != nil {
			return fmt.Errorf("expanding dependency '%s': %s", name, err)
		}

		config.Variants[stage] = *dependency
	}
	return nil
}

// GetVariant retrieves a requested *VariantConfig from the main config
//
func GetVariant(config *Config, name string) (*VariantConfig, error) {
	variant := new(VariantConfig)

	variant.Merge(config.Variants[name])

	return variant, nil
}

// ReadYAMLConfig converts YAML bytes to json and returns new Config struct.
//
func ReadYAMLConfig(data []byte) (*Config, error) {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	return ReadConfig(jsonData)
}

// ReadConfig unmarshals the given YAML bytes into a new Config struct.
//
func ReadConfig(data []byte) (*Config, error) {
	var (
		version VersionConfig
		config  Config
	)

	// Unmarshal config version first for pre-validation
	err := json.Unmarshal(data, &version)

	if err != nil {
		return nil, err
	}

	if err = Validate(version); err != nil {
		return nil, err
	}

	// Unmarshal the default config
	json.Unmarshal([]byte(DefaultConfig), &config)

	// And finally strictly decode the entire user-provided config
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	err = dec.Decode(&config)

	if err != nil {
		return nil, err
	}

	err = Validate(config)

	return &config, err
}

// ReadConfigFile unmarshals the given YAML file contents into a Config
// struct.
//
func ReadConfigFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return ReadYAMLConfig(data)
}
