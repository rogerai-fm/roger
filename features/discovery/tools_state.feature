# Approved by the founder 2026-11-24. Executable: cmd/rogerai-broker/tools_state_bdd_test.go.
# Build corrections (2026-11-24, approved by the founder): attribute_sources is on /discover,
# not /market; the "filters" scenario uses the quantization filter (price is not a named
# no_match filter); the regression section at the end was added after review.
#
# The tools mark a client reads is the one routing obeys, and a tools request that no station
# can take says so in a field, not only in a sentence.
#
# Source of the ask: the iOS app's request for model information ("publish one
# tools state per model that matches what routing will do: verified, passthrough or none; run
# the tools check on gpt-oss-120b, qwen-3.8-27b, mistral-large-2512, gemini-3.7-flash").
#
# GROUND TRUTH (origin/main 7cf851ce, and a /market read on 2026-11-24):
#   - There is ONE tools bit per (station, model): toolsVerifiedForLocked (toolcall.go:494).
#     The tool-call test sets it; a station cannot declare it (a declared "tools" is stripped
#     at register). It lapses after toolsVerifiedTTL (90 minutes) without a fresh pass.
#   - Listing and routing already read that same bit:
#       routing: pickFor (tunnel.go:5964), the Tower edge plan (toweredge.go:998), the route
#                explain dry run (explain.go:256);
#       listing: /market capabilities (market.go:158), /discover capabilities and
#                attribute_sources.tools "verified" (market.go:604, :287), /v1/models supported
#                parameters (models.go:223, :235).
#   - Hosted (curated) stations earn the bit by the same test as everyone else. There is no
#     "passthrough" case: no station carries tools without passing the test.
#   - On 2026-11-24 every general chat model carried "tools", including the four the
#     request named. The one-job Wave models carry none.
#   - A tools request that no station can take gets 503 no_match, and the sentence ends
#     "with the tools capability". metadata carries only "filters" (when a filter emptied the
#     pool). A client must read the sentence to learn tools were the reason. The iOS app
#     cannot tell, so it retries without tools and the user loses tools silently.
#
# DECISIONS (each is a ruling to confirm):
#   T1. No "passthrough" state. Every tools mark is a test result. "tools" in capabilities
#       means routing will send a tools request; its absence means it will not.
#   T2. No new per-model tools key. capabilities already equals what routing does (pinned by
#       the parity scenarios below). An "untested" vs "failed" split was considered and
#       dropped: a client acts the same on both (do not send tools).
#   T3. A no_match caused by a capability carries metadata.missing, a list from the closed
#       set ["tools", "vision"], in that order. The code stays no_match and the sentence does
#       not change. "filters" still appears beside it when a filter also applied. missing
#       appears only when a station in the caller's scope is on air for the model and lacks
#       the capability; a model that is off air in that scope gets no missing key.
#   T4. The re-test of the four named models is already done (they carry the mark live). No
#       work is scheduled for them.
#   T5. App side, not this spec: on a missing ["tools"] the app should tell the person tools
#       are off for that model before it retries without them.

