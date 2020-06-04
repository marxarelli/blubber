package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.wikimedia.org/r/blubber/config"
)

func TestPolicyRead(t *testing.T) {
	policy, err := config.ReadYAMLPolicy([]byte(`---
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
		`value: "root", for "variants.foo.runs.as" violates policy rule "ne=root"`,
	)

	policy = config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "base", Rule: "oneof=debian:jessie debian:stretch"},
		},
	}

	assert.EqualError(t,
		policy.Validate(cfg),
		`value: "foo:tag", for "base" violates policy rule "oneof=debian:jessie debian:stretch"`,
	)
}

func TestEnforcementOnFlag(t *testing.T) {
	cfg := config.Config{
		Variants: map[string]config.VariantConfig{
			"production": config.VariantConfig{
				CommonConfig: config.CommonConfig{
					Runs: config.RunsConfig{
						Insecurely: config.Flag{True: true},
					},
				},
			},
		},
	}

	policy := config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "variants.production.runs.insecurely", Rule: "isfalse"},
		},
	}

	assert.Error(t,
		policy.Validate(cfg),
		`value for "variants.production.runs.insecurely" violates policy rule "isfalse"`,
	)

}

func TestResolveJSONPath(t *testing.T) {
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

	val, err := config.ResolveJSONPath("variants.foo.runs.as", cfg)

	if assert.NoError(t, err) {
		assert.Equal(t, "root", val)
	}
}

func TestIncludesPolicyVerification(t *testing.T) {
	variant := "production"
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: docker-registry.wikimedia.org/wikimedia-stretch:latest
    variants:
      dev:
        base: foo
      production:
        includes: [dev]`))

	assert.NoError(t, err)

	err = config.ExpandIncludesAndCopies(cfg, variant)

	assert.NoError(t, err)

	policy := config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "variants.production.base", Rule: "omitempty,startswith=docker-registry.wikimedia.org"},
		},
	}

	assert.Error(t,
		policy.Validate(*cfg),
		`value: foo, for "variants.production.base" violates policy rule "omitempty,startswith=docker-registry.wikimedia.org"`,
	)
}

func TestCopiesPolicyVerification(t *testing.T) {
	variant := "production"
	cfg, err := config.ReadYAMLConfig([]byte(`---
    version: v4
    base: docker-registry.wikimedia.org/wikimedia-stretch:latest
    variants:
      dev:
        base: foo
      pred:
        includes: [dev]
      production:
        copies: [pred]`))

	assert.NoError(t, err)

	err = config.ExpandIncludesAndCopies(cfg, variant)
	assert.NoError(t, err)

	policy := config.Policy{
		Enforcements: []config.Enforcement{
			{Path: "variants.pred.base", Rule: "omitempty,startswith=docker-registry.wikimedia.org"},
		},
	}

	assert.Error(t,
		policy.Validate(*cfg),
		`value: foo, for "variants.pred.base" violates policy rule "omitempty,startswith=docker-registry.wikimedia.org"`,
	)
}
