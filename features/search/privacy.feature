# Search API: what is kept, what is passed on, what is never logged. @security
#
# APPROVED by the founder 2026-11-15. A search query says a lot about a person. The words go from the
# caller to the broker, to ONE station, and from that station to its engine. That path is
# stated plainly to callers (docs + the app copy), and nothing beyond it is kept or forwarded.
#
# Ground truth: cmd/rogerai-broker/audio.go (job.User = b.pseudonym(user, node), the rebuilt
# Job), internal/agent/agent.go (only Content-Type + Authorization forwarded upstream), the
# ledger row shape in internal/store.
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   P1  the broker never writes query text or result text to a log, a ledger row, a metric,
#       the lineage feed or the console. It records model, node, latency, result count, cost.
#   P2  the station receives a per-station pseudonym, never the caller's account, wallet, key,
#       email or IP.
#   P3  the node forwards to its engine only the search parameters. No caller identity, no
#       pseudonym, no broker headers.
#   P4  CSAM evidence preservation (moderation.feature X4) is the ONE place query text is
#       stored, under the existing evidence rules.
#   P5  the public docs and the search station listing say who sees the words: RogerAI's
#       broker and the station you are routed to, and that station's engine.

@search @security @privacy
Feature: Search words go to one station and its engine, and are not kept

  Background:
    Given a consumer with a funded wallet
    And a search station "web-search" on air with engine "brave"

  Rule: no query or result text is kept by the broker (P1)

    # The probe carries letters hex never uses (k, u, q, m, t), so a request id, signature or
    # hash in a stored record cannot match it by chance; "7f3a" alone did, once (2026-11-25).

    Scenario: A successful search leaves no query text behind
      When the consumer searches for "a distinctive phrase kumquat7f3a"
      Then the broker log does not contain "kumquat7f3a"
      And no ledger row, metric row, lineage entry or console event contains "kumquat7f3a"

    Scenario: A failed search leaves no query text behind
      Given the station returns HTTP 502
      When the consumer searches for "a distinctive phrase kumquat7f3a"
      Then the broker log does not contain "kumquat7f3a"

    Scenario: A refused query leaves no query text behind
      When the consumer searches for a flagged phrase containing "kumquat7f3a"
      Then the broker log does not contain "kumquat7f3a"
      And only the moderation category is recorded

    Scenario: Result text is not kept
      Given the station returns a result titled "unique title 9c1e"
      When the consumer searches
      Then no log, ledger row, metric row, lineage entry or console event contains "9c1e"

    Scenario: What IS recorded is enough to bill and debug
      When the consumer searches
      Then the request's records hold the model, node, latency, result count and cost

    Scenario: GET /generation for a search shows routing and billing, not words
      When the consumer searches and then reads GET /generation for that request
      Then it shows the model, the station, the cost and the result count
      And it shows no query or result text

    Scenario: A data export holds no search words
      When the consumer searches and then requests GET /account/export
      Then the export lists the search's model, time and cost
      And the export contains no query or result text

  Rule: the station learns nothing about who asked (P2)

    Scenario: The station receives a pseudonym, not the caller
      When the consumer searches
      Then the job the station receives names the caller only by a pseudonym
      And the job carries no account id, wallet id, email, grant key or IP address

    Scenario: The pseudonym differs per station
      Given two search stations on air
      When the consumer searches once on each
      Then the two stations receive different pseudonyms for the same caller

  Rule: the engine learns only the search (P3)

    Scenario Outline: The node sends its engine only the search parameters
      Given a node serving engine "<engine>" against a recording upstream
      When the broker relays a search job for "tides" from a caller with pseudonym "u-123"
      Then the upstream request carries the query "tides"
      And no upstream header, parameter or body field contains "u-123"
      And no upstream header names the broker, the job id or the caller

      Examples:
        | engine  |
        | brave   |
        | searxng |
        | native  |

  Rule: we say who sees the words (P5)

    Scenario: The search docs say where the words go
      When a reader opens the public search API docs
      Then they say the query goes to RogerAI's broker, to the one station that answers, and to that station's engine
      And they say RogerAI does not keep the query text

    Scenario: The search station listing names each station's engine
      When an anonymous GET /search/stations arrives
      Then every listed station names its engine
