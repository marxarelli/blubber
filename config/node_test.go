package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestNodeConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    node:
      requirements: [package.json]
      env: foo
      use-npm-ci: true
    variants:
      build:
        node:
          requirements: []
          env: bar`))

	if assert.NoError(t, err) {
		assert.Equal(t, config.RequirementsConfig{
			{From: "local", Source: "package.json"},
		}, cfg.Node.Requirements)
		assert.Equal(t, "foo", cfg.Node.Env)
		assert.Equal(t, true, cfg.Node.UseNpmCi.True)

		err = config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "build")

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

func TestNodeConfigInstructionsUseNpmCi(t *testing.T) {
	cfg := config.NodeConfig{
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "package.json"},
		},
		UseNpmCi: config.Flag{True: true},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json"}, "./"},
				build.Run{"npm ci", []string{}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}

func TestNodeConfigInstructionsNonProduction(t *testing.T) {
	cfg := config.NodeConfig{
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "package.json"},
		},
		Env: "foo",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"package.json"}, "./"},
				build.Run{"npm install", []string{}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV": "foo",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNodeConfigInstructionsProduction(t *testing.T) {
	cfg := config.NodeConfig{
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "package.json"},
			{From: "local", Source: "package-lock.json"},
		},
		Env: "production",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			assert.Equal(t,
				[]build.Instruction{
					build.Copy{[]string{"package.json", "package-lock.json"}, "./"},
					build.Run{"npm install", []string{"--only=production"}},
					build.Run{"npm dedupe", []string{}},
				},
				cfg.InstructionsForPhase(build.PhasePreInstall),
			)
		})

		t.Run("AllowDedupeFailure", func(t *testing.T) {
			var cfg2 = config.NodeConfig{
				AllowDedupeFailure: config.Flag{True: true},
			}
			cfg2.Merge(cfg)
			assert.Equal(t,
				[]build.Instruction{
					build.Copy{[]string{"package.json", "package-lock.json"}, "./"},
					build.Run{"npm install", []string{"--only=production"}},
					build.Run{"npm dedupe || echo %s", []string{
						"WARNING: npm dedupe failed, continuing anyways",
					}},
				},
				cfg2.InstructionsForPhase(build.PhasePreInstall),
			)
		})
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"NODE_ENV": "production",
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
					"NODE_ENV": "production",
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
