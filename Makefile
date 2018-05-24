RELEASE_DIR ?= ./_release
TARGETS ?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le windows/amd64 plan9/amd64

PACKAGE := phabricator.wikimedia.org/source/blubber

GO_LDFLAGS := \
  -X $(PACKAGE)/meta.Version=$(shell cat VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(shell git rev-parse --short HEAD)

install:
	go install -v -ldflags "$(GO_LDFLAGS)"

release:
	gox -output="$(RELEASE_DIR)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' -ldflags '$(GO_LDFLAGS)' $(PACKAGE)
	cp LICENSE "$(RELEASE_DIR)"
	for f in "$(RELEASE_DIR)"/*/blubber; do \
		shasum -a 256 "$${f}" | awk '{print $$1}' > "$${f}.sha256"; \
	done

.PHONY: install release
