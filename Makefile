export GOPATH=$(CURDIR)/.build
PACKAGE := phabricator.wikimedia.org/source/blubber.git
export BUILD_DIR=$(GOPATH)/src/$(PACKAGE)
BINARY :=$(CURDIR)/bin/blubber

all: clean bin/blubber

clean:
	rm -rf $(CURDIR)/.build
	rm -rf $(BINARY)

bin/blubber:
	mkdir -p $(dir $(BUILD_DIR))
	ln -s $(CURDIR) $(BUILD_DIR)
	cd $(BUILD_DIR) && go get ./...
	cd $(BUILD_DIR) && go build -v -i
	mkdir -p $(CURDIR)/bin && mv $(BUILD_DIR)/blubber.git $(BINARY)


.PHONY: all clean
