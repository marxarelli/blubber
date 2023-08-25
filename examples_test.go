package main

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	ociarchive "github.com/containers/image/v5/oci/archive"
	memoryblobcache "github.com/containers/image/v5/pkg/blobinfocache/memory"
	imagetypes "github.com/containers/image/v5/types"
	"github.com/cucumber/godog"
	"github.com/google/shlex"
	bkclient "github.com/moby/buildkit/client"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"gitlab.wikimedia.org/repos/releng/blubber/util/imagefs"
)

type ctxKey uint8

const (
	wdKey ctxKey = iota
	clientKey
	imageCfgKey
	imageTarfileKey
	imageFsKey
)

func defineSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^"([\w-\./]+)" as a working directory`, createWorkingDirectory)
	ctx.Step(`^this "([\w-\.]+)"(?: (file|executable))?`, createFile)
	ctx.Step(`^you build (and run )?the "([\w-\.]+)" variant`, buildVariant)
	ctx.Step(`^the image will (not )?have the following files in "([^"]*)"$`, theImageHasTheFollowingFilesIn)
	ctx.Step(`^the image will (not )?have the following files in the default working directory$`, theImageHasTheFollowingFilesInDefaultWorkingDir)
	ctx.Step(`^the image will have the (user|group) "([^"]*)" with (?:UID|GID) (\d+)$`, theImageHasTheEntity)
	ctx.Step(`^the image runtime user will be "([^"]*)"$`, theImageRuntimeUserIs)
	ctx.Step(`^the image entrypoint will be "([^"]*)"$`, theImageEntrypointIs)
	ctx.Step(`^the image will include environment variables$`, theImageEnvironmentContains)
	ctx.Step(`^the image will contain a file "([^"]*)" that looks like$`, theImageContainsFileWithContent)
	ctx.Step(`^the entrypoint will have run successfully$`, noop)
}

func TestExamples(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"examples"},
			TestingT: t,
			Tags:     os.Getenv("BLUBBER_ONLY_EXAMPLES"),
		},
	}

	if os.Getenv("BLUBBER_RUN_EXAMPLES") == "" {
		t.Skip("Skipping acceptance tests (set BLUBBER_RUN_EXAMPLES=1 to run them)")
	} else {
		suite.Run()
	}
}

func initializeScenario(ctx *godog.ScenarioContext) {
	defineSteps(ctx)

	// Clean up any working directory we've created during the scenario
	if os.Getenv("BLUBBER_DEBUG_EXAMPLES") == "" {
		ctx.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			if wd, ok := ctx.Value(wdKey).(*workingDirectory); ok {
				wd.Remove()
			}

			if imageTar, ok := ctx.Value(imageTarfileKey).(*os.File); ok {
				os.Remove(imageTar.Name())
			}

			if client, ok := ctx.Value(clientKey).(*bkclient.Client); ok {
				client.Close()
			}

			return ctx, err
		})
	}
}

func noop(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func createWorkingDirectory(ctx context.Context, srcDir string) (context.Context, error) {
	wd, err := newWorkingDirectory()

	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, wdKey, wd)
	err = wd.CopyFrom(srcDir)

	if err != nil {
		return ctx, errors.Wrapf(err, "failed to create a new working directory from %s", srcDir)
	}

	return ctx, nil
}

func createFile(ctx context.Context, file, fileType string, content []byte) (context.Context, error) {
	return withCtxValue[*workingDirectory](ctx, wdKey, func(wd *workingDirectory) (context.Context, error) {
		var mode os.FileMode
		switch fileType {
		case "executable":
			mode = os.FileMode(0755)
		default:
			mode = os.FileMode(0644)
		}

		return ctx, wd.WriteFile(file, append(content, byte('\n')), mode)
	})
}

