package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
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
				build.CopyFrom{"foo", build.Copy{[]string{"./"}, "./"}},
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
				build.CopyFrom{"foo", build.Copy{[]string{"./"}, "./"}},
				instructions[1],
			)
			assert.Equal(
				t,
				build.CopyFrom{"bar", build.Copy{[]string{"/foo"}, "/bar/"}},
				instructions[2],
			)
		}
	})

	t.Run("regression", func(t *testing.T) {
		cfg := config.RequirementsConfig{}
		err := cfg.UnmarshalJSON([]byte(`[
			".git",
			"Makefile",
			"go.mod",
			"go.sum",
			"api/",
			"build/",
			"cmd/",
			"config/",
			"docker/",
			"meta/",
			"scripts/",
			"vendor/"
		]`))

		if assert.NoError(t, err) {
			instructions := cfg.InstructionsForPhase(build.PhasePreInstall)
			assert.Len(t, instructions, 9)
			assert.Equal(
				t,
				build.Copy{[]string{".git", "Makefile", "go.mod", "go.sum"}, "./"},
				instructions[0],
			)
			assert.Equal(t, build.Copy{[]string{"api/"}, "api/"}, instructions[1])
			assert.Equal(t, build.Copy{[]string{"build/"}, "build/"}, instructions[2])
			assert.Equal(t, build.Copy{[]string{"cmd/"}, "cmd/"}, instructions[3])
			assert.Equal(t, build.Copy{[]string{"config/"}, "config/"}, instructions[4])
			assert.Equal(t, build.Copy{[]string{"docker/"}, "docker/"}, instructions[5])
			assert.Equal(t, build.Copy{[]string{"meta/"}, "meta/"}, instructions[6])
			assert.Equal(t, build.Copy{[]string{"scripts/"}, "scripts/"}, instructions[7])
			assert.Equal(t, build.Copy{[]string{"vendor/"}, "vendor/"}, instructions[8])
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
			assert.Equal(t, "", cfg[0].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[1].From)
			assert.Equal(t, "bar", cfg[1].Source)
			assert.Equal(t, "", cfg[1].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[2].From)
			assert.Equal(t, "xyzzy/plugh", cfg[2].Source)
			assert.Equal(t, "", cfg[2].Destination)
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
			assert.Equal(t, "", cfg[0].Destination)
			assert.Equal(t, "foo", cfg[1].From)
			assert.Equal(t, "", cfg[1].Source)
			assert.Equal(t, "", cfg[1].Destination)
			assert.Equal(t, config.LocalArtifactKeyword, cfg[2].From)
			assert.Equal(t, "bar", cfg[2].Source)
			assert.Equal(t, "", cfg[2].Destination)
			assert.Equal(t, "bar", cfg[3].From)
			assert.Equal(t, "/foo", cfg[3].Source)
			assert.Equal(t, "/bar/", cfg[3].Destination)
		}
	})
}
