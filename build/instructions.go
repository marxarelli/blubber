// Package build defines types and interfaces that could potentially be
// compiled to various external build-tool scripts but share a general
// internal abstraction and rules for escaping.
package build

import (
	"fmt"

	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

// Instruction defines a common interface that all concrete build types must
// implement.
type Instruction interface {
	Compile(*Target) error
}

// Base is a concrete build instruction for declaring the base container image
// to start with.
type Base struct {
	Image string // image identifier
	Stage string // optional internal name used for multi-stage builds
}

// Compile to the given [Target]
func (base Base) Compile(target *Target) error {
	// noop. Handled in Target.Initialize
	return nil
}

// ScratchBase is a concrete build instruction for declaring no base image.
type ScratchBase struct {
	Stage string // optional internal name used for multi-stage builds
}

// Compile to the given [Target]
func (sb ScratchBase) Compile(target *Target) error {
	// noop. Handled in Target.Initialize
	return nil
}

// Run is a concrete build instruction for passing any number of arguments to
// a shell command.
//
// The command string may contain inner argument placeholders using the "%s"
// format verb and will be appended with the quoted values of any arguments
// that remain after interpolation of the command string.
type Run struct {
	Command   string   // command string (e.g. "useradd -d %s -u %s")
	Arguments []string // command arguments both inner and final (e.g. ["/home/user", "123", "user"])
}

// Compile to the given [Target]
func (run Run) Compile(target *Target) error {
	return target.Run(run.Command, run.Arguments...)
}

// RunAll is a concrete build instruction for declaring multiple Run
// instructions that will be executed together in a `cmd1 && cmd2` chain.
type RunAll struct {
	Runs []Run // multiple Run instructions to be executed together
}

// Compile to the given [Target]
func (ra RunAll) Compile(target *Target) error {
	runs := make([][]string, len(ra.Runs))

	for i, run := range ra.Runs {
		runs[i] = []string{run.Command}
		runs[i] = append(runs[i], run.Arguments...)
	}

	return target.RunAll(runs...)
}

// Copy is a concrete build instruction for copying source files/directories
// from the build host into the image.
type Copy struct {
	Sources     []string // source file/directory paths
	Destination string   // destination path
}

// Compile to the given [Target]
func (copy Copy) Compile(target *Target) error {
	return target.CopyFromClient(copy.Sources, copy.Destination)
}

// CopyAs is a concrete build instruction for copying source
// files/directories and setting their ownership to the given UID/GID.
//
// While it can technically wrap any build.Instruction, it is meant to be used
// with build.Copy and build.CopyFrom to enforce file/directory ownership.
type CopyAs struct {
	UID string // owner UID
	GID string // owner GID
	Instruction
}

// Compile to the given [Target]
func (ca CopyAs) Compile(target *Target) error {
	from := ""
	sources := []string{}
	destination := ""

	switch ins := ca.Instruction.(type) {
	case Copy:
		sources = ins.Sources
		destination = ins.Destination
	case CopyFrom:
		from = ins.From
		sources = ins.Copy.Sources
		destination = ins.Copy.Destination
	default:
		return errors.New("a CopyAs may only wrap Copy and CopyFrom")
	}

	opts := []llb.CopyOption{
		llb.WithUser(
			target.ExpandEnv(fmt.Sprintf("%s:%s", ca.UID, ca.GID)),
		),
	}

	if from == "" {
		return target.CopyFromClient(sources, destination, opts...)
	}

	return target.CopyFrom(from, sources, destination, opts...)
}

// CopyFrom is a concrete build instruction for copying source
// files/directories from one variant image to another.
type CopyFrom struct {
	From string // source variant name
	Copy
}

// Compile to the given [Target]
func (cf CopyFrom) Compile(target *Target) error {
	return target.CopyFrom(cf.From, cf.Copy.Sources, cf.Copy.Destination)
}

// EntryPoint is a build instruction for declaring a container's default
// runtime process.
type EntryPoint struct {
	Command []string // command and arguments
}

// Compile to the given [Target]
func (ep EntryPoint) Compile(target *Target) error {
	target.Image.Entrypoint(ep.Command)

	return nil
}

// Env is a concrete build instruction for declaring a container's runtime
// environment variables.
type Env struct {
	Definitions map[string]string // number of key/value pairs
}

// Compile to the given [Target]
func (env Env) Compile(target *Target) error {
	target.Image.AddEnv(env.Definitions)

	return target.AddEnv(env.Definitions)
}

// Label is a concrete build instruction for declaring a number of meta-data
// key/value pairs to be included in the image.
type Label struct {
	Definitions map[string]string // number of meta-data key/value pairs
}

// Compile to the given [Target]
func (label Label) Compile(target *Target) error {
	target.Image.AddLabels(label.Definitions)

	return nil
}

// User is a build instruction for setting which user will run future
// commands.
type User struct {
	UID string // user ID
}

// Compile to the given [Target]
func (user User) Compile(target *Target) error {
	uid := user.UID

	if uid == "" {
		// Preserve legacy behavior of an uninitialized User being == root
		uid = "0"
	}

	target.Image.User(uid)

	return target.User(uid)
}

// WorkingDirectory is a build instruction for defining the working directory
// for future command and entrypoint instructions.
type WorkingDirectory struct {
	Path string // working directory path
}

// Compile to the given [Target]
func (wd WorkingDirectory) Compile(target *Target) error {
	target.Image.WorkingDirectory(wd.Path)

	return target.WorkingDirectory(wd.Path)
}

// StringArg is a build instruction defining a build-time replaceable argument
// with a string value.
type StringArg struct {
	Name    string // argument name
	Default string // argument default value
}

// Compile to the given [Target]
func (arg StringArg) Compile(target *Target) error {
	return target.ExposeBuildArg(arg.Name, arg.Default)
}

// UintArg is a build instruction defining a build-time replaceable argument
// with an integer value.
type UintArg struct {
	Name    string // argument name
	Default uint   // argument default value
}

// Compile to the given [Target]
func (arg UintArg) Compile(target *Target) error {
	return target.ExposeBuildArg(arg.Name, fmt.Sprintf("%d", arg.Default))
}
