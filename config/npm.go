package config

import (
	"path"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

const tempNpmInstallDir = "/tmp/node-deps/"

type NpmConfig struct {
	Install Flag `yaml:"install"`
	Env string `yaml:"env"`
}

func (npm *NpmConfig) Merge(npm2 NpmConfig) {
	npm.Install.Merge(npm2.Install)

	if npm2.Env != "" {
		npm.Env = npm2.Env
	}
}

func (npm NpmConfig) InstructionsForPhase(phase build.Phase) []build.Instruction{
	if npm.Install.True {
		switch phase {
		case build.PhasePreInstall:
			npmCmd := "npm install"

			if npm.Env == "production" {
				npmCmd += " --production && npm dedupe"
			}

			return []build.Instruction{
				{build.Run, []string{"mkdir -p ", tempNpmInstallDir}},
				{build.Copy, []string{"package.json", tempNpmInstallDir}},
				{build.Run, []string{"cd ", tempNpmInstallDir, " && ", npmCmd}},
			}
		case build.PhasePostInstall:
			return []build.Instruction{
				{build.Run, []string{"mv ", path.Join(tempNpmInstallDir, "node_modules"), " ./"}},
			}
		}
	}

	return []build.Instruction{}
}
