# Search API: the provider's node. @security
#
# APPROVED by the founder 2026-11-15. A provider's `roger share` node takes a search job from the broker,
# asks its engine, turns the engine's reply into the normalized result list and returns it with
# a signed receipt. Founder ruling 2026-11-15: both built-in adapters (Brave, SearXNG) AND a
# native backend that speaks our contract directly (native_provider.feature).
#
# Ground truth: internal/agent/agent.go serve() (cleanUpstreamPath, isAllowedUpstreamPath,
# header discipline), features/voice/relay_allowlist.feature (the node is the trust boundary),
# internal/harness/search.go braveSearch (the Brave call, no retry on 429, body bound).
#
# ENGINE MAPPINGS
#   brave    GET https://api.search.brave.com/res/v1/web/search (or an owner override)
#            q, count, freshness = pd | pw | pm | py, country. Key in X-Subscription-Token.
#            web.results[]: title, url, description -> snippet, page_age -> published.
#   searxng  GET <upstream>/search?q=..&format=json[&time_range=day|week|month|year]
#            results[]: title, url, content -> snippet, publishedDate -> published.
#            SearXNG has no count or country parameter: the node trims to count, country is
#            not sent (the station listing says the station ignores country).
#   native   POST <upstream>/v1/search with the normalized request; the reply is already
#            normalized. Optional key as "Authorization: Bearer <key>".
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   N1  "/v1/search" joins the node's upstream path allowlist. A search job is never forwarded
#       verbatim: the node calls its engine through the adapter for the job's model.
#   N2  the engine key never leaves the node: not in registration, heartbeat, receipt, result,
#       error body or log line.
#   N3  one engine call per job, bounded at 10 seconds and 512 KiB. No retry, ever: a 429 must
#       not be answered by hammering a rate-limited engine.
#   N4  the node normalizes, but the broker re-validates everything (result_validation.feature).
#       The node is a convenience, not a trust boundary for the caller.

