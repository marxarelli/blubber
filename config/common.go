package config

type CommonConfig struct {
	Base string `json:base`
	Apt AptConfig `json:apt`
	Npm NpmConfig `json:npm`
	Run RunConfig `json:run`
	EntryPoint []string `json:entrypoint`
}

func (cc1 *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc1.Base = cc2.Base
	}

	cc1.Apt.Merge(cc2.Apt)
	cc1.Npm.Merge(cc2.Npm)
	cc1.Run.Merge(cc2.Run)

	if len(cc1.EntryPoint) < 1 {
		cc1.EntryPoint = cc2.EntryPoint
	}
}
