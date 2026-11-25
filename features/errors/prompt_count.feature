# Approved by the founder 2026-11-24. Executable: cmd/rogerai-broker/prompt_count_bdd_test.go.
# Build corrections (2026-11-24, approved by the founder): the hold is sized from body bytes and
# is not a consumer of the count (GROUND TRUTH, D4, the "cost estimate" row); the default
# max_tokens is 1097, not 1096 (the "+1" floor); context_exceeded.feature's docs line now says
# images are not counted.
#
# How the broker counts a request against a model's window, written down, and counted closer
# to what the model really reads, so a request that cannot fit is refused at the broker with
# context_exceeded instead of failing at the station.
#
# Source of the ask: the iOS handoff on model information for clients, item 1 ("say how the
# broker counts a request against ctx, and whether tools are counted; decide which count is
# the truth").
#
# GROUND TRUTH (origin/main 7cf851ce):
#   - One measure, approxPromptTokens / promptTokensFrom (strikes.go:407-420), feeds the window
#     gate at pick time, the context_exceeded numbers, the overflow strike guard
#     (plausiblyOverflows), the cost estimate price ranks on (and the dry run's
#     cost_estimate.min), the default max_tokens, the speed fit and the per-request TPM share.
#     The HOLD is not one of them: it is sized from the body's bytes, an upper bound
#     (holdCostSized), and this change leaves it alone. (Corrected 2026-11-24 during the build.)
#   - It is (bytes of promptScan text) / 4 + 1. promptScan (bodydoc.go:271) collects:
#       * every message's string content, and the text of every text part;
#       * every string VALUE inside tools[].function and functions[] (names, descriptions,
#         nested parameter descriptions, enum values, and also the words "object", "string").
#     It does NOT collect: an assistant message's tool_calls (function name and arguments),
#     the legacy function_call, the property NAMES of a tool schema (they are object keys),
#     JSON punctuation, image parts, or the chat template the station's server adds.
#   - Correction to the record: context_exceeded.feature (D5) and the handoff both say tools
#     are not counted. The tools' string values ARE counted, since the tools-array screening
#     change. Their property names and all earlier tool calls are not.
#   - metadata.counted is "text" today (context_exceeded.feature, approved and shipped).
#   - The billing recount (features/money/recount_billing.feature) also reads promptText. It
#     bills min(station's claim, recount), so a higher recount can rise toward an honest claim
#     and never past it.
#   - Unparsable or non-chat bodies fall back to body bytes / 4 + 1.
#
# MEASURED 2026-11-24 (one open-weight chat model, its own chat template and tokenizer; a
# system line, a user turn, one earlier tool call and its result, plus N tool definitions):
#
#     tools | broker today | model's real count | real / broker
#       0   |       64     |        206         |     3.2
#      10   |      489     |        853         |     1.7
#      35   |     1564     |       2403         |     1.5
#
#   Counting the whole tools array as compact JSON OVER-counted on this model (0.65 to 0.73):
#   its template writes tools more tersely than JSON. The rule in D1 measured 1.29 to 1.35 on
#   all three rows, steady, and never above the real count.
#
# DECISIONS (each is a ruling to confirm):
#   D1. Counted from now on: everything counted today, PLUS
#         - each assistant tool_calls entry's function name and arguments,
#         - the legacy function_call name and arguments,
#         - the property names inside each tool's parameters schema,
#         - 4 tokens per message (role and separators).
#   D2. Still not counted: image parts (their cost differs by model), the template's fixed
#       preamble (differs by model), JSON punctuation and whitespace, tool_call ids. The
#       station's own server stays the final judge; its refusal already comes back as
#       context_exceeded with source "station" (PR 21, PR 28).
#   D3. The rate stays bytes / 4. The count is a gate and an estimate, never a billing number.
#       The rule must not count ABOVE what a model reads: an over-count refuses requests that
#       fit. Under-counting by about a third is accepted.
#   D4. One count everywhere: every consumer listed in GROUND TRUTH reads the new count. No
#       second measure is added. The hold is sized from body bytes and does not change. The
#       billing recount (money) reads the new screen text: it moves toward the station's
#       honest claim, never past it.
#   D5. metadata.counted stays "text". Its meaning is documented as "the request's text,
#       including tool definitions and earlier tool calls". No client parses a new value, and
#       the approved context_exceeded scenario is unchanged.
#   D6. The public API reference says how the count is made, that it is an estimate that
#       usually runs about a third under the model's own count, and that a client should keep
#       headroom.
#   D7. Moderation reads the same text (features/moderation/tool_calls_screening.feature, a
#       separate approval).
#   D8. A real tokenizer on the hot path is NOT used. The broker's tokenizer service exists for
#       the off-path recount; calling it per request adds latency to every relay, and most
#       models' tokenizers are not loaded in it.

