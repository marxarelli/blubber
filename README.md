![Blubber](docs/logo-400.png)

Blubber is a highly opinionated abstraction for container build configurations
and a command-line compiler which currently supports outputting multi-stage
Dockerfiles. It aims to provide a handful of declarative constructs that
accomplish build configuration in a more secure and determinate way than
running ad-hoc commands.

## Example configuration

```yaml
version: v4
base: debian:jessie
apt:
  packages: [libjpeg, libyaml]
lives:
  in: /srv/service
runs:
  environment:
    FOO: bar
    BAR: baz

variants:
  build:
    apt:
      packages: [libjpeg-dev, libyaml-dev]
    node:
      requirements: [package.json, package-lock.json]
    copies: [local]

  development:
    includes: [build]

  test:
    includes: [build]
    apt:
      packages: [chromium]
    entrypoint: [npm, test]

  prep:
    includes: [build]
    node:
      env: production

  production:
    base: debian:jessie-slim
    node:
      env: production
    copies: [prep]
    entrypoint: [node, server.js]
```

## Variants

Blubber supports a concept of composeable configuration variants for defining
slightly different container images while still maintaining a sufficient
degree of parity between them. For example, images for development and testing
may require some development and debugging packages which you wouldn't want in
production lest they contain vulnerabilities and somehow end up linked or
included in the application runtime.

Properties declared at the top level are shared among all variants unless
redefined, and one variant can include the properties of others. Some
properties, like `apt:packages` are combined when inherited or included.

In the example configuration, the `test` variant when expanded effectively
becomes:

```yaml
version: v4
base: debian:jessie
apt:
  packages: [libjpeg, libyaml, libjpeg-dev, libyaml-dev, chromium]
node:
  dependencies: true
runs:
  in: /srv/service
  as: runuser
  uid: 666
  gid: 666
entrypoint: [npm, test]
```

## Artifacts

When trying to ensure optimally sized Docker images for production, there's a
common pattern that has emerged which is essentially to use one image for
building an application and copying the resulting build artifacts to another
much more optimized image, using the latter for production.

The Docker community has responded to this need by implementing
[multi-stage builds](https://github.com/moby/moby/pull/32063) and Blubber
makes use of this with its `copies` configuration property.

In the example configuration, the `production` variant declares artifacts to
be copied over from the result of building the `prep` image.

## Builders key

As an alternative to specifying the various builder keys (`node`, `python`, `php` and `builder`),
it is possible to group builders in a list under the `builders` key. This offers two advantages:
 * It defines an order of execution for the builders. Associated instructions will be generated in
the order in which the builders appear in the file
 * It makes it possible to specify multiple custom builders. In this case, the `builder` key is
replaced by `custom`

Similarly to other configuration keys, `builders` appearing at the top level of the file will be
applied to all variant configurations. Builder keys appearing both at the top level and in a variant,
will be merged; whereas builders present only at the top level will be placed first in the execution
order.

For a particular variant, `builders` and the standalone builder keys are mutually exclusive, but
different styles can be used for different variants. However, note that top level definitions are
applied to all variants, so using one style at the top level precludes the use of the other for all
variants.

The example configuration rewritten to use `builders` becomes:

```yaml
version: v4
base: debian:jessie
apt:
  packages: [libjpeg, libyaml]
lives:
  in: /srv/service
runs:
  environment:
    FOO: bar
    BAR: baz

variants:
  build:
    apt:
      packages: [libjpeg-dev, libyaml-dev]
    builders:
      - node:
          requirements: [package.json, package-lock.json]
    copies: [local]

  development:
    includes: [build]

  test:
    includes: [build]
    apt:
      packages: [chromium]
    entrypoint: [npm, test]

  prep:
    includes: [build]
    builders:
      - node:
          env: production

  production:
    base: debian:jessie-slim
    builders:
      - node:
          env: production
    copies: [prep]
    entrypoint: [node, server.js]
```

See file `examples/blubber.builders.yaml` for a more detailed example with multiple builders.

## Usage

Running the `blubber` command will be produce `Dockerfile` output for the
given variant.

    blubber config.yaml variant

You can see the result of the example configuration by cloning this repo and
running (assuming you have go):

```console
$ make
$ ./blubber examples/blubber.yaml development
$ ./blubber examples/blubber.yaml test
$ ./blubber examples/blubber.yaml production
```

Other examples with different variants can be found under directory `examples`.

## Contribution

If you'd like to make code contributions to Blubber, see
[CONTRIBUTING.md](CONTRIBUTING.md).

## BuildKit frontend for `buildctl` and `docker build`

In addition to a CLI and a microservice, Blubber includes a [BuildKit gRPC
gateway](https://github.com/moby/buildkit#exploring-dockerfiles) that works
with both BuildKit's `buildctl` command and with `docker build`.

To build from Blubber configuration using `buildctl`, do:

```console
$ buildctl build --frontend gateway.v0 \
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.13.1 \
  --local context=. \
  --local dockerfile=. \
  --opt filename=blubber.yaml \
  --opt variant=test
```

If you'd like to build directly with `docker build` (or other toolchains that
invoke it like `docker-compose`), specify a [syntax
directive](https://docs.docker.com/engine/reference/builder/#syntax) at the
top of your Blubber configuration like so.

    # syntax=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.13.1
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
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.13.1 \
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
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.13.1 \
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
  --opt source=docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v0.13.1 \
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
