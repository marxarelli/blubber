package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func ExampleResolveIncludes() {
	cfg, _ := config.ReadConfig([]byte(`---
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

func TestResolveIncludesPreventsInfiniteRecursion(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    variants:
      varA: { includes: [varB] }
      varB: { includes: [varA] }`))

	assert.NoError(t, err)

	_, err2 := config.ResolveIncludes(cfg, "varA")

	assert.EqualError(t, err2, "variant expansion detected loop")
}

func TestMultiLevelIncludes(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
    base: nodejs-slim
    variants:
      build:
        base: nodejs-devel
        node: {env: build}
      development:
        includes: [build]
        node: {env: development}
        entrypoint: [npm, start]
      test:
        includes: [development]
        node: {dependencies: true}
        entrypoint: [npm, test]`))

	assert.NoError(t, err)

	variant, _ := config.ExpandVariant(cfg, "test")

	assert.Equal(t, "nodejs-devel", variant.Base)
	assert.Equal(t, "development", variant.Node.Env)

	devVariant, _ := config.ExpandVariant(cfg, "development")

	assert.True(t, variant.Node.Dependencies.True)
	assert.False(t, devVariant.Node.Dependencies.True)
}

func TestMultiIncludes(t *testing.T) {
	cfg, err := config.ReadConfig([]byte(`---
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