package config

type Config struct {
	*CommonConfig
	Variants map[string]VariantConfig `json:variants`
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
	Includes []string `json:includes`
	Artifacts []ArtifactsConfig `json:artifacts`
	*CommonConfig
}

type ArtifactsConfig struct {
	From string `json:from`
	Source string `json:source`
	Destination string `json:destination`
}
