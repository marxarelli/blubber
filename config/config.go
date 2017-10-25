// Package config provides the internal representation of Blubber
// configuration parsed from YAML. Each configuration type may implement
// its own hooks for injecting build instructions into the compiler.
//
package config

// Config holds the root fields of a Blubber configuration.
//
type Config struct {
	CommonConfig `yaml:",inline"`
	Variants     map[string]VariantConfig `yaml:"variants"`
}
