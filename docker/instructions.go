package docker

import (
	"errors"
	"fmt"
	"strings"

	"gerrit.wikimedia.org/r/blubber/build"
)

// NewInstruction takes a general internal build.Instruction and returns
// a corresponding compilable Docker specific instruction. The given internal
// instruction is partially compiled at this point by calling Compile() which
// applies its own logic for escaping arguments, etc.
//
func NewInstruction(bi build.Instruction) (Instruction, error) {
	i := instruction{arguments: bi.Compile()}

	switch bi.(type) {
	case build.ScratchBase:
		i.name = "FROM"
		i.separator = " AS "
		i.arguments = []string{"scratch", i.arguments[0]}

	case build.Base:
		i.name = "FROM"
		i.separator = " AS "

	case build.Run, build.RunAll:
		i.name = "RUN"

	case build.Copy, build.CopyAs, build.CopyFrom:
		i.name = "COPY"
		i.array = true

		switch bi.(type) {
		case build.CopyAs:
			switch bi.(build.CopyAs).Instruction.(type) {
			case build.Copy:
				i.flags = []string{"chown"}
			case build.CopyFrom:
				i.flags = []string{"chown", "from"}
			}
		case build.CopyFrom:
			i.flags = []string{"from"}
		}

	case build.EntryPoint:
		i.name = "ENTRYPOINT"
		i.array = true

	case build.Env:
		i.name = "ENV"
		i.separator = " "

	case build.Label:
		i.name = "LABEL"
		i.separator = " "

	case build.User:
		i.name = "USER"

	case build.WorkingDirectory:
		i.name = "WORKDIR"

	case build.StringArg, build.UintArg:
		i.name = "ARG"
	}

	if i.name == "" {
		return nil, errors.New("Unable to create Instruction")
	}

	return i, nil
}

// Instruction defines an interface for instruction compilation.
//
type Instruction interface {
	Compile() string
}

type instruction struct {
	name      string   // name (e.g. "RUN")
	flags     []string // flags (e.g. "chown")
	arguments []string // quoted arguments
	separator string   // argument separator
	array     bool     // format arguments as array (enforces ", " separator)
}

// Compile returns a valid Dockerfile line for the instruction.
//
// Output is in the format "<name> <flags> <arguments>", e.g.
// "COPY --chown=123:223 ["foo", "bar"]" and flag values are taken from the
// beginning of the arguments slice.
//
func (ins instruction) Compile() string {
	format := ins.name + " "
	numFlags := len(ins.flags)
	args := make([]interface{}, numFlags+1)

	for i, option := range ins.flags {
		format += "--" + option + "=%s "
		args[i] = ins.arguments[i]
	}

	separator := ins.separator

	if ins.array {
		separator = ", "
		format += "[%s]"
	} else {
		format += "%s"
	}

	format += "\n"
	args[numFlags] = join(ins.arguments[numFlags:], separator)

	return fmt.Sprintf(format, args...)
}

func join(arguments []string, delimiter string) string {
	return removeNewlines(strings.Join(arguments, delimiter))
}

func removeNewlines(instructions string) string {
	out := strings.Replace(instructions, "\n", "\\n", -1)
	return out
}
