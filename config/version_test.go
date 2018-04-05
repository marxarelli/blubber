package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestVersionConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v1
    variants:
      foo: {}`))

	assert.Nil(t, err)

	if assert.NoError(t, err) {
		assert.Equal(t, "v1", cfg.Version)
	}
}

func TestVersionConfigValidation(t *testing.T) {
	t.Run("supported version", func(t *testing.T) {
		err := config.Validate(config.VersionConfig{
			Version: "v1",
		})

		assert.False(t, config.IsValidationError(err))
	})

	t.Run("unsupported version", func(t *testing.T) {
		err := config.Validate(config.VersionConfig{
			Version: "v2",
		})

		if assert.True(t, config.IsValidationError(err)) {
			msg := config.HumanizeValidationError(err)

			assert.Equal(t, `version: config version "v2" is unsupported`, msg)
		}
	})
}
