package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestAptConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    apt:
      packages:
        - libfoo
        - libbar
    variants:
      build:
        apt:
          packages:
            - libfoo-dev`))

	assert.Nil(t, err)

	assert.Equal(t, []string{"libfoo", "libbar"}, cfg.Apt.Packages)

	variant, err := config.ExpandVariant(cfg, "build")

	assert.Nil(t, err)

	assert.Equal(t, []string{"libfoo", "libbar", "libfoo-dev"}, variant.Apt.Packages)
}

func TestAptConfigInstructions(t *testing.T) {
	cfg := config.AptConfig{Packages: []string{"libfoo", "libbar"}}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"DEBIAN_FRONTEND": "noninteractive",
				}},
				build.RunAll{[]build.Run{
					{"apt-get update", []string{}},
					{"apt-get install -y", []string{"libfoo", "libbar"}},
					{"rm -rf /var/lib/apt/lists/*", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestAptConfigValidation(t *testing.T) {
	t.Run("packages", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.AptConfig{
				Packages: []string{
					"f1",
					"foo-fighter",
					"bar+b.az",
					"bar+b.az=0:0.1~foo1-1",
					"bar+b.az/stable",
					"bar+b.az/jessie-wikimedia",
				},
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.AptConfig{
				Packages: []string{
					"f1",
					"foo fighter",
					"bar_baz",
					"bar=0.1*bad version",
					"bar/0bad_release",
				},
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, strings.Join([]string{
					`packages[1]: "foo fighter" is not a valid Debian package name`,
					`packages[2]: "bar_baz" is not a valid Debian package name`,
					`packages[3]: "bar=0.1*bad version" is not a valid Debian package name`,
					`packages[4]: "bar/0bad_release" is not a valid Debian package name`,
				}, "\n"), msg)
			}
		})
	})
}
