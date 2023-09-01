// Package buildkit implements a compiler for turning Blubber configuration
// into a valid llb.State graph.
package buildkit

import (
	"context"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

// Compile takes a parsed config.Config and a configured variant name and
// returns a compiled build.Target
func Compile(
	ctx context.Context,
	bo *BuildOptions,
	cfg *config.Config,
	platform *oci.Platform,
) (*build.Target, error) {

	variants, err := cfg.CopiesDepGraph.GetDeps(bo.Variant)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get variant dependencies")
	}

	variants = append(variants, bo.Variant)

	targets := build.TargetGroup{}
	vcfgs := make(map[string]*config.VariantConfig, len(variants))

	var finalTarget *build.Target

	for _, variant := range variants {
		vcfg, err := config.GetVariant(cfg, variant)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to get variant %s", variant)
		}

		vcfgs[variant] = vcfg

		finalTarget = targets.NewTarget(variant, vcfg.Base, platform, bo.Options)
	}

	err = targets.InitializeAll(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch base images for some targets")
	}

	for _, target := range targets {
		for _, phase := range build.Phases() {
			for _, instruction := range vcfgs[target.Name].InstructionsForPhase(phase) {
				err := instruction.Compile(target)

				if err != nil {
					// TODO if the build.Instruction interface were expanded to include a
					// method that returned the config from which it originated (and even
					// file/line from the source config), we could provide the user with a
					// nicer error message here
					return nil, errors.Wrap(err, "failed to compile instruction")
				}
			}
		}
	}

	if bo != nil && bo.RunEntrypoint {
		finalTarget.RunEntrypoint(bo.EntrypointArgs, bo.RunEnvironment)
	}

	return finalTarget, nil
}
