package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func ExampleBuildIncludesDepGraph() {
	cfg, _ := config.ReadYAMLConfig([]byte(`---
    version: v4
    variants:
      varA: { includes: [varB, varC] }
      varB: { includes: [varD, varE] }
      varC: {}
      varD: { includes: [varF] }
      varE: {}
      varF: {}`))

	config.BuildIncludesDepGraph(cfg)
	includes, _ := cfg.IncludesDepGraph.GetDeps("varA")

	fmt.Printf("%v\n", includes)

	// Output: [varF varD varE varB varC]
}

func TestReadYAMLConfigErrorsOnUnknownYAML(t *testing.T) {
	_, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    newphone: whodis
    variants:
      foo: {}`))

	assert.EqualError(t,
		err,
		`json: unknown field "newphone"`)
}

func TestReadYAMLConfigValidateVersionBeforeStrictUnmarshal(t *testing.T) {
	_, err := config.ReadYAMLConfig([]byte(`---
    version: foo
    newphone: whodis
    variants:
      foo: {}`))

	if assert.True(t, config.IsValidationError(err)) {
		msg := config.HumanizeValidationError(err)

		assert.Equal(t, `version: config version "foo" is unsupported`, msg)
	}
}

func TestResolveIncludesPreventsInfiniteRecursion(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    variants:
      varA: { includes: [varB] }
      varB: { includes: [varA] }`))

	assert.NoError(t, err)

	config.BuildIncludesDepGraph(cfg)
	_, err2 := cfg.IncludesDepGraph.GetDeps("varA")

	assert.EqualError(t, err2, "Detected dependency graph cycle at 'varA'")
}

func TestMultiLevelIncludes(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo-slim
    variants:
      build:
        base: foo-devel
        runs: { as: foo }
      development:
        includes: [build]
        runs: { uid: 123 }
      test:
        includes: [development]
        runs: { insecurely: true }`))

	if assert.NoError(t, err) {
		config.BuildIncludesDepGraph(cfg)
		dev, _ := config.ExpandVariant(cfg, "development")

		assert.Equal(t, "foo-devel", dev.Base)
		assert.Equal(t, "foo", dev.Runs.As)
		assert.Equal(t, uint(123), dev.Runs.UID)

		test, _ := config.ExpandVariant(cfg, "test")

		assert.Equal(t, "foo-devel", test.Base)
		assert.Equal(t, "foo", test.Runs.As)
		assert.Equal(t, uint(123), test.Runs.UID)

		assert.True(t, test.Runs.Insecurely.True)
	}
}

func TestCopiesIncludes(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo-slim
    variants:
      blah:
        base: foo-test
      build:
        base: foo-devel
        runs: { as: foo }
      development:
        includes: [build]
        runs: { uid: 123 }
      test:
        includes: [blah]
        copies: [development]
        runs: { insecurely: true }`))

	if assert.NoError(t, err) {

		_ = config.ExpandIncludesAndCopies(cfg, "test")
		test, _ := config.GetVariant(cfg, "test")
		dev, _ := config.GetVariant(cfg, "development")

		assert.Equal(t, "foo-devel", dev.Base)
		assert.Equal(t, "foo", dev.Runs.As)

		assert.Equal(t, "foo-test", test.Base)
		assert.Equal(t, "runuser", test.Runs.As)

		assert.True(t, test.Runs.Insecurely.True)
	}
}

func TestExpandIncludesAndCopies_TransitiveAndMergedCopies(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo
    variants:
      build: {}
      foo:
        copies:
          - from: build
            source: /foo
      bar:
        includes: [foo]
        copies:
          - from: build
            source: /bar`))

	if assert.NoError(t, err) {

		err = config.ExpandIncludesAndCopies(cfg, "bar")

		if assert.NoError(t, err) {
			bar, _ := config.GetVariant(cfg, "bar")

			assert.Len(t, bar.Copies, 2)

			assert.Equal(t, bar.Copies[0].From, "build")
			assert.Equal(t, bar.Copies[0].Source, "/foo")

			assert.Equal(t, bar.Copies[1].From, "build")
			assert.Equal(t, bar.Copies[1].Source, "/bar")
		}
	}
}

func TestMultiIncludes(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    variants:
      mammal:
        base: neutral
      human:
        base: moral
        includes: [mammal]
      lizard:
        base: immoral
      lizardman:
        includes: [human, lizard]`))

	if assert.NoError(t, err) {
		config.BuildIncludesDepGraph(cfg)
		variant, err := config.ExpandVariant(cfg, "lizardman")

		if assert.NoError(t, err) {
			assert.Equal(t, "immoral", variant.Base)
		}
	}
}

func TestGetVariant(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: foo-slim
    variants:
      build:
        base: foo-devel
        runs: { as: foo }
      development:
        includes: [build]
        runs: { uid: 123 }
      test:
        includes: [development]
        runs: { insecurely: true }`))

	assert.NoError(t, err)

	config.BuildIncludesDepGraph(cfg)
	dev, _ := config.GetVariant(cfg, "development")
	assert.Equal(t, "", dev.Base)
	assert.Equal(t, "", dev.Runs.As)
	assert.Equal(t, uint(123), dev.Runs.UID)

	err = config.ExpandIncludesAndCopies(cfg, "development")
	assert.Nil(t, err)
	_, err = config.GetVariant(cfg, "development")
	assert.NoError(t, err)

	dev, _ = config.GetVariant(cfg, "development")
	assert.Equal(t, "foo-devel", dev.Base)
	assert.Equal(t, "foo", dev.Runs.As)
	assert.Equal(t, uint(123), dev.Runs.UID)

}
