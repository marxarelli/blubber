Blubber is a BuildKit frontend for building application container images from
a minimal set of declarative constructs in YAML. Its focus is on
composability, determinism, cache efficiency, and secure default behaviors.

## Examples

To skip to the examples, see the feature files in the [examples](https://gitlab.wikimedia.org/repos/releng/blubber/-/blob/main/examples/)
directory. The examples are implemented as executable Cucumber tests to ensure
Blubber is always working as expected by users.

## Concepts

### Variants

Blubber supports a concept of composeable configuration variants for defining
slightly different container images while still maintaining a sufficient
degree of parity between them. For example, images for development and testing
may require some development and debugging packages which you wouldn't want in
production lest they contain vulnerabilities and somehow end up linked or
included in the application runtime.

See the [copying from other variants example](https://gitlab.wikimedia.org/repos/releng/blubber/-/blob/main/examples/05-copying-from-other-variants.feature).

### Builders

Builders represent a discrete process and a set of files that is needed to
produce an application artifact.

See the [builders example](https://gitlab.wikimedia.org/repos/releng/blubber/-/blob/main/examples/04-builders.feature).

When defining multiple builders, be sure to use the `builders` field to ensure
an explicit ordering.

Similarly to other configuration keys, `builders` appearing at the top level
of the file will be applied to all variant configurations. Builder keys
appearing both at the top level and in a variant, will be merged; whereas
builders present only at the top level will be placed first in the execution
order.

For a particular variant, `builders` and the standalone builder keys are
mutually exclusive, but different styles can be used for different variants.
However, note that top level definitions are applied to all variants, so using
one style at the top level precludes the use of the other for all variants.

## Usage

Blubber used to include both a CLI and microservice for transpiling to
Dockerfile text. It is now exclusively a [BuildKit
frontend](https://github.com/moby/buildkit#exploring-dockerfiles) that works
with both BuildKit's `buildctl` command and with `docker build` directly.

To build from Blubber configuration using `buildctl`, do:

```console
$ buildctl build --frontend gateway.v0 \
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.22.0 \
  --local context=. \
  --local dockerfile=. \
  --opt filename=blubber.yaml \
  --opt variant=test
```

If you'd like to build directly with `docker build` (or other toolchains that
invoke it like `docker-compose`), specify a [syntax
directive](https://docs.docker.com/engine/reference/builder/#syntax) at the
top of your Blubber configuration like so.

    # syntax=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.22.0
    version: v4
    variants:
      my-variant:
      [...]

And invoke `docker build --target my-variant -f blubber.yaml .`. Note that
Docker must have BuildKit enabled as the default builder. You can also use
`docker buildx` which always uses BuildKit.

Docker's [build-time arguments][build_args] are also supported, including those
used to provide proxies to build processes.

```console
buildctl build --frontend gateway.v0 \
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.22.0 \
  --opt build-arg:http_proxy=http://proxy.example \
  --opt variant=pulls-in-stuff-from-the-internet
  ...
```

[build_args]: https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg

### Additional options for the Buildkit frontend

The following options can be passed via command line (via `--opt`) to configure the build process:
 * `run-variant`: bool. Instructs Blubber to run the target variant's entrypoint (if any) as part
of the BuildKit image build process
 * `entrypoint-args`: JSON array. List of additional arguments for the entrypoint
 * `run-variant-env`: JSON object of key/value pairs to set in the environment when `run-variant` is true.

Example usage:

```console
$ buildctl build --frontend gateway.v0 \
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.22.0 \
  --local context=. \
  --local dockerfile=. \
  --opt filename=blubber.yaml \
  --opt variant=test \
  --opt run-variant=true \
  --opt entrypoint-args='["extraParam1", "extraParam2"]' \
  --opt run-variant-env='{"SOME_VARIABLE": "somevalue"}'
  ...
```

### Building for multiple platforms

Blubber's BuildKit frontend supports building for multiple platforms at once
and publishing a single manifest index for the given platforms (aka a "fat"
manifest). See the [OCI Image Index Specification][oci-image-index] for
details.

Note that your build process must be aware of the [environment
variables][multi-platform-env-vars] set for multi-platform builds in order to
perform any cross-compilation needed.

Example usage:

```console
$ buildctl build --frontend gateway.v0 \
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.22.0 \
  --local context=. \
  --local dockerfile=. \
  --opt filename=blubber.yaml \
  --opt variant=production \
  --opt platform=linux/amd64,linux/arm64 \
  --output type=image,name=my/multi-platform-app:v1.0,push=true

$ docker manifest inspect my/multi-platform-app:v1.0
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
   "manifests": [
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": <n>,
         "digest": "sha256:<digest>",
         "platform": {
            "architecture": "amd64",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": <n>,
         "digest": "sha256:<digest>",
         "platform": {
            "architecture": "arm64",
            "os": "linux"
         }
      }
   ]
}
```

[multi-platform-env-vars]: https://docs.docker.com/build/building/multi-platform/#building-multi-platform-images
[oci-image-index]: https://github.com/opencontainers/image-spec/blob/main/image-index.md