@search @node @security
Feature: A share node answers search jobs through its engine and never leaks its key

  Rule: the path allowlist (N1)

    Scenario: A search job is accepted by a node serving a search offer
      Given a share node serving "web-search" with engine "searxng"
      When the broker relays a job with path "/v1/search" for model "web-search"
      Then the node calls its SearXNG engine once
      And the node returns a normalized result list with a signed receipt

    Scenario Outline: Cosmetic variants of the search path are normalized
      Given a share node serving "web-search" with engine "searxng"
      When the broker relays a job with path "<path>" for model "web-search"
      Then the node calls its SearXNG engine once

      Examples:
        | path         |
        | /v1/search/  |
        | /v1//search  |

    Scenario Outline: Lookalike paths are refused without touching the engine
      Given a share node serving "web-search" with engine "searxng"
      When the broker relays a job with path "<path>"
      Then the node refuses with a 404 "unsupported path"
      And the engine is never called

      Examples:
        | path                  |
        | /search               |
        | /v1/Search            |
        | /v1/search/../admin   |
        | /v1/search/extra      |
        | /v1/searchx           |
        | /v2/search            |

    Scenario: A search job for a model this node does not offer as search is refused
      Given a share node serving chat model "llama3.2" and search "web-search"
      When the broker relays a job with path "/v1/search" for model "llama3.2"
      Then the node refuses the job without calling any backend

    Scenario: A chat job is never sent to a search engine
      Given a share node serving chat model "llama3.2" and search "web-search"
      When the broker relays a chat job for model "web-search"
      Then the search engine is never called

    Scenario Outline: A streaming job on a search share is never piped to the engine
      # Regression (pre-push audit 2026-11-21): a job whose body said "stream": true skipped
      # the path guard and was posted to the engine's address with the key as a bearer token.
      Given a share node serving "web-search" with engine "native"
      When the broker relays a streaming job with path "<path>"
      Then the node refuses with a 404 "unsupported path"
      And the engine is never called

      Examples:
        | path                 |
        | /v1/chat/completions |
        | /v1/audio/speech     |
        |                      |

    Scenario: A search job marked as a stream is served as one ordinary search
      Given a share node serving "web-search" with engine "native"
      When the broker relays a streaming job with path "/v1/search"
      Then the node calls its engine once with the normalized search and no "stream" field
      And the node returns a normalized result list with a signed receipt

    Scenario: Existing chat and voice paths are byte-identical after the change
      Given a share node serving chat, tts and stt models
      When the broker relays one job of each
      Then each is forwarded exactly as before this change

  Rule: the Brave adapter

    Scenario: A Brave search sends the mapped parameters and the key header
      Given a share node serving "web-search" with engine "brave" against a recording upstream
      When the broker relays a search for "tides" with count 7, freshness "week" and country "DE"
      Then the upstream receives GET with q "tides", count 7, freshness "pw" and country "DE"
      And the upstream receives the key in header "X-Subscription-Token"

    Scenario Outline: Freshness maps to Brave's codes
      Given a share node serving "web-search" with engine "brave" against a recording upstream
      When the broker relays a search with freshness "<freshness>"
      Then the upstream receives freshness "<code>"

      Examples:
        | freshness | code     |
        | day       | pd       |
        | week      | pw       |
        | month     | pm       |
        | year      | py       |
        | (absent)  | (absent) |

    Scenario: A Brave reply is mapped to the normalized fields
      Given a Brave upstream that returns web.results with title, url, description and page_age
      When the node serves a search
      Then each result's snippet is the description and published is the page_age

    Scenario: A Brave reply with no web section is zero results, not an error
      Given a Brave upstream that returns {"query": {...}} and no "web" key
      When the node serves a search
      Then the node returns {"results": []}

  Rule: the SearXNG adapter

    Scenario: A SearXNG search asks for JSON and the mapped time range
      Given a share node serving "web-search" with engine "searxng" against a recording upstream
      When the broker relays a search for "tides" with freshness "month"
      Then the upstream receives GET /search with q "tides", format "json" and time_range "month"

    Scenario: SearXNG results are trimmed to the requested count by the node
      Given a SearXNG upstream that returns 30 results
      When the broker relays a search with count 5
      Then the node returns 5 results in the upstream's order

    Scenario: Country is not sent to SearXNG
      Given a share node serving "web-search" with engine "searxng" against a recording upstream
      When the broker relays a search with country "DE"
      Then the upstream request carries no country or language parameter

    Scenario: A SearXNG with JSON output turned off is reported plainly
      Given a SearXNG upstream that answers 403 to format=json
      When the node serves a search
      Then the node returns an error saying the engine's JSON output is turned off

  Rule: the native adapter

    Scenario: A native search posts the normalized request as JSON
      Given a share node serving "go-docs" with engine "native" against a recording upstream
      When the broker relays a search for "generics" with count 3
      Then the upstream receives POST /v1/search with {"query":"generics","count":3}

    Scenario: A native backend's key is sent as a bearer token
      Given a native backend configured with a key
      When the node serves a search
      Then the upstream receives header "Authorization: Bearer <key>"

    Scenario: A native backend with no key gets no Authorization header
      Given a native backend configured without a key
      When the node serves a search
      Then the upstream request has no Authorization header

  Rule: the key never leaves the node (N2)

    Scenario Outline: The engine key appears nowhere outside the engine call
      Given a share node serving engine "<engine>" with key "sk-test-5e3c"
      When the node registers, heartbeats and serves a search that <outcome>
      Then neither the registration, heartbeat, receipt, result, error body nor any log line contains "sk-test-5e3c"

      Examples:
        | engine  | outcome                          |
        | brave   | succeeds                         |
        | brave   | fails with HTTP 401              |
        | brave   | fails with HTTP 422              |
        | brave   | times out                        |
        | native  | succeeds                         |
        | native  | fails with an error echoing the key |

    Scenario: An upstream error body is not passed back raw
      Given an upstream that fails with HTTP 500 and a body holding its own stack trace
      When the node serves a search
      Then the node's error carries the status and a short fixed reason, not the upstream body

  Rule: one bounded call, no retry (N3)

    Scenario Outline: Upstream failures map to station errors without retry
      Given an upstream that <behaves>
      When the node serves a search
      Then the upstream was called exactly once
      And the node returns an error whose reason is "<reason>"

      Examples:
        | behaves                         | reason                       |
        | answers HTTP 429                | engine rate limited          |
        | answers HTTP 401                | engine refused the key       |
        | answers HTTP 500                | engine error (status 500)    |
        | does not answer within 10s      | engine timed out             |
        | returns a 600 KiB body          | engine reply too large       |
        | returns a body that is not JSON | engine reply unreadable      |
        | refuses the connection          | engine unreachable           |

    Scenario: The node's engine timeout is under the broker's relay wait
      Then the node's engine timeout is shorter than the broker's non-stream relay wait

    Scenario: An engine redirect is not followed to another host
      Given an upstream that answers 302 to a different host
      When the node serves a search
      Then the node does not contact the other host
      And the node returns an error whose reason is "engine error (status 302)"

  Rule: the node normalizes, the broker still checks (N4)

    Scenario: The node drops non-http results before replying
      Given an upstream that returns a "javascript:" result and two good ones
      When the node serves a search
      Then the node's reply holds the two good results

    Scenario: A node running modified code still cannot get bad results past the broker
      Given a node that returns "javascript:" results unchanged
      When a consumer searches through it
      Then the broker drops them (result_validation.feature)
