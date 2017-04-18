package config

type ConfigType struct {
	Base string `json:base`
	Dependencies DependenciesType `json:dependencies`
	Build BuildVariantType `json:build`
	Test TestVariantType `json:test`
}

type BuildVariantType struct {
	Dependencies DependenciesType `json:dependencies`
}

type TestVariantType struct {
	Dependencies DependenciesType `json:dependencies`
}

type DependenciesType struct {
	Apt []string `json:apt`
	NpmInclude string `json:npminclude`
}
