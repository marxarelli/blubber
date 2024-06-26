package buildkit

import (
	"context"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/frontend/gateway/client"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

const (
	dockerignoreFilename = ".dockerignore"
	configLang           = "YAML"
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
	bc, err := dockerui.NewClient(c)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create dockerui client")
	}

	buildOptions, err := ParseBuildOptions(bc.BuildOpts())

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse build options")
	}

	// Inherit the dockerui client configuration to ensure docker toolchain
	// compatibility.
	buildOptions.BuildArgs = bc.Config.BuildArgs
	buildOptions.Labels = bc.Config.Labels
	buildOptions.TargetPlatforms = bc.Config.TargetPlatforms

	if len(bc.Config.BuildPlatforms) > 0 {
		buildOptions.BuildPlatform = bc.Config.BuildPlatforms[0]
	}

	if bc.Config.Target != "" {
		buildOptions.Variant = bc.Config.Target
	}

	buildOptions.BuildContext = func(ctx context.Context) (*llb.State, error) {
		return bc.MainContext(ctx)
	}

	cfg, err := readBlubberConfig(ctx, bc)

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

	rb, err := bc.Build(
		ctx,
		func(ctx context.Context, platform *oci.Platform, idx int) (
			client.Reference,
			*dockerspec.DockerOCIImage,
			*dockerspec.DockerOCIImage,
			error,
		) {
			target, err := Compile(ctx, buildOptions, cfg, platform)

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to compile target")
			}

			def, img, err := target.Marshal(ctx)

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to marshal target")
			}

			res, err := c.Solve(ctx, client.SolveRequest{
				Definition:   def.ToPB(),
				CacheImports: bc.CacheImports,
			})

			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "failed to solve")
			}

			ref, err := res.SingleRef()
			if err != nil {
				return nil, nil, nil, err
			}

			dimg := dockerspec.DockerOCIImage{
				Image: *img,
				Config: dockerspec.DockerOCIImageConfig{
					ImageConfig: img.Config,
				},
			}

			return ref, &dimg, nil, nil
		},
	)

	return rb.Finalize()
}

func readBlubberConfig(ctx context.Context, bc *dockerui.Client) (*config.Config, error) {
	cfgSrc, err := bc.ReadEntrypoint(ctx, configLang)
	if err != nil {
		return nil, err
	}

	cfg, err := config.ReadYAMLConfig(cfgSrc.Data)
	if err != nil {
		if config.IsValidationError(err) {
			return nil, errors.Wrapf(err, "config is invalid:\n%v", config.HumanizeValidationError(err))
		}

		return nil, errors.Wrap(err, "error reading config")
	}

	return cfg, nil
}
