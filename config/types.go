package config

type Config struct {
	*CommonConfig
	Variants []VariantConfig `json:variants`
}

type CommonConfig struct {
	Base string `json:base`
	Apt AptConfig `json:apt`
	Npm NpmConfig `json:npm`
	EntryPoint []string `json:entrypoint`
}

type AptConfig struct {
	Packages []string `json:packages`
}

type NpmConfig struct {
  Use string `json:use`
}

type VariantConfig struct {
	Name string `json:name`
	Includes []string `json:includes`
	Artifacts []ArtifactsConfig `json:artifacts`
	*CommonConfig
}

type ArtifactsConfig struct {
	From string `json:name`
	Source string `json:source`
	Destination string `json:destination`
}
