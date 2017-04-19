package config

type Config struct {
	CommonConfig
	Variants map[string]VariantConfig `json:variants`
}

type CommandCompileable interface {
	Commands() []string
}
