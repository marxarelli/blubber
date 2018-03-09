package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// ArtifactsConfig declares files and directories to be copied from one
// variant's container to another during the build.
//
// The most common use of such "multi-stage" builds is to compile and test
// software using one variant image that contains a comprehensive set of
// development dependencies, and copy the software binaries or production only
// source files over into a smaller image that contains only production
// dependencies. For a shorthand configuration of this precise pattern, use
// VariantConfig.Copies.
//
type ArtifactsConfig struct {
	From        string `yaml:"from" validate:"required,variantref"` // source variant from which to copy
	Source      string `yaml:"source" validate:"required"`          // source variant path from which to copy
	Destination string `yaml:"destination" validate:"required"`     // destination path within current variant
}

// InstructionsForPhase injects instructions into the given build phase that
// copy configured artifacts.
//
// PhaseInstall
//
// Injects build.CopyFrom instructions for the configured source and
// destination paths.
//
func (ac ArtifactsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhaseInstall:
		return []build.Instruction{
			build.CopyFrom{ac.From, build.Copy{[]string{ac.Source}, ac.Destination}},
		}
	}

	return []build.Instruction{}
}
