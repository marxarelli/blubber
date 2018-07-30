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
    variants:
      build:
        builder: [make, -f, Makefile]`))

	if assert.NoError(t, err) {
		variant, err := config.ExpandVariant(cfg, "build")

		assert.Equal(t, []string{"make", "-f", "Makefile"}, variant.Builder)

		assert.Nil(t, err)
	}
}

func TestBuilderConfigInstructions(t *testing.T) {
	cfg := config.BuilderConfig{Builder: []string{"make", "-f", "Makefile"}}

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Run{
					"make",
					[]string{"-f", "Makefile"},
				},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
