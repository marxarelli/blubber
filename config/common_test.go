package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestCommonConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: fooimage
    entrypoint: ["/bin/foo"]
    variants:
      build: {}`))

	assert.Nil(t, err)

	assert.Equal(t, "fooimage", cfg.Base)
	assert.Equal(t, []string{"/bin/foo"}, cfg.EntryPoint)

	err = config.ExpandIncludesAndCopies(cfg, "build")
	variant, err := config.GetVariant(cfg, "build")

	assert.Equal(t, "fooimage", variant.Base)
	assert.Equal(t, []string{"/bin/foo"}, variant.EntryPoint)
}

// Ensure that entrypoints inherit correctly
//
func TestEntryPointMerge(t *testing.T) {
	foo := config.CommonConfig{EntryPoint: []string{"/bin/foo"}}
	bar := config.CommonConfig{EntryPoint: []string{"/bin/bar"}}
	foo.Merge(bar)
	assert.Equal(t, []string{"/bin/bar"}, foo.EntryPoint)
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

				assert.Equal(t, `base: "foo fighter" is not a valid image reference`, msg)
			}
		})
	})
}
