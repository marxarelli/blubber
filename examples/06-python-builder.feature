Feature: Python builder
  Blubber supports a specialized Python builder for easy and consistent
  dependency installation and setup for Python projects.

  Background:
    Given "examples/hello-world-python" as a working directory

  Scenario: Installing Python application dependencies
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: python:3.10-bullseye
          builders:
            - python:
                version: python3
                requirements: [requirements.txt]
          copies: [local]
          entrypoint: [python3, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully

  Scenario: Installing Python application dependencies via Poetry
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: python:3.10-bullseye
          builders:
            - python:
                version: python3
                poetry:
                  version: ==1.5.1
                requirements: [pyproject.toml, poetry.lock]
          copies: [local]
          entrypoint: [poetry, run, python3, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully
