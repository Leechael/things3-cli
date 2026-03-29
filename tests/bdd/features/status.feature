Feature: status
  Scenario: check status
    Given the CLI is available
    When I run "things3-cli status --json"
    Then command succeeds
