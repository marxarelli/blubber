package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func ExampleResolveIncludes() {
	cfg, _ := config.ReadYAMLConfig([]byte(`---
    version: v3
    variants:
      varA: { includes: [varB, varC] }
      varB: { includes: [varD, varE] }
      varC: {}
      varD: { includes: [varF] }
      varE: {}
      varF: {}`))

	includes, _ := config.ResolveIncludes(cfg, "varA")

	fmt.Printf("%v\n", includes)

	// Output: [varF varD varE varB varC varA]
}

func TestReadYAMLConfigErrorsOnUnknownYAML(t *testing.T) {
	_, err := config.ReadYAMLConfig([]byte(`---
    version: v3
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
    version: v3
    variants:
      varA: { includes: [varB] }
      varB: { includes: [varA] }`))

	assert.NoError(t, err)

	_, err2 := config.ResolveIncludes(cfg, "varA")

	assert.EqualError(t, err2, "variant expansion detected loop")
}

func TestMultiLevelIncludes(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v3
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

func TestMultiIncludes(t *testing.T) {
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v3
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
		variant, err := config.ExpandVariant(cfg, "lizardman")

		if assert.NoError(t, err) {
			assert.Equal(t, "immoral", variant.Base)
		}
	}
}
