package config

import (
	"strconv"
	"phabricator.wikimedia.org/source/blubber.git/build"
)

type RunsConfig struct {
	In string `yaml:"in"`
	As string `yaml:"as"`
	Uid int `yaml:"uid"`
	Gid int `yaml:"gid"`
}

func (run *RunsConfig) Merge(run2 RunsConfig) {
	if run2.In != "" { run.In = run2.In }
	if run2.As != "" { run.As = run2.As }
	if run2.Uid != 0 { run.Uid = run2.Uid }
	if run2.Gid != 0 { run.Gid = run2.Gid }
}

func (run RunsConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := []build.Instruction{}

	switch phase {
	case build.PhasePrivileged:
		if run.In != "" {
			ins = append(ins, build.Instruction{build.Run, []string{"mkdir -p ", run.In}})
		}

		if run.As != "" {
			ins = append(ins, build.Instruction{build.Run, []string{
				"groupadd -o -g ", strconv.Itoa(run.Gid), " -r ", run.As, " && ",
				"useradd -o -m -d /home/", run.As, " -r -g ", run.As,
				" -u ", strconv.Itoa(run.Uid), " ", run.As,
			}})

			if run.In != "" {
				ins = append(ins, build.Instruction{build.Run, []string{
					"chown ", run.As, ":", run.As, " ", run.In,

				}})
			}
		}
	case build.PhasePrivilegeDropped:
		if run.As != "" {
			ins = append(ins, build.Instruction{build.Env, []string{
				"HOME=\"/home/" + run.As + "\"",
			}})
		}
	}

	return ins
}
