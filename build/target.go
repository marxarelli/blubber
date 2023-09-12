package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	emojiExternal = "🌐"
	emojiImage    = "📦"
	emojiLocal    = "📂"
	emojiShell    = "🖥️"
)

// Target is used during compilation to keep track of build arguments, the
// compiled LLB state, image configuration, and dependent targets.
type Target struct {
	Name         string
	Base         string
	Image        *TargetImage
	Options      *Options
	state        llb.State
	image        *oci.Image
	platform     *oci.Platform
	dependencies *TargetGroup
}

// NewTarget constructs a [Target] using the given arguments and defaults
func NewTarget(name string, base string, platform *oci.Platform, options *Options) *Target {
	state := llb.NewState(nil)

	target := &Target{
		Name:     name,
		Base:     base,
		state:    state,
		platform: platform,
		Options:  options,
	}
	target.image = newImage(target.Platform())
	target.Image = &TargetImage{target: target}
	target.dependencies = &TargetGroup{target}

	return target
}

// BuildEnv returns a full set of environment variables that should be added to
// build-time container processes.
func (target *Target) BuildEnv() map[string]string {
	targetPlatform := target.Platform()
	buildPlatform := target.BuildPlatform()

	return map[string]string{
		// Provide the same environment variables that Docker does for build and
		// target platform
		// see https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
		"BUILDPLATFORM":  platforms.Format(buildPlatform),
		"BUILDOS":        buildPlatform.OS,
		"BUILDARCH":      buildPlatform.Architecture,
		"BUILDVARIANT":   buildPlatform.Variant,
		"TARGETPLATFORM": platforms.Format(targetPlatform),
		"TARGETOS":       targetPlatform.OS,
		"TARGETARCH":     targetPlatform.Architecture,
		"TARGETVARIANT":  targetPlatform.Variant,
	}
}

// Initialize performs preprocessing steps, resolving the base image config,
// adding build-time environment variables, etc.
func (target *Target) Initialize(ctx context.Context) error {
	if target.Base != "" {
		ref, err := reference.ParseNormalizedNamed(target.Base)

		if err != nil {
			return errors.Wrapf(err, "failed to parse stage name %q", target.Base)
		}

		// Note this is based on implementation in upstream's Dockerfile2LLB
		// TODO figure out why removing a specific digest is necessary when
		// resolving an image. Perhaps it's to allow the resolver to find the right
		// platform-specific image in what could be a manifest list?
		resolveName := reference.TagNameOnly(ref).String()
		platform := target.Platform()

		digest, config, err := target.Options.MetaResolver.ResolveImageConfig(ctx, resolveName, llb.ResolveImageConfigOpt{
			Platform:     &platform,
			LogName:      target.Logf("resolving image metadata for %s", resolveName),
			ResolverType: llb.ResolverTypeRegistry,
		})

		if digest != "" {
			refWithDigest, err := reference.WithDigest(ref, digest)

			if err != nil {
				return errors.Wrap(err, "failed to get digest from ref")
			}

			target.Base = refWithDigest.String()
		}

		var img oci.Image
		if err := json.Unmarshal(config, &img); err != nil {
			return errors.Wrap(err, "failed to parse image config")
		}

		target.image = &img
		target.image.Created = nil

		target.state = llb.Image(
			target.Base,
			llb.Platform(target.Platform()),
			target.Describef("%s %s", emojiExternal, target.Base),
		)
	}

	// Set up our initial state using meta data from the image config. This
	// includes environment variables, the working directory, and the default
	// build process owner (user)
	for _, env := range target.image.Config.Env {
		k, v := parseKeyValue(env)
		target.state = target.state.AddEnv(k, v)
	}

	target.WorkingDirectory(target.image.Config.WorkingDir)

	if target.image.Config.User != "" {
		target.User(target.image.Config.User)
	}

	// Add environment variables that specify the build and target platform
	target.AddEnv(target.BuildEnv())

	// Initialize labels if not already inherited
	if target.image.Config.Labels == nil {
		target.image.Config.Labels = map[string]string{}
	}

	// Add labels from build options
	for k, v := range target.Options.Labels {
		target.image.Config.Labels[k] = v
	}

	return nil
}

