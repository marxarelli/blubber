package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestArtifactsConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v3
    base: foo
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

	t.Run("PhaseInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.CopyFrom{
				"foo",
				build.Copy{[]string{"/source/path"}, "/destination/path"},
			}},
			cfg.InstructionsForPhase(build.PhaseInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestArtifactsConfigValidation(t *testing.T) {
	t.Run("from", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v3
        variants:
          build: {}
          foo:
            artifacts:
              - from: build
                source: /foo
                destination: /bar`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("missing", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v3
        variants:
          build: {}
          foo:
            artifacts:
              - from: ~
                source: /foo
                destination: /bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `from: is required`, msg)
			}
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v3
        variants:
          build: {}
          foo:
            artifacts:
              - from: foo bar
                source: /foo
                destination: /bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `from: references an unknown variant "foo bar"`, msg)
			}
		})
	})
}
