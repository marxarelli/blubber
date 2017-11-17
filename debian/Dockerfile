FROM debian:buster

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -y \
  build-essential \
  dh-golang \
  git \
  git-buildpackage \
  golang-any \
  golang-go

RUN mkdir -p /src/blubber
WORKDIR /src/blubber
COPY . .

ENTRYPOINT ["/bin/make", "-f", "debian/Makefile"]
