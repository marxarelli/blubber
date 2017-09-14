package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

type AptConfig struct {
	Packages []string `yaml:"packages"`
}

func (apt *AptConfig) Merge(apt2 AptConfig) {
	apt.Packages = append(apt.Packages, apt2.Packages...)
}

func (apt AptConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if len(apt.Packages) > 0 {
		switch phase {
		case build.PhasePrivileged:
			return []build.Instruction{
				build.RunAll{[]build.Run{
					{"apt-get update", []string{}},
					{"apt-get install -y", apt.Packages},
					{"rm -rf /var/lib/apt/lists/*", []string{}},
				}},
			}
		}
	}

	return []build.Instruction{}
}
