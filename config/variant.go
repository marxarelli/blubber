package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// VariantConfig holds configuration fields for each defined build variant.
//
type VariantConfig struct {
	Includes     []string          `yaml:"includes"`  // references to one or more
	Copies       string            `yaml:"copies"`    // copy standard artifacts from another variant
	Artifacts    []ArtifactsConfig `yaml:"artifacts"` // non-standard artifact configuration
	CommonConfig `yaml:",inline"`
}

// Merge takes another VariantConfig and overwrites this struct's fields.
// Artifacts are merged additively.
//
func (vc *VariantConfig) Merge(vc2 VariantConfig) {
	vc.Copies = vc2.Copies
	vc.Artifacts = append(vc.Artifacts, vc2.Artifacts...)
	vc.CommonConfig.Merge(vc2.CommonConfig)
}

// InstructionsForPhase injects build instructions related to artifact
// copying, volume definition or copying of application files, and all common
// configuration.
//
// PhaseInstall
//
// If VariantConfig.Copies is not set, either copy in application files or
// define a shared volume. Otherwise, delegate to
// ArtifactsConfig.InstructionsForPhase.
//
func (vc *VariantConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := vc.CommonConfig.InstructionsForPhase(phase)
	ainstructions := []build.Instruction{}

	for _, artifact := range vc.allArtifacts() {
		ainstructions = append(ainstructions, artifact.InstructionsForPhase(phase)...)
	}

	instructions = append(ainstructions, instructions...)

	switch phase {
	case build.PhaseInstall:
		if vc.Copies == "" {
			if vc.SharedVolume.True {
				instructions = append(instructions, build.Volume{vc.Runs.In})
			} else {
				instructions = append(instructions, build.Copy{[]string{"."}, "."})
			}
		}
	}

	return instructions
}

// VariantDependencies returns all unique names of other variants that are
// referenced in the VariantConfig.Artifacts configuration.
//
func (vc *VariantConfig) VariantDependencies() []string {
	// get unique set of variant dependencies based on artifacts
	existing := map[string]bool{}
	dependencies := []string{}

	for _, artifact := range vc.allArtifacts() {
		if dependency := artifact.From; dependency != "" && !existing[dependency] {
			existing[dependency] = true
			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies
}

func (vc *VariantConfig) allArtifacts() []ArtifactsConfig {
	return append(vc.defaultArtifacts(), vc.Artifacts...)
}

func (vc *VariantConfig) defaultArtifacts() []ArtifactsConfig {
	if vc.Copies != "" {
		return []ArtifactsConfig{
			{
				From:        vc.Copies,
				Source:      vc.Runs.In,
				Destination: vc.Runs.In,
			},
			{
				From:        vc.Copies,
				Source:      LocalLibPrefix,
				Destination: LocalLibPrefix,
			},
		}
	}

	return []ArtifactsConfig{}
}
