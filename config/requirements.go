package config

import (
	"encoding/json"
	"errors"
	"path"

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

			// Sanitize Source and Destination paths
			srcFile := path.Clean(artifact.Source)
			destDir := path.Clean(artifact.Destination)
			destLen := len(artifact.Destination)

			if artifact.Destination == "" {
				// Preserve legacy behavior from build.SyncFiles for implicit
				// destinations. The legacy behavior is to copy the file to
				// a path matching the path it was imported from. Long form
				// configuration allows the user to override this.
				destDir = path.Dir(srcFile) + "/"
			} else if artifact.Destination[destLen-1:] == "/" {

				destDir = destDir + "/"
			}

			_, haveSeenDest := artifacts[artifact.From][destDir]
			if !haveSeenDest {
				// First time seeing this Destination for this From:
				// - remeber the order it was seen in the config
				destOrder[artifact.From] = append(
					destOrder[artifact.From],
					destDir,
				)
				// - make a slice to track related Source values
				artifacts[artifact.From][destDir] = []string{}
			}

			artifacts[artifact.From][destDir] = append(
				artifacts[artifact.From][destDir],
				srcFile,
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

// NewFromShort creates a legacy short form requirements artifact.
//
func NewFromShort(source string) ArtifactsConfig {
	return ArtifactsConfig{
		From:   LocalArtifactKeyword,
		Source: source,
		// Preserve legacy behavior from build.SyncFiles for implicit
		// destinations. The legacy behavior is to copy the file to
		// a path matching the path it was imported from. Long form
		// configuration allows the user to override this.
		Destination: path.Dir(path.Clean(source)) + "/",
	}
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
			(*rc)[i] = NewFromShort(source)
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
			longhand[i] = NewFromShort(source)
		}
	}
	*rc = RequirementsConfig(longhand)

	return nil
}
