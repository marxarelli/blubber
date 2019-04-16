package buildkit

import (
	"context"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/gateway/client"

	"gerrit.wikimedia.org/r/blubber/config"
)

const (
	localNameConfig   = "dockerfile"
	keyConfigPath     = "filename"
	keyVariant        = "variant"
	defaultVariant    = "test"
	defaultConfigPath = ".pipeline/blubber.yaml"
)

// Build handles BuildKit client requests for the Blubber gateway.
//
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts

	variant := opts[keyVariant]

	if variant == "" {
		variant = defaultVariant
	}

	cfg, err := readBlubberConfig(ctx, c)

	if err != nil {
		return nil, err
	}

	st, err := CompileToLLB(ctx, cfg, variant)

	if err != nil {
		return nil, err
	}

	def, err := st.Marshal()

	if err != nil {
		return nil, err
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func readBlubberConfig(ctx context.Context, c client.Client) (*config.Config, error) {
	opts := c.BuildOpts().Opts

	configPath := opts[keyConfigPath]
	if configPath == "" {
		configPath = defaultConfigPath
	}

	st := llb.Local(localNameConfig,
		llb.IncludePatterns([]string{configPath}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.SharedKeyHint(defaultConfigPath),
	)

	def, err := st.Marshal()

	if err != nil {
		return nil, err
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})

	if err != nil {
		return nil, err
	}

	ref, err := res.SingleRef()

	if err != nil {
		return nil, err
	}

	cfgBytes, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: configPath,
	})

	if err != nil {
		return nil, err
	}

	cfg, err := config.ReadYAMLConfig(cfgBytes)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
