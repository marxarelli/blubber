package config_test

import (
	"testing"
	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestRunsConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    runs:
      as: someuser
      in: /some/directory
      uid: 666
      gid: 777
      environment: { FOO: bar }
    variants:
      development: {}`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.Equal(t, "someuser", variant.Runs.As)
	assert.Equal(t, "/some/directory", variant.Runs.In)
	assert.Equal(t, 666, variant.Runs.Uid)
	assert.Equal(t, 777, variant.Runs.Gid)
	assert.Equal(t, map[string]string{"FOO": "bar"}, variant.Runs.Environment)
}

func TestRunsHomeWithUser(t *testing.T) {
	runs := config.RunsConfig{As: "someuser"}

	assert.Equal(t, "/home/someuser", runs.Home())
}

func TestRunsHomeWithoutUser(t *testing.T) {
	runs := config.RunsConfig{}

	assert.Equal(t, "/root", runs.Home())
}

func TestEnvironmentDefinitionsIsSortedAndQuoted(t *testing.T) {
	runs := config.RunsConfig{
		Environment: map[string]string{
			"fooname": "foovalue",
			"barname": "barvalue",
			"quxname": "quxvalue",
		},
	}

	assert.Equal(t, []string{
		`barname="barvalue"`,
		`fooname="foovalue"`,
		`quxname="quxvalue"`,
	}, runs.EnvironmentDefinitions())
}
