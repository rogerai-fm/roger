# Approved by the founder 2026-11-23. Executable: cmd/rogerai-broker/model_info_bdd_test.go.
#
# GET /market says how large a request each model takes, so a client can cut a request to fit
# from the one feed it already reads. Today the window is only on /discover (per station) and
# on /v1/models (context_length, rogerai.ctx_max).
#
# Source of the ask: the iOS handoff BROKER-MODEL-INFO-FOR-CLIENTS-HANDOFF-2026-11-23 (item 3).
#
# GROUND TRUTH (origin/main 37369958):
#   - marketView (market.go:385) has no window field.
#   - /v1/models already publishes rogerai.ctx_max, "DECLARED windows only" (models.go:193).
#   - The context-window refusal quotes maxDeclaredCtxLocked: widest DECLARED window, live
#     public stations only.
#   - /market counts curated stations apart (curated_providers) but they are on the row.
#
# DECISIONS:
#   D1. The field is "ctx_max", the name /v1/models already uses. The app asked for "max_ctx".
#       One name on both feeds is less to remember.
#   D2. ctx_max is the widest DECLARED window among the row's online public stations, human
#       and curated. It is the same number the context_exceeded refusal reports.
#   D3. When every online station's window is an estimate, ctx_max is omitted and
#       "ctx_estimated": true is sent with "ctx_max_estimated" holding the widest estimate.
#       A client may use it to cut a request but the broker does not refuse on it.
#   D4. No max_output in this change. No station declares an output cap today, so there is
#       nothing true to publish. It needs a declared field first.

Feature: /market publishes each model's widest context window

  Background:
    Given a broker with an empty in-memory node registry

  Scenario: A model's row carries the declared window
    Given station "s1" is on air for "m" with a declared window of 16384
    When a consumer GETs /market
    Then the "m" row has ctx_max 16384

  Scenario: The widest declared window wins
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with a declared window of 16384
    When a consumer GETs /market
    Then the "m" row has ctx_max 16384

  Scenario: An offline station's window is not counted
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" declared a window of 131072 for "m" and stopped heartbeating
    When a consumer GETs /market
    Then the "m" row has ctx_max 4096

  Scenario: A private station's window is not counted
    Given station "s1" is on air for "m" with a declared window of 4096
    And private station "p1" is on air for "m" with a declared window of 131072
    When a consumer GETs /market
    Then the "m" row has ctx_max 4096

  Scenario: A curated station's window counts
    Given curated station "c1" is on air for "big/model" with a declared window of 1000000
    When a consumer GETs /market
    Then the "big/model" row has ctx_max 1000000

  Scenario: An estimated window does not beat a declared one
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with an estimated window of 32768
    When a consumer GETs /market
    Then the "m" row has ctx_max 4096
    And the "m" row has no "ctx_estimated" key

  Scenario: Only estimated windows are marked as estimates
    Given station "s1" is on air for "m" with an estimated window of 4096
    When a consumer GETs /market
    Then the "m" row has no "ctx_max" key
    And the "m" row has ctx_estimated true and ctx_max_estimated 4096

  Scenario: A voice model has no window
    Given station "v1" is on air for "voice" with modality "tts"
    When a consumer GETs /market
    Then the "voice" row has no "ctx_max" key
    And the "voice" row has no "ctx_estimated" key
    And the "voice" row has no "ctx_max_estimated" key

  # Regression (review, 2026-11-23): the chat check used the first offer seen for the model.
  Scenario: A voice station sharing a chat model's id does not hide or set the window
    Given station "s1" is on air for "m" with a declared window of 8192
    And station "v1" is on air for "m" with modality "tts"
    When a consumer GETs /market 12 times across cache expiries
    Then every time the "m" row has ctx_max 8192

  Scenario: A banned station's window is not counted
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with a declared window of 32768
    And "s2" is banned
    When a consumer GETs /market
    Then the "m" row has ctx_max 4096

  # The undetected-window sentinel (~333k) is not a size anyone should cut a request to.
  Scenario: An estimated window is capped at 32768
    Given station "s1" is on air for "m" with an estimated window of 333333
    When a consumer GETs /market
    Then the "m" row has ctx_estimated true and ctx_max_estimated 32768

  Scenario: A station that declares no window adds nothing
    Given station "s1" is on air for "m" with a window of 0
    When a consumer GETs /market
    Then the "m" row has no "ctx_max" key

  Scenario: /market and /v1/models agree
    Given station "s1" is on air for "m" with a declared window of 4096
    And station "s2" is on air for "m" with a declared window of 16384
    When a consumer GETs /market and GETs /v1/models
    Then the "m" row's ctx_max equals the "m" entry's rogerai.ctx_max

  Scenario: /market and the context_exceeded refusal agree
    Given station "s1" is on air for "m" with a declared window of 4096
    And "u-1" is a funded consumer
    When a consumer GETs /market
    And "u-1" relays for "m" with a prompt of about 5000 tokens
    Then the "m" row's ctx_max equals error.metadata.window

  Scenario: The model filter keeps the field
    Given station "s1" is on air for "m" with a declared window of 16384
    When a consumer GETs /market?model=m
    Then the "m" row has ctx_max 16384

  Scenario: Existing fields are unchanged
    Given station "s1" is on air for "m" with a declared window of 16384
    When a consumer GETs /market
    Then every field the "m" row carried before this change is present with the same value

  Scenario: A window change shows after the cache expires
    Given station "s1" is on air for "m" with a declared window of 4096
    And a consumer GETs /market
    When "s1" re-registers "m" with a declared window of 16384
    And the public market cache expires
    And a consumer GETs /market
    Then the "m" row has ctx_max 16384

  Scenario: A station cannot declare an absurd window into the feed
    Given station "s1" registers "m" with a declared window of -5
    When a consumer GETs /market
    Then the "m" row has no "ctx_max" key

  Scenario: The API reference documents the field
    Then the public API reference describes ctx_max on /market as the widest declared window among online stations
