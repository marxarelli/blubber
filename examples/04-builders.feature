Feature: Builders
  Use cases will often involve more than simply copying files to the image.
  Many will require a build step: some process that creates additional files
  in the image's local filesystem needed at runtime.

  @set3
  Scenario: Compiling an application from source
    Given "examples/hello-world-c" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:bullseye
          apt:
            packages:
              - gcc
              - libc6-dev
          builder:
            requirements: [main.c]
            command: [gcc, -o, hello, main.c]
          entrypoint: [./hello]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello |

  @set4
  Scenario: Compiling an application using multiple builders
    Given "examples/hello-world-go" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: golang:1.18
          builders:
            - custom:
                requirements: [go.mod, go.sum]
                command: [go, mod, download]
            - custom:
                requirements: [main.go]
                command: [go, build, .]
          entrypoint: [./hello-world-go]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello-world-go |
