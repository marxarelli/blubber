package config

type CommonConfig struct {
	Base string `yaml:"base"`
	Apt AptConfig `yaml:"apt"`
	Npm NpmConfig `yaml:"npm"`
	Run RunConfig `yaml:"run"`
	SharedVolume bool `yaml:"sharedvolume"`
	EntryPoint []string `yaml:"entrypoint"`
}

func (cc1 *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc1.Base = cc2.Base
	}

	cc1.Apt.Merge(cc2.Apt)
	cc1.Npm.Merge(cc2.Npm)
	cc1.Run.Merge(cc2.Run)

	cc1.SharedVolume = cc1.SharedVolume || cc2.SharedVolume

	if len(cc1.EntryPoint) < 1 {
		cc1.EntryPoint = cc2.EntryPoint
	}
}
