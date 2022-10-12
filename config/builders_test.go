package config_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestBuildersConfigYAML(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    builders:
      - python:
          version: python2.7
          requirements: [ requirements.txt ]
      - node:
          requirements: [ package.json, package-lock.json ]
          use-npm-ci: false
      - custom:
          command: [ make, deps ]
          requirements: [ Makefile, vendor ]`))

	if assert.NoError(t, err) {
		expectedPythonConfig := config.PythonConfig{
			Version: "python2.7",
			Requirements: config.RequirementsConfig{
				{From: "local", Source: "requirements.txt"},
			},
		}
		expectedNodeConfig := config.NodeConfig{
			Requirements: config.RequirementsConfig{
				{From: "local", Source: "package.json"},
				{From: "local", Source: "package-lock.json"},
			},
			UseNpmCi: config.Flag{
				True: false,
				Set:  true,
			},
		}
		expectedCustomBuilderConfig := config.BuilderConfig{
			Command: []string{"make", "deps"},
			Requirements: config.RequirementsConfig{
				{From: "local", Source: "Makefile"},
				{From: "local", Source: "vendor"},
			},
		}

		// Builder configs should appear in the order they were declared
		assert.Equal(t, expectedPythonConfig, cfg.Builders[0])
		assert.Equal(t, expectedNodeConfig, cfg.Builders[1])
		assert.Equal(t, expectedCustomBuilderConfig, cfg.Builders[2])
	}
}

func TestRepeatedConfigurations(t *testing.T) {
	for _, rc := range repeatedConfigs {
		t.Run(rc.typeOfRepetition, func(t *testing.T) {
			cfg, err := config.ReadYAMLConfig([]byte(rc.config))

			if rc.expectedBuilders == nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, rc.expectedBuilders, cfg.Builders)
			}
		})
	}
}

var repeatedConfigs = []struct {
	typeOfRepetition string
	config           string
	expectedBuilders config.BuildersConfig
}{
	{"Multiple instances of one type of predefined builder are not allowed", `---
        version: v4
        base: foo
        builders:
          - python:
              version: python2.7
              requirements: [ requirements.txt ]
          - node:
              requirements: [ package.json, package-lock.json ]
              use-npm-ci: false
          - python:
              version: python3
              requirements: [ requirements-redux.txt ]`,
		nil},
	{"Multiple instances of custom builders are allowed", `---
       version: v4
       base: foo
       builders:
         - python:
             version: python2.7
             requirements: [ requirements.txt ]
         - custom:
             command: [ make, deps ]
             requirements: [ Makefile, vendor ]
         - custom:
             command: [ npm, run-script, build:vue ]
             requirements: [ vue.config.js ]`,
		config.BuildersConfig{
			config.PythonConfig{
				Version: "python2.7",
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "requirements.txt"},
				},
			},
			config.BuilderConfig{
				Command: []string{"make", "deps"},
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "Makefile"},
					{From: "local", Source: "vendor"},
				},
			},
			config.BuilderConfig{
				Command: []string{"npm", "run-script", "build:vue"},
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "vue.config.js"},
				},
			},
		},
	},
}

func TestBuildersDisallowingFields(t *testing.T) {
	configTemplate := `---
    version: v4
    base: foo
    %s
    builders:
      - custom:
          command: [ make, deps ]
          requirements: [ Makefile, vendor ]`

	t.Run("Builders key cannot be specified together with", func(t *testing.T) {
		for _, df := range disallowingFields {
			t.Run(df.builderType, func(t *testing.T) {
				_, err := config.ReadYAMLConfig([]byte(fmt.Sprintf(configTemplate, df.builderConfig)))
				assert.Contains(t, err.Error(), "notallowedwith")
			})
		}
	})
}

var disallowingFields = []struct {
	builderType   string
	builderConfig string
}{
	{"Python key", `python:
      poetry:
        version: ==1.0.10
      requirements: [pyproject.toml, poetry.lock]`},
	{"Node key", `node:
      requirements: [package.json, package-lock.json]
      use-npm-ci: false`},
	{"Php key", `php:
      requirements: [composer.json]`},
	{"Builder key", `builder:
      command: [make, deps]
      requirements: [Makefile, vendor]`},
}

func TestBuildersConfigMerge(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    builders:
      - node:
          use-npm-ci: true
      - python:
          version: python2.7
      - php:
          requirements: [composer.json]
    variants:
      build:
        builders:
          - python:
              requirements: [requirements.txt]
          - node:
              requirements: [package.json, package-lock.json]
          - custom:
              command: [make, deps]
              requirements: [Makefile, vendor]`))

	if assert.NoError(t, err) {
		err = config.ExpandIncludesAndCopies(cfg, "build")
		assert.Nil(t, err)

		expectedBuildersConfig := config.BuildersConfig{
			config.PhpConfig{
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "composer.json"},
				},
			},
			config.PythonConfig{
				Version: "python2.7",
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "requirements.txt"},
				},
			},
			config.NodeConfig{
				UseNpmCi: config.Flag{
					True: true,
					Set:  true,
				},
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "package.json"},
					{From: "local", Source: "package-lock.json"},
				},
			},
			config.BuilderConfig{
				Command: []string{"make", "deps"},
				Requirements: config.RequirementsConfig{
					{From: "local", Source: "Makefile"},
					{From: "local", Source: "vendor"},
				},
			},
		}

		assert.Equal(t, expectedBuildersConfig, cfg.Variants["build"].Builders)
	}
}
