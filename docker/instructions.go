package docker

import (
	"errors"
	"fmt"
	"strings"

	"phabricator.wikimedia.org/source/blubber/build"
)

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

type DockerRun struct{ abstractDockerInstruction }

func (dr DockerRun) Compile() string {
	return fmt.Sprintf(
		"RUN %s\n",
		join(dr.arguments, ""))
}

type DockerCopy struct{ abstractDockerInstruction }

func (dc DockerCopy) Compile() string {
	return fmt.Sprintf(
		"COPY [%s]\n",
		join(dc.arguments, ", "))
}

type DockerCopyFrom struct{ abstractDockerInstruction }

func (dcf DockerCopyFrom) Compile() string {
	return fmt.Sprintf(
		"COPY --from=%s [%s]\n",
		dcf.arguments[0],
		join(dcf.arguments[1:], ", "))
}

type DockerEnv struct{ abstractDockerInstruction }

func (de DockerEnv) Compile() string {
	return fmt.Sprintf(
		"ENV %s\n",
		join(de.arguments, " "))
}

// DockerLabel represents a concrete LABEL instruction
type DockerLabel struct{ abstractDockerInstruction }

// Compile returns multiple key="value" arguments as a single LABEL string
func (dl DockerLabel) Compile() string {
	return fmt.Sprintf(
		"LABEL %s\n",
		join(dl.arguments, " "))
}

type DockerVolume struct{ abstractDockerInstruction }

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