package config

import (
	"bytes"
	"encoding/json"
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// BuildersConfig holds the configuration of multiple different builders. The order in which they
// appear in the slice, is the order in which their associated instructions are generated
type BuildersConfig []build.PhaseCompileable

type builderEntry struct {
	PythonBuilder *PythonConfig  `json:"python"`
	NodeBuilder   *NodeConfig    `json:"node"`
	PhpBuilder    *PhpConfig     `json:"php"`
	CustomBuilder *BuilderConfig `json:"custom"`
}

// Merge takes another BuildersConfig and merges its fields into this one's, with the following rules:
//
// * Builders common to both will be merged, with builders from bc2 taking precedence. They will
// retain their relative positions as defined by bc2
//
// * Non-common builders in bc will be placed before builders of bc2 (custom builders are considered
// non-common)
func (bc *BuildersConfig) Merge(bc2 BuildersConfig) {
	bc2BuildersType2Pos := make(map[string]int, len(bc2))
	for i, b := range bc2 {
		switch b.(type) {
		case PythonConfig:
			bc2BuildersType2Pos["python"] = i
		case NodeConfig:
			bc2BuildersType2Pos["node"] = i
		case PhpConfig:
			bc2BuildersType2Pos["php"] = i
		}
	}

	leadingBuilders := []build.PhaseCompileable{}
	for _, bi := range *bc {
		switch bi.(type) {
		case PythonConfig:
			if b2Pos, ok := bc2BuildersType2Pos["python"]; ok {
				b := bi.(PythonConfig)
				b2 := bc2[b2Pos].(PythonConfig)
				b.Merge(b2)
				bc2[b2Pos] = b
			} else {
				leadingBuilders = append(leadingBuilders, bi)
			}
		case NodeConfig:
			if b2Pos, ok := bc2BuildersType2Pos["node"]; ok {
				b := bi.(NodeConfig)
				b2 := bc2[b2Pos].(NodeConfig)
				b.Merge(b2)
				bc2[b2Pos] = b
			} else {
				leadingBuilders = append(leadingBuilders, bi)
			}
		case PhpConfig:
			if b2Pos, ok := bc2BuildersType2Pos["php"]; ok {
				b := bi.(PhpConfig)
				b2 := bc2[b2Pos].(PhpConfig)
				b.Merge(b2)
				bc2[b2Pos] = b
			} else {
				leadingBuilders = append(leadingBuilders, bi)
			}
		default:
			leadingBuilders = append(leadingBuilders, bi)
		}
	}

	trailingBuilders := bc2
	*bc = append(leadingBuilders, trailingBuilders...)
}

// InstructionsForPhase injects instructions into the given build phase for builder. The relative
// order of each instruction set is the same as the order of the builders
func (bc BuildersConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}
	for _, builder := range bc {
		instructions = append(instructions, builder.InstructionsForPhase(phase)...)
	}
	return instructions
}

// UnmarshalJSON implements json.Unmarshaler to manually handle different builder types
func (bc *BuildersConfig) UnmarshalJSON(unmarshal []byte) error {
	builderEntries := []builderEntry{}
	dec := json.NewDecoder(bytes.NewReader(unmarshal))
	dec.DisallowUnknownFields()
	err := dec.Decode(&builderEntries)

	builders := make(BuildersConfig, len(builderEntries))
	for i, be := range builderEntries {
		if be.PythonBuilder != nil {
			builders[i] = *be.PythonBuilder
		}
		if be.NodeBuilder != nil {
			builders[i] = *be.NodeBuilder
		}
		if be.PhpBuilder != nil {
			builders[i] = *be.PhpBuilder
		}
		if be.CustomBuilder != nil {
			builders[i] = *be.CustomBuilder
		}
	}
	*bc = builders

	return err
}
