package buildkit

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/containerd/containerd/platforms"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/frontend/gateway/client"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

const (
	keyCacheFrom      = "cache-from"    // for registry only. deprecated in favor of keyCacheImports
	keyCacheImports   = "cache-imports" // JSON representation of []CacheOptionsEntry
	keyConfigPath     = "filename"
	keyEntrypointArgs = "entrypoint-args"
	keyRunEntrypoint  = "run-variant"
	keyRunEnvironment = "run-variant-env"
	keyTarget         = "target"
	keyTargetPlatform = "platform"
	keyVariant        = "variant"

	// Support some of the same build-arg: options that buildkit's dockerfile
	// frontend supports, such as setting proxies, etc.
	// e.g. `buildctl ... --opt build-arg:http_proxy=http://foo`
	// See https://github.com/moby/buildkit/blob/81b6ff2c55565bdcb9f0dbcff52515f7c7bb429c/frontend/dockerfile/docs/reference.md#predefined-args
	buildArgPrefix = "build-arg:"
)

// BuildOptions contains options specific to the BuildKit frontend as well as
// options for the [build.Target].
type BuildOptions struct {
	// CacheOptions specifies caches to be imported prior to the build
	CacheOptions []client.CacheOptionsEntry

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
// client options.
func ParseBuildOptions(clientBuildOpts client.BuildOpts) (*BuildOptions, error) {
	opts := clientBuildOpts.Opts
	bo := BuildOptions{Options: build.NewOptions()}

	var err error

	for k, v := range opts {
		switch k {
		case keyConfigPath:
			bo.ConfigPath = opts[k]
		case keyTarget, keyVariant:
			bo.Variant = opts[k]
		case keyTargetPlatform:
			bo.TargetPlatforms, err = parsePlatforms(v)

			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse target platforms: %q", v)
			}
		case keyRunEntrypoint:
			runVariant, err := strconv.ParseBool(v)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to parse run-variant option")
			}
			bo.RunEntrypoint = runVariant
		case keyEntrypointArgs:
			var cmd []string
			err = json.Unmarshal([]byte(v), &cmd)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse extra args for entrypoint: %q", v)
			}
			bo.EntrypointArgs = cmd
		case keyRunEnvironment:
			var env map[string]string
			err = json.Unmarshal([]byte(v), &env)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to parse %s: %q", keyRunEnvironment, v)
			}
			bo.RunEnvironment = env
		}
	}

	bo.CacheOptions, err = parseCacheOptions(opts)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse cache options")
	}

	// Prefer the first worker's platform as the build platform
	if workers := clientBuildOpts.Workers; len(workers) > 0 && len(workers[0].Platforms) > 0 {
		bo.BuildPlatform = &workers[0].Platforms[0]
	}

	bo.SessionID = clientBuildOpts.SessionID
	bo.BuildArgs = filterOpts(opts, buildArgPrefix)

	return &bo, nil
}

// parseCacheOptions handles given cache imports. Note that clients may give
// these options in two different ways, either as `cache-imports` or
// `cache-from`. The latter is used for registry based cache imports.
// See https://github.com/moby/buildkit/blob/v0.10/client/solve.go#L477
//
// TODO the master branch of buildkit removes the legacy `cache-from` key, so
// we should be able to eventually deprecate it, but that will involve
// dropping support for older buildctl and docker buildx clients.
func parseCacheOptions(opts map[string]string) ([]client.CacheOptionsEntry, error) {
	var cacheImports []client.CacheOptionsEntry
	// new API
	if cacheImportsStr := opts[keyCacheImports]; cacheImportsStr != "" {
		var cacheImportsUM []controlapi.CacheOptionsEntry
		if err := json.Unmarshal([]byte(cacheImportsStr), &cacheImportsUM); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal %s (%q)", keyCacheImports, cacheImportsStr)
		}
		for _, um := range cacheImportsUM {
			cacheImports = append(cacheImports, client.CacheOptionsEntry{Type: um.Type, Attrs: um.Attrs})
		}
	}
	// old API
	if cacheFromStr := opts[keyCacheFrom]; cacheFromStr != "" {
		cacheFrom := strings.Split(cacheFromStr, ",")
		for _, s := range cacheFrom {
			im := client.CacheOptionsEntry{
				Type: "registry",
				Attrs: map[string]string{
					"ref": s,
				},
			}
			// FIXME(AkihiroSuda): skip append if already exists
			cacheImports = append(cacheImports, im)
		}
	}

	return cacheImports, nil
}

func parsePlatforms(v string) ([]*oci.Platform, error) {
	var pp []*oci.Platform
	for _, v := range strings.Split(v, ",") {
		p, err := platforms.Parse(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse target platform %s", v)
		}
		p = platforms.Normalize(p)
		pp = append(pp, &p)
	}
	return pp, nil
}

func filterOpts(opts map[string]string, prefix string) map[string]string {
	filtered := map[string]string{}

	for k, v := range opts {
		if strings.HasPrefix(k, prefix) {
			filtered[strings.TrimPrefix(k, prefix)] = v
		}
	}

	return filtered
}
