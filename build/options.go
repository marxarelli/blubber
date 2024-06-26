package build

import (
	"context"

	"github.com/containerd/containerd/platforms"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	defaultBuildContext = "context"
	defaultVariant      = "test"
)

// Options stores options to configure the build process.
type Options struct {
	// Function that returns the initial llb.State for the main build context.
	BuildContext ContextResolver

	// The target variant
	Variant string

	// Resolver used to fetch images
	MetaResolver llb.ImageMetaResolver

	// Build-time arguments
	BuildArgs map[string]string

	// Extra labels to add to the result
	Labels map[string]string

	// Build platform
	BuildPlatform oci.Platform

	// Target platforms
	TargetPlatforms []oci.Platform
}

// NewOptions creates a new Options with default values assigned
func NewOptions() *Options {
	defaultPlatform := platforms.DefaultSpec()

	return &Options{
		BuildContext: func(ctx context.Context) (*llb.State, error) {
			localCtx := llb.Local(defaultBuildContext, llb.SharedKeyHint(defaultBuildContext))
			return &localCtx, nil
		},
		BuildPlatform:   defaultPlatform,
		MetaResolver:    imagemetaresolver.Default(),
		TargetPlatforms: []oci.Platform{defaultPlatform},
		Variant:         defaultVariant,
		BuildArgs:       map[string]string{},
		Labels:          map[string]string{},
	}
}

// MultiPlatform returns whether the build options contain multiple target
// platforms.
func (opts *Options) MultiPlatform() bool {
	return len(opts.TargetPlatforms) > 1
}
