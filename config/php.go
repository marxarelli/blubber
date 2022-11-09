package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// PhpConfig holds configuration for whether/how to install php packages.
//
type PhpConfig struct {
	// Install requirements from given files
	Requirements RequirementsConfig `json:"requirements" validate:"omitempty,unique,dive"`

	// Whether to use the no-dev flag
	Production Flag `json:"production"`
}

// Dependencies returns variant dependencies.
//
func (pc PhpConfig) Dependencies() []string {
	return pc.Requirements.Dependencies()
}

// Merge takes another PhpConfig and merges its fields into this one's,
// overwriting the requirements files.
//
func (pc *PhpConfig) Merge(pc2 PhpConfig) {
	pc.Production.Merge(pc2.Production)

	if pc2.Requirements != nil {
		pc.Requirements = pc2.Requirements
	}
}

// InstructionsForPhase injects instructions into the build related to PHP
// dependency installation.
//
// PhasePreInstall
//
// Installs Php package dependencies declared in composer files into the
// application directory. Installing dependencies during the build.PhasePreInstall
// phase allows a compiler implementation (e.g. Docker) to produce cache-efficient
// output so only changes to composer json will invalidate these steps of the image
// build.
//
func (pc PhpConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := pc.Requirements.InstructionsForPhase(phase)

	switch phase {
	case build.PhasePreInstall:
		var composerInstall build.RunAll
		if len(pc.Requirements) > 0 {

			composerInstall = build.RunAll{[]build.Run{
				{"composer install", []string{"--no-scripts"}},
			}}

			if pc.Production.True {
				composerInstall.Runs[0].Arguments = append(composerInstall.Runs[0].Arguments, "--no-dev")
			}

			ins = append(ins, composerInstall)
		}

	}

	return ins
}
