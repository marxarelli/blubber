package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// BuilderConfig contains configuration for the definition of an arbitrary
// build command and the files required to successfully execute the command.
//
type BuilderConfig struct {
	Command      []string `json:"command"`
	Requirements []string `json:"requirements"`
}

// Merge takes another BuilderConfig and merges its fields into this one's,
// overwriting the builder command.
//
func (bc *BuilderConfig) Merge(bc2 BuilderConfig) {
	if bc2.Command != nil {
		bc.Command = bc2.Command
	}

	if bc2.Requirements != nil {
		bc.Requirements = bc2.Requirements
	}
}

// InstructionsForPhase injects instructions into the build related to
// builder commands and required files.
//
// PhasePreInstall
//
// Creates directories for requirements files, copies in requirements files,
// and runs the builder command.
//
func (bc BuilderConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if len(bc.Command) == 0 {
		return []build.Instruction{}
	}

	switch phase {
	case build.PhasePreInstall:
		syncs := build.SyncFiles(bc.Requirements, ".")
		run := build.Run{Command: bc.Command[0]}

		if len(bc.Command) > 1 {
			run.Arguments = bc.Command[1:]
		}

		return append(syncs, run)
	}

	return []build.Instruction{}
}
