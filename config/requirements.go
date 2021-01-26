package config

import (
	"encoding/json"
	"errors"

	"gerrit.wikimedia.org/r/blubber/build"
)

// RequirementsConfig holds configuration for which files to copy into the
// variant from local and other variant sources.
//
type RequirementsConfig []ArtifactsConfig

// InstructionsForPhase injects instructions into the given build phase that
// copy configured artifacts.
//
// PhasePreInstall
//
// In the case of a "local" build context copy, simply return a build.Copy
// with the configured source and destination. In the case of a variant copy,
// return a build.CopyFrom instruction for the variant name, source and
// destination paths.
func (rc RequirementsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	instructions := []build.Instruction{}

	switch phase {
	case build.PhasePreInstall:
		// Map of artifacts grouped by From and Destination
		artifacts := map[string]map[string][]string{}
		// Set of From values in input order
		fromOrder := []string{}
		// Map of sets of Destination values in input order grouped by From
		destOrder := map[string][]string{}

		for _, artifact := range rc {
			_, haveSeenFrom := artifacts[artifact.From]
			if !haveSeenFrom {
				// First time seeing this From:
				// - remember the order it was seen in the config
				fromOrder = append(fromOrder, artifact.From)
				// - make a slice to track related Destination values
				destOrder[artifact.From] = []string{}
				// - make a map to track discovered Source values grouped by
				// Destination
				artifacts[artifact.From] = map[string][]string{}
			}

			src := artifact.NormalizedSource()
			dest := artifact.NormalizedDestination()

			_, haveSeenDest := artifacts[artifact.From][dest]
			if !haveSeenDest {
				// First time seeing this Destination for this From:
				// - remeber the order it was seen in the config
				destOrder[artifact.From] = append(
					destOrder[artifact.From],
					dest,
				)
				// - make a slice to track related Source values
				artifacts[artifact.From][dest] = []string{}
			}

			artifacts[artifact.From][dest] = append(
				artifacts[artifact.From][dest],
				src,
			)
		}

		for _, from := range fromOrder {
			for _, dest := range destOrder[from] {
				copy := build.Copy{artifacts[from][dest], dest}
				if from == LocalArtifactKeyword || from == "" {
					instructions = append(instructions, copy)
				} else {
					instructions = append(
						instructions,
						build.CopyFrom{from, copy},
					)
				}
			}
		}
	}

	return instructions
}

// IsUnmarshalTypeError returns true if the provided error is of type
// json.UnmarshalTypeError.
//
func IsUnmarshalTypeError(err error) bool {
	_, ok := err.(*json.UnmarshalTypeError)
	return ok
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand and
// longhand requirements configuration.
//
func (rc *RequirementsConfig) UnmarshalJSON(unmarshal []byte) error {
	shorthand := []string{}
	err := json.Unmarshal(unmarshal, &shorthand)

	if err == nil {
		// Input was entirely in short form
		*rc = make(RequirementsConfig, len(shorthand))

		for i, source := range shorthand {
			(*rc)[i] = NewArtifactsConfigFromSource(source)
		}

		return nil
	}

	// We treat UnmarshalTypeError as a soft error. It means that some part of
	// the input could not be matched to the target interface. Other errors
	// indicate severly malformed input, so we will propigate the error.
	if !IsUnmarshalTypeError(err) {
		return err
	}

	longhand := []ArtifactsConfig{}
	err = json.Unmarshal(unmarshal, &longhand)

	if err == nil {
		// Input was entirely in long form
		*rc = RequirementsConfig(longhand)

		return nil
	}

	if !IsUnmarshalTypeError(err) {
		return err
	}

	if len(shorthand) != len(longhand) {
		return errors.New("mismatched unmarshal results")
	}

	// Input was mixed short and long form. Walk the short form results and
	// turn any non-empty strings into ArtifactsConfig values in the same slot
	// of the long form results.
	for i, source := range shorthand {
		if source != "" {
			longhand[i] = NewArtifactsConfigFromSource(source)
		}
	}
	*rc = RequirementsConfig(longhand)

	return nil
}
