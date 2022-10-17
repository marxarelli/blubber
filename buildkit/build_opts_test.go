package buildkit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildOptsParsing(t *testing.T) {
	extraOpts, _ := ParseExtraOptions(map[string]string{
		"run-variant":     "true",
		"entrypoint-args": `["param1", "param2"]`,
	})

	assert.Equal(t,
		ExtraBuildOptions{
			runEntrypoint:  true,
			entrypointArgs: []string{"param1", "param2"},
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
