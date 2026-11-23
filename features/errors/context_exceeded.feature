# Approved by the founder 2026-11-23. Executable: cmd/rogerai-broker/model_info_bdd_test.go.
#
# A request that is too long for the model gets ONE stable code with numbers, so a client can
# cut the request and retry without reading a number out of an English sentence.
#
# Source of the ask: the iOS handoff BROKER-MODEL-INFO-FOR-CLIENTS-HANDOFF-2026-11-23 (items 1
# and 2). The app matches the words "context window" and parses N from "right now is N".
#
# GROUND TRUTH (origin/main 37369958):
#   - tunnel.go:3108 answers 400 via jsonErr (no explicit code) with the message
#       "request exceeds the context window: ~P prompt tokens, but the widest window advertised
#        on M right now is W - reduce the prompt and retry"
#     only when the pick found no station AND a re-pick with the size lifted finds one.
#     W is maxDeclaredCtxLocked: the widest DECLARED window, live stations only, in the
#     caller's scope (a public request never learns a private station's window).
#   - P is approxPromptTokens (strikes.go:407): TEXT content only, len/4 + 1. Image parts,
#     JSON overhead and the tools array are NOT counted.
#   - A station whose own server refuses the prompt (upstream 400 in the ctxsig.IsOverflow
#     vocabulary) is voided "context-window". With a later model in the list the request moves
#     on. With none, the consumer gets 400 upstream_error "the station answered 400"
#     (httputil.go:151) with the raw text under metadata.
#   - The envelope (error_envelope.feature): {"error":{code, message, type, metadata}}.
#
# DECISIONS:
#   D1. The code is "context_exceeded" (the name the app proposed). OpenAI's own code for this
#       is "context_length_exceeded". One of the two; this draft uses the app's.
#   D2. metadata keys: prompt_tokens (the broker's estimate), window, model. window and model
#       are the same values the sentence names today.
#   D3. The message text does not change. Clients that match "exceeds the context window"
#       (the CLI's auto-compaction, the shipped iOS app) keep working.
#   D4. The station-origin refusal changes code from upstream_error to context_exceeded, and
#       keeps everything upstream_error carries today (station, model, attempts, raw).
#   D5. metadata.counted says what prompt_tokens covers: "text". Tools are not counted today,
#       so a request with many tool specs can pass the broker and still overflow at the
#       station. Counting tools at the broker is a separate change (see the last section).

