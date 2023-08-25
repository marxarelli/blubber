Feature: Defining the build and runtime environment

  When it comes to the application directory, the builder user/UID and
  group/GID, and environment variables, Blubber behaves with a set of opinionated
  default values. However, you can tweak any one of these values.

  Background:
    Given "examples/hello-world" as a working directory

  @set4
  Scenario: Rely on the default values
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build the "hello" variant
    Then the image will have the user "somebody" with UID 65533
    And the image will have the group "somebody" with GID 65533
    And the image will have the user "runuser" with UID 900
    And the image will have the group "runuser" with GID 900
    And the image will have the following files in "/srv/app"
      | owner | group | name      |
      | 65533 | 65533 | README.md |
      | 65533 | 65533 | hello.sh  |
    And the image runtime user will be "900"
    And the image entrypoint will be "./hello.sh"

  @set1
  Scenario: Customize the application location and file owner
    Given this "blubber.yaml"
      """
      version: v4
      lives:
        in: /usr/local/app
        as: appowner
        uid: 1234
        gid: 1235
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build the "hello" variant
    Then the image will have the user "appowner" with UID 1234
    And the image will have the group "appowner" with GID 1235
    And the image will have the following files in "/usr/local/app"
      | owner | group | name      |
      | 1234 | 1235 | README.md |
      | 1234 | 1235 | hello.sh  |

  @set2
  Scenario: Customize the runtime process owner
    Given this "blubber.yaml"
      """
      version: v4
      runs:
        as: apprunner
        uid: 4321
        gid: 4320
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build the "hello" variant
    Then the image will have the user "apprunner" with UID 4321
    And the image will have the group "apprunner" with GID 4320
    And the image runtime user will be "4321"

  @set3
  Scenario: Disable the unprivileged runtime process owner
    Given this "blubber.yaml"
      """
      version: v4
      lives:
        uid: 1234
      runs:
        insecurely: true
        uid: 4321
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build the "hello" variant
    Then the image runtime user will be "1234"

  @set4
  Scenario: Defining extra environment variables
    Given this "blubber.yaml"
      """
      version: v4
      runs:
        environment:
          FOO: bar
          BAZ: qux
      variants:
        hello:
          base: debian:bullseye
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build the "hello" variant
    Then the image will include environment variables
      | FOO=bar |
      | BAZ=qux |
