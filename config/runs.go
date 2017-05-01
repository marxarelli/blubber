package config

import (
	"strconv"
	"strings"
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

func (run RunsConfig) Commands() []string {
	cmds := []string{}

	if run.In != "" {
		cmds = append(cmds, strings.Join([]string{"mkdir -p", run.In}, " "))
	}

	if run.As != "" {
		cmd := []string{
			"groupadd -o -g", strconv.Itoa(run.Gid), "-r", run.As, "&&",
			"useradd -o -m -r -g", run.As, "-u", strconv.Itoa(run.Uid), run.As,
		}

		cmds = append(cmds, strings.Join(cmd, " "))

		if run.In != "" {
			owner := strings.Join([]string{run.As, ":", run.As}, "")
			cmds = append(cmds, strings.Join([]string{"chown", owner, run.In}, " "))
		}
	}

	return cmds
}
