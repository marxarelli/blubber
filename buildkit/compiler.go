// Package buildkit implements a compiler for turning Blubber configuration
// into a valid llb.State graph.
//
package buildkit

import (
	"bufio"
	"bytes"
	"context"
	"github.com/moby/buildkit/client/llb"
	d2llb "github.com/moby/buildkit/frontend/dockerfile/dockerfile2llb"
	binfotypes "github.com/moby/buildkit/util/buildinfo/types"

	"gitlab.wikimedia.org/repos/releng/blubber/config"
	"gitlab.wikimedia.org/repos/releng/blubber/docker"
)

// CompileToLLB takes a parsed config.Config and a configured variant name and
// returns an llb.State graph.
//
func CompileToLLB(
	ctx context.Context,
	ebo *ExtraBuildOptions,
	cfg *config.Config,
	variant string,
	convertOpts d2llb.ConvertOpt,
) (*llb.State, *d2llb.Image, *binfotypes.BuildInfo, error) {
	buffer, err := docker.Compile(cfg, variant)

	if err != nil {
		return nil, nil, nil, err
	}

	state, image, buildinfo, err := d2llb.Dockerfile2LLB(ctx, buffer.Bytes(), convertOpts)

	if err != nil {
		return nil, nil, nil, err
	}

	targetVariant := cfg.Variants[variant]
	state = postProcessLLB(ebo, state, &targetVariant)

	return state, image, buildinfo, nil
}

// Compile takes a parsed config.Config and a configured variant name and
// returns an llb.State graph as JSON.
//
func Compile(cfg *config.Config, variant string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	ctx := context.Background()

	state, _, _, err := CompileToLLB(ctx, nil, cfg, variant, d2llb.ConvertOpt{})

	if err != nil {
		return nil, err
	}

	def, err := state.Marshal(ctx)

	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(buffer)
	err = llb.WriteTo(def, writer)

	if err != nil {
		return nil, err
	}

	err = writer.Flush()

	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func postProcessLLB(
	ebo *ExtraBuildOptions,
	state *llb.State,
	targetVariant *config.VariantConfig,
) *llb.State {
	newState := *state

	entryPoint := targetVariant.EntryPoint
	if entryPoint != nil && ebo != nil && ebo.RunEntrypoint() {
		newState = state.Run(
			llb.Args(append(entryPoint, ebo.EntrypointArgs()...)),
			disableCacheForOp(),
		).Root()
	}

	return &newState
}

type runOptionFunc func(*llb.ExecInfo)

func (fn runOptionFunc) SetRunOption(ei *llb.ExecInfo) {
	fn(ei)
}

func disableCacheForOp() llb.RunOption {
	return runOptionFunc(func(ei *llb.ExecInfo) {
		ei.Constraints.Metadata.IgnoreCache = true
	})
}
