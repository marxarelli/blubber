package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestBuilderConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    builder:
      command: [make, -f, Makefile, test]
      requirements: [Makefile]
    variants:
      test: {}
      build:
        builder:
          command: [make]
          requirements: []`))

	if assert.NoError(t, err) {
		err := config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make", "-f", "Makefile", "test"}, variant.Builder.Command)
			assert.Equal(t, config.RequirementsConfig{
				{
					From:        config.LocalArtifactKeyword,
					Source:      "Makefile",
					Destination: "./",
				},
			}, variant.Builder.Requirements)
		}

		err = config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		variant, err = config.GetVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make"}, variant.Builder.Command)
			assert.Equal(t, config.RequirementsConfig{}, variant.Builder.Requirements)
		}
	}
}

func TestBuilderConfigInstructions(t *testing.T) {
	cfg := config.BuilderConfig{Command: []string{"make", "-f", "Makefile"}}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Run{
					"make",
					[]string{"-f", "Makefile"},
				},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}

func TestBuilderConfigInstructionsWithRequirements(t *testing.T) {
	cfg := config.BuilderConfig{
		Command: []string{"make", "-f", "Makefile", "foo"},
		Requirements: config.RequirementsConfig{
			{
				From:        config.LocalArtifactKeyword,
				Source:      "Makefile",
				Destination: "",
			},
			{
				From:        config.LocalArtifactKeyword,
				Source:      "foo",
				Destination: "",
			},
			{
				From:        config.LocalArtifactKeyword,
				Source:      "bar/baz",
				Destination: "",
			},
		},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{[]string{"Makefile", "foo"}, "./"},
				build.Copy{[]string{"bar/baz"}, "bar/"},
				build.Run{
					"make",
					[]string{"-f", "Makefile", "foo"},
				},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}
