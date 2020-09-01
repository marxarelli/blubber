package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestPhpConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    php:
      requirements: [composer.json]
    variants:
      build:
        php:
          requirements: []`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"composer.json"}, cfg.Php.Requirements)

		err = config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Empty(t, variant.Php.Requirements)
		}
	}
}

func TestPhpConfigInstructionsNoRequirements(t *testing.T) {
	cfg := config.PhpConfig{}

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
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestPhpConfigInstructions(t *testing.T) {
	cfg := config.PhpConfig{Requirements: []string{"composer.json"}}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"composer.json"}, "./"},
				build.RunAll{[]build.Run{
					{"composer install", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestPhpConfigInstructionsProduction(t * testing.T) {
	cfg := config.PhpConfig{Requirements: []string{"composer.json"}, Production: config.Flag{True: true}}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"composer.json"}, "./"},
				build.RunAll{[]build.Run{
					{"composer install", []string{"--no-dev"}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}