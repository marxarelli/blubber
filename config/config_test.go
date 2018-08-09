package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v3
    variants:
      foo: {}`))

	if assert.NoError(t, err) {
		assert.Equal(t, "v3", cfg.Version)
		assert.Contains(t, cfg.Variants, "foo")
		assert.IsType(t, config.VariantConfig{}, cfg.Variants["foo"])
	}
}

func TestConfigValidation(t *testing.T) {
	t.Run("variants", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.Config{
				VersionConfig: config.VersionConfig{Version: "v3"},
				Variants: map[string]config.VariantConfig{
					"build": config.VariantConfig{},
					"foo":   config.VariantConfig{},
				},
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.Config{
				VersionConfig: config.VersionConfig{Version: "v3"},
				Variants: map[string]config.VariantConfig{
					"build foo": config.VariantConfig{},
					"foo bar":   config.VariantConfig{},
				},
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `variants: contains a bad variant name`, msg)
			}
		})
	})
}
