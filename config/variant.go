package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// VariantConfig holds configuration fields for each defined build variant.
//
type VariantConfig struct {
	Includes     []string          `yaml:"includes" validate:"dive,variantref"`    // other variants
	Copies       string            `yaml:"copies" validate:"omitempty,variantref"` // copy artifacts from variant
	Artifacts    []ArtifactsConfig `yaml:"artifacts" validate:"dive"`              // artifact configuration
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
	var switchUser string

	switch phase {
	case build.PhasePrivileged:
		switchUser = "root"

	case build.PhasePrivilegeDropped:
		switchUser = vc.Lives.As
		instructions = build.ApplyUser(vc.Lives.UID, vc.Lives.GID, instructions)

	case build.PhasePreInstall:
		instructions = build.ApplyUser(vc.Lives.UID, vc.Lives.GID, instructions)

	case build.PhaseInstall:
		if vc.Copies == "" {
			if vc.SharedVolume.True {
				instructions = append(instructions, build.Volume{vc.Lives.In})
			} else {
				instructions = append(instructions, build.Copy{[]string{"."}, "."})
			}
		}

		instructions = build.ApplyUser(vc.Lives.UID, vc.Lives.GID, instructions)

	case build.PhasePostInstall:
		switchUser = vc.Runs.As
		instructions = build.ApplyUser(vc.Runs.UID, vc.Runs.GID, instructions)

		if len(vc.EntryPoint) > 0 {
			instructions = append(instructions, build.EntryPoint{vc.EntryPoint})
		}
	}

	if switchUser != "" {
		instructions = append(
			[]build.Instruction{
				build.User{switchUser},
				build.Home(switchUser),
			},
			instructions...,
		)
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
				Source:      vc.Lives.In,
				Destination: vc.Lives.In,
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
