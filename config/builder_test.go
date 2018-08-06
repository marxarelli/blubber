package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestBuilderConfigYAML(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v2
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
		variant, err := config.ExpandVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make", "-f", "Makefile", "test"}, variant.Builder.Command)
			assert.Equal(t, []string{"Makefile"}, variant.Builder.Requirements)
		}

		variant, err = config.ExpandVariant(cfg, "build")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"make"}, variant.Builder.Command)
			assert.Equal(t, []string{}, variant.Builder.Requirements)
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
		Command:      []string{"make", "-f", "Makefile", "foo"},
		Requirements: []string{"Makefile", "foo", "bar/baz"},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Run{"mkdir -p", []string{"bar/"}},
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