func buildVariant(ctx context.Context, andRun string, variant string) (context.Context, error) {
	runVariant := andRun != ""

	return withCtxValue[*workingDirectory](ctx, wdKey, func(wd *workingDirectory) (context.Context, error) {
		var err error

		blubberImage := os.Getenv("BLUBBER_TEST_IMAGE")

		if blubberImage == "" {
			return ctx, errors.New("you must set BLUBBER_TEST_IMAGE with the blubber frontend ref to run these tests")
		}

		// Attempt to retrieve an existing client first
		client, ok := ctx.Value(clientKey).(*bkclient.Client)

		if !ok {
			client, err = bkclient.New(ctx, os.Getenv("BUILDKIT_HOST"))

			if err != nil {
				return ctx, err
			}

			ctx = context.WithValue(ctx, clientKey, client)
		}

		tmptar, err := os.CreateTemp("", "blubber.oci.*.tar")

		if err != nil {
			return ctx, err
		}

		ctx = context.WithValue(ctx, imageTarfileKey, tmptar)

		solveOpt := bkclient.SolveOpt{
			Frontend: "gateway.v0",
			FrontendAttrs: map[string]string{
				"source":   blubberImage,
				"filename": "blubber.yaml",
				"variant":  variant,
				"no-cache": "",
				"platform": "linux/amd64",
			},
			LocalDirs: map[string]string{
				"context":    wd.Path,
				"dockerfile": wd.Path,
			},
			Exports: []bkclient.ExportEntry{
				{
					Type: bkclient.ExporterOCI,
					Output: func(_ map[string]string) (io.WriteCloser, error) {
						return tmptar, nil
					},
				},
			},
		}

		if runVariant {
			solveOpt.FrontendAttrs["run-variant"] = "true"
		}

		_, err = client.Build(ctx, solveOpt, "buildctl", func(ctx context.Context, c gateway.Client) (*gateway.Result, error) {
			return c.Solve(ctx, gateway.SolveRequest{
				Frontend:    solveOpt.Frontend,
				FrontendOpt: solveOpt.FrontendAttrs,
			})
		}, nil)

		if err != nil {
			return ctx, errors.Wrapf(err, "failed to build variant %s", variant)
		}

		tmptar.Close()

		// Save the image filesystem for future assertions
		ref, err := ociarchive.ParseReference(tmptar.Name())

		if err != nil {
			return ctx, errors.Wrapf(err, "failed to get image reference for OCI tarball %s", tmptar.Name())
		}

		sys := &imagetypes.SystemContext{
			OSChoice:           "linux",
			ArchitectureChoice: "amd64",
		}

		cache := memoryblobcache.New()

		img, err := ref.NewImage(ctx, sys)

		if err != nil {
			return ctx, errors.Wrapf(err, "failed to get image from ref %s", ref.StringWithinTransport())
		}

		cfg, err := img.OCIConfig(ctx)

		if err != nil {
			return ctx, errors.Wrap(err, "failed to get image config")
		}

		return context.WithValue(context.WithValue(ctx, imageCfgKey, cfg), imageFsKey, imagefs.New(ctx, ref, sys, cache)), err
	})
}

