package build

import (
	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	defaultBuildContext  = "context"
	defaultConfigContext = "dockerfile"
	defaultVariant       = "test"
	defaultConfigPath    = ".pipeline/blubber.yaml"
)

// Options stores options to configure the build process.
type Options struct {
	// Name of the client's local context that contains the source files
	ClientBuildContext string

	// Name of the client's local context that contains the build config
	// (blubber.yaml)
	ClientConfigContext string

	// Path to the build config, relative to the ConfigContext
	ConfigPath string

	// The target variant
	Variant string

	// Resolver used to fetch images
	MetaResolver llb.ImageMetaResolver

	// Build-time arguments
	BuildArgs map[string]string

	// Extra labels to add to the result
	Labels map[string]string

	// Session ID
	SessionID string

	// Files to be excluded from local context copy operations
	Excludes []string

	// Build platform
	BuildPlatform *oci.Platform

	// Target platforms
	TargetPlatforms []*oci.Platform
}

// NewOptions creates a new Options with default values assigned
func NewOptions() *Options {
	defaultPlatform := platforms.DefaultSpec()

	return &Options{
		ClientConfigContext: defaultConfigContext,
		ClientBuildContext:  defaultBuildContext,
		BuildPlatform:       &defaultPlatform,
		ConfigPath:          defaultConfigPath,
		MetaResolver:        imagemetaresolver.Default(),
		TargetPlatforms:     []*oci.Platform{&defaultPlatform},
		Variant:             defaultVariant,
		BuildArgs:           map[string]string{},
		Labels:              map[string]string{},
		Excludes:            []string{},
	}
}

// MultiPlatform returns whether the build options contain multiple target
// platforms.
func (opts *Options) MultiPlatform() bool {
	return len(opts.TargetPlatforms) > 1
}
