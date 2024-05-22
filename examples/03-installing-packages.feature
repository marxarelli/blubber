Feature: Installing packages
  Often variants will need to install additional software to satisfy build or
  runtime dependencies. You can have APT packages installed using the `apt`
  directives.

  Background:
    Given "examples/hello-world-c" as a working directory

  @set1
  Scenario: Install additional build dependencies
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:bullseye
          apt:
            packages:
              - gcc
              - libc6-dev
      """
    When you build the "build" variant
    Then the image will have the following files in "/usr/bin"
      | gcc |

  @set2
  Scenario: Install from additional APT sources
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: docker-registry.wikimedia.org/golang1.21:1.21-1-20240609
          apt:
            sources:
              - url: https://apt.wikimedia.org/wikimedia
                distribution: bullseye-wikimedia
                components:
                  - thirdparty/amd-rocm54
            packages:
              bullseye-wikimedia: # you may use an explicit distribution/release name like so
                - fake-libgcc-7-dev
      """
    When you build the "build" variant
    Then the image will have the following files in "/usr/share/doc/fake-libgcc-7-dev"
      | copyright     |
