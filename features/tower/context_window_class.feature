# Approved by the founder 2026-11-23 (D1 to D6 as written, D4 accepted: an old broker reads the new class as a generic failure).
# Founder rulings of 2026-11-23, after the build: the anonymous scenario was corrected, D4 and D5 were
# reworded, and the four readings D7 to D10 were confirmed and given scenarios.
# BUILD STATUS: BUILT. internal/station/execute.go and edge_sealed.go (the class and the operator's line),
# cmd/rogerai/relayfabric.go (the operator's terminal), the broker's relay (its reading of the class),
# internal/localplane and internal/agent/localtower.go and localtower_stream.go (the self-hosted plane),
# docs/tower.md. Executable: internal/station/context_window_class_bdd_test.go,
# internal/agent/context_window_class_bdd_test.go, and the broker's own runner for the broker scenarios.
#
# A station behind a Tower that refuses a prompt as too long is answered with the same
# context_exceeded code a direct station gets (features/errors/context_exceeded.feature).
# The founder ruled yes on 2026-11-23. This file is the how.
#
# GROUND TRUTH (origin/main df8bda4b plus PR #21):
#   - A Tower is blind. When the model behind a Station replies non-2xx, only a failure CLASS
#     crosses it, never the model's words (internal/station/execute.go, UpstreamStatusError.Class):
#       "the model did not answer (status 429, retry-after 6)"
#       "the model did not answer (status 400)"
#   - The broker reads the two numbers from the class and answers the consumer "the station
#     behind the tower replied 400". A class it cannot read is a generic failure.
#   - So on this path the broker cannot tell "too long" from any other 400.
#   - The Station does see the model's words locally (UpstreamStatusError.Detail, kept for the
#     operator's log).
#   - internal/ctxsig.IsOverflow is the one list of overflow spellings; the direct path and
#     the client's auto-compaction both use it.
#
# DECISIONS IN THIS DRAFT (each is a ruling to confirm):
#   D1. The Station classifies, because only it sees the text. When the model replies 400 and
#       ctxsig.IsOverflow matches the body, the class gains one fixed word:
#         "the model did not answer (status 400, context-window)"
#       Still no model words cross the Tower. One bit does.
#   D2. The broker treats that class as void reason "context-window", exactly as on the
#       direct path: with a later model in the list the request moves on; with none the
#       consumer gets 400 context_exceeded with source "station", the broker's own
#       prompt_tokens, and window = the declared window on the Tower's routable row.
#   D3. Order of release. The broker learns the new class first (a deploy). Stations gain it
#       with the next roger release. Until a Station updates, it sends the old class and the
#       consumer gets today's generic Tower failure. Nothing breaks in either order except
#       D4.
#   D4. An OLD broker reading the NEW class cannot read it, and answers a generic failure. Accepted. The word stays inside the brackets. The
#       self-hosted plane is not affected: it never used the class channel before this change.
#   D5. Trust. A Tower operator is a third party and could send the class falsely to push a
#       request to another model. The broker applies the same plausibility guard as the direct
#       path. A claim it does not believe is a Station fault on the Tower's record.
#   D6. Only status 400 can carry the word. Any other status with it is read as that status
#       alone.
#   D7. A believed claim is a fact about the prompt: it records no Station fault. A claim that
#       is not believed records one.
#   D8. The refusal on the Tower path never names the Tower, for any caller.
#       Accepted residual (ruled 2026-11-23): on a stream that was already committed when the
#       refusal arrives, the Tower id is visible in the X-RogerAI-Relay header and the usage
#       record's node, as on every successful Tower stream. The error object never names it.
#   D9. A claim that is not believed and has no later model to move to is still answered
#       context_exceeded.
#   D10. The self-hosted plane sends the code without numbers: it counts no tokens and knows
#       no window.

