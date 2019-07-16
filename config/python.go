package config

import (
	"gerrit.wikimedia.org/r/blubber/build"
)

// PythonLibPrefix is the path to installed dependency wheels.
//
const PythonLibPrefix = LocalLibPrefix + "/python"

// PythonSitePackages is the path to installed Python packages.
//
const PythonSitePackages = PythonLibPrefix + "/site-packages"

// PythonSiteBin is the path to installed Python packages bin files.
//
const PythonSiteBin = PythonSitePackages + "/bin"

// PythonConfig holds configuration fields related to pre-installation of project
// dependencies via PIP.
//
type PythonConfig struct {
	Version       string   `json:"version"`         // Python binary to use when installing dependencies
	Requirements  []string `json:"requirements"`    // install requirements from given files
	UseSystemFlag bool     `json:"use-system-flag"` // Inject the --system flag into the install command (T227919)
}

// Merge takes another PythonConfig and merges its fields into this one's,
// overwriting both the dependencies flag and requirements.
//
func (pc *PythonConfig) Merge(pc2 PythonConfig) {
	if pc2.Version != "" {
		pc.Version = pc2.Version
	}

	if pc2.Requirements != nil {
		pc.Requirements = pc2.Requirements
	}

	if pc2.UseSystemFlag {
		pc.UseSystemFlag = true
	}
}

// InstructionsForPhase injects instructions into the build related to Python
// dependency installation.
//
// PhasePrivileged
//
// Ensures that the newest versions of setuptools, wheel, tox, and pip are
// installed.
//
// PhasePreInstall
//
// Sets up Python wheels under the shared library directory (/opt/lib/python)
// for dependencies found in the declared requirements files. Installing
// dependencies during the build.PhasePreInstall phase allows a compiler
// implementation (e.g. Docker) to produce cache-efficient output so only
// changes to the given requirements files will invalidate these steps of the
// image build.
//
// Injects build.Env instructions for PIP_WHEEL_DIR and PIP_FIND_LINKS that
// will cause future executions of `pip install` (and by extension, `tox`) to
// consider packages from the shared library directory first.
//
// PhasePostInstall
//
// Injects a build.Env instruction for PIP_NO_INDEX that will cause future
// executions of `pip install` and `tox` to consider _only_ packages from the
// shared library directory, helping to speed up image builds by reducing
// network requests from said commands.
//
func (pc PythonConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	if pc.Version != "" {
		switch phase {
		case build.PhasePrivileged:
			if pc.Requirements != nil {
				return []build.Instruction{build.RunAll{[]build.Run{
					{pc.version(), []string{"-m", "easy_install", "pip"}},
					{pc.version(), []string{"-m", "pip", "install", "-U", "setuptools", "wheel", "tox"}},
				}}}
			}

		case build.PhasePreInstall:
			if pc.Requirements != nil {
				ins := []build.Instruction{
					build.Env{map[string]string{
						"PIP_WHEEL_DIR":  PythonLibPrefix,
						"PIP_FIND_LINKS": "file://" + PythonLibPrefix,
					}},
					build.CreateDirectory(PythonLibPrefix),
				}

				ins = append(ins, build.SyncFiles(pc.Requirements, ".")...)

				if args := pc.RequirementsArgs(); len(args) > 0 {
					installCmd := append([]string{"-m", "pip", "install", "--target"}, PythonSitePackages)
					if pc.UseSystemFlag {
						installCmd = InsertElement(installCmd, "--system", PosOf(installCmd, "install") + 1)
					}
					ins = append(ins, build.RunAll{[]build.Run{
						{pc.version(), append([]string{"-m", "pip", "wheel"}, args...)},
						{pc.version(), append(installCmd, args...)},
					}})
				}

				return ins
			}

		case build.PhasePostInstall:
			env := build.Env{map[string]string{
				"PYTHONPATH": PythonSitePackages,
				"PATH":       PythonSiteBin + ":${PATH}",
			}}

			if pc.Requirements != nil {
				env.Definitions["PIP_NO_INDEX"] = "1"
			}

			return []build.Instruction{env}
		}
	}

	return []build.Instruction{}
}

// RequirementsArgs returns the configured requirements as pip `-r` arguments.
//
func (pc PythonConfig) RequirementsArgs() []string {
	args := make([]string, len(pc.Requirements)*2)

	for i, req := range pc.Requirements {
		args[i*2] = "-r"
		args[(i*2)+1] = req
	}

	return args
}

func (pc PythonConfig) version() string {
	if pc.Version == "" {
		return "python"
	}

	return pc.Version
}

// InsertElement - insert el into slice at pos
func InsertElement(slice []string, el string, pos int) []string {
	slice = append(slice, "")
	copy(slice[pos+1:], slice[pos:])
	slice[pos] = el
	return slice
}

// PosOf - find position of an element in a slice
func PosOf(slice []string, el string) int {
	for p, v := range slice {
		if v == el {
			return p
		}
	}
	return -1
}
