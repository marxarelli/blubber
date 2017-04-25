package config

import (
	"bytes"
)

type NpmConfig struct {
	Install bool `yaml:"install"`
	Env string `yaml:"env"`
}

func (npm *NpmConfig) Merge(npm2 NpmConfig) {
	npm.Install = npm.Install || npm2.Install

	if npm2.Env != "" {
		npm.Env = npm2.Env
	}
}

func (npm NpmConfig) Commands() []string {
	if !npm.Install {
		return []string{}
	}

	buffer := new(bytes.Buffer)

	buffer.WriteString("npm install")

	if npm.Env == "production" {
		buffer.WriteString(" --production && npm dedupe")
	}

	return []string{buffer.String()}
}
