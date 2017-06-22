package docker_test

import (
	"bytes"
	"testing"
	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
	"phabricator.wikimedia.org/source/blubber.git/docker"
)

func TestSingleStageHasNoName(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo/bar
    variants:
      development: {}`))

	assert.Nil(t, err)

	dockerfile := docker.Compile(cfg, "development").String()

	assert.Contains(t, dockerfile, "FROM foo/bar\n")
}

func TestMultiStageIncludesStageNames(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo/bar
    variants:
      build: {}
      production:
        artifacts:
          - from: build
            source: .
            destination: .`))

	assert.Nil(t, err)

	dockerfile := docker.Compile(cfg, "production").String()

	assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
	assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")
}

func TestCompileInstructionCopy(t *testing.T) {
	buffer := new(bytes.Buffer)
	instruction := build.Instruction{build.Copy, []string{"foo", "bar"}}

	docker.CompileInstruction(buffer, instruction)

	assert.Equal(t, "COPY [\"foo\", \"bar\"]\n", buffer.String())
}

func TestCompileInstructionEnv(t *testing.T) {
	buffer := new(bytes.Buffer)
	instruction := build.Instruction{build.Env, []string{"foo=bar", "baz=qux"}}

	docker.CompileInstruction(buffer, instruction)

	assert.Equal(t, "ENV foo=bar \\\n    baz=qux\n", buffer.String())
}
