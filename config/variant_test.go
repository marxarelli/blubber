package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
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

func TestVariantLoops(t *testing.T) {
	cfg := config.Config{
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{Includes: []string{"bar"}},
			"bar": config.VariantConfig{Includes: []string{"foo"}}}}

	cfgTwo := config.Config{
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{},
			"bar": config.VariantConfig{Includes: []string{"foo"}}}}

	// Configuration that contains a loop in "Includes" should error
	_, err := config.ExpandVariant(&cfg, "bar")
	assert.NotNil(t, err)

	_, errTwo := config.ExpandVariant(&cfgTwo, "bar")
	assert.Nil(t, errTwo)
}

func TestMultiLevelIncludes(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: nodejs-slim
    variants:
      build:
        base: nodejs-devel
        node: {env: build}
      development:
        includes: [build]
        node: {env: development}
        entrypoint: [npm, start]
      test:
        includes: [development]
        node: {dependencies: true}
        entrypoint: [npm, test]`))

	assert.Nil(t, err)

	variant, _ := config.ExpandVariant(cfg, "test")

	assert.Equal(t, "nodejs-devel", variant.Base)
	assert.Equal(t, "development", variant.Node.Env)

	devVariant, _ := config.ExpandVariant(cfg, "development")

	assert.True(t, variant.Node.Dependencies.True)
	assert.False(t, devVariant.Node.Dependencies.True)
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
