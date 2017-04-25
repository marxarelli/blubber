package config

type VariantConfig struct {
	Includes []string `yaml:"includes"`
	Artifacts []ArtifactsConfig `yaml:"artifacts"`
	CommonConfig `yaml:",inline"`
}

func (vc1 *VariantConfig) Merge(vc2 VariantConfig) {
	vc1.Artifacts = append(vc1.Artifacts, vc2.Artifacts...)
	vc1.CommonConfig.Merge(vc2.CommonConfig)
}
