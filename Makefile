PACKAGE := phabricator.wikimedia.org/source/blubber

GO_LDFLAGS := \
  -X $(PACKAGE)/meta.Version=$(shell cat VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(shell git rev-parse --short HEAD)

install:
	go install -v -ldflags "$(GO_LDFLAGS)"

.PHONY: install
