package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestNpmConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    npm:
      install: true
      env: foo
    variants:
      build:
        npm:
          install: false
          env: bar`))

	assert.Nil(t, err)

	assert.Equal(t, true, cfg.Npm.Install.True)
	assert.Equal(t, "foo", cfg.Npm.Env)

	variant, err := config.ExpandVariant(cfg, "build")

	assert.Equal(t, false, variant.Npm.Install.True)
	assert.Equal(t, "bar", variant.Npm.Env)
}

func TestNpmConfigInstructionsNoInstall(t *testing.T) {
	cfg := config.NpmConfig{Install: config.Flag{True: false}}

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

func TestNpmConfigInstructionsNonProduction(t *testing.T) {
	cfg := config.NpmConfig{Install: config.Flag{True: true}, Env: "foo"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				{build.Run, []string{"mkdir -p /tmp/node-deps/"}},
				{build.Copy, []string{"package.json", "/tmp/node-deps/"}},
				{build.Run, []string{"cd /tmp/node-deps/ && npm install"}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				{build.Run, []string{"mv /tmp/node-deps/node_modules ./"}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestNpmConfigInstructionsProduction(t *testing.T) {
	cfg := config.NpmConfig{Install: config.Flag{True: true}, Env: "production"}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				{build.Run, []string{"mkdir -p /tmp/node-deps/"}},
				{build.Copy, []string{"package.json", "/tmp/node-deps/"}},
				{build.Run, []string{"cd /tmp/node-deps/ && npm install --production && npm dedupe"}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				{build.Run, []string{"mv /tmp/node-deps/node_modules ./"}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
