@minutiae
Feature: Minutiae

  These scenarios describe some of the details of Blubber's image
  creation.

  Background:
    Given "examples/hello-world" as a working directory

  @set1
  Scenario: Blubber embeds its version as a label
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello: {}
      """
    When you build the "hello" variant
    Then the image will include labels
      | blubber.version |

  @set2
  Scenario: Blubber embeds the target variant name as a label
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello: {}
      """
    When you build the "hello" variant
    Then the image will include labels
      | blubber.variant | hello |
