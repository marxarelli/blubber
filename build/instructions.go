package build

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Instruction interface {
	Compile() []string
}

type Run struct {
	Command   string
	Arguments []string
}

func (run Run) Compile() []string {
	numInnerArgs := strings.Count(run.Command, `%`) - strings.Count(run.Command, `%%`)
	command := sprintf(run.Command, run.Arguments[0:numInnerArgs])

	if len(run.Arguments) > numInnerArgs {
		command += " " + strings.Join(quoteAll(run.Arguments[numInnerArgs:]), " ")
	}

	return []string{command}
}

type RunAll struct {
	Runs []Run
}

func (runAll RunAll) Compile() []string {
	commands := make([]string, len(runAll.Runs))

	for i, run := range runAll.Runs {
		commands[i] = run.Compile()[0]
	}

	return []string{strings.Join(commands, " && ")}
}

type Copy struct {
	Sources     []string
	Destination string
}

func (copy Copy) Compile() []string {
	return append(quoteAll(copy.Sources), quote(copy.Destination))
}

type CopyFrom struct {
	From string
	Copy
}

func (cf CopyFrom) Compile() []string {
	return append([]string{cf.From}, cf.Copy.Compile()...)
}

type Env struct {
	Definitions map[string]string
}

func (env Env) Compile() []string {
	defs := make([]string, 0, len(env.Definitions))
	names := make([]string, 0, len(env.Definitions))

	for name := range env.Definitions {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		defs = append(defs, name+"="+quote(env.Definitions[name]))
	}

	return defs
}

func quote(arg string) string {
	return strconv.Quote(arg)
}

func quoteAll(arguments []string) []string {
	quoted := make([]string, len(arguments))

	for i, arg := range arguments {
		quoted[i] = quote(arg)
	}

	return quoted
}

func sprintf(format string, arguments []string) string {
	args := make([]interface{}, len(arguments))

	for i, v := range arguments {
		args[i] = quote(v)
	}

	return fmt.Sprintf(format, args...)
}
