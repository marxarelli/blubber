package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v1
    variants:
      foo: {}`))

	if assert.NoError(t, err) {
		assert.Equal(t, "v1", cfg.Version)
		assert.Contains(t, cfg.Variants, "foo")
		assert.IsType(t, config.VariantConfig{}, cfg.Variants["foo"])
	}
}

func TestConfigValidation(t *testing.T) {
	t.Run("variants", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        variants:
          build foo: {}
          foo bar: {}`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `variants: contains a bad variant name`, msg)
			}
		})
	})
}