// ExposeBuildArg looks for a build argument and adds an environment variable
// for it to the target build state. If a build argument is not found, the
// given default value is used.
func (target *Target) ExposeBuildArg(name string, defaultValue string) error {
	value := defaultValue

	if v, ok := target.Options.BuildArgs[name]; ok {
		value = v
	}

	target.state = target.state.AddEnv(name, value)

	return nil
}

// AddEnv adds environment variables to the target build state.
//
// Note these variables will only be defined for build time processes. To add
// environment variables to the resulting image config (for runtime processes
// when a container is later run), use [Image.AddEnv].
func (target *Target) AddEnv(definitions map[string]string) error {
	for _, k := range sortedKeys(definitions) {
		v, ok := definitions[k]
		if ok {
			target.state = target.state.AddEnv(k, target.ExpandEnv(v))
		}
	}
	return nil
}

// Describef returns an llb.ConstraintsOpt that describes a compile operation
// to the end user
func (target *Target) Describef(msg string, v ...interface{}) llb.ConstraintsOpt {
	return llb.WithCustomName(target.Logf(msg, v...))
}

// ClientBuildDir returns the llb.State for the user's build context
func (target *Target) ClientBuildDir() llb.State {
	return llb.Local(
		target.Options.ClientBuildContext,
		llb.SessionID(target.Options.SessionID),
		llb.ExcludePatterns(target.Options.Excludes),
		llb.SharedKeyHint(target.Options.ClientBuildContext),
		target.Describef("%s [client build directory]", emojiLocal),
	)
}

// CopyFromClient copies one or more sources from the client filesystem to the
// given destination on the target filesystem
func (target *Target) CopyFromClient(sources []string, destination string, options ...llb.CopyOption) error {
	return target.copy(sources, destination, "", options)
}

// CopyFrom copies one or more sources from the given dependency to the given
// destination on the target filesystem
func (target *Target) CopyFrom(from string, sources []string, destination string, options ...llb.CopyOption) error {
	if from == "" {
		return errors.Errorf("from may be not empty")
	}

	return target.copy(sources, destination, from, options)
}

func (target *Target) copy(sources []string, destination string, from string, options []llb.CopyOption) error {
	// If there is more than 1 file being copied, the destination must be a
	// directory ending with "/"
	if len(sources) > 1 && !strings.HasSuffix(destination, "/") {
		destination = destination + "/"
	}

	// Prepare a FileAction with one copy operation for each source
	var fa *llb.FileAction

	// Use the same default behavior as the Dockerfile frontend
	copyOpts := []llb.CopyOption{
		&llb.CopyInfo{
			FollowSymlinks:      true,
			CopyDirContentsOnly: true,
			AttemptUnpack:       false,
			CreateDestPath:      true,
			AllowWildcard:       true,
			AllowEmptyWildcard:  true,
		},
	}
	fileOpts := []llb.ConstraintsOpt{}

	// Default to using the client's local build context as the source
	// filesystem unless an explicit one was given
	fromState := target.ClientBuildDir()

	if from == "" {
		fileOpts = append(
			fileOpts,
			target.Describef("%s %+v -> %+v", emojiLocal, sources, destination),
		)
	} else {
		fileOpts = append(
			fileOpts,
			target.Describef("%s {%s}%+v -> %+v", emojiImage, from, sources, destination),
		)

		// We're copying from some other state filesystem, either a variant
		// dependency or an external image
		if dep, ok := target.dependencies.Find(from); ok {
			fromState = dep.state
		} else {
			fromState = llb.Image(
				from,
				llb.Platform(target.Platform()),
				target.Describef("%s %s", emojiExternal, from),
			)
		}
	}

	copyOpts = append(copyOpts, options...)

	for _, src := range sources {
		if fa == nil {
			fa = llb.Copy(fromState, src, destination, copyOpts...)
		} else {
			fa = fa.Copy(fromState, src, destination, copyOpts...)
		}
	}

	target.state = target.state.File(fa, fileOpts...)
	return nil
}

