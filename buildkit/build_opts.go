package buildkit

import (
	"encoding/json"
	"github.com/pkg/errors"
	"strconv"
)

const runEntrypointKey = "run-variant"
const entrypointArgsKey = "entrypoint-args"

// ExtraBuildOptions stores options to configure the build process. These are not BuildKit options,
// but additional configuration implemented by the BuildKit frontend
type ExtraBuildOptions struct {
	// Whether to run the target's variant entrypoint during the build process. The entrypoint
	// command will be executed by BuildKit while creating the image
	runEntrypoint bool
	// Additional arguments to be added to the entrypoint command
	entrypointArgs []string
}

// RunEntrypoint returns ebo.runEntrypoint
func (ebo *ExtraBuildOptions) RunEntrypoint() bool {
	return ebo.runEntrypoint
}

// EntrypointArgs returns ebo.entrypointArgs
func (ebo *ExtraBuildOptions) EntrypointArgs() []string {
	return ebo.entrypointArgs
}

// ParseExtraOptions parses and returns a newly created ExtraBuildOption
func ParseExtraOptions(ops map[string]string) (*ExtraBuildOptions, error) {
	var ebo ExtraBuildOptions
	var err error

	for k, v := range ops {
		switch k {
		case runEntrypointKey:
			runVariant, err := strconv.ParseBool(v)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to parse run-variant option")
			}
			ebo.runEntrypoint = runVariant
		case entrypointArgsKey:
			var cmd []string
			err = json.Unmarshal([]byte(v), &cmd)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse extra args for entrypoint: %q", v)
			}
			ebo.entrypointArgs = cmd
		}
	}

	return &ebo, nil
}
