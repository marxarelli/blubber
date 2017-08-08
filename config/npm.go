package config

import (
	"path"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

const tempNpmInstallDir = "/tmp/node-deps/"

type NpmConfig struct {
	Install Flag   `yaml:"install"`
	Env     string `yaml:"env"`
}

func (npm *NpmConfig) Merge(npm2 NpmConfig) {
	npm.Install.Merge(npm2.Install)

	if npm2.Env != "" {
		npm.Env = npm2.Env
	}
}

func (npm NpmConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if npm.Install.True {
		switch phase {
		case build.PhasePreInstall:
			npmInstall := build.RunAll{[]build.Run{
				{"cd", []string{tempNpmInstallDir}},
				{"npm install", []string{}},
			}}

			if npm.Env == "production" {
				npmInstall.Runs[1].Arguments = []string{"--production"}
				npmInstall.Runs = append(npmInstall.Runs,
					build.Run{"npm dedupe", []string{}},
				)
			}

			return []build.Instruction{
				build.Run{"mkdir -p", []string{tempNpmInstallDir}},
				build.Copy{[]string{"package.json"}, tempNpmInstallDir},
				npmInstall,
			}
		case build.PhasePostInstall:
			return []build.Instruction{
				build.Run{"mv", []string{path.Join(tempNpmInstallDir, "node_modules"), "./"}},
			}
		}
	}

	return []build.Instruction{}
}
