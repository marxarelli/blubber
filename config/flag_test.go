package config_test

import (
	"testing"
	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestFlagOverwrite(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    npm: { install: true }
    sharedvolume: false
    variants:
      development:
        sharedvolume: true
        npm: { install: false }`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.False(t, variant.Npm.Install.True)
	assert.True(t, variant.SharedVolume.True)
}
