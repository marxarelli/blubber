package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"phabricator.wikimedia.org/source/blubber/config"
)

func TestPolicyRead(t *testing.T) {
	policy, err := config.ReadPolicy([]byte(`---
    enforcements:
      - path: variants.production.runs.as
        rule: ne=root
      - path: base
        rule: oneof=debian:jessie debian:stretch`))

	if assert.NoError(t, err) {
		if assert.Len(t, policy.Enforcements, 2) {
			assert.Equal(t, "variants.production.runs.as", policy.Enforcements[0].Path)
			assert.Equal(t, "ne=root", policy.Enforcements[0].Rule)

			assert.Equal(t, "base", policy.Enforcements[1].Path)
			assert.Equal(t, "oneof=debian:jessie debian:stretch", policy.Enforcements[1].Rule)
		}
	}
}

func TestPolicyValidate(t *testing.T) {
	cfg := config.Config{
		CommonConfig: config.CommonConfig{
			Base: "foo:tag",
		},
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Runs: config.RunsConfig{
						UserConfig: config.UserConfig{
							As: "root",
						},
					},
				},
			},
		},
	}

	policy := config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "variants.foo.runs.as", Rule: "ne=root"},
		},
	}

	assert.EqualError(t,
		policy.Validate(cfg),
		`value for "variants.foo.runs.as" violates policy rule "ne=root"`,
	)

	policy = config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "base", Rule: "oneof=debian:jessie debian:stretch"},
		},
	}

	assert.EqualError(t,
		policy.Validate(cfg),
		`value for "base" violates policy rule "oneof=debian:jessie debian:stretch"`,
	)
}

func TestEnforcementOnFlag(t *testing.T) {
	cfg := config.Config{
		Variants: map[string]config.VariantConfig{
			"production": config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Node: config.NodeConfig{
						Dependencies: config.Flag{True: true},
					},
				},
			},
		},
	}

	policy := config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "variants.production.node.dependencies", Rule: "isfalse"},
		},
	}

	assert.Error(t,
		policy.Validate(cfg),
		`value for "variants.production.node.dependencies" violates policy rule "isfalse"`,
	)

}

func TestResolveYAMLPath(t *testing.T) {
	cfg := config.Config{
		Variants: map[string]config.VariantConfig{
			"foo": config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Runs: config.RunsConfig{
						UserConfig: config.UserConfig{
							As: "root",
						},
					},
				},
			},
		},
	}

	val, err := config.ResolveYAMLPath("variants.foo.runs.as", cfg)

	if assert.NoError(t, err) {
		assert.Equal(t, "root", val)
	}
}
