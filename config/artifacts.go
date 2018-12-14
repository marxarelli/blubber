package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// ArtifactsConfig declares files and directories to be copied from one place
// to another during the build, either from the "local" build context or from
// another variant.
//
// The most common use of the latter such "multi-stage" build is to compile
// and test the application using one variant image that contains a
// comprehensive set of development dependencies, and copy the application
// binaries or production only source files over into a smaller image that
// contains only production dependencies.
//
type ArtifactsConfig struct {
	From        string `json:"from" validate:"required,variantref"`
	Source      string `json:"source" validate:"requiredwith=destination,relativelocal"`
	Destination string `json:"destination" validate:"requiredwith=source,relativelocal"`
}

// Expand returns the longhand configured artifact and/or the default
// artifacts for any configured by shorthand notation (i.e. on the `From`
// field).
//
func (ac ArtifactsConfig) Expand(appDirectory string) []ArtifactsConfig {
	// check for shorthand configuration and return its expanded form
	if ac.From != "" && ac.Source == "" && ac.Destination == "" {
		if ac.From == LocalArtifactKeyword {
			return []ArtifactsConfig{
				{
					From:        ac.From,
					Source:      ".",
					Destination: ".",
				},
			}
		}

		return []ArtifactsConfig{
			{
				From:        ac.From,
				Source:      appDirectory,
				Destination: appDirectory,
			},
			{
				From:        ac.From,
				Source:      LocalLibPrefix,
				Destination: LocalLibPrefix,
			},
		}
	}

	return []ArtifactsConfig{ac}
}

// InstructionsForPhase injects instructions into the given build phase that
// copy configured artifacts.
//
// PhaseInstall
//
// In the case of a "local" build context copy, simply return a build.Copy
// with the configured source and destination. In the case of a variant copy,
// return a build.CopyFrom instruction for the variant name, source and
// destination paths.
//
func (ac ArtifactsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhaseInstall:
		copy := build.Copy{[]string{ac.Source}, ac.Destination}

		if ac.From == LocalArtifactKeyword {
			return []build.Instruction{copy}
		}

		return []build.Instruction{
			build.CopyFrom{ac.From, copy},
		}
	}

	return []build.Instruction{}
}
