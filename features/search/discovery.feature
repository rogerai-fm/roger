# Search API: how search stations are listed in the marketplace.
#
# APPROVED by the founder 2026-11-15. A search station is not a model: it has no context window and no
# tokens per second. It gets its own public listing, shaped like /voices, and rides the
# existing discovery and category surfaces with its modality and engine visible.
#
# Ground truth: cmd/rogerai-broker/main.go routes (/voices, /discover, /market, /categories,
# /v1/models), features/voice/discovery.feature (metadata only, never a node address).
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   L1  GET /search/stations is public and anonymous. Each entry: id (namespaced
#       "@station/model"), model, station label, engine, scope, price_per_1k_queries (USD at
#       credit = $1), free, latency_ms (median of recent searches, null until there are 10),
#       ignores (the request fields the engine ignores, e.g. ["country"] for SearXNG).
#   L2  it never exposes a node hostname, bridge URL, IP or key.
#   L3  /discover carries modality "search", unit "query", engine and scope on search offers.
#       FOUNDER RULING 2026-11-23: search stations ride /discover and /market for every
#       caller, the same way voice stations do. Each client sets them aside by modality.
#       A client released before search existed shows one as a chat band until it is upgraded.
#   L4  /v1/models is an OpenAI model catalog and does NOT list search offers.
#   L5  /categories lists subcategory "search" under "retrieval" with its live station count.

@search @discovery
Feature: Search stations are listed for people and agents, with no node addresses

  Rule: GET /search/stations (L1, L2)

    Scenario: The listing is anonymous
      Given a search station "@acme" offering "web-search" with engine "brave" at 6000 credits per 1,000,000 queries
      When an anonymous GET /search/stations arrives
      Then the response status is 200
      And it lists "@acme/web-search" with engine "brave", price_per_1k_queries 6 and free false

    Scenario: A free station is marked free
      Given a search station "@zed" offering "web-search" with engine "searxng" at 0 credits per 1,000,000 queries
      When an anonymous GET /search/stations arrives
      Then "@zed/web-search" has free true and price_per_1k_queries 0

    Scenario: SearXNG stations say they ignore country
      Given a search station with engine "searxng"
      When an anonymous GET /search/stations arrives
      Then that station's ignores is ["country"]

    Scenario: Brave and native stations ignore nothing by default
      Given a search station with engine "brave" and one with engine "native"
      When an anonymous GET /search/stations arrives
      Then both have ignores []

    Scenario: latency_ms is null until there are 10 searches, then the median
      Given a new search station
      When an anonymous GET /search/stations arrives
      Then its latency_ms is null
      When the station has served 10 searches taking 100 to 1000 ms
      Then its latency_ms is the median of those 10

    Scenario: Only search stations are listed
      Given a chat node, a tts node and a search station on air
      When an anonymous GET /search/stations arrives
      Then only the search station is listed

    Scenario: An off-air or banned station is not listed
      Given a search station that went off air and one whose node is banned
      When an anonymous GET /search/stations arrives
      Then neither is listed

    Scenario: A private-band search station is not listed publicly
      Given a search station shared only on a private band
      When an anonymous GET /search/stations arrives
      Then it is not listed

    Scenario: The listing exposes no node address or key
      Given a search station whose bridge URL is "https://abc123.trycloudflare.com", local address "192.168.1.50:8888" and engine key "sk-test-5e3c"
      When an anonymous GET /search/stations arrives
      Then the body contains none of those values
      And no field exposes a node hostname or IP

    Scenario: The listing is cheap to serve and rate limited like other public listings
      When one IP sends GET /search/stations above the public listing limit
      Then the excess is refused with 429 and a Retry-After header

    Scenario: Only GET is accepted
      When a POST /search/stations arrives
      Then the response status is 405

  Rule: the shared discovery surfaces (L3, L4, L5)

    Scenario: /discover carries the search fields
      Given a search station "web-search" with engine "native" and scope "the open web"
      When GET /discover arrives
      Then that offer has modality "search", unit "query", engine "native" and scope "the open web"

    Scenario: /v1/models does not list search offers
      Given a chat node offering "llama3.2" and a search station "web-search"
      When GET /v1/models arrives
      Then "llama3.2" is listed and "web-search" is not

    Scenario: /v1/models/<search id> is 404
      Given a search station "web-search"
      When GET /v1/models/web-search arrives
      Then the response status is 404

    Scenario: /categories counts search supply
      Given two search stations on air
      When GET /categories arrives
      Then category "retrieval" lists subcategory "search"
      And category "retrieval" counts the 2 search stations among its providers

  Rule: providers see their search stations like any other

    Scenario: The owner's /stations shows a search station with its usage
      Given an owner whose node serves "web-search"
      When the owner reads GET /stations
      Then "web-search" is listed with modality "search", engine, status and query count
      And the engine key is not in the response
