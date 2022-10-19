# Contributing to Blubber

`blubber` is an open source project maintained by Wikimedia Foundation's
Release Engineering Team and developed primarily to support a continuous
delivery pipeline for MediaWiki and related applications. We will, however,
consider any contribution that advances the project in a way that is valuable
to both users inside and outside of WMF and our communities.

## Requirements

 1. `go` >= 1.17 and related tools
    * To install on rpm style systems: `sudo dnf install golang golang-godoc`
    * To install on apt style systems: `sudo apt install golang golang-golang-x-tools`
    * To install on macOS use [Homebrew](https://brew.sh) and run:
      `brew install go`
    * You can run `go version` to check the golang version.
    * If your distro's go package is too old or unavailable,
      [download](https://golang.org/dl/) a newer golang version.
 2. An account at [gerrit.wikimedia.org](https://gerrit.wikimedia.org)
    * See the [guide](https://www.mediawiki.org/wiki/Gerrit/Getting_started)
      on mediawiki.org for setup instructions.
 3. (optional) `gox` is used for cross-compiling binary releases.
    * To install `gox` use `go get github.com/mitchellh/gox`.
 4. (optional) `golint` is used in `make lint` for code checking.
    * To install `golint` use `go get -u golang.org/x/lint/golint`
    * More info at: https://github.com/golang/lint

## Get the source

Use `go get` to install the source from our Git repo into `src` under your
`GOPATH`. By default, this will be `~/go/src`.

    go get gitlab.wikimedia.org/repos/releng/blubber

Symlink it to a different directory if you'd prefer not to work from your
`GOPATH`. For example:

    cd ~/Projects
    ln -s ~/go/src/gitlab.wikimedia.org/repos/releng/blubber
    cd blubber # yay.

## Have a read through the documentation

If you haven't already seen the [README.md](README.md), check it out.

Run `godoc -http :9999` and peruse the HTML generated from inline docs
at `localhost:9999/pkg/gitlab.wikimedia.org/repos/releng/blubber`.

## Running tests and linters

Tests and linters for packages/files you've changed will automatically run
when you submit your changes to Gerrit for review. You can also run them
locally using the `Makefile`:

    make lint # to run all linters
    make unit # or all unit tests
    make test # or all linters and unit tests

    go test -run TestFuncName ./... # to run a single test function

Alternatively you can run the test inside a Blubber built image
(`.pipeline/blubber.yaml`) using our Docker build-kit:

    make test-docker

## Getting your changes reviewed and merged

Push your changes to Gerrit for review. See the
[guide](https://www.mediawiki.org/wiki/Gerrit/Tutorial#How_to_submit_a_patch)
on mediawiki.org on how to correctly prepare and submit a patch.

## Releases

The `release` target of the `Makefile` in this repository uses `gox` to
cross-compile binary releases of Blubber.

    make release

## Testing and debugging the BuildKit frontend

Debugging the gRPC BuildKit gateway (`cmd/blubber-buildkit`) can be difficult
as stack traces do not surface from the user-facing tools like `docker build`
or `buildctl`. The easiest way to get access to the gateway's logging is to
start `buildkitd` in a container and use `docker logs -f` to tail its logs
while building.

Start `buildkitd` in a Docker container.

    docker run -d --name buildkitd --privileged moby/buildkit:latest
    export BUILDKIT_HOST=docker-container://buildkitd

Build the buildkit gateway image and tag it for distribution to a
registry accessible by `buildkitd`.

    ./blubber .pipeline/blubber.yaml buildkit \
      | docker build -t my-docker-io-account/blubber-buildkit -f - .

Publish the gateway image. (Requires that you first auth with `docker login`.)

    docker push my-docker-io-account/blubber-buildkit

Tail `buildkitd` logs in a terminal.

    docker logs -f buildkitd

Build a configuration using `buildctl` and the published gateway image.

    buildctl build \
        --frontend gateway.v0 \
        --opt source=my-docker-io-account/blubber-buildkit \
        --local context=. \
        --local dockerfile=. \
        --opt filename=.pipeline/blubber.yaml \
        --opt variant=test

If a fatal occurs, you should now see the full error and stack trace in the
Docker logs.
