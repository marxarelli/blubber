// Package build defines types and interfaces that could potentially be
// compiled to various external build-tool scripts but share a general
// internal abstraction and rules for escaping.
//
package build

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Instruction defines a common interface that all concrete build types must
// implement.
//
type Instruction interface {
	Compile() []string
}

// Run is a concrete build instruction for passing any number of arguments to
// a shell command.
//
// The command string may contain inner argument placeholders using the "%s"
// format verb and will be appended with the quoted values of any arguments
// that remain after interpolation of the command string.
//
type Run struct {
	Command   string   // command string (e.g. "useradd -d %s -u %s")
	Arguments []string // command arguments both inner and final (e.g. ["/home/user", "123", "user"])
}

// Compile quotes all arguments, interpolates the command string with inner
// arguments, and appends the final arguments.
//
func (run Run) Compile() []string {
	numInnerArgs := strings.Count(run.Command, `%`) - strings.Count(run.Command, `%%`)
	command := sprintf(run.Command, run.Arguments[0:numInnerArgs])

	if len(run.Arguments) > numInnerArgs {
		command += " " + strings.Join(quoteAll(run.Arguments[numInnerArgs:]), " ")
	}

	return []string{command}
}

// RunAll is a concrete build instruction for declaring multiple Run
// instructions that will be executed together in a `cmd1 && cmd2` chain.
//
type RunAll struct {
	Runs []Run // multiple Run instructions to be executed together
}

// Compile concatenates all individually compiled Run instructions into a
// single command.
//
func (runAll RunAll) Compile() []string {
	commands := make([]string, len(runAll.Runs))

	for i, run := range runAll.Runs {
		commands[i] = run.Compile()[0]
	}

	return []string{strings.Join(commands, " && ")}
}

// Copy is a concrete build instruction for copying source files/directories
// from the build host into the image.
//
type Copy struct {
	Sources     []string // source file/directory paths
	Destination string   // destination path
}

// Compile quotes the defined source files/directories and destination.
//
func (copy Copy) Compile() []string {
	dest := copy.Destination

	// If there is more than 1 file being copied, the destination must be a
	// directory ending with "/"
	if len(copy.Sources) > 1 && !strings.HasSuffix(copy.Destination, "/") {
		dest = dest + "/"
	}

	return append(quoteAll(copy.Sources), quote(dest))
}

// CopyAs is a concrete build instruction for copying source
// files/directories and setting their ownership to the given UID/GID.
//
// While it can technically wrap any build.Instruction, it is meant to be used
// with build.Copy and build.CopyFrom to enforce file/directory ownership.
//
type CopyAs struct {
	UID uint // owner UID
	GID uint // owner GID
	Instruction
}

// Compile returns the variant name unquoted and all quoted CopyAs instruction
// fields.
//
func (ca CopyAs) Compile() []string {
	return append([]string{fmt.Sprintf("%d:%d", ca.UID, ca.GID)}, ca.Instruction.Compile()...)
}

// CopyFrom is a concrete build instruction for copying source
// files/directories from one variant image to another.
//
type CopyFrom struct {
	From string // source variant name
	Copy
}

// Compile returns the variant name unquoted and all quoted Copy instruction
// fields.
//
func (cf CopyFrom) Compile() []string {
	return append([]string{cf.From}, cf.Copy.Compile()...)
}

// EntryPoint is a build instruction for declaring a container's default
// runtime process.
type EntryPoint struct {
	Command []string // command and arguments
}

// Compile returns the quoted entrypoint command and arguments.
//
func (ep EntryPoint) Compile() []string {
	return quoteAll(ep.Command)
}

// Env is a concrete build instruction for declaring a container's runtime
// environment variables.
//
type Env struct {
	Definitions map[string]string // number of key/value pairs
}

// Compile returns the key/value pairs as a number of `key="value"` strings
// where the values are properly quoted and the slice is ordered by the keys.
//
func (env Env) Compile() []string {
	return compileSortedKeyValues(env.Definitions)
}

// Label is a concrete build instruction for declaring a number of meta-data
// key/value pairs to be included in the image.
//
type Label struct {
	Definitions map[string]string // number of meta-data key/value pairs
}

// Compile returns the key/value pairs as a number of `key="value"` strings
// where the values are properly quoted and the slice is ordered by the keys.
//
func (label Label) Compile() []string {
	return compileSortedKeyValues(label.Definitions)
}

// User is a build instruction for setting which user will run future
// commands.
//
type User struct {
	Name string // user name
}

// Compile returns the quoted user name.
//
func (user User) Compile() []string {
	return []string{quote(user.Name)}
}

// WorkingDirectory is a build instruction for defining the working directory
// for future command and entrypoint instructions.
//
type WorkingDirectory struct {
	Path string // working directory path
}

// Compile returns the quoted working directory path.
//
func (wd WorkingDirectory) Compile() []string {
	return []string{quote(wd.Path)}
}

func compileSortedKeyValues(keyValues map[string]string) []string {
	defs := make([]string, 0, len(keyValues))
	names := make([]string, 0, len(keyValues))

	for name := range keyValues {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		defs = append(defs, name+"="+quote(keyValues[name]))
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
