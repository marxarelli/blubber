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

## Usage

Running the `blubber` command will be produce `Dockerfile` output for the
given variant.

    blubber config.yaml variant

You can see the result of the example configuration by cloning this repo and
running (assuming you have go):

    make
    ./blubber blubber.example.yaml development
    ./blubber blubber.example.yaml test
    ./blubber blubber.example.yaml production

## Contribution

If you'd like to make code contributions to Blubber, see
[CONTRIBUTING.md](CONTRIBUTING.md).

## BuildKit frontend for `buildctl` and `docker build`

In addition to a CLI and a microservice, Blubber includes a [BuildKit gRPC
gateway](https://github.com/moby/buildkit#exploring-dockerfiles) that works
with both BuildKit's `buildctl` command and with `docker build`.

To build from Blubber configuration using `buildctl`, do:

    buildctl build --frontend gateway.v0 \
      --opt source=docker-registry.wikimedia.org/wikimedia/blubber-buildkit:0.9.0 \
      --local context=. \
      --local dockerfile=. \
      --opt filename=blubber.yaml
      --opt variant=test

If you'd like to build directly with `docker build` (or other toolchains that
invoke it like `docker-compose`), specify a [syntax
directive](https://docs.docker.com/engine/reference/builder/#syntax) at the
top of your Blubber configuration like so.

    # syntax=docker-registry.wikimedia.org/wikimedia/blubber-buildkit:0.9.0
    version: v4
    variants:
      my-variant:
      [...]

And invoke `docker build --target my-variant -f blubber.yaml .`. Note that
Docker must have BuildKit enabled as the default builder. You can also use
`docker buildx` which always uses BuildKit.
