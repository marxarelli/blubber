package docker_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
	"phabricator.wikimedia.org/source/blubber/docker"
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

	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
}

func TestMultipleArtifactsFromSameStage(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
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

	dockerfile := docker.Compile(cfg, "production").String()

	assert.Contains(t, dockerfile, "FROM foo/bar AS build\n")
	assert.Contains(t, dockerfile, "FROM foo/bar AS production\n")

	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS build\n"))
	assert.Equal(t, 1, strings.Count(dockerfile, "FROM foo/bar AS production\n"))
}
