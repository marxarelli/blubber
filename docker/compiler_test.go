package docker_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
	"gerrit.wikimedia.org/r/blubber/docker"
	"gerrit.wikimedia.org/r/blubber/meta"
)

func TestSingleStageHasNoName(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo/bar
    variants:
      development: {}`))

	assert.Nil(t, err)

	err = config.ExpandIncludesAndCopies(cfg, "development")
	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "development")
	dockerfile := dockerOut.String()

	assert.Contains(t, dockerfile, "FROM foo/bar\n")
}

func TestMultiStageIncludesStageNames(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo/bar
    variants:
      build: {}
      production:
        copies:
          - from: build
            source: .
            destination: .`))

	if assert.NoError(t, err) {
		err = config.ExpandIncludesAndCopies(cfg, "production")
		assert.Nil(t, err)

		dockerOut, _ := docker.Compile(cfg, "production")
		dockerfile := dockerOut.String()

		assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
		assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")

		assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
		assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
	}
}

func TestMultipleArtifactsFromSameStage(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo/bar
    variants:
      build: {}
      production:
        copies:
          - from: build
            source: .
            destination: .
          - from: build
            source: bar
            destination: bar`))

	assert.Nil(t, err)

	err = config.ExpandIncludesAndCopies(cfg, "production")
	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "production")
	dockerfile := dockerOut.String()

	assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
	assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")

	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
}

// T254629, T259069
func TestMultiLevelArtifacts(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`
---
version: v4
base: foo/bar
variants:
  one:
    copies: [local]
  two:
    copies: [one]
  three:
    copies: [two]
        `))

	assert.Nil(t, err)

	err = config.ExpandIncludesAndCopies(cfg, "three")
	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "three")
	dockerfile := dockerOut.String()

	// There should be exactly 3 stages
	assert.Equal(t, 3, strings.Count(dockerfile, "FROM "))
	// Verify that both stages one and two are built in addition to the requested variant.
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS one\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS two\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS three\n"))
}

func TestMetaDataLabels(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo/bar
    variants:
      development: {}`))

	assert.Nil(t, err)

	config.ExpandIncludesAndCopies(cfg, "development")
	dockerOut, _ := docker.Compile(cfg, "development")
	dockerfile := dockerOut.String()

	version := meta.FullVersion()

	assert.Contains(t, dockerfile,
		"LABEL blubber.variant=\"development\" blubber.version=\""+version+"\"\n",
	)
}
