package buildkit

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"

	"gerrit.wikimedia.org/r/blubber/config"
)

const (
	localNameConfig   = "dockerfile"
	keyConfigPath     = "filename"
	keyTarget         = "target"
	keyVariant        = "variant"
	defaultVariant    = "test"
	defaultConfigPath = ".pipeline/blubber.yaml"

	// Support the dockerfile frontend's build-arg: options which include, but
	// are not limited to, setting proxies.
	// e.g. `buildctl ... --opt build-arg:http_proxy=http://foo`
	// See https://github.com/moby/buildkit/blob/81b6ff2c55565bdcb9f0dbcff52515f7c7bb429c/frontend/dockerfile/docs/reference.md#predefined-args
	buildArgPrefix = "build-arg:"
)

// Build handles BuildKit client requests for the Blubber gateway.
//
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts

	variant := opts[keyVariant]

	if variant == "" {
		variant = opts[keyTarget]
	}

	if variant == "" {
		variant = defaultVariant
	}

	cfg, err := readBlubberConfig(ctx, c)

	if err != nil {
		return nil, errors.Wrap(err, "failed to read blubber config")
	}

	err = config.ExpandIncludesAndCopies(cfg, variant)

	if err != nil {
		return nil, errors.Wrap(err, "failed to expand includes and copies")
	}

	st, image, err := CompileToLLB(ctx, cfg, variant, filterOpts(opts, buildArgPrefix))

	if err != nil {
		return nil, errors.Wrap(err, "failed to compile to LLB state")
	}

	imageConfig, err := json.Marshal(image)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal image config")
	}

	def, err := st.Marshal(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal definition")
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to solve")
	}

	res.AddMeta(exptypes.ExporterImageConfigKey, imageConfig)

	return res, nil
}

func readBlubberConfig(ctx context.Context, c client.Client) (*config.Config, error) {
	opts := c.BuildOpts().Opts

	configPath := opts[keyConfigPath]
	if configPath == "" {
		configPath = defaultConfigPath
	}

	st := llb.Local(localNameConfig,
		llb.FollowPaths([]string{configPath}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint(configPath),
	)

	def, err := st.Marshal(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal definition")
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to solve to load config")
	}

	ref, err := res.SingleRef()

	if err != nil {
		return nil, errors.Wrap(err, "failed to get ")
	}

	cfgBytes, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: configPath,
	})

	if err != nil {
		return nil, err
	}

	cfg, err := config.ReadYAMLConfig(cfgBytes)

	if err != nil {
		if config.IsValidationError(err) {
			return nil, errors.Wrapf(err, "config is invalid:\n%v", config.HumanizeValidationError(err))
		}

		return nil, errors.Wrap(err, "error reading config")
	}

	return cfg, nil
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
