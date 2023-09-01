package buildkit_test

import (
	"testing"

	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/stretchr/testify/require"

	"gitlab.wikimedia.org/repos/releng/blubber/buildkit"
)

func TestBuildOptsParsing(t *testing.T) {
	buildOpts, _ := buildkit.ParseBuildOptions(
		client.BuildOpts{
			Opts: map[string]string{
				"run-variant":     "true",
				"entrypoint-args": `["param1", "param2"]`,
				"run-variant-env": `{"KEY": "Value"}`,
			},
		},
	)

	require.True(t, buildOpts.RunEntrypoint)
	require.Equal(t, []string{"param1", "param2"}, buildOpts.EntrypointArgs)
	require.Equal(t, map[string]string{"KEY": "Value"}, buildOpts.RunEnvironment)
}

func TestWrongEntrypointCmdFormat(t *testing.T) {
	_, err := buildkit.ParseBuildOptions(
		client.BuildOpts{
			Opts: map[string]string{
				"entrypoint-args": `["param1"}`,
			},
		},
	)

	require.Error(t, err)
}

func TestBadRunVariantEnv(t *testing.T) {
	_, err := buildkit.ParseBuildOptions(
		client.BuildOpts{
			Opts: map[string]string{
				"run-variant-env": `KEY=VALUE`,
			},
		},
	)

	require.Error(t, err)
}
