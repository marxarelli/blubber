package config_test

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"phabricator.wikimedia.org/source/blubber.git/build"
	"phabricator.wikimedia.org/source/blubber.git/config"
)

func TestRunsConfig(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    runs:
      as: someuser
      in: /some/directory
      uid: 666
      gid: 777
      environment: { FOO: bar }
    variants:
      development: {}`))

	assert.Nil(t, err)

	variant, err := config.ExpandVariant(cfg, "development")

	assert.Nil(t, err)

	assert.Equal(t, "someuser", variant.Runs.As)
	assert.Equal(t, "/some/directory", variant.Runs.In)
	assert.Equal(t, 666, variant.Runs.Uid)
	assert.Equal(t, 777, variant.Runs.Gid)
	assert.Equal(t, map[string]string{"FOO": "bar"}, variant.Runs.Environment)
}

func TestRunsHomeWithUser(t *testing.T) {
	runs := config.RunsConfig{As: "someuser"}

	assert.Equal(t, "/home/someuser", runs.Home())
}

func TestRunsHomeWithoutUser(t *testing.T) {
	runs := config.RunsConfig{}

	assert.Equal(t, "/root", runs.Home())
}

func TestRunsConfigInstructions(t *testing.T) {
	cfg := config.RunsConfig{
		As:  "someuser",
		In:  "/some/directory",
		Uid: 666,
		Gid: 777,
		Environment: map[string]string{
			"fooname": "foovalue",
			"barname": "barvalue",
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{build.RunAll{[]build.Run{
				{"mkdir -p", []string{"/some/directory"}},
				{"groupadd -o -g %s -r", []string{"777", "someuser"}},
				{"useradd -o -m -d %s -r -g %s -u %s", []string{"/home/someuser", "someuser", "666", "someuser"}},
				{"chown %s:%s", []string{"someuser", "someuser", "/some/directory"}},
			}}},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{"HOME": "/home/someuser"}},
				build.Env{map[string]string{"barname": "barvalue", "fooname": "foovalue"}},
			},
			cfg.InstructionsForPhase(build.PhasePrivilegeDropped),
		)

		t.Run("with empty Environment", func(t *testing.T) {
			cfg.Environment = map[string]string{}

			assert.Equal(t,
				[]build.Instruction{
					build.Env{map[string]string{"HOME": "/home/someuser"}},
				},
				cfg.InstructionsForPhase(build.PhasePrivilegeDropped),
			)
		})
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}
