package config

import (
	"path"
	"phabricator.wikimedia.org/source/blubber/build"
)

// NodeConfig holds configuration fields related to the Node environment and
// whether/how to install NPM packages.
//
type NodeConfig struct {
	Dependencies Flag   `yaml:"dependencies"` // install dependencies declared in package.json
	Env          string `yaml:"env"`          // environment name ("production" install)
}

// Merge takes another NodeConfig and merges its fields into this one's,
// overwriting both the environment and dependencies flag.
//
func (nc *NodeConfig) Merge(nc2 NodeConfig) {
	nc.Dependencies.Merge(nc2.Dependencies)

	if nc2.Env != "" {
		nc.Env = nc2.Env
	}
}

// InstructionsForPhase injects instructions into the build related to Node
// dependency installation and setting of the NODE_ENV, NODE_PATH, and PATH
// environment variables.
//
// PhasePreInstall
//
// Installs Node package dependencies declared in package.json into the shared
// library directory (/opt/lib). Only production related packages are install
// if NodeConfig.Env is set to "production" in which case `npm dedupe` is also
// run. Installing dependencies during the build.PhasePreInstall phase allows
// a compiler implementation (e.g. Docker) to produce cache-efficient output
// so only changes to package.json will invalidate these steps of the image
// build.
//
// PhasePostInstall
//
// Injects build.Env instructions for NODE_ENV, NODE_PATH, and PATH, setting
// the environment according to the configuration, ensuring that packages
// installed during build.PhasePreInstall are found by Node, and that any
// installed binaries are found by shells.
//
func (nc NodeConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	switch phase {
	case build.PhasePreInstall:
		if nc.Dependencies.True {
			npmInstall := build.RunAll{[]build.Run{
				{"cd", []string{LocalLibPrefix}},
				{"npm install", []string{}},
			}}

			if nc.Env == "production" {
				npmInstall.Runs[1].Arguments = []string{"--production"}
				npmInstall.Runs = append(npmInstall.Runs,
					build.Run{"npm dedupe", []string{}},
				)
			}

			return []build.Instruction{
				build.Copy{[]string{"package.json"}, LocalLibPrefix},
				npmInstall,
			}
		}
	case build.PhasePostInstall:
		if nc.Env != "" || nc.Dependencies.True {
			return []build.Instruction{build.Env{map[string]string{
				"NODE_ENV":  nc.Env,
				"NODE_PATH": path.Join(LocalLibPrefix, "node_modules"),
				"PATH":      path.Join(LocalLibPrefix, "node_modules", ".bin") + ":${PATH}",
			}}}
		}
	}

	return []build.Instruction{}
}
