package docker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/docker"
)

func TestBase(t *testing.T) {
	i := build.Base{Image: "foo", Stage: "bar"}
	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "FROM foo AS bar\n", di.Compile())
	}
}

func TestScratchBase(t *testing.T) {
	i := build.ScratchBase{Stage: "bar"}
	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "FROM scratch AS bar\n", di.Compile())
	}
}

func TestRun(t *testing.T) {
	i := build.Run{"echo", []string{"hello"}}
	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "RUN echo \"hello\"\n", di.Compile())
	}
}

func TestRunAll(t *testing.T) {
	i := build.RunAll{[]build.Run{
		{"echo", []string{"hello"}},
		{"echo", []string{"yo"}},
	}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "RUN echo \"hello\" && echo \"yo\"\n", di.Compile())
	}
}

func TestCopy(t *testing.T) {
	i := build.Copy{[]string{"foo1", "foo2"}, "bar"}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "COPY [\"foo1\", \"foo2\", \"bar/\"]\n", di.Compile())
	}
}

func TestCopyAs(t *testing.T) {
	t.Run("with Copy", func(t *testing.T) {
		i := build.CopyAs{"123", "124", build.Copy{[]string{"foo1", "foo2"}, "bar"}}

		di, err := docker.NewInstruction(i)

		if assert.NoError(t, err) {
			assert.Equal(t, "COPY --chown=123:124 [\"foo1\", \"foo2\", \"bar/\"]\n", di.Compile())
		}
	})

	t.Run("with CopyFrom", func(t *testing.T) {
		i := build.CopyAs{"123", "124", build.CopyFrom{"foo", build.Copy{[]string{"foo1", "foo2"}, "bar"}}}

		di, err := docker.NewInstruction(i)

		if assert.NoError(t, err) {
			assert.Equal(t, "COPY --chown=123:124 --from=foo [\"foo1\", \"foo2\", \"bar/\"]\n", di.Compile())
		}
	})
}

func TestCopyFrom(t *testing.T) {
	i := build.CopyFrom{"foo", build.Copy{[]string{"foo1", "foo2"}, "bar"}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "COPY --from=foo [\"foo1\", \"foo2\", \"bar/\"]\n", di.Compile())
	}
}

func TestEntryPoint(t *testing.T) {
	i := build.EntryPoint{[]string{"foo", "bar"}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "ENTRYPOINT [\"foo\", \"bar\"]\n", di.Compile())
	}
}

func TestEnv(t *testing.T) {
	i := build.Env{map[string]string{"foo": "bar", "bar": "foo"}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "ENV bar=\"foo\" foo=\"bar\"\n", di.Compile())
	}
}

func TestLabel(t *testing.T) {
	i := build.Label{map[string]string{"foo": "bar", "bar": "foo"}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "LABEL bar=\"foo\" foo=\"bar\"\n", di.Compile())
	}
}

func TestUser(t *testing.T) {
	i := build.User{UID: "1000"}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "USER 1000\n", di.Compile())
	}
}

func TestWorkingDirectory(t *testing.T) {
	i := build.WorkingDirectory{"/foo/dir"}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "WORKDIR \"/foo/dir\"\n", di.Compile())
	}
}

func TestStringArg(t *testing.T) {
	i := build.StringArg{"RUNS_AS", "runuser"}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "ARG RUNS_AS=\"runuser\"\n", di.Compile())
	}
}

func TestUintArg(t *testing.T) {
	i := build.UintArg{"RUNS_UID", 900}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "ARG RUNS_UID=900\n", di.Compile())
	}
}

func TestEscapeRun(t *testing.T) {
	i := build.Run{"/bin/true\nRUN echo HACKED!", []string{}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "RUN /bin/true\\nRUN echo HACKED!\n", di.Compile())
	}
}

func TestEscapeCopy(t *testing.T) {
	i := build.Copy{[]string{"file.a", "file.b"}, "dest"}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "COPY [\"file.a\", \"file.b\", \"dest/\"]\n", di.Compile())
	}
}

func TestEscapeEnv(t *testing.T) {
	i := build.Env{map[string]string{"a": "b\nRUN echo HACKED!"}}

	di, err := docker.NewInstruction(i)

	if assert.NoError(t, err) {
		assert.Equal(t, "ENV a=\"b\\nRUN echo HACKED!\"\n", di.Compile())
	}
}
