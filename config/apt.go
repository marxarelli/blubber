package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// AptConfig represents configuration pertaining to package installation from
// existing APT sources.
//
type AptConfig struct {
	Packages []string `json:"packages" validate:"dive,debianpackage"`
}

// Merge takes another AptConfig and combines the packages declared within
// with the packages of this AptConfig.
//
func (apt *AptConfig) Merge(apt2 AptConfig) {
	apt.Packages = append(apt.Packages, apt2.Packages...)
}

// InstructionsForPhase injects build instructions that will install the
// declared packages during the privileged phase.
//
// PhasePrivileged
//
// Updates the APT cache, installs configured packages, and cleans up.
//
func (apt AptConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if len(apt.Packages) > 0 {
		switch phase {
		case build.PhasePrivileged:
			return []build.Instruction{
				build.Env{map[string]string{
					"DEBIAN_FRONTEND": "noninteractive",
				}},
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
