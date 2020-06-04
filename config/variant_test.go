package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestVariantConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    variants:
      build:
        copies: [local]
      production:
        copies:
          - from: build
            source: /foo/src
            destination: /foo/dst
          - from: build
            source: /bar/src
            destination: /bar/dst`))

	if assert.NoError(t, err) {
		err := config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Len(t, variant.Copies, 1)
		}

		err = config.ExpandIncludesAndCopies(cfg, "production")
		assert.Nil(t, err)

		variant, err = config.GetVariant(cfg, "production")

		if assert.NoError(t, err) {
			assert.Len(t, variant.Copies, 2)
		}
	}
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
	err := config.ExpandIncludesAndCopies(&cfg, "bar")
	assert.Error(t, err)

	errTwo := config.ExpandIncludesAndCopies(&cfgTwo, "bar")
	assert.NoError(t, errTwo)
}

func TestVariantConfigInstructions(t *testing.T) {
	t.Run("PhaseInstall", func(t *testing.T) {
		t.Run("without copies", func(t *testing.T) {
			cfg := config.VariantConfig{}

			assert.Empty(t, cfg.InstructionsForPhase(build.PhaseInstall))
		})

		t.Run("with copies", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Lives: config.LivesConfig{UserConfig: config.UserConfig{UID: 123, GID: 223}},
				},
				Copies: config.CopiesConfig{
					{From: "local"},
					{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.CopyAs{123, 223, build.Copy{[]string{"."}, "."}},
					build.CopyAs{123, 223, build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}}},
				},
				cfg.InstructionsForPhase(build.PhaseInstall),
			)
		})
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
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

		t.Run("without Runs.Insecurely", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Lives: config.LivesConfig{
						UserConfig: config.UserConfig{
							As: "foouser",
						},
					},
					Runs: config.RunsConfig{
						Insecurely: config.Flag{True: false},
						UserConfig: config.UserConfig{
							As: "baruser",
						},
					},
					EntryPoint: []string{"/foo", "bar"},
				},
			}

			assert.Equal(t,
				[]build.Instruction{
					build.User{"baruser"},
					build.Env{map[string]string{"HOME": "/home/baruser"}},
					build.EntryPoint{[]string{"/foo", "bar"}},
				},
				cfg.InstructionsForPhase(build.PhasePostInstall),
			)
		})

		t.Run("with Runs.Insecurely", func(t *testing.T) {
			cfg := config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Lives: config.LivesConfig{
						UserConfig: config.UserConfig{
							As: "foouser",
						},
					},
					Runs: config.RunsConfig{
						Insecurely: config.Flag{True: true},
						UserConfig: config.UserConfig{
							As: "baruser",
						},
					},
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
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: { includes: [build] }`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
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

		t.Run("should not contain duplicates", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo: { copies: [foo, bar, foo] }`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `copies: cannot contain duplicates`, msg)
			}
		})
	})
}
