package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func TestLivesConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v1
    base: foo
    lives:
      in: /some/directory
      as: foouser
      uid: 123
      gid: 223
    variants:
      development: {}`))

	if assert.NoError(t, err) {
		assert.Equal(t, "/some/directory", cfg.Lives.In)
		assert.Equal(t, "foouser", cfg.Lives.As)
		assert.Equal(t, uint(123), cfg.Lives.UID)
		assert.Equal(t, uint(223), cfg.Lives.GID)

		variant, err := config.ExpandVariant(cfg, "development")

		if assert.NoError(t, err) {
			assert.Equal(t, "/some/directory", variant.Lives.In)
			assert.Equal(t, "foouser", variant.Lives.As)
			assert.Equal(t, uint(123), variant.Lives.UID)
			assert.Equal(t, uint(223), variant.Lives.GID)
		}
	}
}

func TestLivesConfigDefaults(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    version: v1
    base: foo`))

	if assert.NoError(t, err) {
		assert.Equal(t, "somebody", cfg.Lives.As)
		assert.Equal(t, uint(65533), cfg.Lives.UID)
		assert.Equal(t, uint(65533), cfg.Lives.GID)
	}
}

func TestLivesConfigInstructions(t *testing.T) {
	cfg := config.LivesConfig{
		In: "/some/directory",
		UserConfig: config.UserConfig{
			As:  "foouser",
			UID: 123,
			GID: 223,
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.RunAll{[]build.Run{
				{"groupadd -o -g %s -r", []string{"223", "foouser"}},
				{"useradd -o -m -d %s -r -g %s -u %s", []string{"/home/foouser", "foouser", "123", "foouser"}},
				{"mkdir -p", []string{"/some/directory"}},
				{"chown %s:%s", []string{"123", "223", "/some/directory"}},
				{"mkdir -p", []string{"/opt/lib"}},
				{"chown %s:%s", []string{"123", "223", "/opt/lib"}},
			}}},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.WorkingDirectory{"/some/directory"},
			},
			cfg.InstructionsForPhase(build.PhasePrivilegeDropped),
		)
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestLivesConfigValidation(t *testing.T) {
	t.Run("in", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          in: /foo`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("non-root", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          in: /`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "/" is not a valid absolute non-root path`, msg)
			}
		})

		t.Run("non-root tricky", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          in: /foo/..`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "/foo/.." is not a valid absolute non-root path`, msg)
			}
		})

		t.Run("absolute", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          in: foo/bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `in: "foo/bar" is not a valid absolute non-root path`, msg)
			}
		})
	})

	t.Run("as", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          as: foo-bar.baz`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("optional", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives: {}`))

			assert.False(t, config.IsValidationError(err))
		})

		t.Run("no spaces", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          as: foo bar`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "foo bar" is not a valid user name`, msg)
			}
		})

		t.Run("long enough", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          as: fo`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "fo" is not a valid user name`, msg)
			}
		})

		t.Run("not root", func(t *testing.T) {
			_, err := config.ReadConfig([]byte(`---
        version: v1
        lives:
          as: root`))

			if assert.True(t, config.IsValidationError(err)) {
				msg := config.HumanizeValidationError(err)

				assert.Equal(t, `as: "root" is not a valid user name`, msg)
			}
		})
	})
}
