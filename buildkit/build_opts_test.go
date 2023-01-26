package buildkit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildOptsParsing(t *testing.T) {
	extraOpts, _ := ParseExtraOptions(map[string]string{
		"run-variant":     "true",
		"entrypoint-args": `["param1", "param2"]`,
		"run-variant-env": `{"KEY": "Value"}`,
	})

	assert.Equal(t,
		ExtraBuildOptions{
			runEntrypoint:         true,
			entrypointArgs:        []string{"param1", "param2"},
			runVariantEnvironment: map[string]string{"KEY": "Value"},
		},
		*extraOpts,
	)
}

func TestWrongEntrypointCmdFormat(t *testing.T) {
	_, err := ParseExtraOptions(map[string]string{
		"entrypoint-args": `["param1"}`,
	})

	assert.NotNil(t, err)
}

func TestBadRunVariantEnv(t *testing.T) {
	_, err := ParseExtraOptions(map[string]string{
		"run-variant-env": `KEY=VALUE`,
	})

	assert.NotNil(t, err)
}
