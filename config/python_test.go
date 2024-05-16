package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
	"gitlab.wikimedia.org/repos/releng/blubber/config"
)

func TestPythonConfigYAMLMerge(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    python:
      version: python2.7
      requirements: [requirements.txt]
      tox-version: 4.11.1
    variants:
      test:
        python:
          version: python3
          requirements: [other-requirements.txt, requirements-test.txt]
          tox-version: 4.11.2
          use-system-site-packages: true`))

	if assert.NoError(t, err) {
		assert.Equal(t, config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		}, cfg.Python.Requirements)
		assert.Equal(t, "python2.7", cfg.Python.Version)

		err = config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, config.RequirementsConfig{
				{From: "local", Source: "other-requirements.txt"},
				{From: "local", Source: "requirements-test.txt"},
			}, variant.Python.Requirements)
			assert.Equal(t, "python3", variant.Python.Version)
			assert.Equal(t, true, variant.Python.UseSystemSitePackages.True)
			assert.Equal(t, "4.11.2", variant.Python.ToxVersion)
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
		assert.Equal(t, config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		}, cfg.Python.Requirements)

		err = config.ExpandIncludesAndCopies(cfg, "test")
		assert.Nil(t, err)

		variant, err := config.GetVariant(cfg, "test")

		if assert.NoError(t, err) {
			assert.Equal(t, config.RequirementsConfig{}, variant.Python.Requirements)
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
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
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
		Version: "python2.7",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
			{From: "local", Source: "requirements-test.txt"},
			{From: "local", Source: "docs/requirements.txt"},
		},
	}

	t.Run("PhasePrivileged", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivileged))
	})

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{Sources: []string{"requirements.txt", "requirements-test.txt"}, Destination: "./"},
				build.Copy{Sources: []string{"docs/requirements.txt"}, Destination: "docs/"},
				build.Run{Command: "python2.7", Arguments: []string{"-m", "venv", "/opt/lib/venv"}},
				build.Env{Definitions: map[string]string{"PATH": "/opt/lib/venv/bin:$PATH", "VIRTUAL_ENV": "/opt/lib/venv"}},
				build.RunAll{Runs: []build.Run{
					{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
					{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-U", "wheel", "tox", "pip<21"}}}},
				build.Run{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-r", "requirements.txt", "-r", "requirements-test.txt", "-r", "docs/requirements.txt"}}},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePostInstall))
	})
}

func TestPythonConfigUseSystemSitePackages(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python2.7",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
			{From: "local", Source: "requirements-test.txt"},
			{From: "local", Source: "docs/requirements.txt"},
		},
		UseSystemSitePackages: config.Flag{True: true},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t, []build.Instruction{
			build.Copy{Sources: []string{"requirements.txt", "requirements-test.txt"}, Destination: "./"},
			build.Copy{Sources: []string{"docs/requirements.txt"}, Destination: "docs/"},
			build.Run{Command: "python2.7", Arguments: []string{"-m", "venv", "/opt/lib/venv", "--system-site-packages"}},
			build.Env{Definitions: map[string]string{"PATH": "/opt/lib/venv/bin:$PATH", "VIRTUAL_ENV": "/opt/lib/venv"}},
			build.RunAll{Runs: []build.Run{
				{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
				{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-U", "wheel", "tox", "pip<21"}}}},
			build.Run{Command: "python2.7", Arguments: []string{"-m", "pip", "install", "-r", "requirements.txt", "-r", "requirements-test.txt", "-r", "docs/requirements.txt"}}},
			cfg.InstructionsForPhase(build.PhasePreInstall))
	})
}

func TestPythonConfigUseNoDepsFlag(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3.9",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		},
		UseNoDepsFlag: config.Flag{True: true},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{Sources: []string{"requirements.txt"}, Destination: "./"},
				build.Run{Command: "python3.9", Arguments: []string{"-m", "venv", "/opt/lib/venv"}},
				build.Env{Definitions: map[string]string{"PATH": "/opt/lib/venv/bin:$PATH", "VIRTUAL_ENV": "/opt/lib/venv"}},
				build.RunAll{Runs: []build.Run{
					{Command: "python3.9", Arguments: []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
					{Command: "python3.9", Arguments: []string{"-m", "pip", "install", "-U", "wheel", "tox", "pip"}}}},
				build.Run{Command: "python3.9", Arguments: []string{"-m", "pip", "install", "--no-deps", "-r", "requirements.txt"}}},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Run{"python3.9", []string{"-m", "pip", "check"}},
			},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})

}

func TestPythonConfigRequirementsArgs(t *testing.T) {
	cfg := config.PythonConfig{
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "foo"},
			{From: "local", Source: "bar"},
			{From: "local", Source: "baz/qux"},
		},
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

func TestPythonConfigToxVersion(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "requirements.txt"},
		},
		ToxVersion: "1.23.4",
	}

	t.Run("tox version honored", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{Sources: []string{"requirements.txt"}, Destination: "./"},
				build.Run{Command: "python3", Arguments: []string{"-m", "venv", "/opt/lib/venv"}},
				build.Env{Definitions: map[string]string{"PATH": "/opt/lib/venv/bin:$PATH", "VIRTUAL_ENV": "/opt/lib/venv"}},
				build.RunAll{Runs: []build.Run{
					{Command: "python3", Arguments: []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
					{Command: "python3", Arguments: []string{"-m", "pip", "install", "-U", "wheel", "tox==1.23.4", "pip"}}}},
				build.Run{Command: "python3", Arguments: []string{"-m", "pip", "install", "-r", "requirements.txt"}}},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})
}

func TestPythonConfigInstructionsWithPoetry(t *testing.T) {
	cfg := config.PythonConfig{
		Version: "python3",
		Requirements: config.RequirementsConfig{
			{From: "local", Source: "pyproject.toml"},
			{From: "local", Source: "poetry.lock"},
		},
		Poetry: config.PoetryConfig{Version: "==10.0.1"},
	}

	t.Run("PhasePreInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{
				build.Copy{Sources: []string{"pyproject.toml", "poetry.lock"}, Destination: "./"},
				build.Run{Command: "python3", Arguments: []string{"-m", "venv", "/opt/lib/venv"}},
				build.Env{Definitions: map[string]string{"PATH": "/opt/lib/venv/bin:$PATH", "VIRTUAL_ENV": "/opt/lib/venv"}},
				build.RunAll{Runs: []build.Run{
					{Command: "python3", Arguments: []string{"-m", "pip", "install", "-U", "setuptools!=60.9.0"}},
					{Command: "python3", Arguments: []string{"-m", "pip", "install", "-U", "wheel", "tox", "pip"}}}},
				build.Env{Definitions: map[string]string{"POETRY_VIRTUALENVS_PATH": "/opt/lib/poetry"}},
				build.Run{Command: "python3", Arguments: []string{"-m", "pip", "install", "-U", "poetry==10.0.1"}},
				build.Run{Command: "mkdir -p", Arguments: []string{"/opt/lib/poetry"}},
				build.Run{Command: "poetry", Arguments: []string{"install", "--no-root", "--no-dev"}}},
			cfg.InstructionsForPhase(build.PhasePreInstall),
		)
	})

	t.Run("PhasePrivilegeDropped", func(t *testing.T) {
		assert.Empty(t, cfg.InstructionsForPhase(build.PhasePrivilegeDropped))
	})

	t.Run("PhasePostInstall", func(t *testing.T) {
		assert.Equal(t,
			[]build.Instruction{},
			cfg.InstructionsForPhase(build.PhasePostInstall),
		)
	})
}
