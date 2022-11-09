package config

import (
	"path"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
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
	From        string `json:"from" validate:"required,artifactfrom"`
	Source      string `json:"source" validate:"requiredwith=destination,relativelocal"`
	Destination string `json:"destination"`
}

// NewArtifactsConfigFromSource creates an local ArtifactsConfig from the
// given source. This helps to support legacy requirements definitions.
//
func NewArtifactsConfigFromSource(source string) ArtifactsConfig {
	return ArtifactsConfig{
		From:   LocalArtifactKeyword,
		Source: source,
	}
}

// Dependencies returns variant dependencies.
//
func (ac ArtifactsConfig) Dependencies() []string {
	if ac.From != "" && ac.From != LocalArtifactKeyword {
		return []string{ac.From}
	}
	return []string{}
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

	// if destination is empty, use the source value
	if ac.Destination == "" {
		return []ArtifactsConfig{
			{From: ac.From, Source: ac.Source, Destination: ac.Source},
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

// EffectiveDestination returns the destination as a file path that amounts to
// the location of the artifact after a copy is performed.
//
// If the destination is a file path or source is a directory, the effective
// destination is identical to the normalized destination path.
//
// If the source is a file path (e.g. "foo/bar"), and the destination is a
// directory (e.g. "foo2/"), the effective destination is the directory + the
// base name of the source file (e.g. "foo2/bar").
//
func (ac ArtifactsConfig) EffectiveDestination() string {
	dest := ac.NormalizedDestination()

	if !isDir(dest) || isDir(ac.Source) {
		return dest
	}

	base := path.Base(ac.Source)

	if dest == "./" {
		return base
	}

	return dest + base
}

// NormalizedDestination returns the destination defaulted to the source
// directory and sanitized by path.Clean but with any original trailing '/'
// retained to indicate a directory path.
//
func (ac ArtifactsConfig) NormalizedDestination() string {
	// Default behavior is to derive Destination from Source
	if ac.Destination == "" {
		// If source is a directory, use it as is
		if ac.Source != "" && isDir(ac.Source) {
			return ac.Source
		}

		// Otherwise, use the source directory
		return path.Dir(ac.NormalizedSource()) + "/"
	}

	dest := path.Clean(ac.Destination)

	if dest != "/" && (isDir(ac.Destination) || isDir(ac.Source)) {
		dest += "/"
	}

	return dest
}

// NormalizedSource returns the source sanitized by path.Clean but retaining
// any terminating "/" to denote a directory.
//
func (ac ArtifactsConfig) NormalizedSource() string {
	cleaned := path.Clean(ac.Source)

	if cleaned != "/" && isDir(ac.Source) {
		return cleaned + "/"
	}

	return cleaned
}

func isDir(aPath string) bool {
	return path.Clean(aPath) == "." || aPath[len(aPath)-1:] == "/"
}
