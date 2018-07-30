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
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
    base: foo/bar
    variants:
      development: {}`))

	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "development")
	dockerfile := dockerOut.String()

	assert.Contains(t, dockerfile, "FROM foo/bar\n")
}

func TestMultiStageIncludesStageNames(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
    base: foo/bar
    variants:
      build: {}
      production:
        artifacts:
          - from: build
            source: .
            destination: .`))

	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "production")
	dockerfile := dockerOut.String()

	assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
	assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")

	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
}

func TestMultipleArtifactsFromSameStage(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
    base: foo/bar
    variants:
      build: {}
      production:
        artifacts:
          - from: build
            source: .
            destination: .
          - from: build
            source: bar
            destination: bar`))

	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "production")
	dockerfile := dockerOut.String()

	assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
	assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")

	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
}

func TestMetaDataLabels(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
    base: foo/bar
    variants:
      development: {}`))

	assert.Nil(t, err)

	dockerOut, _ := docker.Compile(cfg, "development")
	dockerfile := dockerOut.String()

	version := meta.FullVersion()

	assert.Contains(t, dockerfile,
		"LABEL blubber.variant=\"development\" blubber.version=\""+version+"\"\n",
	)
}
