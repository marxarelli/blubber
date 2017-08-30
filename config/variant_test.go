package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestVariantConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    variants:
      build: {}
      production:
        copies: build
        artifacts:
          - from: build
            source: /foo/src
            destination: /foo/dst
          - from: build
            source: /bar/src
            destination: /bar/dst`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "production")

	assert.Nil(t, err)

	assert.Equal(t, "build", variant.Copies)
	assert.Len(t, variant.Artifacts, 2)
}

func TestVariantDependencies(t *testing.T) {
	cfg := config.VariantConfig{
		Copies: "foo",
		Artifacts: []config.ArtifactsConfig{
			{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
		},
	}

	assert.Equal(t, []string{"foo", "build"}, cfg.VariantDependencies())
}

func TestVariantConfigInstructions(t *testing.T) {
	cfg := config.VariantConfig{
		CommonConfig: config.CommonConfig{Runs: config.RunsConfig{In: "/srv/service"}},
		Copies:       "foo",
		Artifacts: []config.ArtifactsConfig{
			{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
		},
	}

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.CopyFrom{"foo", build.Copy{[]string{"/srv/service"}, "/srv/service"}},
				build.CopyFrom{"foo", build.Copy{[]string{config.LocalLibPrefix}, config.LocalLibPrefix}},
				build.CopyFrom{"build", build.Copy{[]string{"/foo/src"}, "/foo/dst"}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
