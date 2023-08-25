Feature: Copying from other variants
  Build dependencies can add up quickly, and in most cases you don't want them
  around in your production image if they're not needed at runtime. You can
  copy files from a build variant into a more minimal production variant in a
  couple of different ways.

  Background:
    Given "examples/hello-world-go" as a working directory

  @set1
  Scenario: Copying from another variant using "copies.from"
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: golang:1.18
          lives:
            in: /src
          builders:
            - custom:
                requirements: [go.mod, go.sum]
                command: [go, mod, download]
            - custom:
                requirements: [main.go]
                command: [go, build, .]
        production:
          base: ~ # Go binaries are statically linked, so we can even use a scratch image here
          copies:
            - from: build
              source: /src/hello-world-go
              destination: /hello-world
          entrypoint: [/hello-world]
      """
    When you build the "production" variant
    Then the image will have the following files in "/"
      | hello-world |
