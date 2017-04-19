package config

type VariantConfig struct {
	Includes []string `json:includes`
	Artifacts []ArtifactsConfig `json:artifacts`
	CommonConfig
}

func (vc1 *VariantConfig) Merge(vc2 VariantConfig) {
	vc1.Artifacts = append(vc1.Artifacts, vc2.Artifacts...)
	vc1.CommonConfig.Merge(vc2.CommonConfig)
}
