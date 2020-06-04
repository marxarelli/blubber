package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestArtifactsConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    variants:
      build: {}
      production:
        copies:
          - from: build
            source: /foo/src
            destination: /foo/dst
          - from: build
            source: /bar/src
            destination: /bar/dst`))

	if assert.NoError(t, err) {
		err := config.ExpandIncludesAndCopies(cfg, "production")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "production")

		if assert.NoError(t, err) {
			assert.Len(t, variant.Copies, 2)

			assert.Contains(t,
				variant.Copies,
				config.ArtifactsConfig{From: "build", Source: "/foo/src", Destination: "/foo/dst"},
			)
			assert.Contains(t,
				variant.Copies,
				config.ArtifactsConfig{From: "build", Source: "/bar/src", Destination: "/bar/dst"},
			)
		}
	}
}

func TestArtifactsConfigInstructions(t *testing.T) {
	cfg := config.ArtifactsConfig{
		From:        "foo",
		Source:      "/source/path",
		Destination: "/destination/path",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhaseInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.CopyFrom{
				"foo",
				build.Copy{[]string{"/source/path"}, "/destination/path"},
			}},
			cfg.InstructionsForPhase(build.PhaseInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestArtifactsConfigValidation(t *testing.T) {
	t.Run("from", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo:
            copies:
              - from: build
                source: /foo
                destination: /bar`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("missing", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo:
            copies:
              - from: ~
                source: /foo
                destination: /bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `from: is required`, msg)
			}
		})

		t.Run("bad", func(t *testing.T) {
			_, err := config.ReadYAMLConfig([]byte(`---
        version: v4
        variants:
          build: {}
          foo:
            copies:
              - from: foo bar
                source: /foo
                destination: /bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `from: references an unknown variant "foo bar"`, msg)
			}
		})
	})

	t.Run("from: variant", func(t *testing.T) {
		t.Run("source", func(t *testing.T) {
			t.Run("with no destination given can be empty", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            build: {}
            foo:
              copies:
                - from: build`))

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("with destination given must not be empty", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            build: {}
            foo:
              copies:
                - from: build
                  destination: /bar`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `source: is required if "destination" is also set`, msg)
				}
			})
		})

		t.Run("destination", func(t *testing.T) {
			t.Run("with no source given can be empty", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            build: {}
            foo:
              copies:
                - from: build`))

				assert.False(t, config.IsValidationError(err))
			})

			t.Run("with source given must not be empty", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            build: {}
            foo:
              copies:
                - from: build
                  source: /bar`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `destination: is required if "source" is also set`, msg)
				}
			})
		})

	})

	t.Run("from: local", func(t *testing.T) {
		t.Run("source", func(t *testing.T) {
			t.Run("must be a relative path", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            foo:
              copies:
                - from: local
                  source: /bad/path
                  destination: ./foo`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `source: path must be relative when "from" is "local"`, msg)
				}
			})

			t.Run("must not use ../", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            foo:
              copies:
                - from: local
                  source: ./funny/../../business
                  destination: ./foo`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `source: path must be relative when "from" is "local"`, msg)
				}
			})
		})

		t.Run("destination", func(t *testing.T) {
			t.Run("must be a relative path", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            foo:
              copies:
                - from: local
                  source: ./foo
                  destination: /bad/path`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `destination: path must be relative when "from" is "local"`, msg)
				}
			})

			t.Run("must not use ../", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            foo:
              copies:
                - from: local
                  source: ./foo
                  destination: ./funny/../../business`))

				if assert.True(t, config.IsValidationError(err)) {
					msg := config.HumanizeValidationError(err)

					assert.Equal(t, `destination: path must be relative when "from" is "local"`, msg)
				}
			})
		})
	})
}
