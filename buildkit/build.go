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
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
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
// When performing a multi-platform build, the final exported manifest will be
// an OCI image index (aka "fat" manifest) and multiple sub manifests will be
// created for each platform that contain the actual image layers.
//
// See https://github.com/opencontainers/image-spec/blob/main/image-index.md
//
// For a single platform build, the export will be a normal single manifest
// with image layers.
//
// See https://github.com/opencontainers/image-spec/blob/main/manifest.md
//
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	buildOpts := c.BuildOpts()
	opts := buildOpts.Opts

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

	// Defer to dockerfile2llb on the default platform by passing nil
	targetPlatforms := []*ocispecs.Platform{nil}

	// Parse any given target platform(s)
	if platform, exists := opts[keyTargetPlatform]; exists && platform != "" {
		targetPlatforms, err = parsePlatforms(platform)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse target platforms %s", platform)
		}
	}

	isMultiPlatform := len(targetPlatforms) > 1
	exportPlatforms := &exptypes.Platforms{
		Platforms: make([]exptypes.Platform, len(targetPlatforms)),
	}
	finalResult := client.NewResult()

	eg, ctx := errgroup.WithContext(ctx)

	// Solve for target platforms in parallel
	for i, tp := range targetPlatforms {
		func(i int, platform *ocispecs.Platform) {
			eg.Go(func() (err error) {
				st, image, bi, err := CompileToLLB(
					ctx,
					extraOpts,
					cfg,
					variant,
					d2llb.ConvertOpt{
						MetaResolver:   c,
						SessionID:      buildOpts.SessionID,
						BuildArgs:      filterOpts(opts, buildArgPrefix),
						Excludes:       excludes,
						TargetPlatform: platform,
						PrefixPlatform: isMultiPlatform,
					},
				)

				if err != nil {
					return errors.Wrap(err, "failed to compile to LLB state")
				}

				imageConfig, err := json.Marshal(image)

				if err != nil {
					return errors.Wrapf(err, "failed to marshal image config")
				}

				def, err := st.Marshal(ctx)

				if err != nil {
					return errors.Wrap(err, "failed to marshal definition")
				}

				result, err := c.Solve(ctx, client.SolveRequest{
					Definition: def.ToPB(),
				})

				if err != nil {
					return errors.Wrap(err, "failed to solve")
				}

				ref, err := result.SingleRef()
				if err != nil {
					return err
				}

				buildinfo, err := json.Marshal(bi)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal build info")
				}

				if !isMultiPlatform {
					finalResult.AddMeta(exptypes.ExporterImageConfigKey, imageConfig)
					finalResult.AddMeta(exptypes.ExporterBuildInfo, buildinfo)
					finalResult.SetRef(ref)
				} else {
					p := platforms.DefaultSpec()
					if platform != nil {
						p = *platform
					}

					k := platforms.Format(p)
					finalResult.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, k), imageConfig)
					finalResult.AddMeta(fmt.Sprintf("%s/%s", exptypes.ExporterBuildInfo, k), buildinfo)
					finalResult.AddRef(k, ref)
					exportPlatforms.Platforms[i] = exptypes.Platform{
						ID:       k,
						Platform: p,
					}
				}
				return nil
			})
		}(i, tp)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if isMultiPlatform {
		dt, err := json.Marshal(exportPlatforms)
		if err != nil {
			return nil, err
		}
		finalResult.AddMeta(exptypes.ExporterPlatformsKey, dt)
	}

	return finalResult, nil
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

func parsePlatforms(v string) ([]*ocispecs.Platform, error) {
	var pp []*ocispecs.Platform
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
