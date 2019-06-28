package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestFlagMerge(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    runs: { insecurely: true }
    variants:
      development:
        runs: { insecurely: false }`))

	if assert.NoError(t, err) {
		variant, err := config.ExpandVariant(cfg, "development")

		if assert.NoError(t, err) {
			assert.False(t, variant.Runs.Insecurely.True)
		}
	}
}
