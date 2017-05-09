package config

import (
	"github.com/marxarelli/blubber/build"
)

type CommonConfig struct {
	Base string `yaml:"base"`
	Apt AptConfig `yaml:"apt"`
	Npm NpmConfig `yaml:"npm"`
	Runs RunsConfig `yaml:"runs"`
	SharedVolume bool `yaml:"sharedvolume"`
	EntryPoint []string `yaml:"entrypoint"`
}

func (cc1 *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc1.Base = cc2.Base
	}

	cc1.Apt.Merge(cc2.Apt)
	cc1.Npm.Merge(cc2.Npm)
	cc1.Runs.Merge(cc2.Runs)

	cc1.SharedVolume = cc1.SharedVolume || cc2.SharedVolume

	if len(cc1.EntryPoint) < 1 {
		cc1.EntryPoint = cc2.EntryPoint
	}
}


func (cc *CommonConfig) PhaseCompileableConfig() []build.PhaseCompileable {
	return []build.PhaseCompileable{cc.Apt, cc.Npm, cc.Runs}
}

func (cc *CommonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	for _, phaseCompileable := range cc.PhaseCompileableConfig() {
		instructions = append(instructions, phaseCompileable.InstructionsForPhase(phase)...)
	}

	return instructions
}
