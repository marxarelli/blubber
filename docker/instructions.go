package docker

import (
	"errors"
	"fmt"
	"strings"

	"phabricator.wikimedia.org/source/blubber/build"
)

// NewInstruction takes a general internal build.Instruction and returns
// a corresponding compilable Docker specific instruction. The given internal
// instruction is partially compiled at this point by calling Compile() which
// applies its own logic for escaping arguments, etc.
//
func NewInstruction(instruction build.Instruction) (Instruction, error) {
	switch instruction.(type) {
	case build.Run, build.RunAll:
		var dockerInstruction Run
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Copy:
		var dockerInstruction Copy
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.CopyAs:
		var dockerInstruction CopyAs
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.CopyFrom:
		var dockerInstruction CopyFrom
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Env:
		var dockerInstruction Env
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Label:
		var dockerInstruction Label
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.User:
		var dockerInstruction User
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	case build.Volume:
		var dockerInstruction Volume
		dockerInstruction.arguments = instruction.Compile()
		return dockerInstruction, nil
	}

	return nil, errors.New("Unable to create Instruction")
}

// Instruction defines an interface for instruction compilation.
//
type Instruction interface {
	Compile() string
	Arguments() []string
}

type abstractInstruction struct {
	arguments []string
}

func (di abstractInstruction) Arguments() []string {
	return di.arguments
}

// Run compiles into a RUN instruction.
//
type Run struct{ abstractInstruction }

// Compile compiles RUN instructions.
//
func (dr Run) Compile() string {
	return fmt.Sprintf(
		"RUN %s\n",
		join(dr.arguments, ""))
}

// Copy compiles into a COPY instruction.
//
type Copy struct{ abstractInstruction }

// Compile compiles COPY instructions.
//
func (dc Copy) Compile() string {
	return fmt.Sprintf(
		"COPY [%s]\n",
		join(dc.arguments, ", "))
}

// CopyAs compiles into a COPY --chown instruction.
//
type CopyAs struct{ abstractInstruction }

// Compile compiles COPY --chown instructions.
//
func (dca CopyAs) Compile() string {
	return fmt.Sprintf(
		"COPY --chown=%s [%s]\n",
		dca.arguments[0],
		join(dca.arguments[1:], ", "))
}

// CopyFrom compiles into a COPY --from instruction.
//
type CopyFrom struct{ abstractInstruction }

// Compile compiles COPY --from instructions.
//
func (dcf CopyFrom) Compile() string {
	return fmt.Sprintf(
		"COPY --from=%s [%s]\n",
		dcf.arguments[0],
		join(dcf.arguments[1:], ", "))
}

// Env compiles into a ENV instruction.
//
type Env struct{ abstractInstruction }

// Compile compiles ENV instructions.
//
func (de Env) Compile() string {
	return fmt.Sprintf(
		"ENV %s\n",
		join(de.arguments, " "))
}

// Label compiles into a LABEL instruction.
//
type Label struct{ abstractInstruction }

// Compile returns multiple key="value" arguments as a single LABEL
// instruction.
//
func (dl Label) Compile() string {
	return fmt.Sprintf(
		"LABEL %s\n",
		join(dl.arguments, " "))
}

// User compiles into a USER instruction.
//
type User struct{ abstractInstruction }

// Compile compiles USER instructions.
//
func (du User) Compile() string {
	return fmt.Sprintf(
		"USER %s\n",
		join(du.arguments, ", "))
}

// Volume compiles into a VOLUME instruction.
//
type Volume struct{ abstractInstruction }

// Compile compiles VOLUME instructions.
//
func (dv Volume) Compile() string {
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
