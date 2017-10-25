package docker

import (
	"errors"
	"fmt"
	"strings"

	"phabricator.wikimedia.org/source/blubber/build"
)

// NewDockerInstruction takes a general internal build.Instruction and returns
// a corresponding compilable Docker specific instruction. The given internal
// instruction is partially compiled at this point by calling Compile() which
// applies its own logic for escaping arguments, etc.
//
func NewDockerInstruction(instruction build.Instruction) (DockerInstruction, error) {
	switch instruction.(type) {
	case build.Run, build.RunAll:
		var dockerInstruction DockerRun
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Copy:
		var dockerInstruction DockerCopy
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.CopyFrom:
		var dockerInstruction DockerCopyFrom
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Env:
		var dockerInstruction DockerEnv
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Label:
		var dockerInstruction DockerLabel
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Volume:
		var dockerInstruction DockerVolume
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	}

	return nil, errors.New("Unable to create DockerInstruction")
}

// DockerInstruction defines an interface for instruction compilation.
//
type DockerInstruction interface {
	Compile() string
	Arguments() []string
}

type abstractDockerInstruction struct {
	arguments []string
}

func (di abstractDockerInstruction) Arguments() []string {
	return di.arguments
}

// DockerRun compiles into a RUN instruction.
//
type DockerRun struct{ abstractDockerInstruction }

// Compile compiles RUN instructions.
//
func (dr DockerRun) Compile() string {
	return fmt.Sprintf(
		"RUN %s\n",
		join(dr.arguments, ""))
}

// DockerCopy compiles into a COPY instruction.
//
type DockerCopy struct{ abstractDockerInstruction }

// Compile compiles COPY instructions.
//
func (dc DockerCopy) Compile() string {
	return fmt.Sprintf(
		"COPY [%s]\n",
		join(dc.arguments, ", "))
}

// DockerCopyFrom compiles into a COPY --from instruction.
//
type DockerCopyFrom struct{ abstractDockerInstruction }

// Compile compiles COPY --from instructions.
//
func (dcf DockerCopyFrom) Compile() string {
	return fmt.Sprintf(
		"COPY --from=%s [%s]\n",
		dcf.arguments[0],
		join(dcf.arguments[1:], ", "))
}

// DockerEnv compiles into a ENV instruction.
//
type DockerEnv struct{ abstractDockerInstruction }

// Compile compiles ENV instructions.
//
func (de DockerEnv) Compile() string {
	return fmt.Sprintf(
		"ENV %s\n",
		join(de.arguments, " "))
}

// DockerLabel compiles into a LABEL instruction.
//
type DockerLabel struct{ abstractDockerInstruction }

// Compile returns multiple key="value" arguments as a single LABEL
// instruction.
//
func (dl DockerLabel) Compile() string {
	return fmt.Sprintf(
		"LABEL %s\n",
		join(dl.arguments, " "))
}

// DockerVolume compiles into a VOLUME instruction.
//
type DockerVolume struct{ abstractDockerInstruction }

// Compile compiles VOLUME instructions.
//
func (dv DockerVolume) Compile() string {
	return fmt.Sprintf(
		"VOLUME [%s]\n",
		join(dv.arguments, ", "))
}

func join(arguments []string, delimiter string) string {
	return removeNewlines(strings.Join(arguments, delimiter))
}

func removeNewlines(instructions string) string {
	out := strings.Replace(instructions, "\n", "\\n", -1)
	return out
}
