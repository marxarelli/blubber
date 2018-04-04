package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// DefaultConfig contains YAML that is applied before the user's
// configuration.
//
const DefaultConfig = `---
lives:
  in: /srv/app
  as: somebody
  uid: 65533
  gid: 65533
runs:
  as: runuser
  uid: 900
  gid: 900`

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

// ReadConfig unmarshals the given YAML bytes into a new Config struct.
//
func ReadConfig(data []byte) (*Config, error) {
	var (
		version VersionConfig
		config  Config
	)

	// Unmarshal (un-strictly) config version first for pre-validation
	err := yaml.Unmarshal(data, &version)

	if err != nil {
		return nil, err
	}

	if err = Validate(version); err != nil {
		return nil, err
	}

	// Unmarshal the default config
	yaml.Unmarshal([]byte(DefaultConfig), &config)

	// And finally strictly unmarshal the entire user-provided config
	err = yaml.UnmarshalStrict(data, &config)

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

	return ReadConfig(data)
}
