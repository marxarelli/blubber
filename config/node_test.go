package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestNodeConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v3
    base: foo
    node:
      requirements: [package.json]
      env: foo
    variants:
      build:
        node:
          requirements: []
          env: bar`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"package.json"}, cfg.Node.Requirements)
		assert.Equal(t, "foo", cfg.Node.Env)

		variant, err := config.ExpandVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Empty(t, variant.Node.Requirements)
			assert.Equal(t, "bar", variant.Node.Env)
		}
	}
}

func TestNodeConfigInstructionsNoRequirements(t *testing.T) {
	cfg := config.NodeConfig{}

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

func TestNodeConfigInstructionsNonProduction(t *testing.T) {
	cfg := config.NodeConfig{Requirements: []string{"package.json"}, Env: "foo"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json"}, "/opt/lib/"},
				build.RunAll{[]build.Run{
					{"cd", []string{"/opt/lib"}},
					{"npm install", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "foo",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigInstructionsProduction(t *testing.T) {
	cfg := config.NodeConfig{Requirements: []string{"package.json", "package-lock.json"}, Env: "production"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json", "package-lock.json"}, "/opt/lib/"},
				build.RunAll{[]build.Run{
					{"cd", []string{"/opt/lib"}},
					{"npm install", []string{"--production"}},
					{"npm dedupe", []string{}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "production",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigInstructionsEnvironmentOnly(t *testing.T) {
	cfg := config.NodeConfig{Env: "production"}

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
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV":  "production",
					"NODE_PATH": "/opt/lib/node_modules",
					"PATH":      "/opt/lib/node_modules/.bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigValidation(t *testing.T) {
	t.Run("env", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.NodeConfig{
				Env: "production",
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			err := config.Validate(config.NodeConfig{})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.NodeConfig{
				Env: "foo bar",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `env: "foo bar" is not a valid Node environment name`, msg)
			}
		})
	})
}
