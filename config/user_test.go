package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestUserConfigValidation(t *testing.T) {
	t.Run("as", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.UserConfig{
				As: "foo-bar.baz",
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			err := config.Validate(config.UserConfig{})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("no spaces", func(t *testing.T) {
			err := config.Validate(config.UserConfig{
				As: "foo bar",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "foo bar" is not a valid user name`, msg)
			}
		})

		t.Run("long enough", func(t *testing.T) {
			err := config.Validate(config.UserConfig{
				As: "fo",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "fo" is not a valid user name`, msg)
			}
		})

		t.Run("not root", func(t *testing.T) {
			err := config.Validate(config.UserConfig{
				As: "root",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "root" is not a valid user name`, msg)
			}
		})
	})
}
