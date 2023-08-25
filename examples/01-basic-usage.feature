Feature: Basic usage

  The most basic operations when building an image are: starting with a given
  base image, copying local files to the image filesystem, and specifying some
  defaults for how to run a container based on the final image.

  Blubber supports these operations and adds some basic constraints
  around who owns the files and runtime process.

  Background:
    Given "examples/hello-world" as a working directory

  @set1
  Scenario: Copy in a script and run it as the container entry point
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: debian:bullseye    # start with a debian system
          copies: [local]          # copy our working directory to the default application path
          entrypoint: [./hello.sh] # run ./hello.sh when a container is started using this image
      """
    When you build the "hello" variant
    Then the image will have the following files in the default working directory
      | README.md |
      | hello.sh  |
    And the image runtime user will be "900"
    And the image entrypoint will be "./hello.sh"

  @set2
  Scenario: Blubber respects .dockerignore files
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
      """
    And this ".dockerignore"
      """
      /README.md
      """
    When you build the "hello" variant
    Then the image will have the following files in the default working directory
      | hello.sh  |
    And the image will not have the following files in the default working directory
      | README.md |

  @set3
  Scenario: Variants can include one another
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
        hey:
          includes: [hello]
      """
    When you build the "hey" variant
    Then the image will have the following files in the default working directory
      | README.md |
      | hello.sh  |