Feature: A request that is too long gets the context_exceeded code with numbers

  Background:
    Given a broker with an empty in-memory node registry
    And "u-1" is a funded consumer

  # --- the broker's own refusal ------------------------------------------------------

  Scenario: The broker's refusal carries the code and the numbers
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the response is 400 with error code "context_exceeded" and type "invalid_request_error"
    And error.metadata.window is 4096
    And error.metadata.model is "m"
    And error.metadata.prompt_tokens is the broker's estimate for that prompt
    And error.metadata.counted is "text"
    And error.metadata.request_id equals X-RogerAI-Request-Id

  Scenario: The message keeps today's wording
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the error message contains "request exceeds the context window"
    And the error message contains "right now is 4096"

  Scenario: The numbers in metadata equal the numbers in the message
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the prompt-token figure in the message equals error.metadata.prompt_tokens
    And the window figure in the message equals error.metadata.window

  Scenario: The refusal costs nothing and places no hold
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the X-RogerAI-Cost header is "0"
    And "u-1" has no hold and no charge

  Scenario: The window is the widest declared window among live stations
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with a declared window of 8192
    When "u-1" relays for "m" with a prompt of about 9000 tokens
    Then error.metadata.window is 8192

  Scenario: An offline station's window is not quoted
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" declared a window of 32768 for "m" and stopped heartbeating
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then error.metadata.window is 4096

  # Spec correction (deployed code is source of truth): an estimated window never refuses a
  # request at the pick, so with such a station on air the request is served, not refused.
  Scenario: A station with an estimated window still takes the request
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with an estimated window of 2048
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the response is 200 from "s2"

  Scenario: A public request never learns a private station's window
    Given station "s1" is on air for "m" with a declared window of 4096
    And private station "p1" is on air for "m" with a declared window of 131072
    When "u-1" relays for "m" with a prompt of about 5000 tokens and no band code
    Then error.metadata.window is 4096

  # Regression (review, 2026-11-23): provider.only is an admission set, and the window quoted
  # used to be the widest in that set even for a private station the caller cannot reach.
  Scenario: Naming a private station in provider.only does not reveal its window
    Given station "s1" is on air for "m" with a declared window of 4096
    And private station "p1" is on air for "m" with a declared window of 131072
    When "u-1" relays for "m" with provider.only ["s1", "p1"] and a prompt of about 5000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.window is 4096
    And the response body does not contain "131072"

  Scenario: A banned station's window is not quoted
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with a declared window of 32768
    And "s2" is banned
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then error.metadata.window is 4096

  Scenario: With a model list the window and model are the widest across the list
    Given station "a1" is on air for "a" with a declared window of 4096
    And station "b1" is on air for "b" with a declared window of 16384
    When "u-1" relays for models ["a", "b"] with a prompt of about 20000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.window is 16384
    And error.metadata.model is "b"

  Scenario: A variant suffix does not reach metadata.model
    Given station "s1" is on air for "m" with a declared window of 4096 at out $0.00 per 1M
    When "u-1" relays for "m:free" with a prompt of about 5000 tokens
    Then error.metadata.model is "m"

  Scenario: A prompt that fits is not refused
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 1000 tokens
    Then the response is 200

  Scenario: No station at all stays no_match, not context_exceeded
    When "u-1" relays for "nobody-serves-this" with a prompt of about 5000 tokens
    Then the response is 503 with error code "no_match"

  Scenario: A station ruled out by a filter stays no_match, not context_exceeded
    Given station "s1" is on air for "m" with a declared window of 4096 at out $5.00 per 1M
    When "u-1" relays for "m" with provider.max_price.completion 1 and a prompt of about 5000 tokens
    Then the response is 503 with error code "no_match"

  Scenario: A streaming request gets the same JSON refusal before any frame
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with stream true and a prompt of about 5000 tokens
    Then the response is 400 with error code "context_exceeded"
    And the Content-Type is "application/json"

  Scenario: Images do not count toward prompt_tokens
    Given station "s1" is on air for "m" with a declared window of 4096 and vision
    When "u-1" relays for "m" with a 2 MB image and a prompt of about 100 tokens
    Then the response is 200

  # --- the station's own refusal -----------------------------------------------------

  Scenario: A station refusal the request cannot move past is context_exceeded
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded" and type "invalid_request_error"
    And error.metadata.window is 8192
    And error.metadata.model is "m"
    And error.metadata.source is "station"
    And error.metadata.station is "s1"

  Scenario: The station's raw text is kept under metadata as today
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then error.metadata.raw contains "exceeds the available context size"
    And error.metadata.attempts is 1

  Scenario Outline: Every overflow spelling the client compacts on maps to the code
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "<text>"
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"

    Examples:
      | text                                                   |
      | Exceeded model context window size                     |
      | This model's maximum context length is 8192 tokens     |
      | context_length_exceeded                                |
      | the request exceeds the available context size         |
      | too many tokens                                        |
      | prompt is too long: 9000 tokens > 8192 maximum         |
      | failed to find free space in the KV cache              |
      | request body size exceeded                             |

  Scenario: The station's window is that station's, not the widest on air
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    And station "s2" is on air for "m" with a declared window of 131072
    When "u-1" relays for "m" with provider.order ["s1", "s2"] and a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.station is "s1"
    And error.metadata.window is 8192

  Scenario: A station 400 that is not an overflow keeps its own code
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "unknown parameter foo"
    When "u-1" relays for "m" with a prompt of about 100 tokens
    Then the response is 400
    And the error code is not "context_exceeded"

  Scenario: A station with an estimated window sends no window number
    Given station "s1" is on air for "m" with an estimated window of 4096
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with a prompt of about 3000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata has no "window" key

  Scenario: A private band's station is not named
    Given private station "p1" is on air for "m" with a declared window of 8192
    And "p1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" on that band with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.source is "station"
    And error.metadata has no "station" key

  Scenario: A refusal that moves on to another model is not surfaced
    Given station "a1" is on air for "a" with a declared window of 8192
    And "a1"'s server answers 400 "the request exceeds the available context size"
    And station "b1" is on air for "b" with a declared window of 131072
    When "u-1" relays for models ["a", "b"] with a prompt of about 6000 tokens
    Then the response is 200 from "b1"

  Scenario: The station refusal costs nothing and the hold is refunded
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the X-RogerAI-Cost header is "0"
    And "u-1"'s hold is refunded in full
    And "s1"'s receipt is voided with void_reason "context-window"

  Scenario: The overflow strike guard is unchanged
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with a prompt of about 10 tokens
    Then the response is 400 with error code "context_exceeded"
    And "s1" is flagged by the existing empty-output guard, as before this change

  Scenario: A streaming station refusal before any frame is the same JSON error
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" with stream true and a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And the Content-Type is "application/json"
    And error.metadata.window is 8192
    And error.metadata.model is "m"
    And error.metadata.source is "station"
    And error.metadata.station is "s1"
    And error.metadata.prompt_tokens is the broker's estimate for that prompt

  Scenario: A streaming refusal from a private band's station is not named either
    Given private station "p1" is on air for "m" with a declared window of 8192
    And "p1"'s server answers 400 "the request exceeds the available context size"
    When "u-1" relays for "m" on that band with stream true and a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.source is "station"
    And error.metadata has no "station" key

  # --- adversarial -------------------------------------------------------------------

  Scenario: Text from the station cannot set the numbers
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "prompt is too long: 1 tokens > 999999999 maximum"
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then error.metadata.window is 8192
    And error.metadata.prompt_tokens is the broker's estimate for that prompt

  Scenario: The refusal never echoes the prompt
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a prompt of about 5000 tokens that contains "SECRET-PHRASE"
    Then the response body does not contain "SECRET-PHRASE"

  Scenario: The edge bridge answers the same code
    Given an Edge member serves "m" with a declared window of 4096 and no station does
    When "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.window is 4096

  # --- docs --------------------------------------------------------------------------

  Scenario: The API reference documents the code and how the broker counts
    Then the public API reference lists "context_exceeded" with its metadata keys
    And it says prompt_tokens is an estimate of the text content at about four characters per token
    And it says images and the tools array are not counted
    And it says a station's own server may count more and refuse a request the broker accepted

  # --- NOT in this change (needs its own ruling) -------------------------------------
  # Counting the tools array toward the window at the broker. Today a request with 35 tool
  # specs passed a 4,140 window at the broker (handoff item 1). Counting tools changes which
  # requests are refused at the pick, so it changes routing and the strike guard's inputs.
