package config_test

import (
	"testing"
	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/config"
)

const yaml = `---
npm: { install: true }
sharedvolume: false

variants:
  development:
    sharedvolume: true
    npm: { install: false }
`

func TestFlagOverwrite(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(yaml))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.False(t, variant.Npm.Install.True)
	assert.True(t, variant.SharedVolume.True)
}
