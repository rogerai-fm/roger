# Search stations need a signed-in owner.
#
# APPROVED by the founder 2026-11-15. Founder ruling 2026-11-15: every search station, public or private,
# free or paid, needs a signed-in owner. Search results are third-party web content that the
# broker screens and strikes on (moderation.feature X4); a strike must land on an account, and
# every station a caller's words can reach must be attributable.
#
# This amends share_cli.feature's free defaults ("SearXNG and native default to free"): the
# price still defaults to free, but going on air needs `roger login`.
#
# Ground truth: tunnel.go register (gated := priced || private), cmd/rogerai/share_search.go.
#
# DECISIONS PINNED BY THIS SPEC (approved 2026-11-15):
#   OR1 the broker refuses a registration carrying a search offer unless it resolves to a
#       signed-in, unbanned owner: 401 when unsigned, 403 when signed but not linked.
#   OR2 `roger share --search` checks this before the engine check and says how to fix it.
#   OR3 every search response and listing entry names the owner's station; provider.station is
#       never empty.

@search @security
Feature: A search station is always attributable to a signed-in owner

  Scenario Outline: The broker refuses an anonymous search registration
    Given a node with <identity>
    When it registers a free search offer
    Then the registration is refused with status <status>
    And the message says a search station requires `roger login`

    Examples:
      | identity                         | status |
      | no signature                     | 401    |
      | a signature by an unlinked key   | 403    |

  Scenario: A banned owner cannot put a search station on air
    Given an owner whose account is banned
    When their node registers a search offer
    Then the registration is refused with status 403

  Scenario: A signed-in owner's free search station goes on air
    Given a signed-in owner "@acme"
    When their node registers a free search offer
    Then the search offer is on air and listed under "@acme"

  Scenario: A node with a chat offer and a search offer is refused when anonymous
    Given a node with no signature
    When it registers a free chat offer and a free search offer together
    Then the registration is refused
    And the message names the search offer

  Scenario: Anonymous free chat registration is unchanged
    Given a node with no signature
    When it registers a free chat offer
    Then the registration is accepted

  @cli
  Scenario: roger share --search says to log in before it checks the engine
    Given roger is not logged in
    When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888"
    Then the command exits non-zero
    And the message says "a search station needs a signed-in owner: run `roger login` first"
    And no check query is sent

  Scenario: Every served search names the owner's station
    Given a signed-in owner "@acme" with a search station "web-search" on air
    When a consumer searches "web-search"
    Then the response provider station is "@acme"
