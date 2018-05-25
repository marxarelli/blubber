![Blubber](http://tyler.zone/blubber.png)

**Very experimental proof of concept.**

Blubber is a highly opinionated abstraction for container build configurations
and a command-line compiler which currently supports outputting multi-stage
Dockerfiles. It aims to provide a handful of declarative constructs that
accomplish build configuration in a more secure and determinate way than
running ad-hoc commands.

## Example configuration

```yaml
version: v2
base: debian:jessie
apt:
  packages: [libjpeg, libyaml]
runs:
  in: /srv/service
  as: runuser
  uid: 666
  gid: 666
  environment:
    FOO: bar
    BAR: baz

variants:
  build:
    apt:
      packages: [libjpeg-dev, libyaml-dev]
    node:
      requirements: [package.json, package-lock.json]

  development:
    includes: [build]
    sharedvolume: true

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
    copies: prep
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
version: v2
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
makes use of this with its `artifacts` configuration property.

In the example configuration, the `production` variant declares artifacts to
be copied over from the result of building the `test` image.

## Usage

Running the `blubber` command will be produce `Dockerfile` output for the
given variant.

    blubber config.yaml variant

You can see the result of the example configuration by cloning this repo and
running (assuming you have go):

    make
    ./bin/blubber blubber blubber.example.yaml development
    ./bin/blubber blubber blubber.example.yaml test
    ./bin/blubber blubber blubber.example.yaml production

## Contribution

If you'd like to make code contributions to Blubber, see
[CONTRIBUTING.md](CONTRIBUTING.md).
