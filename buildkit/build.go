package buildkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/moby/buildkit/frontend/gateway/client"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

const (
	dockerignoreFilename = ".dockerignore"
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
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	buildOptions, err := ParseBuildOptions(c.BuildOpts())

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse build options")
	}

	cfg, err := readBlubberConfig(ctx, c, buildOptions)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to read blubber config")
	}

	err = config.ExpandIncludesAndCopies(cfg, buildOptions.Variant)

	if err != nil {
		if config.IsValidationError(err) {
			err = errors.New(config.HumanizeValidationError(err))
		}
		return nil, errors.Wrap(err, "failed to expand includes and copies")
	}

	buildOptions.Excludes, err = readDockerExcludes(ctx, c, buildOptions)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(`failed to read "%s"`, dockerignoreFilename))
	}

	exportPlatforms := &exptypes.Platforms{
		Platforms: make([]exptypes.Platform, len(buildOptions.TargetPlatforms)),
	}
	finalResult := client.NewResult()

	eg, ctx := errgroup.WithContext(ctx)

	// Solve for all target platforms in parallel
	for i, tp := range buildOptions.TargetPlatforms {
		func(i int, platform *oci.Platform) {
			eg.Go(func() (err error) {
				result, err := buildImage(ctx, c, buildOptions, cfg, platform)

				if err != nil {
					return errors.Wrap(err, "failed to build image")
				}

				result.AddToClientResult(finalResult)
				exportPlatforms.Platforms[i] = *result.ExportPlatform

				return nil
			})
		}(i, tp)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if buildOptions.MultiPlatform() {
		dt, err := json.Marshal(exportPlatforms)
		if err != nil {
			return nil, err
		}
		finalResult.AddMeta(exptypes.ExporterPlatformsKey, dt)
	}

	return finalResult, nil
}

// Represents the result of a single image build
type buildResult struct {
	// Reference to built image
	Reference client.Reference

	// Image metadata and runtime config
	ImageConfig []byte

	// Target platform
	Platform *oci.Platform

	// Whether this is a result for a multi-platform build
	MultiPlatform bool

	// Exportable platform information (platform and platform ID)
	ExportPlatform *exptypes.Platform
}

// Adds the build result to the final client result.
//
// For multi-platform builds, _add_ to an aggregate final result. Each
// platform-specific image reference, image config, and build info will be
// keyed by platform name. For OCI format outputs, there will be a single
// multi-platform manifest _index_ which references each platform-specific
// manifest.
//
// For single-platform builds, set the final result's reference,
// image config, and build info. For OCI format outputs, there will
// be a single manifest
func (br *buildResult) AddToClientResult(cr *client.Result) {
	if br.MultiPlatform {
		cr.AddMeta(
			fmt.Sprintf("%s/%s", exptypes.ExporterImageConfigKey, br.ExportPlatform.ID),
			br.ImageConfig,
		)
		cr.AddRef(br.ExportPlatform.ID, br.Reference)
	} else {
		cr.AddMeta(exptypes.ExporterImageConfigKey, br.ImageConfig)
		cr.SetRef(br.Reference)
	}
}

// Builds a given variant and returns the resulting image reference, image
// config, and build info.
func buildImage(
	ctx context.Context,
	c client.Client,
	buildOpts *BuildOptions,
	cfg *config.Config,
	targetPlatform *oci.Platform,
) (*buildResult, error) {

	result := buildResult{
		Platform:      targetPlatform,
		MultiPlatform: buildOpts.MultiPlatform(),
	}

	target, err := Compile(ctx, buildOpts, cfg, targetPlatform)

	if err != nil {
		return nil, errors.Wrap(err, "failed to compile target")
	}

	def, imgConfig, err := target.Marshal(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal target")
	}

	result.ImageConfig = imgConfig

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition:   def.ToPB(),
		CacheImports: buildOpts.CacheOptions,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to solve")
	}

	result.Reference, err = res.SingleRef()
	if err != nil {
		return nil, err
	}

	// Add platform-specific export info for the result that can later be used
	// in multi-platform results
	result.ExportPlatform = &exptypes.Platform{
		Platform: *targetPlatform,
		ID:       platforms.Format(*targetPlatform),
	}

	return &result, nil
}

func readBlubberConfig(ctx context.Context, c client.Client, opts *BuildOptions) (*config.Config, error) {
	cfgBytes, err := readFileFromLocal(ctx, c, opts.ClientConfigContext, opts.ConfigPath, true)
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

func readDockerExcludes(ctx context.Context, c client.Client, opts *BuildOptions) ([]string, error) {
	dockerignoreBytes, err := readFileFromLocal(ctx, c, opts.ClientBuildContext, dockerignoreFilename, false)
	if err != nil {
		return nil, err
	}

	excludes, err := dockerignore.ReadAll(bytes.NewBuffer(dockerignoreBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dockerignore")
	}

	return excludes, nil
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
		llb.SharedKeyHint(localCtx+"-"+filepath),
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
