package config

import (
	"sort"
	"strconv"

	"phabricator.wikimedia.org/source/blubber.git/build"
)

type RunsConfig struct {
	In          string            `yaml:"in"`
	As          string            `yaml:"as"`
	Uid         int               `yaml:"uid"`
	Gid         int               `yaml:"gid"`
	Environment map[string]string `yaml:"environment"`
}

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

func (run RunsConfig) Home() string {
	if run.As == "" {
		return "/root"
	} else {
		return "/home/" + run.As
	}
}

func (run RunsConfig) EnvironmentDefinitions() []string {
	defs := make([]string, 0, len(run.Environment))
	names := make([]string, 0, len(run.Environment))

	for name := range run.Environment {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		defs = append(defs, name+"="+strconv.Quote(run.Environment[name]))
	}

	return defs
}

func (run RunsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := []build.Instruction{}

	switch phase {
	case build.PhasePrivileged:
		if run.In != "" {
			ins = append(ins, build.Instruction{build.Run, []string{
				"mkdir -p ", run.In,
			}})
		}

		if run.As != "" {
			ins = append(ins, build.Instruction{build.Run, []string{
				"groupadd -o -g ", strconv.Itoa(run.Gid), " -r ", run.As, " && ",
				"useradd -o -m -d ", strconv.Quote(run.Home()), " -r -g ", run.As,
				" -u ", strconv.Itoa(run.Uid), " ", run.As,
			}})

			if run.In != "" {
				ins = append(ins, build.Instruction{build.Run, []string{
					"chown ", run.As, ":", run.As, " ", run.In,
				}})
			}
		}
	case build.PhasePrivilegeDropped:
		ins = append(ins, build.Instruction{build.Env, []string{
			"HOME=" + strconv.Quote(run.Home()),
		}})

		ins = append(ins, build.Instruction{build.Env, run.EnvironmentDefinitions()})
	}

	return ins
}
