package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestConfigValidation(t *testing.T) {
	t.Run("variants", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        variants:
          build: {}
          foo: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
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
