package config

import (
	"strconv"

	"phabricator.wikimedia.org/source/blubber/build"
)

// LocalLibPrefix declares the shared directory into which application level
// dependencies will be installed.
//
const LocalLibPrefix = "/opt/lib"

// RunsConfig holds configuration fields related to the application's
// runtime environment.
//
type RunsConfig struct {
	In          string            `yaml:"in"`          // working directory
	As          string            `yaml:"as"`          // unprivileged user
	Uid         int               `yaml:"uid"`         // unprivileged UID
	Gid         int               `yaml:"gid"`         // unprivileged GID
	Environment map[string]string `yaml:"environment"` // environment variables
}

// Merge takes another RunsConfig and overwrites this struct's fields. All
// fields except Environment are overwritten if set. The latter is an additive
// merge.
//
func (run *RunsConfig) Merge(run2 RunsConfig) {
	if run2.In != "" {
		run.In = run2.In
	}
	if run2.As != "" {
		run.As = run2.As
	}
	if run2.Uid != 0 {
		run.Uid = run2.Uid
	}
	if run2.Gid != 0 {
		run.Gid = run2.Gid
	}

	if run.Environment == nil {
		run.Environment = make(map[string]string)
	}

	for name, value := range run2.Environment {
		run.Environment[name] = value
	}
}

// Home returns the home directory for the configured user, or /root if no
// user is set.
//
func (run RunsConfig) Home() string {
	if run.As == "" {
		return "/root"
	} else {
		return "/home/" + run.As
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
// Injects build.Env instructions for the user home directory and all
// names/values defined by RunsConfig.Environment.
//
func (run RunsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := []build.Instruction{}

	switch phase {
	case build.PhasePrivileged:
		runAll := build.RunAll{[]build.Run{
			{"mkdir -p", []string{LocalLibPrefix}},
		}}

		if run.In != "" {
			runAll.Runs = append(runAll.Runs,
				build.Run{"mkdir -p", []string{run.In}},
			)
		}

		if run.As != "" {
			runAll.Runs = append(runAll.Runs,
				build.Run{"groupadd -o -g %s -r",
					[]string{strconv.Itoa(run.Gid), run.As}},
				build.Run{"useradd -o -m -d %s -r -g %s -u %s",
					[]string{run.Home(), run.As, strconv.Itoa(run.Uid), run.As}},
				build.Run{"chown %s:%s",
					[]string{run.As, run.As, LocalLibPrefix}},
			)

			if run.In != "" {
				runAll.Runs = append(runAll.Runs,
					build.Run{"chown %s:%s",
						[]string{run.As, run.As, run.In}},
				)
			}
		}

		if len(runAll.Runs) > 0 {
			ins = append(ins, runAll)
		}
	case build.PhasePrivilegeDropped:
		ins = append(ins, build.Env{map[string]string{"HOME": run.Home()}})

		if len(run.Environment) > 0 {
			ins = append(ins, build.Env{run.Environment})
		}
	}

	return ins
}
