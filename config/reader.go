package config

import (
	"bytes"
	"encoding/json"
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

// ExpandVariant merges a named variant with a config. It also attempts to
// recursively expand any included variants in the expanded variant.
//
func ExpandVariant(config *Config, name string) (*VariantConfig, error) {
	expanded := new(VariantConfig)
	expanded.CommonConfig.Merge(config.CommonConfig)

	includes, err := config.IncludesDepGraph.GetDeps(name)

	if err != nil {
		return nil, err
	}

	for _, include := range includes {
		expanded.Merge(config.Variants[include])
	}

	expanded.Merge(config.Variants[name])

	return expanded, nil
}

// ExpandIncludesAndCopies resolves 'includes' for the
// specified variant.  It also expands any variants that are referenced
// directly or indirectly via 'copies' directives.
// This should be run before policy verfication
// so that the policy enforcement is applied to the final blubber spec
//
func ExpandIncludesAndCopies(config *Config, name string) error {
	BuildIncludesDepGraph(config)
	buildCopiesDepGraph(config)

	vcfg, err := ExpandVariant(config, name)

	if err != nil {
		return fmt.Errorf("processing includes for variant '%s': %s", name, err)
	}

	config.Variants[name] = *vcfg

	copiesDeps, err := config.CopiesDepGraph.GetDeps(name)

	if err != nil {
		return fmt.Errorf("processing copies for variant '%s': %s", name, err)
	}

	for _, stage := range copiesDeps {
		// Skip the pseudo variant "local"
		if stage == LocalArtifactKeyword {
			continue
		}

		vcfg, err := ExpandVariant(config, stage)
		if err != nil {
			return fmt.Errorf("processing includes for variant '%s': %s", stage, err)
		}

		config.Variants[stage] = *vcfg
	}

	return nil
}

// BuildIncludesDepGraph constructs the 'includes' dependency graph
func BuildIncludesDepGraph(config *Config) {
	graph := NewDepGraph()

	for variant, vcfg := range config.Variants {
		graph.EnsureNode(variant)
		for _, include := range vcfg.Includes {
			graph.AddDependency(variant, include)
		}
	}

	config.IncludesDepGraph = graph
}

func buildCopiesDepGraph(config *Config) {
	graph := NewDepGraph()

	graph.EnsureNode(LocalArtifactKeyword)

	for variant, vcfg := range config.Variants {
		graph.EnsureNode(variant)
		for _, ac := range vcfg.Copies {
			graph.AddDependency(variant, ac.From)
		}
	}

	config.CopiesDepGraph = graph
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