// theImageHasTheFollowingFilesIn can be used with any of the following
// table columns:
//
//	| owner | group | name               |
//	| 123   | 123   | some-dir/some-file |
//	| 123   | 123   | some-other-file    |
//
// Or a very simple listing:
//
//	| some-dir/some-file |
//	| some-other-file    |
func theImageHasTheFollowingFilesIn(ctx context.Context, not string, dir string, files *godog.Table) (context.Context, error) {
	negate := false
	if not != "" {
		negate = true
	}

	return withImageFS(ctx, func(image fs.FS) (context.Context, error) {
		headers := map[int]string{}

		for i, row := range files.Rows {
			if len(row.Cells) == 1 {
				path := filepath.Join(dir, row.Cells[0].Value)
				file, err := image.Open(path)

				if err == nil {
					defer file.Close()

					if negate {
						return ctx, errors.Errorf("file %s exists in the image and it should not", path)
					}
				} else if !negate {
					return ctx, errors.Wrapf(err, "file %s does not exist in the image", path)
				}
			} else if i == 0 {
				for j, cell := range row.Cells {
					headers[j] = cell.Value
				}
			} else {
				fields := make(map[string]string, len(row.Cells))

				for j, cell := range row.Cells {
					fields[headers[j]] = cell.Value
				}

				name, ok := fields["name"]

				if !ok {
					return ctx, errors.New("a file table must have a name column")
				}

				delete(fields, "name")

				path := filepath.Join(dir, name)
				file, err := image.Open(path)

				if err != nil {
					return ctx, errors.Wrapf(err, "file %s does not exist in the image", path)
				}

				defer file.Close()
				info, err := file.Stat()

				if err != nil {
					return ctx, errors.Wrapf(err, "failed to stat %s", path)
				}

				header, ok := info.Sys().(*tar.Header)
				if !ok {
					return ctx, errors.Errorf("failed to get tar header for %s", path)
				}

				for field, value := range fields {
					switch field {
					case "owner", "group":
						expected, _ := strconv.Atoi(value)
						var actual int

						if field == "owner" {
							actual = header.Uid
						} else {
							actual = header.Gid
						}

						if actual != expected {
							return ctx, errors.Errorf("expected file %s to have %s %d, but it has %d", path, field, expected, actual)
						}
					default:
						return ctx, errors.Errorf("unknown table field %s", field)
					}
				}
			}
		}

		return ctx, nil
	})
}

func theImageHasTheFollowingFilesInDefaultWorkingDir(ctx context.Context, not string, files *godog.Table) (context.Context, error) {
	return theImageHasTheFollowingFilesIn(ctx, not, "/srv/app", files)
}

func theImageContainsFileWithContent(ctx context.Context, path, content string) (context.Context, error) {
	return withImageFileData(ctx, path, func(data []byte) (context.Context, error) {
		dataStr := string(data)
		if !(content == dataStr || (content+"\n") == dataStr) {
			return ctx, errors.Errorf("content of %s doesn't match", path)
		}

		return ctx, nil
	})
}

func theImageHasTheEntity(ctx context.Context, ent, name, id string) (context.Context, error) {
	source := "/etc/passwd"
	if ent == "group" {
		source = "/etc/group"
	}

	return withImageFile(ctx, source, func(reader io.Reader) (context.Context, error) {
		found := false

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			record := strings.Split(scanner.Text(), ":")

			if len(record) >= 3 {
				if record[0] == name && record[2] == id {
					found = true
					break
				}
			}
		}

		err := scanner.Err()
		if err != nil {
			return ctx, err
		}

		if !found {
			return ctx, errors.Errorf("%s %s with id %s not found in %s", ent, name, id, source)
		}

		return ctx, nil
	})
}

func theImageRuntimeUserIs(ctx context.Context, user string) (context.Context, error) {
	return withCtxValue[*ociv1.Image](ctx, imageCfgKey, func(image *ociv1.Image) (context.Context, error) {
		if image.Config.User != user {
			return ctx, errors.Errorf("expected image user to be %s but got %s", user, image.Config.User)
		}

		return ctx, nil
	})
}

func theImageEntrypointIs(ctx context.Context, entrypoint string) (context.Context, error) {
	return withCtxValue[*ociv1.Image](ctx, imageCfgKey, func(image *ociv1.Image) (context.Context, error) {
		if !(len(image.Config.Entrypoint) == 1 && image.Config.Entrypoint[0] == entrypoint) {
			return ctx, errors.Errorf("expected entrypoint to be [%s] but got %s", entrypoint, image.Config.Entrypoint)
		}

		return ctx, nil
	})
}

