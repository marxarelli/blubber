package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestRequirementsInstructionsForPhase(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`["foo", "bar"]`))

		if assert.NoError(t, err) {
			instructions := cfg.InstructionsForPhase(build.PhasePreInstall)
			assert.Len(t, instructions, 1)
			assert.Equal(
				t,
				build.Copy{[]string{"foo", "bar"}, "./"},
				instructions[0],
			)
		}
	})

	t.Run("objects", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`[
			{ "from": "foo" },
			{ "from": "bar", "source": "/foo", "destination": "/bar/" }
		]`))

		if assert.NoError(t, err) {
			instructions := cfg.InstructionsForPhase(build.PhasePreInstall)
			assert.Len(t, instructions, 2)
			assert.Equal(
				t,
				build.CopyFrom{"foo", build.Copy{[]string{"."}, "./"}},
				instructions[0],
			)
			assert.Equal(
				t,
				build.CopyFrom{"bar", build.Copy{[]string{"/foo"}, "/bar/"}},
				instructions[1],
			)
		}
	})

	t.Run("mixed", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`[
			"foo",
			{ "from": "foo" },
			"bar",
			{ "from": "bar", "source": "/foo", "destination": "/bar/" }
		]`))

		if assert.NoError(t, err) {
			instructions := cfg.InstructionsForPhase(build.PhasePreInstall)
			assert.Len(t, instructions, 3)
			assert.Equal(
				t,
				build.Copy{[]string{"foo", "bar"}, "./"},
				instructions[0],
			)
			assert.Equal(
				t,
				build.CopyFrom{"foo", build.Copy{[]string{"."}, "./"}},
				instructions[1],
			)
			assert.Equal(
				t,
				build.CopyFrom{"bar", build.Copy{[]string{"/foo"}, "/bar/"}},
				instructions[2],
			)
		}
	})
}
func TestRequirementsConfigUnmarshalJSON(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`["foo", "bar", "xyzzy/plugh"]`))

		if assert.NoError(t, err) {
			assert.Len(t, cfg, 3)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[0].From)
			assert.Equal(t, "foo", cfg[0].Source)
			assert.Equal(t, "./", cfg[0].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[1].From)
			assert.Equal(t, "bar", cfg[1].Source)
			assert.Equal(t, "./", cfg[1].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[2].From)
			assert.Equal(t, "xyzzy/plugh", cfg[2].Source)
			assert.Equal(t, "xyzzy/", cfg[2].Destination)
		}
	})

	t.Run("objects", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`[{ "from": "foo" }, { "from": "bar", "source": "/foo", "destination": "/bar" }]`))

		if assert.NoError(t, err) {
			assert.Len(t, cfg, 2)
			assert.Equal(t, "foo", cfg[0].From)
			assert.Equal(t, "", cfg[0].Source)
			assert.Equal(t, "", cfg[0].Destination)
			assert.Equal(t, "bar", cfg[1].From)
			assert.Equal(t, "/foo", cfg[1].Source)
			assert.Equal(t, "/bar", cfg[1].Destination)
		}
	})

	t.Run("mixed", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`[
			"foo",
			{ "from": "foo" },
			"bar",
			{ "from": "bar", "source": "/foo", "destination": "/bar/" }
		]`))

		if assert.NoError(t, err) {
			assert.Len(t, cfg, 4)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[0].From)
			assert.Equal(t, "foo", cfg[0].Source)
			assert.Equal(t, "./", cfg[0].Destination)
			assert.Equal(t, "foo", cfg[1].From)
			assert.Equal(t, "", cfg[1].Source)
			assert.Equal(t, "", cfg[1].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[2].From)
			assert.Equal(t, "bar", cfg[2].Source)
			assert.Equal(t, "./", cfg[2].Destination)
			assert.Equal(t, "bar", cfg[3].From)
			assert.Equal(t, "/foo", cfg[3].Source)
			assert.Equal(t, "/bar/", cfg[3].Destination)
		}
	})
}
