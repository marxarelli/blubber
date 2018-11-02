package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// VariantConfig holds configuration fields for each defined build variant.
//
type VariantConfig struct {
	Includes     []string          `json:"includes" validate:"dive,variantref"`    // other variants
	Copies       string            `json:"copies" validate:"omitempty,variantref"` // copy artifacts from variant
	Artifacts    []ArtifactsConfig `json:"artifacts" validate:"dive"`              // artifact configuration
	CommonConfig `json:",inline"`
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
// copying, copying of application files, and all common configuration.
//
// PhaseInstall
//
// If VariantConfig.Copies is not set, copy in application files. Otherwise,
// delegate to ArtifactsConfig.InstructionsForPhase.
//
func (vc *VariantConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := vc.CommonConfig.InstructionsForPhase(phase)

	var switchUser string
	var uid, gid uint

	switch phase {
	case build.PhasePrivileged:
		switchUser = "root"

	case build.PhasePrivilegeDropped:
		switchUser = vc.Lives.As
		uid, gid = vc.Lives.UID, vc.Lives.GID

	case build.PhasePreInstall:
		uid, gid = vc.Lives.UID, vc.Lives.GID

	case build.PhaseInstall:
		uid, gid = vc.Lives.UID, vc.Lives.GID

		if vc.Copies == "" {
			instructions = append(instructions, build.Copy{[]string{"."}, "."})
		}

	case build.PhasePostInstall:
		if vc.Runs.Insecurely.True {
			uid, gid = vc.Lives.UID, vc.Lives.GID
		} else {
			switchUser = vc.Runs.As
			uid, gid = vc.Runs.UID, vc.Runs.GID
		}

		if len(vc.EntryPoint) > 0 {
			instructions = append(instructions, build.EntryPoint{vc.EntryPoint})
		}
	}

	for _, artifact := range vc.allArtifacts() {
		instructions = append(instructions, artifact.InstructionsForPhase(phase)...)
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

	if uid != 0 {
		instructions = build.ApplyUser(uid, gid, instructions)
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
