package build_test

import (
	"testing"

	"github.com/moby/buildkit/solver/pb"
	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/util/testtarget"
)

func TestBase(t *testing.T) {
	target := testtarget.NewTarget("foo")

	i := build.Base{Image: "foo", Stage: "bar"}
	err := i.Compile(target)

	if assert.NoError(t, err) {
		// base is a noop
	}
}

func TestScratchBase(t *testing.T) {
	target := testtarget.NewTarget("foo")

	i := build.ScratchBase{Stage: "bar"}
	err := i.Compile(target)

	if assert.NoError(t, err) {
		// base is a noop
	}
}

func TestRun(t *testing.T) {
	t.Run("with only outer arguments", func(t *testing.T) {
		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.Run{"echo %s %s", []string{"foo bar", "baz"}},
		)

		_, eops := req.ContainsNExecOps(1)

		req.Equal(
			[]string{"/bin/sh", "-c", `echo "foo bar" "baz"`},
			eops[0].Exec.Meta.Args,
		)
	})

	t.Run("with inner and outer arguments", func(t *testing.T) {
		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.Run{"useradd -d %s -u %s", []string{"/foo", "666", "bar"}},
		)

		_, eops := req.ContainsNExecOps(1)

		req.Equal(
			[]string{"/bin/sh", "-c", `useradd -d "/foo" -u "666" "bar"`},
			eops[0].Exec.Meta.Args,
		)
	})
}

func TestRunAll(t *testing.T) {
	_, req := testtarget.Compile(t,
		testtarget.NewTargets("foo"),
		build.RunAll{[]build.Run{
			{"echo %s", []string{"foo"}},
			{"cat %s", []string{"/bar"}},
			{"baz", []string{}},
		}},
	)

	_, eops := req.ContainsNExecOps(1)

	req.Equal(
		[]string{"/bin/sh", "-c", `echo "foo" && cat "/bar" && baz`},
		eops[0].Exec.Meta.Args,
	)
}

func TestCopy(t *testing.T) {
	sources := []string{"source1", "source2"}

	_, req := testtarget.Compile(t,
		testtarget.NewTargets("foo"),
		build.Copy{sources, "dest"},
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

func TestCopyAs(t *testing.T) {
	t.Run("wrapping Copy", func(t *testing.T) {
		sources := []string{"source1", "source2"}

		_, req := testtarget.Compile(t,
			testtarget.NewTargets("foo"),
			build.CopyAs{
				"123", "321",
				build.Copy{sources, "dest/"},
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

			req.Equal(uint32(123), copy.Owner.User.GetByID())
			req.Equal(uint32(321), copy.Owner.Group.GetByID())
		}
	})

	t.Run("wrapping CopyFrom", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("bar", "foo"),
			func(bar *build.Target) {
				bar.WorkingDirectory("/srv/bar")
				bar.RunShell("last op")
			},
			func(foo *build.Target) {
				foo.WorkingDirectory("/srv/foo")
				i := build.CopyAs{
					"123", "321",
					build.CopyFrom{"bar", build.Copy{[]string{"source1"}, "dest"}},
				}
				i.Compile(foo)
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

		req.Equal(uint32(123), copy.Owner.User.GetByID())
		req.Equal(uint32(321), copy.Owner.Group.GetByID())
	})

	t.Run("referencing build arguments", func(t *testing.T) {
		_, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				i := build.CopyAs{
					"$LIVES_UID", "$LIVES_GID",
					build.Copy{[]string{"source1"}, "dest/"},
				}
				target.ExposeBuildArg("LIVES_UID", "123")
				target.ExposeBuildArg("LIVES_GID", "321")
				i.Compile(target)
			},
		)

		fops, fileOps := req.ContainsNFileOps(1)
		inputs := req.HasValidInputs(fops[0])
		req.Len(inputs, 2)

		_, copies := req.ContainsNCopyActions(fileOps[0], 1)

		req.IsType((*pb.Op_Source)(nil), inputs[1].Op)
		req.Equal("local://context", inputs[1].Op.(*pb.Op_Source).Source.Identifier)

		copy := copies[0].Copy
		req.Equal("/source1", copy.Src)
		req.Equal("/srv/app/dest/", copy.Dest)

		req.Equal(uint32(123), copy.Owner.User.GetByID())
		req.Equal(uint32(321), copy.Owner.Group.GetByID())
	})
}

func TestEntryPoint(t *testing.T) {
	image, req := testtarget.Compile(t,
		testtarget.NewTargets("foo"),
		build.EntryPoint{[]string{"/bin/foo", "bar", "baz"}},
	)

	t.Run("configures image", func(t *testing.T) {
		req.Equal([]string{"/bin/foo", "bar", "baz"}, image.Config.Entrypoint)
	})
}

func TestEnv(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			i := build.Env{map[string]string{
				"fooname": "foovalue",
				"barname": "barvalue",
			}}
			i.Compile(target)
			target.RunShell("foo")
		},
	)

	t.Run("configures image", func(t *testing.T) {
		req.Contains(image.Config.Env, "fooname=foovalue")
		req.Contains(image.Config.Env, "barname=barvalue")
	})

	t.Run("affects exec ops", func(t *testing.T) {
		_, execOps := req.ContainsNExecOps(1)
		req.Contains(execOps[0].Exec.Meta.Env, "fooname=foovalue")
		req.Contains(execOps[0].Exec.Meta.Env, "barname=barvalue")
	})
}

