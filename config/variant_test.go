package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestVariantConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    variants:
      build: {}
      production:
        copies: build
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

	assert.Equal(t, "build", variant.Copies)
	assert.Len(t, variant.Artifacts, 2)
}

func TestVariantDependencies(t *testing.T) {
	cfg := config.VariantConfig{
		Copies: "foo",
		Artifacts: []config.ArtifactsConfig{
			{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
		},
	}

	assert.Equal(t, []string{"foo", "build"}, cfg.VariantDependencies())
}

func TestVariantConfigInstructions(t *testing.T) {
	t.Run("PhaseInstall", func(t *testing.T) {
		t.Run("copies", func(t *testing.T) {
			cfg := config.VariantConfig{Copies: "foo"}

			assert.Empty(t, cfg.InstructionsForPhase(build.PhaseInstall))
		})

		t.Run("shared volume", func(t *testing.T) {
			cfg := config.VariantConfig{}
			cfg.Runs.In = "/srv/service"
			cfg.SharedVolume.True = true

			assert.Equal(t,
				[]build.Instruction{
					build.Volume{"/srv/service"},
				},
				cfg.InstructionsForPhase(build.PhaseInstall),
			)
		})

		t.Run("standard source copy", func(t *testing.T) {
			cfg := config.VariantConfig{}

			assert.Equal(t,
				[]build.Instruction{
					build.Copy{[]string{"."}, "."},
				},
				cfg.InstructionsForPhase(build.PhaseInstall),
			)
		})
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		t.Run("for copies and artifacts", func(t *testing.T) {
			cfg := config.VariantConfig{
				Copies: "foo",
				Artifacts: []config.ArtifactsConfig{
					{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
				},
				CommonConfig: config.CommonConfig{Runs: config.RunsConfig{In: "/srv/service"}},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.CopyFrom{"foo", build.Copy{[]string{"/srv/service"}, "/srv/service"}},
					build.CopyFrom{"foo", build.Copy{[]string{config.LocalLibPrefix}, config.LocalLibPrefix}},
					build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})

		t.Run("for just artifacts", func(t *testing.T) {
			cfg := config.VariantConfig{
				Artifacts: []config.ArtifactsConfig{
					{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
				},
				CommonConfig: config.CommonConfig{Runs: config.RunsConfig{In: "/srv/service"}},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})
	})
}
