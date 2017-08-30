package config

import (
	"phabricator.wikimedia.org/source/blubber.git/build"
)

type ArtifactsConfig struct {
	From        string `yaml:"from"`
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

func (ac ArtifactsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePostInstall:
		return []build.Instruction{
			build.CopyFrom{ac.From, build.Copy{[]string{ac.Source}, ac.Destination}},
		}
	}

	return []build.Instruction{}
}
