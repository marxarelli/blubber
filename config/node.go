package config

import (
	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// NodeConfig holds configuration fields related to the Node environment and
// whether/how to install NPM packages.
//
type NodeConfig struct {
	// Install requirements from given files
	Requirements RequirementsConfig `json:"requirements" validate:"omitempty,unique,dive"`

	// Environment name ("production" install)
	Env string `json:"env" validate:"omitempty,nodeenv"`

	// Whether to run npm ci
	UseNpmCi Flag `json:"use-npm-ci"`
}

// Merge takes another NodeConfig and merges its fields into this one's,
// overwriting useNpmCi, the environment, and the requirements files.
//
func (nc *NodeConfig) Merge(nc2 NodeConfig) {
	nc.UseNpmCi.Merge(nc2.UseNpmCi)

	if nc2.Requirements != nil {
		nc.Requirements = nc2.Requirements
	}

	if nc2.Env != "" {
		nc.Env = nc2.Env
	}
}

// InstructionsForPhase injects instructions into the build related to Node
// dependency installation and setting of the NODE_ENV.
//
// PhasePreInstall
//
// Installs Node package dependencies declared in requirements files into the
// application directory. Only production related packages are install if
// NodeConfig.Env is set to "production" in which case `npm dedupe` is also
// run. Installing dependencies during the build.PhasePreInstall phase allows
// a compiler implementation (e.g. Docker) to produce cache-efficient output
// so only changes to package.json will invalidate these steps of the image
// build.
//
// PhasePostInstall
//
// Injects build.Env instructions for NODE_ENV, setting the environment
// according to the configuration.
//
func (nc NodeConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := nc.Requirements.InstructionsForPhase(phase)

	switch phase {
	case build.PhasePreInstall:
		if len(nc.Requirements) > 0 {
			var npmInstall build.RunAll
			if nc.UseNpmCi.True {
				npmInstall = build.RunAll{[]build.Run{
					{"npm ci", []string{}},
				}}
			} else {
				npmInstall = build.RunAll{[]build.Run{
					{"npm install", []string{}},
				}}
			}

			if nc.Env == "production" {
				npmInstall.Runs[0].Arguments = []string{"--only=production"}
				npmInstall.Runs = append(npmInstall.Runs,
					build.Run{"npm dedupe", []string{}},
				)
			}

			ins = append(ins, npmInstall)
		}
	case build.PhasePostInstall:
		if nc.Env != "" || len(nc.Requirements) > 0 {
			ins = append(ins, build.Env{map[string]string{
				"NODE_ENV": nc.Env,
			}})
		}
	}

	return ins
}
