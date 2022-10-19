SHELL := /bin/bash
RELEASE_DIR ?= ./_release
TARGETS ?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le windows/amd64 plan9/amd64
VERSION = $(shell cat VERSION)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
FULLVERSION = $(VERSION)-$(GIT_COMMIT)

PACKAGE := gitlab.wikimedia.org/repos/releng/blubber

GO_LIST_GOFILES := '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}{{range .XTestGoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}'
GO_PACKAGES = $(shell go list ./...)

GO_LDFLAGS = \
  -X $(PACKAGE)/meta.Version=$(VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(GIT_COMMIT)

# go build/install commands
#
GO_BUILD = go build -v -ldflags "$(GO_LDFLAGS)"
GO_INSTALL = go install -v -ldflags "$(GO_LDFLAGS)"

# Respect TARGET* variables defined by docker
# see https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
GOOS = $(TARGETOS)
GOARCH = $(TARGETARCH)
export GOOS
export GOARCH

all: code blubber blubberoid blubber-buildkit

blubber:
	$(GO_BUILD) ./cmd/blubber

blubberoid:
	$(GO_BUILD) ./cmd/blubberoid

blubber-buildkit: download
	$(GO_BUILD) ./cmd/blubber-buildkit

code:
	go generate $(GO_PACKAGES)

clean:
	go clean $(GO_PACKAGES) || true
	rm -f blubber blubberoid || true

download:
	go mod download

install: all
	$(GO_INSTALL) $(GO_PACKAGES)

install-tools: download
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

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
	@test -z "$$(grep '[^[:blank:]]' .lint-gofmt.diff)" || (echo "gofmt found errors:"; cat .lint-gofmt.diff; exit 1)
	golint -set_exit_status $(GO_PACKAGES)
	go vet -composites=false $(GO_PACKAGES)

unit:
	go test -cover -ldflags "$(GO_LDFLAGS)" $(GO_PACKAGES)

blubber-buildkit-docker:
	DOCKER_BUILDKIT=1 docker build --pull=false -f .pipeline/blubber.yaml --target buildkit -t localhost/blubber-buildkit .
	@echo Buildkit Docker image built
	@echo It can be used locally in a .pipeline/blubber.yaml with:
	@echo '   # syntax = localhost/blubber-buildkit'

test-docker:
	DOCKER_BUILDKIT=1 docker build -f .pipeline/blubber.yaml --target test -t blubber/test .
	docker run -it --rm blubber/test

test: unit lint

FULLVERSION:
	@echo $(FULLVERSION) > FULLVERSION
