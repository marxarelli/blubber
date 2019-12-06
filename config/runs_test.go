package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestRunsConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    runs:
      as: someuser
      insecurely: true
      uid: 666
      gid: 777
      environment: { FOO: bar }
    variants:
      development: {}`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.Equal(t, "someuser", variant.Runs.As)
	assert.Equal(t, true, variant.Runs.Insecurely.True)
	assert.Equal(t, uint(666), variant.Runs.UID)
	assert.Equal(t, uint(777), variant.Runs.GID)
	assert.Equal(t, map[string]string{"FOO": "bar"}, variant.Runs.Environment)
}

func TestRunsConfigInstructions(t *testing.T) {
	cfg := config.RunsConfig{
		UserConfig: config.UserConfig{
			As:  "someuser",
			UID: 666,
			GID: 777,
		},
		Environment: map[string]string{
			"fooname": "foovalue",
			"barname": "barvalue",
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.RunAll{[]build.Run{
				{"groupadd -o -g %s -r", []string{"777", "someuser"}},
				{"useradd -l -o -m -d %s -r -g %s -u %s", []string{"/home/someuser", "someuser", "666", "someuser"}},
			}}},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{"barname": "barvalue", "fooname": "foovalue"}},
			},
			cfg.InstructionsForPhase(build.PhasePrivilegeDropped),
		)

		t.Run("with empty Environment", func(t *testing.T) {
			cfg.Environment = map[string]string{}

			assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
		})
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestRunsConfigValidation(t *testing.T) {
	t.Run("environment", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.RunsConfig{
				Environment: map[string]string{
					"foo":          "bar",
					"f1oo":         "bar",
					"FOO":          "bar",
					"foo_fighter":  "bar",
					"FOO_FIGHTER":  "bar",
					"_FOO_FIGHTER": "bar",
				},
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			err := config.Validate(config.RunsConfig{})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			t.Run("spaces", func(t *testing.T) {
				err := config.Validate(config.RunsConfig{
					Environment: map[string]string{
						"foo fighter": "bar",
					},
				})

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `environment: contains invalid environment variable names`, msg)
				}
			})

			t.Run("dashes", func(t *testing.T) {
				err := config.Validate(config.RunsConfig{
					Environment: map[string]string{
						"foo-fighter": "bar",
					},
				})

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `environment: contains invalid environment variable names`, msg)
				}
			})

			t.Run("dots", func(t *testing.T) {
				err := config.Validate(config.RunsConfig{
					Environment: map[string]string{
						"foo.fighter": "bar",
					},
				})

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `environment: contains invalid environment variable names`, msg)
				}
			})

			t.Run("starts with number", func(t *testing.T) {
				err := config.Validate(config.RunsConfig{
					Environment: map[string]string{
						"1foo": "bar",
					},
				})

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `environment: contains invalid environment variable names`, msg)
				}
			})
		})
	})
}
