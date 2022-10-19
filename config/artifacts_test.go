package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
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

func TestArtifactsConfigExpand(t *testing.T) {
	t.Run("local with no source/destination", func(t *testing.T) {
		cfg := config.ArtifactsConfig{From: "local"}

		assert.Equal(t, []config.ArtifactsConfig{
			{From: "local", Source: ".", Destination: "."},
		}, cfg.Expand("/app/dir"))
	})

	t.Run("variant with no source/destination", func(t *testing.T) {
		cfg := config.ArtifactsConfig{From: "foo"}

		assert.Equal(t, []config.ArtifactsConfig{
			{From: "foo", Source: "/app/dir", Destination: "/app/dir"},
			{From: "foo", Source: "/opt/lib", Destination: "/opt/lib"},
		}, cfg.Expand("/app/dir"))
	})

	t.Run("variant with source/destination", func(t *testing.T) {
		cfg := config.ArtifactsConfig{From: "foo", Source: "./foo/dir", Destination: "./bar/dir"}

		assert.Equal(t, []config.ArtifactsConfig{
			{From: "foo", Source: "./foo/dir", Destination: "./bar/dir"},
		}, cfg.Expand("/app/dir"))
	})

	t.Run("source but no destination", func(t *testing.T) {
		cfg := config.ArtifactsConfig{From: "foo", Source: "./foo/dir"}

		assert.Equal(t, []config.ArtifactsConfig{
			{From: "foo", Source: "./foo/dir", Destination: "./foo/dir"},
		}, cfg.Expand("/app/dir"))
	})
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

				assert.Equal(t, `from: "foo bar" is not a valid image reference or known variant`, msg)
			}
		})
	})

	t.Run("from: variant|imageref", func(t *testing.T) {
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
			t.Run("can be empty", func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(`---
          version: v4
          variants:
            build: {}
            foo:
              copies:
                - from: build
                  source: /bar`))

				assert.False(t, config.IsValidationError(err))
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

func TestArtifactsEffectiveDestination(t *testing.T) {
	t.Run("where source is a file and destination is a file", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "local",
			Source:      "foo/bar",
			Destination: "foo2/bar2",
		}

		assert.Equal(t, "foo2/bar2", artifact.EffectiveDestination())
	})

	t.Run("where source is a directory", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "local",
			Source:      "foo/bar/",
			Destination: "foo2/bar2",
		}

		assert.Equal(t, "foo2/bar2/", artifact.EffectiveDestination())
	})

	t.Run("where source is a file and destination is a directory", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "local",
			Source:      "foo/bar",
			Destination: "foo2/bar2/",
		}

		assert.Equal(t, "foo2/bar2/bar", artifact.EffectiveDestination())
	})

	t.Run("where source is a file and destination is ./", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "local",
			Source:      "foo/bar",
			Destination: ".",
		}

		assert.Equal(t, "bar", artifact.EffectiveDestination())
	})

	t.Run("where destination is /", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "foo",
			Source:      "foo/bar",
			Destination: "/",
		}

		assert.Equal(t, "/bar", artifact.EffectiveDestination())
	})

	t.Run("where source is /", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "foo",
			Source:      "/",
			Destination: "foo",
		}

		assert.Equal(t, "foo/", artifact.EffectiveDestination())
	})

	t.Run("where source and destination are /", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From:        "foo",
			Source:      "/",
			Destination: "/",
		}

		assert.Equal(t, "/", artifact.EffectiveDestination())
	})
}

func TestArtifactsNormalizedDestination(t *testing.T) {
	t.Run("where source and destination are omitted", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			From: "local",
		}

		assert.Equal(t, "./", artifact.NormalizedDestination())
	})

	t.Run("where destination is omitted", func(t *testing.T) {
		t.Run("and source is a directory", func(t *testing.T) {
			artifact := config.ArtifactsConfig{
				Source: "foo/dir/",
			}

			assert.Equal(t, "foo/dir/", artifact.NormalizedDestination())
		})

		t.Run("and source is a file", func(t *testing.T) {
			artifact := config.ArtifactsConfig{
				Source: "foo/dir",
			}

			assert.Equal(t, "foo/", artifact.NormalizedDestination())
		})
	})

	t.Run("where destination is present", func(t *testing.T) {
		t.Run("and destination is a directory", func(t *testing.T) {
			artifact := config.ArtifactsConfig{
				Destination: "foo/dir/",
			}

			assert.Equal(t, "foo/dir/", artifact.NormalizedDestination())
		})

		t.Run("and source is a directory", func(t *testing.T) {
			artifact := config.ArtifactsConfig{
				Source:      "foo/dir/",
				Destination: "foo/dir",
			}

			assert.Equal(t, "foo/dir/", artifact.NormalizedDestination())
		})

		t.Run("and destination is a file", func(t *testing.T) {
			artifact := config.ArtifactsConfig{
				Source:      "foo",
				Destination: "foo/bar",
			}

			assert.Equal(t, "foo/bar", artifact.NormalizedDestination())
		})
	})
}

func TestArtifactsNormalizedSource(t *testing.T) {
	t.Run("where source is a directory", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			Source: "foo/../bar//",
		}

		assert.Equal(t, "bar/", artifact.NormalizedSource())
	})

	t.Run("where source is a file", func(t *testing.T) {
		artifact := config.ArtifactsConfig{
			Source: "foo/../bar",
		}

		assert.Equal(t, "bar", artifact.NormalizedSource())
	})
}