Feature: A Tower station's too-long refusal crosses as a class and is answered context_exceeded

  # --- the Station ---------------------------------------------------------------------

  Scenario Outline: The Station names the class for every overflow spelling
    Given a Station whose model replies 400 with the body "<text>"
    When the Station serves a sealed attempt
    Then the failure that crosses the Tower is "the model did not answer (status 400, context-window)"

    Examples:
      | text                                               |
      | Exceeded model context window size                 |
      | This model's maximum context length is 8192 tokens |
      | context_length_exceeded                            |
      | the request exceeds the available context size     |
      | too many tokens                                    |
      | prompt is too long: 9000 tokens > 8192 maximum     |
      | failed to find free space in the KV cache          |
      | request body size exceeded                         |

  Scenario: The model's words still never cross the Tower
    Given a Station whose model replies 400 with the body "context length exceeded for SECRET-PROMPT-TEXT"
    When the Station serves a sealed attempt
    Then the failure that crosses the Tower does not contain "SECRET-PROMPT-TEXT"
    And the failure that crosses the Tower is "the model did not answer (status 400, context-window)"

  Scenario: A 400 that is not an overflow keeps the plain class
    Given a Station whose model replies 400 with the body "unknown parameter foo"
    When the Station serves a sealed attempt
    Then the failure that crosses the Tower is "the model did not answer (status 400)"

  Scenario Outline: Only a 400 carries the word
    Given a Station whose model replies <status> with the body "context length exceeded"
    When the Station serves a sealed attempt
    Then the failure that crosses the Tower is "the model did not answer (status <status>)"

    Examples:
      | status |
      | 413    |
      | 422    |
      | 500    |
      | 503    |

  Scenario: A 429 with a wait is unchanged
    Given a Station whose model replies 429 with Retry-After 6 and the body "too many tokens per minute"
    When the Station serves a sealed attempt
    Then the failure that crosses the Tower is "the model did not answer (status 429, retry-after 6)"

  Scenario: The operator's log keeps the model's words
    Given a Station whose model replies 400 with the body "the request exceeds the available context size"
    When the Station serves a sealed attempt
    Then the Station's own log line contains "exceeds the available context size"

  # --- the broker ----------------------------------------------------------------------

  Scenario: The broker answers context_exceeded for the class
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded" and type "invalid_request_error"
    And error.metadata.source is "station"
    And error.metadata.window is 8192
    And error.metadata.prompt_tokens is the broker's estimate for that prompt
    And the error message contains "exceeds the context window"

  # Corrected on the founder's ruling of 2026-11-23: an anonymous caller never reaches a Tower.
  Scenario: An anonymous caller is refused before any Tower is asked
    Given an approved Tower "t1" serves "m" free with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    When an anonymous caller relays for "m" with a prompt of about 6000 tokens
    Then the response is 403 with error code "login_required"
    And "t1" received no attempt

  Scenario: With a model list the request moves on to the next model
    Given an approved Tower "t1" serves "a" with a declared window of 8192
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And station "b1" is on air for "b" with a declared window of 131072
    And "u-1" is a funded consumer
    When "u-1" relays for models ["a", "b"] with a prompt of about 6000 tokens
    Then the response is 200 from "b1"
    And the "t1" attempt's void_reason is "context-window"

  Scenario: The refusal costs nothing
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the X-RogerAI-Cost header is "0"
    And "u-1"'s hold is refunded in full
    And the Tower's operator earns nothing for the attempt

  Scenario: A streaming request gets the same JSON refusal before any frame
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with stream true and a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And the Content-Type is "application/json"

  Scenario: A plain 400 class keeps today's answer
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400
    And the error code is not "context_exceeded"
    And the error message contains "the station behind the tower replied 400"

  Scenario: An old Station's class keeps today's answer
    Given an approved Tower "t1" whose Station predates the class and whose model replies 400 with "context length exceeded"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400
    And the error code is not "context_exceeded"

  # --- adversarial ---------------------------------------------------------------------

  # Renamed on the founders' ruling of 2026-11-23: the request still moves to the next model;
  # the guard records the false claim against the Tower.
  Scenario: A Tower that claims the class on a small prompt is recorded as faulty
    Given an approved Tower "t1" serves "a" with a declared window of 8192
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And station "b1" is on air for "b" with a declared window of 131072
    And "u-1" is a funded consumer
    When "u-1" relays for models ["a", "b"] with a prompt of about 10 tokens
    Then the response is 200 from "b1"
    And "t1" has a Station fault recorded for the attempt

  Scenario: A Tower that claims the class for everything loses its place
    Given an approved Tower "t1" whose Station claims the class on every attempt
    When the broker's canary probes "t1"
    Then the canary fails and "t1" is not routable

  Scenario Outline: A malformed class is read as a plain failure, never as the code
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "<class>"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the error code is not "context_exceeded"

    Examples:
      | class                                                              |
      | the model did not answer (status 200, context-window)              |
      | the model did not answer (status 500, context-window)              |
      | the model did not answer (status 400, context-window, extra)       |
      | the model did not answer (status 400, CONTEXT-WINDOW)              |
      | context-window                                                     |
      | the model did not answer (status 400) context-window               |

  Scenario: The class never sets the numbers
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then error.metadata.window is 8192
    And error.metadata.prompt_tokens is the broker's estimate for that prompt

  Scenario: A Tower with an estimated window sends no window number
    Given an approved Tower "t1" serves "m" with an estimated window of 4096 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 3000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata has no "window" key

  # --- the self-hosted plane -----------------------------------------------------------

  Scenario: A Tower-local plane answers its own consumers the same way
    Given a standalone Tower whose Station's model replies 400 with "the request exceeds the available context size"
    When a local consumer relays through the Tower's own plane
    Then the response is 400 with error code "context_exceeded"

  # --- docs ----------------------------------------------------------------------------

  Scenario: The reference documents the class
    Then the Tower operator documentation lists the failure classes that cross a Tower, including "context-window"
    And it says no text from the model crosses

  # --- confirmed readings (founder, 2026-11-23) ---

  Scenario: A believed claim records no Station fault
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And "t1" has no Station fault recorded for the attempt

  Scenario: The refusal never names the Tower for a signed-in caller
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata has no "station" key
    And error.metadata has no "model" key
    And error.metadata has no "raw" key
    And the response names neither "t1" nor the Station behind it

  Scenario: A claim that is not believed and has no later model is still answered context_exceeded
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with a prompt of about 10 tokens
    Then the response is 400 with error code "context_exceeded"
    And "t1" has a Station fault recorded for the attempt

  Scenario: The self-hosted plane sends the code without numbers
    Given a standalone Tower whose Station's model replies 400 with "the request exceeds the available context size"
    When a local consumer relays through the Tower's own plane
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.source is "station"
    And error.metadata has no "prompt_tokens" key
    And error.metadata has no "window" key

  # --- regression (review 2026-11-23) ---

  Scenario: A Tower refusal that arrives after the stream is committed carries the code in its error frame
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And the Station behind "t1" refuses only after the stream is committed
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with stream true and a prompt of about 6000 tokens
    Then the stream was committed before the refusal
    And the stream's error frame has error code "context_exceeded" and type "invalid_request_error"
    And the error frame's metadata has source "station", window 8192 and the broker's estimate for that prompt
    And the error frame names neither "t1" nor the Station behind it
    And the stream ends with a zero-cost usage chunk voided "context-window" and "[DONE]"

  Scenario: A direct station's refusal that arrives after the stream is committed carries the code too
    Given station "s1" is on air for "m" with a declared window of 8192
    And "s1"'s server answers 400 "the request exceeds the available context size" only after the stream is committed
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with stream true and a prompt of about 6000 tokens
    Then the stream was committed before the refusal
    And the stream's error frame has error code "context_exceeded" and type "invalid_request_error"
    And the error frame's metadata has source "station", window 8192 and the broker's estimate for that prompt
    And the stream ends with a zero-cost usage chunk voided "context-window" and "[DONE]"

  Scenario: A streamed refusal before any frame carries the request id
    Given an approved Tower "t1" serves "m" with a declared window of 8192 and no direct station does
    And the Station behind "t1" fails with "the model did not answer (status 400, context-window)"
    And "u-1" is a funded consumer
    When "u-1" relays for "m" with stream true and a prompt of about 6000 tokens
    Then the response is 400 with error code "context_exceeded"
    And error.metadata.request_id equals X-RogerAI-Request-Id
    And error.metadata.window is 8192
