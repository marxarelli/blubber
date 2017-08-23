package config

type VariantConfig struct {
	Includes     []string          `yaml:"includes"`
	Artifacts    []ArtifactsConfig `yaml:"artifacts"`
	CommonConfig `yaml:",inline"`
}

func (vc *VariantConfig) Merge(vc2 VariantConfig) {
	vc.Artifacts = append(vc.Artifacts, vc2.Artifacts...)
	vc.CommonConfig.Merge(vc2.CommonConfig)
}
