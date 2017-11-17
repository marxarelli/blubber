# Makefile used for generating new debian/changelog entries and building the
# package on non-Debian systems. It should be executed from the project's root
# directory and requires Docker.
#
# Generate a new changelog section:
#  make -f debian/Makefile changelog NEW_VERSION=0.2.0-2
#
# Build the package:
#  make -f debian/Makefile package
#
ID := $(shell ls -id . | cut -d' ' -f1)
NAME := "blubber$(ID)"
LABEL := "blubber.tmp=$(ID)"
DF := "debian/Dockerfile"

GIT_USER_NAME ?= $(shell git config user.name)
GIT_USER_EMAIL ?= $(shell git config user.email)

all: package

clean:
	docker ps -aq --filter "label=$(LABEL)" | xargs docker rm
	docker images -q --filter "label=$(LABEL)" | xargs docker rmi

_clean:
	git clean -xfdf -e debian/*
	git submodule foreach git clean -xfdf
	git reset --hard HEAD
	git submodule foreach git reset --hard
	git submodule update --init --recursive

changelog:
ifndef NEW_VERSION
	$(error You must define NEW_VERSION when generating the changelog)
else
	docker rm $(NAME) || true
	docker run --name $(NAME) $(shell docker build --label $(LABEL) -qf $(DF) .) \
	  _clean _changelog NEW_VERSION=$(NEW_VERSION) \
	  GIT_USER_NAME="$(GIT_USER_NAME)" GIT_USER_EMAIL="$(GIT_USER_EMAIL)"
	docker cp $(NAME):/src/blubber/debian/changelog debian/changelog
	docker rm $(NAME)
	git diff -- debian/changelog
endif

_changelog:
	git config user.name "$(GIT_USER_NAME)"
	git config user.email "$(GIT_USER_EMAIL)"
	gbp dch --git-author --ignore-branch --new-version $(NEW_VERSION)

package:
	docker rm $(NAME) || true
	docker run --name $(NAME) $(shell docker build --label $(LABEL) -qf $(DF) .) \
	  _clean _package
	docker diff $(NAME) | awk '$$1 == "A" && $$2 ~ "^/src/blubber_" { print $$2 }' | \
	  while read f; do \
	    docker cp $(NAME):$$f ./ ; \
	  done
	docker rm $(NAME)
	ls -l blubber_*

_package:
	gbp buildpackage --git-ignore-branch --git-ignore-new -uc -us

.PHONY: all
