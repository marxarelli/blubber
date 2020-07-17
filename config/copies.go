package config

import (
	"encoding/json"

	"gerrit.wikimedia.org/r/blubber/build"
)

// LocalArtifactKeyword defines a special keyword indicating
// file/directory artifacts to be copied from the local build host context.
//
const LocalArtifactKeyword = "local"

// CopiesConfig holds configuration for which files to copy into the variant
// from local and other variant sources.
//
type CopiesConfig []ArtifactsConfig

// Expand returns a version of this CopiesConfig with its shorthand
// configurations expanded.
//
func (cc *CopiesConfig) Expand(appDirectory string) CopiesConfig {
	expanded := CopiesConfig{}

	// expand all artifact definitions
	for _, artifact := range *cc {
		expanded = append(expanded, artifact.Expand(appDirectory)...)
	}

	return expanded
}

// InstructionsForPhase delegates to its member ArtifactsConfig.
//
func (cc CopiesConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	for _, artifact := range cc {
		instructions = append(instructions, artifact.InstructionsForPhase(phase)...)
	}

	return instructions
}

// Merge takes another CopiesConfig and overwrites this struct's fields.
//
// Artifacts are merged additively and duplicates are removed. Uniqueness is
// ensured by taking the latest definition over the previous.
//
func (cc *CopiesConfig) Merge(cc2 CopiesConfig) {
	// efficient search of the other CopiesConfig using a map
	newlyDefined := make(map[ArtifactsConfig]bool, len(*cc))
	for _, artifact := range cc2 {
		newlyDefined[artifact] = true
	}

	dupesRemoved := CopiesConfig{}

	// omit any previously defined artifacts that are among the new
	for _, artifact := range *cc {
		if !newlyDefined[artifact] {
			dupesRemoved = append(dupesRemoved, artifact)
		}
	}

	*cc = append(dupesRemoved, cc2...)
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand and
// longhand copies configuration.
//
func (cc *CopiesConfig) UnmarshalJSON(unmarshal []byte) error {
	shorthand := []string{}

	err := json.Unmarshal(unmarshal, &shorthand)

	if err == nil {
		*cc = make(CopiesConfig, len(shorthand))

		for i, variant := range shorthand {
			(*cc)[i] = ArtifactsConfig{From: variant}
		}

		return nil
	}

	longhand := []ArtifactsConfig{}

	err = json.Unmarshal(unmarshal, &longhand)

	if err == nil {
		*cc = CopiesConfig(longhand)

		return nil
	}

	return err
}