func TestLabel(t *testing.T) {
	image, req := testtarget.Compile(t,
		testtarget.NewTargets("foo"),
		build.Label{map[string]string{
			"fooname": "foovalue",
			"barname": "barvalue",
		}},
	)

	t.Run("configures image", func(t *testing.T) {
		req.NotNil(image.Config.Labels)
		req.Contains(image.Config.Labels, "fooname")
		req.Equal("foovalue", image.Config.Labels["fooname"])

		req.Contains(image.Config.Labels, "barname")
		req.Equal("barvalue", image.Config.Labels["barname"])
	})
}

func TestUser(t *testing.T) {
	t.Run("with UID", func(t *testing.T) {
		image, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				i := build.User{UID: "1000"}
				i.Compile(target)
				target.RunShell("foo")
			},
		)

		t.Run("configures image", func(t *testing.T) {
			req.Equal("1000", image.Config.User)
		})

		t.Run("affects exec ops", func(t *testing.T) {
			_, execOps := req.ContainsNExecOps(1)
			req.Equal("1000", execOps[0].Exec.Meta.User)
		})
	})

	t.Run("defaults to root", func(t *testing.T) {
		image, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				i := build.User{}
				i.Compile(target)
				target.RunShell("foo")
			},
		)

		t.Run("configures image", func(t *testing.T) {
			req.Equal("0", image.Config.User)
		})

		t.Run("affects exec ops", func(t *testing.T) {
			_, execOps := req.ContainsNExecOps(1)
			req.Equal("0", execOps[0].Exec.Meta.User)
		})
	})

	t.Run("supports variables", func(t *testing.T) {
		image, req := testtarget.Setup(t,
			testtarget.NewTargets("foo"),
			func(target *build.Target) {
				target.ExposeBuildArg("RUNS_UID", "666")

				i := build.User{UID: "$RUNS_UID"}
				i.Compile(target)

				target.RunShell("foo")
			},
		)

		t.Run("configures image", func(t *testing.T) {
			req.Equal("666", image.Config.User)
		})

		t.Run("affects exec ops", func(t *testing.T) {
			_, execOps := req.ContainsNExecOps(1)
			req.Equal("666", execOps[0].Exec.Meta.User)
		})
	})
}

func TestWorkingDirectory(t *testing.T) {
	image, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			i := build.WorkingDirectory{"/foo/path"}
			i.Compile(target)
			target.RunShell("foo")
		},
	)

	t.Run("configures image", func(t *testing.T) {
		req.Equal("/foo/path", image.Config.WorkingDir)
	})

	t.Run("affects exec ops", func(t *testing.T) {
		_, execOps := req.ContainsNExecOps(1)
		req.Equal("/foo/path", execOps[0].Exec.Meta.Cwd)
	})
}

func TestStringArg(t *testing.T) {
	_, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			i := build.StringArg{"RUNS_AS", "runuser"}
			i.Compile(target)
			target.RunShell("foo")
		},
	)

	t.Run("affects exec ops", func(t *testing.T) {
		_, execOps := req.ContainsNExecOps(1)
		req.Contains(execOps[0].Exec.Meta.Env, "RUNS_AS=runuser")
	})

	t.Run("expands variables", func(t *testing.T) {
		req.ExpandsEnv("RUNS_AS is runuser", "RUNS_AS is $RUNS_AS")
	})
}

func TestUintArg(t *testing.T) {
	_, req := testtarget.Setup(t,
		testtarget.NewTargets("foo"),
		func(target *build.Target) {
			i := build.UintArg{"RUNS_UID", 900}
			i.Compile(target)
			target.RunShell("foo")
		},
	)

	t.Run("affects exec ops", func(t *testing.T) {
		_, execOps := req.ContainsNExecOps(1)
		req.Contains(execOps[0].Exec.Meta.Env, "RUNS_UID=900")
	})

	t.Run("expands variables", func(t *testing.T) {
		req.ExpandsEnv("RUNS_UID is 900", "RUNS_UID is $RUNS_UID")
	})
}
