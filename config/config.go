package config

type Config struct {
	CommonConfig `yaml:",inline"`
	Variants map[string]VariantConfig `yaml:"variants"`
}

type CommandCompileable interface {
	Commands() []string
}
