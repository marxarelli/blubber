package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func TestVariantConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo
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

func TestVariantConfigInstructions(t *testing.T) {
	t.Run("PhaseInstall", func(t *testing.T) {
		t.Run("copies", func(t *testing.T) {
			cfg := config.VariantConfig{Copies: "foo"}

			assert.Empty(t, cfg.InstructionsForPhase(build.PhaseInstall))
		})

		t.Run("shared volume", func(t *testing.T) {
			cfg := config.VariantConfig{}
			cfg.Lives.In = "/srv/service"
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
			cfg.Lives.UID = 123
			cfg.Lives.GID = 223

			assert.Equal(t,
				[]build.Instruction{
					build.CopyAs{123, 223, build.Copy{[]string{"."}, "."}},
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
				CommonConfig: config.CommonConfig{Lives: config.LivesConfig{In: "/srv/service"}},
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
				CommonConfig: config.CommonConfig{Lives: config.LivesConfig{In: "/srv/service"}},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})

		t.Run("with entrypoint", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					EntryPoint: []string{"/foo", "bar"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.EntryPoint{[]string{"/foo", "bar"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})
	})
}

func TestVariantConfigValidation(t *testing.T) {
	t.Run("includes", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: { includes: [build] }`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: { includes: [build, foobuild, foo_build] }`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, strings.Join([]string{
					`includes[1]: references an unknown variant "foobuild"`,
					`includes[2]: references an unknown variant "foo_build"`,
				}, "\n"), msg)
			}
		})
	})

	t.Run("copies", func(t *testing.T) {

		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: { copies: build }`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: { copies: foobuild }`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `copies: references an unknown variant "foobuild"`, msg)
			}
		})
	})
}
