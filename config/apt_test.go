package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func TestAptConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
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
