package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/build"
	"phabricator.wikimedia.org/source/blubber/config"
)

func TestPythonConfigUnmarshalMerge(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo
    python:
      version: python2.7
      requirements: [requirements.txt]
    variants:
      test:
        python:
          version: python3
          requirements: [other-requirements.txt, requirements-test.txt]`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"requirements.txt"}, cfg.Python.Requirements)
		assert.Equal(t, "python2.7", cfg.Python.Version)

		variant, err := config.ExpandVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"other-requirements.txt", "requirements-test.txt"}, variant.Python.Requirements)
			assert.Equal(t, "python3", variant.Python.Version)
		}
	}
}

func TestPythonConfigMergeEmpty(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo
    python:
      requirements: [requirements.txt]
    variants:
      test:
        python:
          requirements: []`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"requirements.txt"}, cfg.Python.Requirements)

		variant, err := config.ExpandVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, variant.Python.Requirements)
		}
	}
}

func TestPythonConfigDoNotMergeNil(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: foo
    python:
      requirements: [requirements.txt]
    variants:
      test:
        python:
          requirements: ~`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"requirements.txt"}, cfg.Python.Requirements)

		variant, err := config.ExpandVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"requirements.txt"}, variant.Python.Requirements)
		}
	}
}

func TestPythonConfigInstructionsNoRequirementsWithVersion(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python2.7",
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "easy_install", "pip"}},
					{"python2.7", []string{"-m", "pip", "install", "-U", "setuptools", "wheel", "tox"}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.RunAll{[]build.Run{
					{"mkdir -p", []string{"/opt/lib/python"}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_NO_INDEX": "1",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestPythonConfigInstructionsNoRequirementsNoVersion(t *testing.T) {
	cfg := config.PythonConfig{}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestPythonConfigInstructionsWithRequirements(t *testing.T) {
	cfg := config.PythonConfig{
		Version:      "python2.7",
		Requirements: []string{"requirements.txt", "requirements-test.txt", "docs/requirements.txt"},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "easy_install", "pip"}},
					{"python2.7", []string{"-m", "pip", "install", "-U", "setuptools", "wheel", "tox"}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePrivileged),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.RunAll{[]build.Run{
					{"mkdir -p", []string{"/opt/lib/python"}},
					{"mkdir -p", []string{"docs/"}},
				}},
				build.Copy{[]string{"requirements.txt", "requirements-test.txt"}, "./"},
				build.Copy{[]string{"docs/requirements.txt"}, "docs/"},
				build.Run{"python2.7", []string{"-m", "pip", "wheel",
					"-r", "requirements.txt",
					"-r", "requirements-test.txt",
					"-r", "docs/requirements.txt",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_NO_INDEX": "1",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestPythonConfigRequirementsByDir(t *testing.T) {
	cfg := config.PythonConfig{
		Requirements: []string{"foo", "./bar", "./c/c-foo", "b/b-foo", "b/b-bar", "a/a-foo"},
	}

	sortedDirs, reqsByDir := cfg.RequirementsByDir()

	assert.Equal(t,
		[]string{
			"./",
			"a/",
			"b/",
			"c/",
		},
		sortedDirs,
	)

	assert.Equal(t,
		map[string][]string{
			"./": []string{"foo", "bar"},
			"c/": []string{"c/c-foo"},
			"b/": []string{"b/b-foo", "b/b-bar"},
			"a/": []string{"a/a-foo"},
		},
		reqsByDir,
	)
}

func TestPythonConfigRequirementsArgs(t *testing.T) {
	cfg := config.PythonConfig{
		Requirements: []string{"foo", "bar", "baz/qux"},
	}

	assert.Equal(t,
		[]string{
			"-r", "foo",
			"-r", "bar",
			"-r", "baz/qux",
		},
		cfg.RequirementsArgs(),
	)
}