// Run executes the given command and arguments using the default shell
func (target *Target) Run(command string, args ...string) error {
	return target.RunAll(append([]string{command}, args...))
}

// RunAll executes the given set of commands and arguments as shell commands
// with logical &&'s (e.g. "/bin/sh -c 'cmd1 arg && cmd2 arg'")
//
// All arguments will be quoted. The command string may contain % formatting
// "verbs" (a la [fmt.Sprintf]) for which substitutions will be made using the
// corresponding leading arguments. The remaining arguments will be appended.
//
// ## Example
//
//	target.RunAll(
//			string{"chown -R %s:%s", "123", "321", "/dir"},
//			[]string{"chmod", "0755", "/dir"},
//	)
//
// Would append a single run operation that executes:
//
//	/bin/sh -c 'chown -R "123":"321" "/dir" && chmod "0755" "/dir"'
func (target *Target) RunAll(runs ...[]string) error {
	commands := make([]string, len(runs))

	for i, run := range runs {
		if len(run) < 1 {
			return errors.New("no run command")
		}

		cmd := run[0]
		args := run[1:]

		// 1. Count the number of % formatting tokens (n) in the command string
		// 2. Format the command string along with n of the leading arguments
		// 3. Append the remaining arguments to the command string
		numInnerArgs := strings.Count(cmd, `%`) - strings.Count(cmd, `%%`)
		command := sprintf(cmd, args[0:numInnerArgs])

		if len(args) > numInnerArgs {
			command += " " + strings.Join(quoteAll(args[numInnerArgs:]), " ")
		}

		commands[i] = command
	}

	return target.RunShell(strings.Join(commands, " && "))
}

// RunShell runs the given command using /bin/sh
func (target *Target) RunShell(command string) error {
	command = target.ExpandEnv(command)
	target.state = target.state.Run(
		llb.Args([]string{"/bin/sh", "-c", command}),
		target.Describef("%s $ %s", emojiShell, command),
	).Root()

	return nil
}

// ExpandEnv substitutes environment variable references in the given string
// for the current values taken from the current target state
func (target *Target) ExpandEnv(subject string) string {
	ctx := context.TODO()

	return os.Expand(subject, func(key string) string {
		val, ok, _ := target.state.GetEnv(ctx, key)

		if ok {
			return val
		}

		return ""
	})
}

// Logf formats logging messages for this target
func (target *Target) Logf(msg string, values ...interface{}) string {
	padding := strings.Repeat(" ", target.dependencies.MaxNameLength()-target.NameLength())

	v := append([]interface{}{}, target, padding)
	v = append(v, values...)

	return fmt.Sprintf("[%s] %s"+msg, v...)
}

// Marshal returns a solveable LLB definition and JSON image configuration for
// this target
func (target *Target) Marshal(ctx context.Context) (*llb.Definition, []byte, error) {
	def, err := target.state.Marshal(ctx)

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal LLB state to protobuf")
	}

	imageCfg, err := json.Marshal(target.image)

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal image config")
	}

	return def, imageCfg, nil
}

// WriteTo marshals the target state to protobuf and writes it to the given
// [io.Writer].
func (target *Target) WriteTo(ctx context.Context, writer io.Writer) error {
	def, _, err := target.Marshal(ctx)

	if err != nil {
		return errors.Wrap(err, "failed to marshal target")
	}

	return llb.WriteTo(def, writer)
}

// String returns the string representation of the target suitable for the
// user
func (target *Target) String() string {
	if target.Options.MultiPlatform() {
		return fmt.Sprintf("%s_{%s}", target.Name, platforms.Format(target.Platform()))
	}
	return target.Name
}

