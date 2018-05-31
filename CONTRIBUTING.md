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
 3. `arcanist` for code review
    * See our [help article](https://www.mediawiki.org/wiki/Phabricator/Arcanist)
      for setup instructions.
 4. An account at [phabricator.wikimedia.org](https://phabricator.wikimedia.org)
    * See our [help article](https://www.mediawiki.org/wiki/Phabricator/Help)
      for setup instructions.
 5. (optional) `gox` is used for cross-compiling binary releases. To
    install `gox` use `go get github.com/mitchellh/gox`.

## Get the source

Use `go get` to install the source from our Git repo into `src` under your
`GOPATH`. By default, this will be `~/go/src`.

    go get phabricator.wikimedia.org/source/blubber

Symlink it to a different directory if you'd prefer not to work from your
`GOPATH`. For example:

    cd ~/Projects
    ln -s ~/go/src/phabricator.wikimedia.org/source/blubber
    cd blubber # yay.

## Initialize submodules in `.arcvendor`

We currently use a submodule for integrating Go testing tools into Arcanist
which will run automatically upon submission to Differential via `arc diff`.

    git submodule update --init

## Have a read through the documentation

If you haven't already seen the [README.md](README.md), check it out.

Run `godoc -http :9999` and peruse the HTML generated from inline docs
at `localhost:9999/pkg/phabricator.wikimedia.org/source/blubber`.

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
when you submit your changes to Differential via `arc diff`. You can also do
this manually.

    arc unit # or
    arc unit --everything # or simply
    go test ./... # or
    go test -run TestFuncName ./... # to run a single test

    arc lint # or
    arc lint --everything

## Getting your changes reviewed

Use `arc diff` to submit your changes to Differential.


## Landing your changes

Once your changes have been accepted, run `arc land` on your local branch to
merge/push the commit and close the diff.

## Releases

The `release` target of the `Makefile` in this repository uses `gox` to
cross-compile binary releases of Blubber.

    make release
