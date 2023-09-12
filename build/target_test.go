package build_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/llbtest"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testmetaresolver"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testtarget"
)

func TestInitialize(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)

	options := build.NewOptions()
	options.BuildPlatform = &oci.Platform{OS: "linux", Architecture: "amd64"}
	options.TargetPlatforms = []*oci.Platform{{OS: "linux", Architecture: "arm64"}}
	options.MetaResolver = testmetaresolver.New(
		"docker-registry.wikimedia.org/foo/base",
		oci.Image{
			Config: oci.ImageConfig{
				User:       "root",
				Env:        []string{"FOO=bar"},
				WorkingDir: "/base/workdir",
				Entrypoint: []string{"/base/entry"},
				Cmd:        []string{"arg1", "arg2"},
				Labels:     map[string]string{"bar.label": "bar"},
			},
		},
	)
	options.Labels["foo.label"] = "foo"

	target := build.NewTarget("foo", "docker-registry.wikimedia.org/foo/base", nil, options)

	req.NoError(target.Initialize(ctx))

	// add a run op to test environment variables, etc.
	target.RunShell("foo")

	def, imageJSON, err := target.Marshal(ctx)
	req.NoError(err)

	llbreq := llbtest.New(t, def)

	var image oci.Image
	req.NoError(json.Unmarshal(imageJSON, &image))

	// base image config should have been inherited
	req.Equal("linux", image.OS)
	req.Equal("arm64", image.Architecture)
	req.Equal("root", image.Config.User)
	req.Equal([]string{"FOO=bar"}, image.Config.Env)
	req.Equal("/base/workdir", image.Config.WorkingDir)
	req.Equal([]string{"/base/entry"}, image.Config.Entrypoint)
	req.Equal([]string{"arg1", "arg2"}, image.Config.Cmd)
	req.Equal(map[string]string{"foo.label": "foo", "bar.label": "bar"}, image.Config.Labels)

	// Assert the correctness of the returned LLB ops. There should be a source
	// op for the base image, and an exec op for the run we added with
	// environment variables, working directory, and user all effected by the
	// base image config. Environment variables should also contain correct
	// information about the build/target platforms
	_, sourceOps := llbreq.ContainsNSourceOps(1)

	req.Equal(
		"docker-image://docker-registry.wikimedia.org/foo/base@sha256:91ec0ba821e5708c27e18b22cda1e7647bef0091a386102237614387de2c7c77",
		sourceOps[0].Source.Identifier,
	)

	_, execOps := llbreq.ContainsNExecOps(1)

	req.Equal([]string{"/bin/sh", "-c", "foo"}, execOps[0].Exec.Meta.Args)

	req.Contains(execOps[0].Exec.Meta.Env, "FOO=bar")
	req.Contains(execOps[0].Exec.Meta.Env, "TARGETPLATFORM=linux/arm64")
	req.Contains(execOps[0].Exec.Meta.Env, "TARGETOS=linux")
	req.Contains(execOps[0].Exec.Meta.Env, "TARGETARCH=arm64")
	req.Contains(execOps[0].Exec.Meta.Env, "TARGETVARIANT=")
	req.Contains(execOps[0].Exec.Meta.Env, "BUILDPLATFORM=linux/amd64")
	req.Contains(execOps[0].Exec.Meta.Env, "BUILDOS=linux")
	req.Contains(execOps[0].Exec.Meta.Env, "BUILDARCH=amd64")
	req.Contains(execOps[0].Exec.Meta.Env, "BUILDVARIANT=")

	req.Equal(execOps[0].Exec.Meta.Cwd, "/base/workdir")
	req.Equal(execOps[0].Exec.Meta.User, "root")
}

func TestBuildEnv(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)

	target := testtarget.NewTarget("foo")
	target.Options.BuildPlatform = &oci.Platform{
		OS:           "linux",
		Architecture: "amd64",
	}
	target.Options.TargetPlatforms = []*oci.Platform{
		{OS: "linux", Architecture: "arm64", Variant: "v8"},
	}

	req.NoError(target.Initialize(ctx))

	req.Equal(
		map[string]string{
			"BUILDPLATFORM":  "linux/amd64",
			"BUILDOS":        "linux",
			"BUILDARCH":      "amd64",
			"BUILDVARIANT":   "",
			"TARGETPLATFORM": "linux/arm64/v8",
			"TARGETOS":       "linux",
			"TARGETARCH":     "arm64",
			"TARGETVARIANT":  "v8",
		},
		target.BuildEnv(),
	)
}

func TestExposeBuildArg(t *testing.T) {
	t.Run("tries build args from options first", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				target.Options.BuildArgs["FOOARG"] = "bar"
				target.ExposeBuildArg("FOOARG", "baz")
				target.RunShell("foo")
			},
		)

		_, execOps := req.ContainsNExecOps(1)

		req.Contains(execOps[0].Exec.Meta.Env, "FOOARG=bar")
	})

	t.Run("uses given default", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				target.ExposeBuildArg("FOOARG", "baz")
				target.RunShell("foo")
			},
		)

		_, execOps := req.ContainsNExecOps(1)

		req.Contains(execOps[0].Exec.Meta.Env, "FOOARG=baz")
	})
}

func TestAddEnv(t *testing.T) {
	_, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.AddEnv(
				map[string]string{
					"FOO": "foo",
					"BAR": "bar",
				},
			)
			target.RunShell("foo")
		},
	)

	_, execOps := req.ContainsNExecOps(1)

	req.Contains(execOps[0].Exec.Meta.Env, "FOO=foo")
	req.Contains(execOps[0].Exec.Meta.Env, "BAR=bar")
}

