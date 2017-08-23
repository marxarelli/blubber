package config

import (
	"path"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

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
				{"cd", []string{LocalLibPrefix}},
				{"npm install", []string{}},
			}}

			if npm.Env == "production" {
				npmInstall.Runs[1].Arguments = []string{"--production"}
				npmInstall.Runs = append(npmInstall.Runs,
					build.Run{"npm dedupe", []string{}},
				)
			}

			return []build.Instruction{
				build.Copy{[]string{"package.json"}, LocalLibPrefix},
				npmInstall,
			}
		case build.PhasePostInstall:
			return []build.Instruction{
				build.Env{map[string]string{"NODE_PATH": path.Join(LocalLibPrefix, "node_modules")}},
			}
		}
	}

	return []build.Instruction{}
}
