package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestCommonConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
    base: fooimage
    sharedvolume: true
    entrypoint: ["/bin/foo"]
    variants:
      build: {}`))

	assert.Nil(t, err)

	assert.Equal(t, "fooimage", cfg.Base)
	assert.Equal(t, true, cfg.SharedVolume.True)
	assert.Equal(t, []string{"/bin/foo"}, cfg.EntryPoint)

	variant, err := config.ExpandVariant(cfg, "build")

	assert.Equal(t, "fooimage", variant.Base)
	assert.Equal(t, true, variant.SharedVolume.True)
	assert.Equal(t, []string{"/bin/foo"}, variant.EntryPoint)
}

func TestCommonConfigValidation(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			err := config.Validate(config.CommonConfig{
				Base: "foo",
			})

			assert.Nil(t, err)
		})

		t.Run("optional", func(t *testing.T) {
			err := config.Validate(config.CommonConfig{
				Base: "",
			})

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("bad", func(t *testing.T) {
			err := config.Validate(config.CommonConfig{
				Base: "foo fighter",
			})

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `base: "foo fighter" is not a valid base image reference`, msg)
			}
		})
	})
}
