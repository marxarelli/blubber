package buildkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/containerd/containerd/platforms"
	controlapi "github.com/moby/buildkit/api/services/control"
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
	keyCacheFrom         = "cache-from"    // for registry only. deprecated in favor of keyCacheImports
	keyCacheImports      = "cache-imports" // JSON representation of []CacheOptionsEntry
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

	// Parse cache imports
	cacheImports, err := parseCacheOptions(opts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cache import options")
	}

	// Default the build platform to the buildkit host's os/arch
	defaultBuildPlatform := platforms.DefaultSpec()

	// But prefer the first worker's platform
	if workers := c.BuildOpts().Workers; len(workers) > 0 && len(workers[0].Platforms) > 0 {
		defaultBuildPlatform = workers[0].Platforms[0]
	}

	buildPlatforms := []ocispecs.Platform{defaultBuildPlatform}

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

	// Solve for all target platforms in parallel
	for i, tp := range targetPlatforms {
		func(i int, platform *ocispecs.Platform) {
			eg.Go(func() (err error) {
				result, err := buildImage(
					ctx,
					c,
					extraOpts,
					cfg,
					variant,
					d2llb.ConvertOpt{
						MetaResolver:   c,
						SessionID:      buildOpts.SessionID,
						BuildArgs:      filterOpts(opts, buildArgPrefix),
						Excludes:       excludes,
						BuildPlatforms: buildPlatforms,
						TargetPlatform: platform,
						PrefixPlatform: isMultiPlatform,
					},
					cacheImports,
				)

				if err != nil {
					return errors.Wrap(err, "failed to build image")
				}

				result.AddToClientResult(finalResult)
				exportPlatforms.Platforms[i] = result.ExportPlatform

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

// Represents the result of a single image build
type buildResult struct {
	// Reference to built image
	Reference client.Reference

	// Image configuration
	ImageConfig []byte

	// Extra build info
	BuildInfo []byte

	// Target platform
	Platform *ocispecs.Platform

	// Whether this is a result for a multi-platform build
	MultiPlatform bool

	// Exportable platform information (platform and platform ID)
	ExportPlatform exptypes.Platform
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
		cr.AddMeta(
			fmt.Sprintf("%s/%s", exptypes.ExporterBuildInfo, br.ExportPlatform.ID),
			br.BuildInfo,
		)
		cr.AddRef(br.ExportPlatform.ID, br.Reference)
	} else {
		cr.AddMeta(exptypes.ExporterImageConfigKey, br.ImageConfig)
		cr.AddMeta(exptypes.ExporterBuildInfo, br.BuildInfo)
		cr.SetRef(br.Reference)
	}
}

// Builds a given variant and returns the resulting image reference, image
// config, and build info.
func buildImage(
	ctx context.Context,
	c client.Client,
	ebo *ExtraBuildOptions,
	cfg *config.Config,
	variant string,
	convertOpts d2llb.ConvertOpt,
	cacheImports []client.CacheOptionsEntry,
) (*buildResult, error) {

	result := buildResult{
		Platform:      convertOpts.TargetPlatform,
		MultiPlatform: convertOpts.PrefixPlatform,
	}

	state, image, bi, err := CompileToLLB(ctx, ebo, cfg, variant, convertOpts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to compile to LLB state")
	}

	result.ImageConfig, err = json.Marshal(image)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal image config")
	}

	def, err := state.Marshal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal definition")
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition:   def.ToPB(),
		CacheImports: cacheImports,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to solve")
	}

	result.Reference, err = res.SingleRef()
	if err != nil {
		return nil, err
	}

	result.BuildInfo, err = json.Marshal(bi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal build info")
	}

	// Add platform-specific export info for the result that can later be used
	// in multi-platform results
	result.ExportPlatform = exptypes.Platform{
		Platform: platforms.DefaultSpec(),
	}

	if result.Platform != nil {
		result.ExportPlatform.Platform = *result.Platform
	}

	result.ExportPlatform.ID = platforms.Format(result.ExportPlatform.Platform)

	return &result, nil
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

// parseCacheOptions handles given cache imports. Note that clients may give
// these options in two different ways, either as `cache-imports` or
// `cache-from`. The latter is used for registry based cache imports.
// See https://github.com/moby/buildkit/blob/v0.10/client/solve.go#L477
//
// TODO the master branch of buildkit removes the legacy `cache-from` key, so
// once they cut a new minor version, we can remove support for it.
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