// BuildPlatform returns either the build platform from the [build.Options] or
// the default platform.
func (target *Target) BuildPlatform() oci.Platform {
	if target.Options.BuildPlatform != nil {
		return *target.Options.BuildPlatform
	}

	// this is a bit of defensive programming as target.Options.BuildPlatform
	// should never be nil in practice
	return platforms.DefaultSpec()
}

// Platform returns either the target platform given explicitly at
// construction or the first target platform in the options.
func (target *Target) Platform() oci.Platform {
	if target.platform != nil {
		return *target.platform
	} else if len(target.Options.TargetPlatforms) > 0 {
		return *target.Options.TargetPlatforms[0]
	}

	// this is a bit of defensive programming as target.Options.TargetPlatforms
	// should never be empty in practice
	return platforms.DefaultSpec()
}

// NameLength returns the number of printable runes in the target name.
func (target *Target) NameLength() int {
	return utf8.RuneCountInString(target.Name)
}

// User sets the effective build time user for the target state.
//
// Note the given user will only affect build time processes. To configure the
// user of the resulting image config (for runtime processes when a container
// is run), use [Image.User].
func (target *Target) User(user string) error {
	target.state = target.state.User(target.ExpandEnv(user))
	return nil
}

// WorkingDirectory sets the build time working directory of the target state.
//
// Note this will only affect build time processes. To configure the working
// directory of the resulting image config (for runtime processes when a
// container is run), use [Image.WorkingDirectory].
func (target *Target) WorkingDirectory(dir string) error {
	target.state = target.state.Dir(dir)
	return nil
}

// RunEntrypoint runs the target's entrypoint
//
// Note that caching is always disabled for this operation.
func (target *Target) RunEntrypoint(args []string, env map[string]string) error {
	runOpts := []llb.RunOption{
		llb.Args(append(target.image.Config.Entrypoint, args...)),
		disableCacheForOp(),
	}

	for _, k := range sortedKeys(env) {
		v, ok := env[k]
		if ok {
			runOpts = append(runOpts, llb.AddEnv(k, v))
		}
	}

	target.state = target.state.Run(runOpts...).Root()

	return nil
}

// TargetGroup provides interfaces for building multiple dependent targets.
type TargetGroup []*Target

// NewTarget creates a new [Target] and sets its dependencies to this
// [TargetGroup].
func (tg *TargetGroup) NewTarget(name string, base string, platform *oci.Platform, options *Options) *Target {
	target := NewTarget(name, base, platform, options)
	target.dependencies = tg

	*tg = append(*tg, target)

	return target
}

// Find returns the target matching the given name or nil if none by that name
// are found.
func (tg *TargetGroup) Find(name string) (*Target, bool) {
	for _, t := range *tg {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

// InitializeAll calls [Target.Initialize] on all targets in the group.
func (tg *TargetGroup) InitializeAll(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, t := range *tg {
		func(target *Target) {
			eg.Go(func() error {
				return target.Initialize(ctx)
			})
		}(t)
	}

	return eg.Wait()
}

// MaxNameLength returns the length of the longest name in the target group.
func (tg *TargetGroup) MaxNameLength() int {
	max := 0

	for _, target := range *tg {
		l := target.NameLength()

		if l > max {
			max = l
		}
	}

	return max
}

func newImage(platform oci.Platform) *oci.Image {
	image := oci.Image{
		Architecture: platform.Architecture,
		OS:           platform.OS,
	}
	image.RootFS.Type = "layers"
	image.Config.WorkingDir = "/"
	image.Config.Env = []string{"PATH=" + system.DefaultPathEnv(platform.OS)}

	return &image
}

func parseKeyValue(env string) (string, string) {
	parts := strings.SplitN(env, "=", 2)
	v := ""
	if len(parts) > 1 {
		v = parts[1]
	}

	return parts[0], v
}

type runOptionFunc func(*llb.ExecInfo)

func (fn runOptionFunc) SetRunOption(ei *llb.ExecInfo) {
	fn(ei)
}

func disableCacheForOp() llb.RunOption {
	return runOptionFunc(func(ei *llb.ExecInfo) {
		ei.Constraints.Metadata.IgnoreCache = true
	})
}
