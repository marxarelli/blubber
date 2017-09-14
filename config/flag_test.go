package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestFlagOverwrite(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    node: { dependencies: true }
    sharedvolume: false
    variants:
      development:
        sharedvolume: true
        node: { dependencies: false }`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.False(t, variant.Node.Dependencies.True)
	assert.True(t, variant.SharedVolume.True)
}
