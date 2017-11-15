export GOPATH=$(CURDIR)/.build
PACKAGE := phabricator.wikimedia.org/source/blubber
export BUILD_DIR=$(GOPATH)/src/$(PACKAGE)
BINARY :=$(CURDIR)/bin/blubber

GO_LDFLAGS := \
  -X $(PACKAGE)/meta.Version=$(shell cat VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(shell git rev-parse --short HEAD)

all: clean bin/blubber

clean:
	rm -rf $(CURDIR)/.build
	rm -rf $(BINARY)

bin/blubber:
	mkdir -p $(dir $(BUILD_DIR))
	ln -s $(CURDIR) $(BUILD_DIR)
	cd $(BUILD_DIR) && go build -v -i -ldflags "$(GO_LDFLAGS)"
	mkdir -p $(CURDIR)/bin && mv $(BUILD_DIR)/blubber $(BINARY)
	rm -rf $(CURDIR)/.build


.PHONY: all clean
