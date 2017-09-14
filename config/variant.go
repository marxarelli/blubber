package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

type VariantConfig struct {
	Includes     []string          `yaml:"includes"`
	Copies       string            `yaml:"copies"`
	Artifacts    []ArtifactsConfig `yaml:"artifacts"`
	CommonConfig `yaml:",inline"`
}

func (vc *VariantConfig) Merge(vc2 VariantConfig) {
	vc.Copies = vc2.Copies
	vc.Artifacts = append(vc.Artifacts, vc2.Artifacts...)
	vc.CommonConfig.Merge(vc2.CommonConfig)
}

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