func TestDescribef(t *testing.T) {
	target := testtarget.NewTarget("foo")
	req := require.New(t)

	opt := target.Describef("msg %s", "value")

	req.Implements((*llb.ConstraintsOpt)(nil), opt)

	var c llb.Constraints
	opt.SetConstraintsOption(&c)

	req.Contains(c.Metadata.Description, "llb.customname")
	req.Equal("[foo] msg value", c.Metadata.Description["llb.customname"])
}

func TestClientBuildDir(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)

	target := testtarget.NewTarget("foo")
	target.Options.ClientBuildContext = "context"
	target.Options.SessionID = "foo-session"
	target.Options.Excludes = []string{"*.log"}

	state := target.ClientBuildDir()

	def, err := state.Marshal(ctx)
	req.NoError(err)

	llbreq := llbtest.New(t, def)

	_, sourceOps := llbreq.ContainsNSourceOps(1)

	req.Equal("local://context", sourceOps[0].Source.Identifier)
	req.Equal(
		map[string]string{
			"local.excludepatterns": "[\"*.log\"]",
			"local.session":         "foo-session",
			"local.sharedkeyhint":   "context",
		},
		sourceOps[0].Source.Attrs,
	)
}

func TestCopyFromClient(t *testing.T) {
	sources := []string{"source1", "source2"}

	_, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.CopyFromClient(
				sources,
				"/srv/app/dest/",
			)
		},
	)

	fops, fileOps := req.ContainsNFileOps(1)
	inputs := req.HasValidInputs(fops[0])
	req.Len(inputs, 2)

	_, copies := req.ContainsNCopyActions(fileOps[0], 2)

	for i, source := range sources {
		req.IsType((*pb.Op_Source)(nil), inputs[1].Op)
		req.Equal("local://context", inputs[1].Op.(*pb.Op_Source).Source.Identifier)

		copy := copies[i].Copy
		req.Equal("/"+source, copy.Src)
		req.Equal("/srv/app/dest/", copy.Dest)
	}
}

func TestCopyFrom(t *testing.T) {
	_, req := testtarget.Setup(t,
		testtarget.NewTargets("bar", "foo"),
		func(bar *build.Target) {
			bar.WorkingDirectory("/srv/bar")
			bar.RunShell("last op")
		},
		func(foo *build.Target) {
			foo.WorkingDirectory("/srv/foo")

			foo.CopyFrom("bar", []string{"source1"}, "dest")
		},
	)

	fops, fileOps := req.ContainsNFileOps(1)
	inputs := req.HasValidInputs(fops[0])
	req.Len(inputs, 2)

	_, copies := req.ContainsNCopyActions(fileOps[0], 1)

	req.IsType((*pb.Op_Source)(nil), inputs[0].Op)
	req.Equal(
		"docker-image://testtarget.test/base/foo@sha256:368a265123d2e737d81ecd3693b714e9ee7db56f72dd4c3c060ad3f8eae58c61",
		inputs[0].Op.(*pb.Op_Source).Source.Identifier,
	)

	// secondary input should be bar's final Op from RunShell("last op") above
	req.IsType((*pb.Op_Exec)(nil), inputs[1].Op)
	req.Equal(
		[]string{"/bin/sh", "-c", "last op"},
		inputs[1].Op.(*pb.Op_Exec).Exec.Meta.Args,
	)

	copy := copies[0].Copy
	req.Equal("/srv/bar/source1", copy.Src)
	req.Equal("/srv/foo/dest", copy.Dest)
}

func TestExpandEnv(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)
	target := testtarget.NewTarget("foo")

	req.NoError(target.Initialize(ctx))

	target.AddEnv(map[string]string{
		"FOO": "foo",
	})

	req.Equal("FOO is foo", target.ExpandEnv("FOO is $FOO"))
}

func TestLogf(t *testing.T) {
	ctx := context.Background()

	t.Run("single platform", func(t *testing.T) {
		req := require.New(t)
		target := testtarget.NewTarget("foo")

		req.NoError(target.Initialize(ctx))

		req.Equal("[foo] bar: baz", target.Logf("bar: %s", "baz"))
	})

	t.Run("multiple targets", func(t *testing.T) {
		req := require.New(t)
		targets := testtarget.NewTargets("foo", "longname")

		req.NoError(targets.InitializeAll(ctx))

		req.Equal("[foo]      bar: baz", targets[0].Logf("bar: %s", "baz"))
		req.Equal("[longname] bar: baz", targets[1].Logf("bar: %s", "baz"))
	})

	t.Run("multiple platforms", func(t *testing.T) {
		req := require.New(t)
		target := testtarget.NewTarget("foo")
		target.Options.TargetPlatforms = []*oci.Platform{
			{OS: "linux", Architecture: "amd64"},
			{OS: "linux", Architecture: "arm64"},
		}

		req.NoError(target.Initialize(ctx))

		req.Equal("[foo_{linux/amd64}] bar: baz", target.Logf("bar: %s", "baz"))
	})
}

func TestRunEntrypoint(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(foo *build.Target) {
			foo.Image.Entrypoint([]string{"/bin/foo", "bar"})
			foo.RunEntrypoint([]string{"baz"}, map[string]string{"FOO": "foo"})
		},
	)

	req.Equal([]string{"/bin/foo", "bar"}, image.Config.Entrypoint)

	_, execOps := req.ContainsNExecOps(1)
	req.Equal([]string{"/bin/foo", "bar", "baz"}, execOps[0].Exec.Meta.Args)
	req.Contains(execOps[0].Exec.Meta.Env, "FOO=foo")
}
