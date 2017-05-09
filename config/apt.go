package config

import (
	"strings"
	"github.com/marxarelli/blubber/build"
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
				{build.Run, []string{
					"apt-get update && apt-get install -y ",
					strings.Join(apt.Packages, " "),
					" && rm -rf /var/lib/apt/lists/*",
				}},
			}
		}
	}

	return []build.Instruction{}
}
