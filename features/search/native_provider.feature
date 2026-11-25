# Search API: native search providers (your own index, no upstream to resell).
#
# APPROVED by the founder 2026-11-15. Founder direction 2026-11-15: beyond Brave and other upstreams,
# "native" providers who run their own index are a route we document clearly and support.
# A native provider runs any service that answers ONE small HTTP contract. roger needs no
# adapter per engine, so a team with a Meilisearch, Typesense, OpenSearch or Elasticsearch
# index, a YaCy peer, or a hand-built crawler can put it on the market the same way.
#
# THE CONTRACT (what the provider's service implements)
#   POST /v1/search
#   request   {"query": string, "count": int 1..20, "freshness"?: "day"|"week"|"month"|"year",
#              "country"?: "US"}                       (Content-Type: application/json)
#   reply 200 {"results": [{"title": string, "url": "https://...", "snippet": string,
#              "published"?: RFC 3339 date or date-time}]}   at most count results, best first
#   errors    any non-2xx; the body is not shown to callers
#   auth      optional "Authorization: Bearer <key>" set by the provider in roger
#   limits    answer within 10 s; reply at most 512 KiB
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   NP1 the contract above is versioned as "search/1" and published in the public docs. A
#       later change that breaks it is "search/2", and roger keeps speaking search/1.
#   NP2 `roger search check --upstream <url> [--upstream-key K]` runs the contract checks
#       against a service WITHOUT going on air, so a provider can build against it locally.
#       It is the same check `roger share` runs (share_cli.feature C5), plus the edge cases.
#   NP3 a native station is labeled "native" everywhere, with its scope line, so a caller can
#       tell an own index from a resold one and a vertical index from a general web one.
#   NP4 the public provider guide covers the three routes (Brave, SearXNG, native), the
#       contract, the privacy duties, pricing, and what running a native index takes.
#   NP5 no reference server ships in this repo. The guide carries a short example instead.

@search @native
Feature: Anyone with their own index can become a native search provider

  Rule: the contract check (NP2)

    Scenario: A conforming service passes every check
      Given a local service that implements search/1 correctly
      When the provider runs "roger search check --upstream http://127.0.0.1:9000"
      Then every check passes
      And the command exits zero
      And nothing is registered with the broker

    Scenario Outline: Each contract rule is checked and named when it fails
      Given a local service that <breaks>
      When the provider runs "roger search check --upstream http://127.0.0.1:9000"
      Then the command exits non-zero
      And the failing check is named "<check>"

      Examples:
        | breaks                                              | check                  |
        | answers 404 on POST /v1/search                      | endpoint               |
        | returns HTML instead of JSON                        | reply is JSON          |
        | returns {"hits": [...]}                             | results array          |
        | returns more results than the count asked for       | respects count         |
        | returns a result with a "javascript:" url           | http(s) urls only      |
        | returns a result with no title                      | every result has a title |
        | returns published "last tuesday"                    | published is RFC 3339  |
        | takes 12 seconds to answer                          | answers within 10 s    |
        | returns a 600 KiB reply                             | reply under 512 KiB    |
        | answers 500 to a query with freshness "week"        | accepts freshness      |
        | answers 500 to a query with Unicode text            | accepts Unicode        |

    Scenario: Soft findings are warnings, not failures
      Given a local service that ignores freshness and country
      When the provider runs "roger search check --upstream http://127.0.0.1:9000"
      Then the command exits zero
      And it warns that freshness and country are ignored

    Scenario: The check sends only fixed test queries
      When the provider runs "roger search check --upstream http://127.0.0.1:9000"
      Then every query sent is one of the check's fixed test strings

    Scenario: The check prints the key as "key set", never the key
      When the provider runs "roger search check --upstream http://127.0.0.1:9000 --upstream-key sk-test-5e3c"
      Then the output does not contain "sk-test-5e3c"

  # Tagged @broker: these two run in the broker suite (they read its listing and routing);
  # the rest of this file runs in the roger CLI suite.
  @broker
  Rule: native stations are labeled for what they are (NP3)

    Scenario: A native station shows its engine and scope in the listing
      Given a native station "go-docs" with scope "Go module docs and the Go blog"
      When an anonymous GET /search/stations arrives
      Then "go-docs" is listed with engine "native" and that scope

    Scenario: A native station is routed like any other search station
      Given a native station "go-docs" on air at 2000 credits per 1,000,000 queries
      When a consumer searches with model "go-docs"
      Then it is answered by that station and charged 0.002000 credits

  Rule: the contract is versioned (NP1)

    Scenario: roger tells the native service which contract version it speaks
      When the node sends a search to a native service
      Then the request carries header "X-Search-Contract: search/1"

  Rule: the provider guide exists and is complete (NP4, NP5)

    Scenario: The public provider guide covers every route and duty
      When a reader opens the public search provider guide
      Then it explains the three routes: Brave, SearXNG and native
      And it gives the search/1 contract exactly as roger checks it
      And it shows a short native example service
      And it states the provider's privacy duties: do not log or sell queries, do not try to identify callers
      And it states that Brave and SearXNG providers answer for those engines' terms
      And it explains how to price per 1,000 queries and how payouts work
      And it lists `roger search check` as the first step for a native provider

    Scenario: The guide's contract and the checker cannot drift
      Then every rule the guide lists as part of search/1 has a check of the same name in roger search check
