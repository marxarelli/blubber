# Contributing to Blubber

`blubber` is an open source project maintained by Wikimedia Foundation's
Release Engineering Team and developed primarily to support a continuous
delivery pipeline for MediaWiki and related applications. We will, however,
consider any contribution that advances the project in a way that is valuable
to both users inside and outside of WMF and our communities.

## Requirements

 1. `go` >= 1.9 (>=1.10 recommended) and related tools
    * To install on rpm style systems: `sudo dnf install golang golang-godoc`
    * To install on apt style systems: `sudo apt install golang golang-golang-x-tools`
    * To install on macOS use [Homebrew](https://brew.sh) and run:
      `brew install go`
    * You can run `go version` to check the golang version.
    * If your distro's go package is too old or unavailable,
      [download](https://golang.org/dl/) a newer golang version.
 2. `dep` for dependency management
    * On macOS, try Homebrew: `brew install dep`
    * [Other](https://golang.github.io/dep/docs/installation.html)
 3. An account at [gerrit.wikimedia.org](https://gerrit.wikimedia.org)
    * See the [guide](https://www.mediawiki.org/wiki/Gerrit/Getting_started)
      on mediawiki.org for setup instructions.
 4. (optional) `gox` is used for cross-compiling binary releases. To
    install `gox` use `go get github.com/mitchellh/gox`.

## Get the source

Use `go get` to install the source from our Git repo into `src` under your
`GOPATH`. By default, this will be `~/go/src`.

    go get gerrit.wikimedia.org/r/blubber

Symlink it to a different directory if you'd prefer not to work from your
`GOPATH`. For example:

    cd ~/Projects
    ln -s ~/go/src/gerrit.wikimedia.org/r/blubber
    cd blubber # yay.

## Have a read through the documentation

If you haven't already seen the [README.md](README.md), check it out.

Run `godoc -http :9999` and peruse the HTML generated from inline docs
at `localhost:9999/pkg/gerrit.wikimedia.org/r/blubber`.

## Installing or updating dependencies

Dealing with Go project dependencies is kind of a moving target at the moment,
but for now we've opted to commit a minimal `vendor` directory which contains
all the required packages. It has been automatically populated by `dep
ensure && dep prune` according to our `Gopkg.toml` and `Gopkg.lock` files.

If you're not making any changes to `Gopkg.toml`, adding, updating, or
removing dependencies, you should already be good to go.

If you do update `Gopkg.toml` to add, update, or remove a dependency, simply
run `dep ensure && dep prune` after doing so, and commit the resulting
`vendor` directory changes.

## Running tests

Tests and linters for packages/files you've changed will automatically run
when you submit your changes to Gerrit for review. You can also run tests
locally by running `go test`.

    go test ./... # for everything, or
    go test -run TestFuncName ./... # to run a single test

## Getting your changes reviewed and merged

Push your changes to Gerrit for review. See the
[guide](https://www.mediawiki.org/wiki/Gerrit/Tutorial#How_to_submit_a_patch)
on mediawiki.org on how to correctly prepare and submit a patch.

## Releases

The `release` target of the `Makefile` in this repository uses `gox` to
cross-compile binary releases of Blubber.

    make release
