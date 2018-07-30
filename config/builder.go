package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// BuilderConfig contains configuration for the definition of an arbitrary
// build command.
//
type BuilderConfig struct {
	Builder []string `yaml:"builder"`
}

// Merge takes another BuilderConfig and merges its fields into this one's,
// overwriting the builder command.
//
func (bc *BuilderConfig) Merge(bc2 BuilderConfig) {
	if len(bc2.Builder) > 0 {
		bc.Builder = bc2.Builder
	}
}

// InstructionsForPhase injects instructions into the build related to
// builder commands.
//
// PhasePostInstall
//
// Runs build command provided for the builder
//
func (bc BuilderConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePostInstall:
		if len(bc.Builder) == 0 {
			return []build.Instruction{}
		}

		run := build.Run{Command: bc.Builder[0]}

		if len(bc.Builder) > 1 {
			run.Arguments = bc.Builder[1:]
		}

		return []build.Instruction{run}
	}

	return []build.Instruction{}
}
