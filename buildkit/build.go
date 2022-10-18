package buildkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	d2llb "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"

	"gerrit.wikimedia.org/r/blubber/config"
)

const (
	localNameConfig      = "dockerfile"
	localNameContext     = "context"
	keyConfigPath        = "filename"
	keyTarget            = "target"
	keyTargetPlatform    = "platform"
	keyVariant           = "variant"
	defaultVariant       = "test"
	defaultConfigPath    = ".pipeline/blubber.yaml"
	dockerignoreFilename = ".dockerignore"

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

	extraOpts, err := ParseExtraOptions(opts)
	if err != nil {
		return nil, err
	}
	cfg, err := readBlubberConfig(ctx, c)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to read blubber config")
	}

	err = config.ExpandIncludesAndCopies(cfg, variant)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to expand includes and copies")
	}

	excludes, err := readDockerExcludes(ctx, c)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(`failed to read "%s"`, dockerignoreFilename))
	}

	convertOpts := d2llb.ConvertOpt{
		BuildArgs: filterOpts(opts, buildArgPrefix),
		Excludes:  excludes,
	}

	if platform, exists := opts[keyTargetPlatform]; exists && platform != "" {
		p, err := platforms.Parse(platform)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse target platform %s", platform)
		}
		p = platforms.Normalize(p)
		convertOpts.TargetPlatform = &p
	}

	st, image, err := CompileToLLB(ctx, extraOpts, cfg, variant, convertOpts)

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

	cfgBytes, err := readFileFromLocal(ctx, c, localNameConfig, configPath, true)
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

func readDockerExcludes(ctx context.Context, c client.Client) ([]string, error) {
	dockerignoreBytes, err := readFileFromLocal(ctx, c, localNameContext, dockerignoreFilename, false)
	if err != nil {
		return nil, err
	}

	excludes, err := dockerignore.ReadAll(bytes.NewBuffer(dockerignoreBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dockerignore")
	}

	return excludes, nil
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

func readFileFromLocal(
	ctx context.Context,
	c client.Client,
	localCtx string,
	filepath string,
	required bool,
) ([]byte, error) {
	st := llb.Local(localCtx,
		llb.SessionID(c.BuildOpts().SessionID),
		llb.FollowPaths([]string{filepath}),
		llb.SharedKeyHint(filepath),
	)

	def, err := st.Marshal(ctx)
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

	// If the file is not required, try to stat it first, and if it doesn't
	// exist, simply return an empty byte slice. If the file is required, we'll
	// save an extra stat call and just try to read it.
	if !required {
		_, err := ref.StatFile(ctx, client.StatRequest{
			Path: filepath,
		})

		if err != nil {
			return []byte{}, nil
		}
	}

	fileBytes, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: filepath,
	})

	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
