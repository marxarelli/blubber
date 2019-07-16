package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/build"
	"gerrit.wikimedia.org/r/blubber/config"
)

func TestPythonConfigYAMLMerge(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    python:
      version: python2.7
      requirements: [requirements.txt]
    variants:
      test:
        python:
          version: python3
          requirements: [other-requirements.txt, requirements-test.txt]
          use-system-flag: true`))

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"requirements.txt"}, cfg.Python.Requirements)
		assert.Equal(t, "python2.7", cfg.Python.Version)

		variant, err := config.ExpandVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"other-requirements.txt", "requirements-test.txt"}, variant.Python.Requirements)
			assert.Equal(t, "python3", variant.Python.Version)
			assert.Equal(t, true, variant.Python.UseSystemFlag)
		}
	}
}

func TestPythonConfigYAMLMergeEmpty(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
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

func TestPythonConfigYAMLDoNotMergeNil(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
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
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePreInstall))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PYTHONPATH": "/opt/lib/python/site-packages",
					"PATH":       "/opt/lib/python/site-packages/bin:${PATH}",
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
				build.Run{"mkdir -p", []string{"/opt/lib/python"}},
				build.Run{"mkdir -p", []string{"docs/"}},
				build.Copy{[]string{"requirements.txt", "requirements-test.txt"}, "./"},
				build.Copy{[]string{"docs/requirements.txt"}, "docs/"},
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "pip", "wheel",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
					{"python2.7", []string{"-m", "pip", "install",
						"--target", "/opt/lib/python/site-packages",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
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
					"PYTHONPATH":   "/opt/lib/python/site-packages",
					"PATH":         "/opt/lib/python/site-packages/bin:${PATH}",
				}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}

func TestPythonConfigUseSystemFlag(t *testing.T) {
	cfg := config.PythonConfig{
		Version:	   "python2.7",
		Requirements:  []string{"requirements.txt", "requirements-test.txt", "docs/requirements.txt"},
		UseSystemFlag: true,
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Env{map[string]string{
					"PIP_WHEEL_DIR":  "/opt/lib/python",
					"PIP_FIND_LINKS": "file:///opt/lib/python",
				}},
				build.Run{"mkdir -p", []string{"/opt/lib/python"}},
				build.Run{"mkdir -p", []string{"docs/"}},
				build.Copy{[]string{"requirements.txt", "requirements-test.txt"}, "./"},
				build.Copy{[]string{"docs/requirements.txt"}, "docs/"},
				build.RunAll{[]build.Run{
					{"python2.7", []string{"-m", "pip", "wheel",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
					{"python2.7", []string{"-m", "pip", "install", "--system",
						"--target", "/opt/lib/python/site-packages",
						"-r", "requirements.txt",
						"-r", "requirements-test.txt",
						"-r", "docs/requirements.txt",
					}},
				}},
			},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
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
