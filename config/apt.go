package config

import (
	"encoding/json"
	"sort"

	"gerrit.wikimedia.org/r/blubber/build"
)

// AptConfig represents configuration pertaining to package installation from
// existing APT sources.
//
type AptConfig struct {
	// Packages keys are the name of the targeted release, or 'default' to
	// specify no target and use the base image's target release,
	// Packages values are a list of the desired packages
	Packages map[string][]string `json:"packages" validate:"dive,keys,omitempty,debianrelease,endkeys,dive,debianpackage"`
}

// DefaultTargetKeyword defines a special keyword indicating
// that the packages to be installed should use the default target release
//
const DefaultTargetKeyword = "default"

// Merge takes another AptConfig and combines the packages declared within
// with the packages of this AptConfig.
//
func (apt *AptConfig) Merge(apt2 AptConfig) {

	if apt2.Packages != nil {
		if apt.Packages == nil {
			apt.Packages = make(map[string][]string)
		}

		for key, pkgs := range apt2.Packages {
			apt.Packages[key] = append(apt.Packages[key], pkgs...)
		}
	}

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
			var ins = []build.Instruction{build.Env{map[string]string{
				"DEBIAN_FRONTEND": "noninteractive",
			}}}
			var runAll = []build.Run{{"apt-get update", []string{}}}
			var targets []string

			// order the targets for the same result each run
			for target := range apt.Packages {
				targets = append(targets, target)
			}
			sort.Strings(targets)

			for _, target := range targets {
				if target == DefaultTargetKeyword {
					runAll = append(runAll, build.Run{"apt-get install -y", apt.Packages[target]})
				} else {
					args := append([]string{target}, apt.Packages[target]...)
					runAll = append(runAll, build.Run{"apt-get install -y -t", args})
				}
			}

			runAll = append(runAll, build.Run{"rm -rf /var/lib/apt/lists/*", []string{}})
			return append(ins, build.RunAll{runAll})
		}
	}

	return []build.Instruction{}
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand and
// longhand apt packages configuration.
//
// Shorthand packages configuration: ["package1", "package2"]
// Longhand packages configuration: { "release1": ["package1, package2"], "release2": ["package3"]}
func (apt *AptConfig) UnmarshalJSON(unmarshal []byte) error {
	apt.Packages = make(map[string][]string)
	longhand := make(map[string]map[string][]string)
	err := json.Unmarshal(unmarshal, &longhand)

	if err == nil {
		for key, pkgs := range longhand["packages"] {
			apt.Packages[key] = append(apt.Packages[key], pkgs...)
		}
		return nil
	}

	shorthand := map[string][]string{}
	err = json.Unmarshal(unmarshal, &shorthand)

	if err == nil {
		// Input was entirely in short form
		apt.Packages[DefaultTargetKeyword] = shorthand["packages"]
		return nil
	}

	return err
}
