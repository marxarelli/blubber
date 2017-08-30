package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestArtifactsConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    variants:
      build: {}
      production:
        artifacts:
          - from: build
            source: /foo/src
            destination: /foo/dst
          - from: build
            source: /bar/src
            destination: /bar/dst`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "production")

	assert.Nil(t, err)

	assert.Len(t, variant.Artifacts, 2)

	assert.Contains(t,
		variant.Artifacts,
		config.ArtifactsConfig{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
	)
	assert.Contains(t,
		variant.Artifacts,
		config.ArtifactsConfig{From: "build", Source: "/bar/src", Destination: "/bar/dst"},
	)
}

func TestArtifactsConfigInstructions(t *testing.T) {
	cfg := config.ArtifactsConfig{
		From:        "foo",
		Source:      "/source/path",
		Destination: "/destination/path",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.CopyFrom{
				"foo",
				build.Copy{[]string{"/source/path"}, "/destination/path"},
			}},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
