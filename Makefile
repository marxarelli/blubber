SHELL := /bin/bash
RELEASE_DIR ?= ./_release
TARGETS ?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le windows/amd64 plan9/amd64

PACKAGE := gerrit.wikimedia.org/r/blubber
REAL_CURDIR := $(shell readlink "$(CURDIR)" || echo "$(CURDIR)")

GO_LIST_GOFILES := '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}{{range .XTestGoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}'
GO_PACKAGES := $(shell go list ./...)

GO_LDFLAGS := \
  -X $(PACKAGE)/meta.Version=$(shell cat VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(shell git rev-parse --short HEAD)

# go build/install commands
#
# workaround bug in case CURDIR is a symlink
# see https://github.com/golang/go/issues/24359
GO_GENERATE := cd "$(REAL_CURDIR)" && go generate
GO_BUILD := cd "$(REAL_CURDIR)" && go build -v -ldflags "$(GO_LDFLAGS)"
GO_INSTALL := cd "$(REAL_CURDIR)" && go install -v -ldflags "$(GO_LDFLAGS)"

all: code blubber blubberoid

blubber:
	$(GO_BUILD) ./cmd/blubber

blubberoid:
	$(GO_BUILD) ./cmd/blubberoid

code:
	$(GO_GENERATE) $(GO_PACKAGES)

clean:
	go clean
	rm -f blubber blubberoid

install: all
	$(GO_INSTALL) $(GO_PACKAGES)

release:
	gox -output="$(RELEASE_DIR)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' -ldflags '$(GO_LDFLAGS)' $(GO_PACKAGES)
	cp LICENSE "$(RELEASE_DIR)"
	for f in "$(RELEASE_DIR)"/*/{blubber,blubberoid}; do \
		shasum -a 256 "$${f}" | awk '{print $$1}' > "$${f}.sha256"; \
	done

lint:
	@echo > .lint-gofmt.diff
	@go list -f $(GO_LIST_GOFILES) $(GO_PACKAGES) | while read f; do \
		gofmt -e -d "$${f}" >> .lint-gofmt.diff; \
	done
	@test -z "$(grep '[^[:blank:]]' .lint-gofmt.diff)" || (echo "gofmt found errors:"; cat .lint-gofmt.diff; exit 1)
	golint -set_exit_status $(GO_PACKAGES)
	go vet -composites=false $(GO_PACKAGES)

unit:
	go test -ldflags "$(GO_LDFLAGS)" $(GO_PACKAGES)

test: unit lint

.PHONY: install release
