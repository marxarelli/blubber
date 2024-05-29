package build_test

import (
	"testing"

	oci "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testtarget"
)

func TestImageEntrypoint(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargetsWithBaseImage(
			[]string{"foo"},
			oci.Image{
				Config: oci.ImageConfig{
					Cmd: []string{"base", "cmd"},
				},
			},
		),
		func(target *build.Target) {
			target.Image.Entrypoint([]string{"/bin/foo", "bar"})
		},
	)

	req.Equal([]string{"/bin/foo", "bar"}, image.Config.Entrypoint)

	// Ensure parity with Dockerfile behavior whereby a new ENTRYPOINT resets
	// the value of CMD inherited from the base image
	req.Nil(image.Config.Cmd)
}

func TestImageUser(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.ExposeBuildArg("RUNS_UID", "123")
			target.Image.User("$RUNS_UID")
		},
	)

	req.Equal("123", image.Config.User)
}

func TestImageWorkingDirectory(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.ExposeBuildArg("FOO", "foo")
			target.Image.WorkingDirectory("/srv/$FOO")
		},
	)

	req.Equal("/srv/foo", image.Config.WorkingDir)
}

func TestImageAddEnv(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.ExposeBuildArg("FOO", "foo")
			target.Image.AddEnv(map[string]string{
				"FOO": "foo",
			})
			target.AddEnv(map[string]string{
				"BAR": `"bar"`,
			})
			target.Image.AddEnv(map[string]string{
				"BAZ": "baz-$FOO",
				"QUX": "qux-$BAR",
				"FOO": "newfoo",
			})
		},
	)

	req.Equal(
		[]string{
			`FOO=newfoo`,
			`BAZ=baz-foo`,
			`QUX=qux-"bar"`,
		},
		image.Config.Env,
	)
}

func TestImageAddLabels(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			target.ExposeBuildArg("FOO", "foo")
			target.AddEnv(map[string]string{
				"BAR": `"bar"`,
			})
			target.Image.AddLabels(map[string]string{
				"BAZ": "baz-$FOO",
				"QUX": "qux-$BAR",
			})
		},
	)

	req.Contains(image.Config.Labels, "BAZ")
	req.Equal(image.Config.Labels["BAZ"], "baz-foo")

	req.Contains(image.Config.Labels, "QUX")
	req.Equal(image.Config.Labels["QUX"], `qux-"bar"`)
}