Feature: The tools mark matches routing, and a refusal names the missing capability

  Background:
    Given a broker with an empty in-memory node registry
    And "u-1" is a funded consumer

  # --- one bit, every surface -------------------------------------------------------------

  Scenario Outline: A model whose station passed the tool test shows tools on every listing
    Given station "s1" is on air for "m" and passed the tool-call test
    When a consumer reads <surface>
    Then "m" shows tools there

    Examples:
      | surface                                              |
      | /market capabilities                                 |
      | /discover attribute_sources.tools as "verified"      |
      | /discover capabilities for "s1"                      |
      | /v1/models supported_parameters includes "tools"     |

  Scenario Outline: A model with no passing station shows tools on no listing
    Given station "s1" is on air for "m" and has not passed the tool-call test
    When a consumer reads <surface>
    Then "m" shows no tools there

    Examples:
      | surface                                              |
      | /market capabilities                                 |
      | /discover attribute_sources for "s1"                 |
      | /discover capabilities for "s1"                      |
      | /v1/models supported_parameters                      |

  Scenario Outline: When a listing shows tools, a tools request is sent; when it does not, it is refused
    Given station "s1" is on air for "m" and <state>
    When "u-1" relays a request for "m" with one function tool on <path>
    Then the outcome is <outcome>

    Examples:
      | state                         | path                   | outcome                 |
      | passed the tool-call test     | the direct relay       | dispatched to "s1"      |
      | passed the tool-call test     | the Tower edge plan    | dispatched to "s1"      |
      | passed the tool-call test     | the route explain      | would dispatch to "s1"  |
      | has not passed the test       | the direct relay       | 503 no_match            |
      | has not passed the test       | the Tower edge plan    | 503 no_match            |
      | has not passed the test       | the route explain      | would answer no_match   |

  Scenario: The mark and routing lapse at the same moment
    Given station "s1" passed the tool-call test for "m" 89 minutes ago and has not been re-tested
    When 2 minutes pass
    Then /market shows no tools for "m"
    And a tools request for "m" is refused with 503 no_match

  Scenario: A hosted station earns the mark only by the test
    Given a curated station is on air for "m" and declares "tools" at register
    And it has not passed the tool-call test
    Then /market shows no tools for "m"
    And a tools request for "m" is refused with 503 no_match

  Scenario: A hosted station that passed the test shows the same mark as any station
    Given a curated station is on air for "m" and passed the tool-call test
    Then /discover attribute_sources.tools for "m" is "verified"

  Scenario: No passthrough value exists on any feed
    Given stations of every kind are on air, some with the mark and some without
    When a consumer reads /market, /discover and /v1/models
    Then no feed carries the value "passthrough" anywhere

  Scenario: A private station's mark does not reach a public listing
    Given only a private station is on air for "m" and it passed the tool-call test
    When a public consumer reads /market
    Then "m" shows no tools there

  # --- the refusal names what is missing ---------------------------------------------------

  Scenario Outline: A capability no_match carries metadata.missing
    Given station "s1" is on air for "m" with <has>
    When "u-1" relays a request for "m" that needs <needs>
    Then the response is 503 with error code "no_match"
    And error.metadata.missing is <missing>

    Examples:
      | has                          | needs              | missing               |
      | no tools and no vision       | tools              | ["tools"]             |
      | no tools and no vision       | vision             | ["vision"]            |
      | no tools and no vision       | tools and vision   | ["tools", "vision"]   |
      | tools only                   | tools and vision   | ["vision"]            |
      | vision only                  | tools and vision   | ["tools"]             |

  Scenario: missing lists each capability some station in scope lacks, even when another has it
    Given station "s1" is on air for "m" with tools and no vision
    And station "s2" is on air for "m" with vision and no tools
    When "u-1" relays a request for "m" that needs tools and vision
    Then error.metadata.missing is ["tools", "vision"]

  Scenario: The sentence does not change
    Given station "s1" is on air for "m" with no tools
    When "u-1" relays a request for "m" with one function tool
    Then the error message ends "with the tools capability"

  # Corrected 2026-11-24 during the build: price is not a named no_match filter
  # (netfilters.go noMatchFilters); the quantization allow-list is.
  Scenario: missing and filters appear together
    Given station "s1" is on air for "m" with no tools and quant "q4_k_m"
    When "u-1" relays a request for "m" with one function tool and provider.quantizations ["fp8"]
    Then error.metadata.missing is ["tools"]
    And error.metadata.filters is ["quantizations"]

  Scenario: A no_match with no capability need carries no missing key
    Given no station is on air for "m"
    When "u-1" relays a plain request for "m"
    Then the response is 503 with error code "no_match"
    And error.metadata has no "missing" key

  Scenario: A grant whose owner is not serving keeps grant_unavailable and no missing key
    Given a grant whose owner has no station on air
    When the grant holder relays a request with one function tool
    Then the error code is "grant_unavailable"
    And error.metadata has no "missing" key

  Scenario: A dry run answers the same missing list
    Given station "s1" is on air for "m" with no tools
    When "u-1" sends a route explain dry run for "m" with one function tool
    Then the dry answer names no_match with missing ["tools"]

  Scenario: A refusal for a missing capability costs nothing
    Given station "s1" is on air for "m" with no tools
    When "u-1" relays a request for "m" with one function tool
    Then the X-RogerAI-Cost header is "0"
    And "s1" is not struck and earns nothing

  # --- adversarial -------------------------------------------------------------------------

  Scenario: A station cannot put a tools mark on a listing by declaring it
    Given station "s1" registers "m" with capabilities ["tools"] and never passes the test
    Then no feed shows tools for "m"

  Scenario: A model that is off air in the caller's scope carries no missing key
    Given no station is on air for "m"
    When "u-1" relays a request for "m" with one function tool
    Then the response is 503 with error code "no_match"
    And error.metadata has no "missing" key

  Scenario: A private station does not make a public refusal say a capability is missing
    Given only a private station is on air for "m" and it has no tools
    When a public consumer relays a request for "m" with one function tool
    Then the response is 503 no_match
    And error.metadata has no "missing" key
    And the response names no station and no private window

  # --- regression (review 2026-11-24): who is in scope -------------------------------------

  Scenario Outline: A station outside the caller's scope does not make a capability missing
    Given station "s2" is on air for "m" with no tools and <state>
    When "u-1" relays a request for "m" with one function tool
    Then the response is 503 with error code "no_match"
    And error.metadata has no "missing" key

    Examples:
      | state                     |
      | has stopped heartbeating  |
      | is banned                 |
      | is excluded by the caller |

  Scenario: A grant whose owner is on air without tools names tools as missing
    Given a grant whose owner is on air for "m" without tools
    When the grant holder relays a request with one function tool
    Then the response is 503 with error code "no_match"
    And error.metadata.missing is ["tools"]

  # --- docs --------------------------------------------------------------------------------

  Scenario: The API reference says what the tools mark means
    Then the public API reference says "tools" in capabilities means routing will send a tools request
    And it says the mark is earned by a test, lapses without a re-test, and cannot be declared
    And it documents error.metadata.missing on no_match
