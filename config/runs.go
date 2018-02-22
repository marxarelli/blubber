package config

import (
	"phabricator.wikimedia.org/source/blubber/build"
)

// RunsConfig holds configuration fields related to the application's
// runtime environment.
//
type RunsConfig struct {
	UserConfig  `yaml:",inline"`
	Environment map[string]string `yaml:"environment" validate:"envvars"` // environment variables
}

// Merge takes another RunsConfig and overwrites this struct's fields. All
// fields except Environment are overwritten if set. The latter is an additive
// merge.
//
func (run *RunsConfig) Merge(run2 RunsConfig) {
	run.UserConfig.Merge(run2.UserConfig)

	if run.Environment == nil {
		run.Environment = make(map[string]string)
	}

	for name, value := range run2.Environment {
		run.Environment[name] = value
	}
}

// InstructionsForPhase injects build instructions related to the runtime
// configuration.
//
// PhasePrivileged
//
// Creates LocalLibPrefix directory and unprivileged user home directory,
// creates the unprivileged user and its group, and sets up directory
// permissions.
//
// PhasePrivilegeDropped
//
// Injects build.Env instructions for all names/values defined by
// RunsConfig.Environment.
//
func (run RunsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePrivileged:
		return []build.Instruction{build.RunAll{
			build.CreateUser(run.As, run.UID, run.GID),
		}}
	case build.PhasePrivilegeDropped:
		if len(run.Environment) > 0 {
			return []build.Instruction{
				build.Env{run.Environment},
			}
		}
	}

	return []build.Instruction{}
}
