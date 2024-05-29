package testtarget

import (
	"context"
	"encoding/json"
	"testing"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/llbtest"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testmetaresolver"
)

// TargetFn represents a [build.Target] callback. See [Setup]
type TargetFn func(*build.Target)

// NewTarget returns a boilerlplate [build.Target] for use in tests.
func NewTarget(name string) *build.Target {
	group := NewTargets(name)
	return group[0]
}

// NewTargets returns a boilerlplate [build.TargetGroup] for use in tests.
func NewTargets(names ...string) build.TargetGroup {
	return NewTargetsWithBaseImage(
		names,
		oci.Image{
			Config: oci.ImageConfig{
				User:       "root",
				Env:        []string{},
				WorkingDir: "/srv/app",
			},
		},
	)
}

// NewTargetsWithBaseImage returns a boilerlplate [build.TargetGroup] for use in
// tests and uses the given oci.Image as the resolved base image.
func NewTargetsWithBaseImage(names []string, baseImage oci.Image) build.TargetGroup {
	var group build.TargetGroup

	for _, name := range names {
		baseImageRef := "testtarget.test/base/" + name

		options := build.NewOptions()
		options.MetaResolver = testmetaresolver.New(
			baseImageRef,
			baseImage,
		)

		group.NewTarget(
			name,
			baseImageRef,
			nil,
			options,
		)
	}

	return group
}

// Setup is a test helper that takes a given [build.TargetGroup],
// intializes it, and calls each given function with the corresponding target
// in the group. Returns an unmarshaled [oci.Image] and [Assertions] for the
// last target for additional assertions.
func Setup(
	t *testing.T,
	targets build.TargetGroup,
	fn ...TargetFn) (*oci.Image, *Assertions) {

	t.Helper()

	ctx := context.TODO()
	req := require.New(t)

	req.Positive(len(targets))
	req.NoError(targets.InitializeAll(ctx))

	for i, target := range targets {
		if i < len(fn) {
			fn[i](target)
		}
	}

	def, imageJSON, err := targets[len(targets)-1].Marshal(ctx)
	req.NoError(err)

	var image oci.Image
	err = json.Unmarshal(imageJSON, &image)
	req.NoError(err)

	return &image, &Assertions{
		LLBAssertions: llbtest.New(t, def),
		Assertions:    require.New(t),
		t:             t,
		target:        targets[len(targets)-1],
	}
}

// Compile is a simplified version of [Setup] without the callback that simply
// compiles the given [build.Instruction] to the first of the given targets.
func Compile(
	t *testing.T,
	targets build.TargetGroup,
	ins build.Instruction) (*oci.Image, *Assertions) {

	t.Helper()

	return Setup(
		t,
		targets,
		func(target *build.Target) {
			require.NoError(t, ins.Compile(target))
		},
	)
}
