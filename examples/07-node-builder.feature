Feature: Node builder
  Blubber supports a specialized Node builder for easy and consistent
  dependency installation and setup for Node projects.

  Background:
    Given "examples/hello-world-node" as a working directory

  Scenario: Installing Node application dependencies
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: node:20-bullseye
          builders:
            - node:
                requirements: [package.json, package-lock.json]
          copies: [local]
          entrypoint: [node, hello.js]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully
