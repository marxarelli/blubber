package buildkit

import (
	"encoding/json"
	"strconv"

	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

const (
	keyVariant        = "variant"
	keyEntrypointArgs = "entrypoint-args"
	keyRunEntrypoint  = "run-variant"
	keyRunEnvironment = "run-variant-env"
)

// BuildOptions contains options specific to the BuildKit frontend as well as
// options for the [build.Target].
type BuildOptions struct {
	// Whether to run the target's variant entrypoint during the build process. The entrypoint
	// command will be executed by BuildKit while creating the image
	RunEntrypoint bool

	// Environment variables to use when running the entrypoint.
	RunEnvironment map[string]string

	// Additional arguments to be added to the entrypoint command
	EntrypointArgs []string

	*build.Options
}

// ParseBuildOptions parses and returns a newly created BuildOptions from the given
// build options.
func ParseBuildOptions(clientBuildOpts client.BuildOpts) (*BuildOptions, error) {
	bo := BuildOptions{Options: build.NewOptions()}

	for k, v := range clientBuildOpts.Opts {
		switch k {
		case keyVariant:
			bo.Variant = v
		case keyRunEntrypoint:
			runVariant, err := strconv.ParseBool(v)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to parse run-variant option")
			}
			bo.RunEntrypoint = runVariant
		case keyEntrypointArgs:
			var cmd []string
			err := json.Unmarshal([]byte(v), &cmd)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse extra args for entrypoint: %q", v)
			}
			bo.EntrypointArgs = cmd
		case keyRunEnvironment:
			var env map[string]string
			err := json.Unmarshal([]byte(v), &env)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse %s: %q", keyRunEnvironment, v)
			}
			bo.RunEnvironment = env
		}
	}

	return &bo, nil
}
