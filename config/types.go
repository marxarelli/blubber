package config

type Config struct {
	CommonConfig
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
	CommonConfig
}

type ArtifactsConfig struct {
	From string `json:from`
	Source string `json:source`
	Destination string `json:destination`
}

func (vc1 *VariantConfig) Merge(vc2 VariantConfig) {
	vc1.Artifacts = append(vc1.Artifacts, vc2.Artifacts...)
	vc1.CommonConfig.Merge(vc2.CommonConfig)
}

func (cc1 *CommonConfig) Merge(cc2 CommonConfig) {
	if cc1.Base == "" {
		cc1.Base = cc2.Base
	}

	cc1.Apt.Packages = append(cc1.Apt.Packages, cc2.Apt.Packages...)

	if cc1.Npm.Use == "" {
		cc1.Npm.Use = cc2.Npm.Use
	}

	if len(cc1.EntryPoint) < 1 {
		cc1.EntryPoint = cc2.EntryPoint
	}
}
