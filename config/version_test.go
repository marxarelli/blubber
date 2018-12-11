package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestVersionConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v3
    variants:
      foo: {}`))

	assert.Nil(t, err)

	if assert.NoError(t, err) {
		assert.Equal(t, "v3", cfg.Version)
	}
}

func TestVersionConfigValidation(t *testing.T) {
	t.Run("supported version", func(t *testing.T) {
		err := config.Validate(config.VersionConfig{
			Version: "v3",
		})

		assert.False(t, config.IsValidationError(err))
	})

	t.Run("unsupported version", func(t *testing.T) {
		err := config.Validate(config.VersionConfig{
			Version: "vX",
		})

		if assert.True(t, config.IsValidationError(err)) {
			msg := config.HumanizeValidationError(err)

			assert.Equal(t, `version: config version "vX" is unsupported`, msg)
		}
	})
}
