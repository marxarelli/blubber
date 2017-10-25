package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// CommonConfig holds the configuration fields common to both the root config
// and each configured variant.
//
type CommonConfig struct {
	Base         string     `yaml:"base"`         // name/path to base image
	Apt          AptConfig  `yaml:"apt"`          // APT related configuration
	Node         NodeConfig `yaml:"node"`         // Node related configuration
	Runs         RunsConfig `yaml:"runs"`         // runtime environment configuration
	SharedVolume Flag       `yaml:"sharedvolume"` // define a volume instead of copying in source files
	EntryPoint   []string   `yaml:"entrypoint"`   // entry-point executable
}

// Merge takes another CommonConfig and merges its fields this one's.
//
func (cc1 *CommonConfig) Merge(cc2 CommonConfig) {
	if cc2.Base != "" {
		cc1.Base = cc2.Base
	}

	cc1.Apt.Merge(cc2.Apt)
	cc1.Node.Merge(cc2.Node)
	cc1.Runs.Merge(cc2.Runs)
	cc1.SharedVolume.Merge(cc2.SharedVolume)

	if len(cc1.EntryPoint) < 1 {
		cc1.EntryPoint = cc2.EntryPoint
	}
}

// PhaseCompileableConfig returns all fields that implement
// build.PhaseCompileable in the order that their instructions should be
// injected.
//
func (cc *CommonConfig) PhaseCompileableConfig() []build.PhaseCompileable {
	return []build.PhaseCompileable{cc.Apt, cc.Node, cc.Runs}
}

// InstructionsForPhase injects instructions into the given build phase for
// each member field that supports it.
//
func (cc *CommonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	for _, phaseCompileable := range cc.PhaseCompileableConfig() {
		instructions = append(instructions, phaseCompileable.InstructionsForPhase(phase)...)
	}

	return instructions
}