func theImageEnvironmentContains(ctx context.Context, envTable *godog.Table) (context.Context, error) {
	envs := make([]string, len(envTable.Rows))

	for i, row := range envTable.Rows {
		for _, cell := range row.Cells {
			envs[i] = cell.Value
		}
	}

	return withCtxValue[*ociv1.Image](ctx, imageCfgKey, func(image *ociv1.Image) (context.Context, error) {
		imageEnvs := parseEnvs(image.Config.Env)

		missing := []string{}
		mismatched := []string{}

		for _, env := range envs {
			k, expected := parseEnv(env)

			if actual, ok := imageEnvs[k]; ok {
				if actual != expected {
					mismatched = append(
						mismatched,
						fmt.Sprintf("%s=%#v (expected) != %s=%#v (actual)", k, expected, k, actual),
					)
				}
			} else {
				missing = append(missing, env)
			}
		}

		if len(missing) > 0 {
			return ctx, errors.Errorf("the image environment is missing environment variables: %v", missing)
		}

		if len(mismatched) > 0 {
			return ctx, errors.Errorf("some image environment variables differ: %s", strings.Join(mismatched, ", "))
		}

		return ctx, nil
	})
}

func withCtxValue[T any](ctx context.Context, key ctxKey, f func(T) (context.Context, error)) (context.Context, error) {
	val, ok := ctx.Value(key).(T)

	if !ok {
		return ctx, errors.New("failed to get the context value")
	}

	return f(val)
}

func withImageFS(ctx context.Context, f func(fs.FS) (context.Context, error)) (context.Context, error) {
	return withCtxValue[imagefs.FS](ctx, imageFsKey, func(image imagefs.FS) (context.Context, error) {
		return f(image.WithContext(ctx))
	})
}

func withImageFile(ctx context.Context, path string, f func(io.Reader) (context.Context, error)) (context.Context, error) {
	return withImageFS(ctx, func(image fs.FS) (context.Context, error) {
		file, err := image.Open(path)

		if err != nil {
			return ctx, errors.Wrapf(err, "failed to open %s from image filesystem", path)
		}

		return f(file)
	})
}

func withImageFileData(ctx context.Context, path string, f func([]byte) (context.Context, error)) (context.Context, error) {
	return withImageFile(ctx, path, func(file io.Reader) (context.Context, error) {
		data, err := io.ReadAll(file)

		if err != nil {
			return ctx, errors.Wrapf(err, "failed to read %s", path)
		}

		return f(data)
	})
}

type workingDirectory struct {
	Path string
}

func newWorkingDirectory() (*workingDirectory, error) {
	path, err := os.MkdirTemp("", "blubber-examples-")

	if err != nil {
		return nil, err
	}

	return &workingDirectory{path}, nil
}

func (wd *workingDirectory) WriteFile(name string, data []byte, mode os.FileMode) error {
	return os.WriteFile(filepath.Join(wd.Path, name), data, mode)
}

func (wd *workingDirectory) Remove() error {
	return os.RemoveAll(wd.Path)
}

func (wd *workingDirectory) CopyFrom(srcDir string) error {
	if srcDir[len(srcDir)-1] != '/' {
		srcDir = srcDir + "/"
	}

	// Note the use of "<src>/." syntax which should work with both BSD and GNU
	// cp to copy the _contents_ of the source directory into the destination
	return exec.Command("cp", "-a", srcDir+".", wd.Path+"/").Run()
}

// parseEnv takes a env variable declaration (which may or may not use double
// quotes to qualify the value) and returns the resulting name and value.
func parseEnv(env string) (string, string) {
	kv := strings.SplitN(env, "=", 2)

	if len(kv) == 2 {
		v, err := shlex.Split(kv[1])

		if err == nil && len(v) > 0 {
			return kv[0], v[0]
		}

		return kv[0], ""
	}

	return env, ""
}

func parseEnvs(env []string) map[string]string {
	m := make(map[string]string, len(env))

	for _, def := range env {
		k, v := parseEnv(def)
		m[k] = v
	}

	return m
}
