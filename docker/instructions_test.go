package docker_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/docker"
)

func TestFactory(t *testing.T) {
	i := build.Instruction{build.Run, []string{"echo hello"}}
	dr, _ := docker.NewDockerInstruction(i)

	var dockerRun docker.DockerRun

	assert.IsType(t, dockerRun, dr)
	assert.Equal(t, dr.Arguments(), i.Arguments)
	assert.NotEmpty(t, dr.Compile())
	assert.Equal(t, "RUN echo hello\n", dr.Compile())
}

func TestEscapeRun(t *testing.T) {
	i := build.Instruction{build.Run, []string{"/bin/true\nRUN echo HACKED!"}}
	dr, _ := docker.NewDockerInstruction(i)

	assert.Equal(t, "RUN /bin/true\\nRUN echo HACKED!\n", dr.Compile())
}

func TestEscapeCopy(t *testing.T) {
	i := build.Instruction{build.Copy, []string{"file.a", "file.b"}}
	dr, _ := docker.NewDockerInstruction(i)

	assert.Equal(t, "COPY [\"file.a\", \"file.b\"]\n", dr.Compile())
}

func TestEscapeEnv(t *testing.T) {
	i := build.Instruction{build.Env, []string{"a=b\nRUN echo HACKED!"}}
	dr, _ := docker.NewDockerInstruction(i)

	assert.Equal(t, "ENV a=b\\nRUN echo HACKED!\n", dr.Compile())
}
