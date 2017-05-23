package config

import (
	"bytes"
	"path"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

const TempNpmInstallDir = "/tmp/node-deps/"

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

func (npm NpmConfig) InstructionsForPhase(phase build.Phase) []build.Instruction{
	if npm.Install {
		switch phase {
		case build.PhasePreInstall:
			npmCmd := new(bytes.Buffer)

			npmCmd.WriteString("npm install")

			if npm.Env == "production" {
				npmCmd.WriteString(" --production && npm dedupe")
			}

			return []build.Instruction{
				{build.Run, []string{"mkdir -p ", TempNpmInstallDir}},
				{build.Copy, []string{"package.json", TempNpmInstallDir}},
				{build.Run, []string{"cd ", TempNpmInstallDir, " && ", npmCmd.String()}},
			}
		case build.PhasePostInstall:
			return []build.Instruction{
				{build.Run, []string{"mv ", path.Join(TempNpmInstallDir, "node_modules"), " ./"}},
			}
		}
	}

	return []build.Instruction{}
}