Feature: The broker counts a request the way the docs say

  Background:
    Given a broker with an empty in-memory node registry
    And "u-1" is a funded consumer

  # --- what is counted --------------------------------------------------------------------

  Scenario Outline: Each piece of a request adds its bytes to the count
    Given a chat request with one user message "hi"
    When the request also carries <piece> of exactly 400 bytes
    Then the broker's count rises by 100 tokens, plus 4 when the piece adds a message

    Examples:
      | piece                                                               |
      | a second user message's string content                              |
      | a text part in a user message's content array                       |
      | a system message's content                                          |
      | a tool message's content                                            |
      | an assistant message's tool_calls[0].function.arguments             |
      | an assistant message's legacy function_call.arguments               |
      | a tool's function.description                                       |
      | a nested parameter description inside a tool's parameters schema    |
      | an enum value inside a tool's parameters schema                     |
      | a property name inside a tool's parameters schema                   |
      | a legacy functions[] entry's description                            |

  Scenario: An assistant tool call's function name is counted
    Given a chat request with one user message "hi"
    When the request also carries an assistant tool call named with 40 bytes and empty arguments
    Then the broker's count rises by 10 tokens and the per-message amount

  Scenario Outline: Each message adds 4 tokens on top of its text
    Given a chat request with <n> messages whose contents are all empty
    Then the broker's count is <count>

    Examples:
      | n  | count |
      | 1  | 5     |
      | 2  | 9     |
      | 10 | 41    |

  Scenario: Bytes are counted, not characters
    Given a chat request whose only message is 100 characters of a 3-byte script
    Then the broker's count is 300 / 4 + 1 + 4

  # --- what is not counted ----------------------------------------------------------------

  Scenario: An image part adds nothing
    Given a chat request with one user message "hi"
    When the request also carries a 2 MB base64 image part
    Then the broker's count does not change

  Scenario: Whitespace and punctuation in the JSON add nothing
    Given two requests with the same messages and tools, one compact and one pretty-printed
    Then the broker's count is equal for both

  Scenario: A tool call id adds nothing
    Given two requests that differ only in an assistant tool call's id length
    Then the broker's count is equal for both

  Scenario: The schema words of a tool definition are not counted twice
    Given a tool whose parameters schema has one property "query" of type "string"
    Then "query" is counted once and the keys "type", "properties" and "required" are not counted as names

  # --- the gate ---------------------------------------------------------------------------

  Scenario: A request whose count equals the window is sent
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a request the broker counts at exactly 4096 tokens
    Then the request is dispatched to "s1"

  Scenario: A request one token over the window is refused at the broker
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a request the broker counts at 4097 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.prompt_tokens is 4097
    And no station is dispatched to

  Scenario: Earlier tool calls can push a request past the window
    Given station "s1" is on air for "m" with a declared window of 4096
    And a request whose message text counts 3000 tokens
    When the request also carries earlier tool calls whose arguments count 1500 tokens
    Then the response is 400 with error code "context_exceeded"
    And no station is dispatched to

  Scenario: Tool property names can push a request past the window
    Given station "s1" is on air for "m" with a declared window of 4096
    And a request with 35 tools that counted under the window before this change
    And the same request counts over 4096 under D1
    Then the response is 400 with error code "context_exceeded"

  Scenario: metadata.counted keeps its value
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" with a request the broker counts at 5000 tokens
    Then error.metadata.counted is "text"

  # --- one count everywhere ---------------------------------------------------------------

  Scenario Outline: Every path uses the same count for the same body
    Given one request body with messages, earlier tool calls and 10 tools
    When it is measured on <path>
    Then the count equals the broker's count for the body

    Examples:
      | path                                     |
      | the direct relay's window gate           |
      | the Tower edge plan                      |
      | the route explain dry run                |
      | the streaming relay                      |
      | the overflow strike guard                |
      | the cost estimate                        |
      | the default max_tokens                   |
      | the per-request TPM share                |

  Scenario: The default max_tokens leaves room for the counted prompt
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" a body with no max_tokens that the broker counts at 3000 tokens
    Then the body sent to "s1" carries a max_tokens no larger than 1097
    # 4096 - 3000 + 1: the count's "+1" is a floor for an empty prompt, not a token, so the
    # default output leaves it out (costest.go defaultMaxTokensSet, unchanged).

  Scenario: The billing recount never bills above the station's claim
    Given a request with earlier tool calls that the broker now counts higher than before
    When the station claims fewer prompt tokens than the broker's recount
    Then the consumer is billed the station's claim

  Scenario: A station that refuses a prompt the broker counted under its window is not struck
    Given station "s1" is on air for "m" with a declared window of 4096
    When "u-1" relays for "m" a request the broker counts at 3000 tokens
    And "s1"'s server refuses it as too long
    Then the consumer gets 400 with error code "context_exceeded" and error.metadata.source "station"
    And "s1" is not struck

  # --- unchanged shapes -------------------------------------------------------------------

  Scenario: A legacy completions body is counted as before
    Given a body {"prompt": "..."} of 4000 bytes
    Then the broker's count is 1001

  Scenario: A body that is not JSON is counted as before
    Given a body of 4000 bytes that is not JSON
    Then the broker's count is 1001

  # --- adversarial ------------------------------------------------------------------------

  Scenario Outline: A malformed tool call is counted without failing the request
    Given a chat request whose assistant message carries tool_calls as <shape>
    When "u-1" relays it
    Then the broker does not fail with a 500
    And the malformed part adds at most its own bytes / 4 to the count

    Examples:
      | shape                                   |
      | a string                                |
      | an array of numbers                     |
      | an entry with no function               |
      | an entry whose arguments is an object   |
      | an entry whose arguments is null        |
      | an entry whose name is a number         |

  Scenario: A tool schema nested past the JSON depth limit does not crash the broker
    Given a tool whose parameters schema nests objects 20000 deep
    When "u-1" relays it
    Then the broker answers 400 and does not crash

  Scenario: One megabyte of tool-call arguments is counted in full
    Given an assistant tool call whose arguments are 1 MB of text
    Then the broker's count includes 262144 tokens for them

  Scenario: Counting adds no new parse of the body
    When a chat request with tools and earlier tool calls is relayed
    Then the body is decoded no more times than before this change

  # --- calibration ------------------------------------------------------------------------

  @live
  Scenario Outline: The count stays under a real model's count
    Given the reference request with <tools> tool definitions and one earlier tool call
    When a local chat model's own template and tokenizer count it
    Then the broker's count is at most the model's count
    And the broker's count is at least half the model's count

    Examples:
      | tools |
      | 0     |
      | 10    |
      | 35    |

  # --- docs -------------------------------------------------------------------------------

  Scenario: The API reference says how a request is counted
    Then the public API reference lists what the broker counts and what it does not
    And it says the count is an estimate that usually runs under the model's own count
    And it says a client should leave headroom below the window
    And it says error.metadata.counted "text" covers tool definitions and earlier tool calls
