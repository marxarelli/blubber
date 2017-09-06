package config

import (
	"path"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

type NodeConfig struct {
	Dependencies Flag   `yaml:"dependencies"`
	Env          string `yaml:"env"`
}

func (nc *NodeConfig) Merge(nc2 NodeConfig) {
	nc.Dependencies.Merge(nc2.Dependencies)

	if nc2.Env != "" {
		nc.Env = nc2.Env
	}
}

func (nc NodeConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePreInstall:
		if nc.Dependencies.True {
			npmInstall := build.RunAll{[]build.Run{
				{"cd", []string{LocalLibPrefix}},
				{"npm install", []string{}},
			}}

			if nc.Env == "production" {
				npmInstall.Runs[1].Arguments = []string{"--production"}
				npmInstall.Runs = append(npmInstall.Runs,
					build.Run{"npm dedupe", []string{}},
				)
			}

			return []build.Instruction{
				build.Copy{[]string{"package.json"}, LocalLibPrefix},
				npmInstall,
			}
		}
	case build.PhasePostInstall:
		if nc.Env != "" || nc.Dependencies.True {
			return []build.Instruction{build.Env{map[string]string{
				"NODE_ENV":  nc.Env,
				"NODE_PATH": path.Join(LocalLibPrefix, "node_modules"),
				"PATH":      path.Join(LocalLibPrefix, "node_modules", ".bin") + ":${PATH}",
			}}}
		}
	}

	return []build.Instruction{}
}
