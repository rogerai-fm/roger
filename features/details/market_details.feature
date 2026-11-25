# Approved by the founder 2026-11-23. Step definitions and code follow in the next change.
#
# GET /market carries each model's card text, one block per model, so a model picker can show
# it from the feed it already reads without joining station rows from /discover.
#
# Source of the ask: the iOS handoff BROKER-MODEL-INFO-FOR-CLIENTS-HANDOFF-2026-11-23 (item 7).
#
# GROUND TRUTH (origin/main df8bda4b):
#   - /v1/models already publishes the block as rogerai.details (models.go:230): FoldDetails
#     over the model's online stations, then about.Block. Keys: source, summary, best_at,
#     not_for, page, base_model, tuning, license, and for declared text declared_by, stations.
#   - Curated notes (source "rogerai") win over declared text (curated_catalog.feature).
#   - The text is sanitized at registration (text_sanitization.feature).
#   - marketView has no details today. computeMarketKeep already walks every online public
#     offer and its taxonomy.
#
# DECISIONS:
#   D1. The /market row gains "details": the SAME block /v1/models calls rogerai.details,
#       key for key. One rule, two feeds. No new text rules.
#   D2. No block when there is no text. A row without details is byte-identical to today.
#   D3. Like every taxonomy field the text is display only: never routed on, never filtered
#       on, never part of the signal.

Feature: /market carries each model's details block

  Background:
    Given the stations and curated notes of features/details/api_cli.feature

  Scenario: A curated model's row carries the curated block
    When a consumer GETs /market
    Then the "demo-1b-fire-decision" row's details is:
      | key        | value                                          |
      | source     | rogerai                                        |
      | summary    | Picks one fire-defense action from site state. |
      | best_at    | Choosing among a fixed list of actions.        |
      | not_for    | Chat. It answers with one action only.         |
      | base_model | demo-1b                                        |
      | tuning     | adapter                                        |
      | license    | Apache-2.0                                     |
      | page       | https://rogerai.fm/research-demo-1b.html#fire  |

  Scenario: A declared model's row uses the band vote
    When a consumer GETs /market
    Then the "someone-7b" row's details.source is "declared"
    And the "someone-7b" row's details.summary is "A 7B chat model."
    And the "someone-7b" row's details.declared_by is 2
    And the "someone-7b" row's details.stations is 3
    And the "someone-7b" row's details has no "page"

  Scenario: A model with no text has no details key
    When a consumer GETs /market
    Then the "mystery-m" row has no "details" key

  Scenario: /market and /v1/models carry the same block for every model
    When a consumer GETs /market and GETs /v1/models
    Then for every model on both, the row's details equals the entry's rogerai.details

  Scenario: Rows without details are unchanged from today
    Given the same offers with every details field removed
    When a consumer GETs /market
    Then the response is byte-identical to the response before this change

  # Fixture corrections (2026-11-23): three stations declare text for "someone-7b", and with
  # all of them gone the model has no row at all. The offline and banned cases use "mystery-m",
  # which keeps its one station with no text.
  Scenario: An offline station's text does not reach the row
    Given a station that declares the summary "Stale words." for "mystery-m" stopped heartbeating
    When a consumer GETs /market
    Then the "mystery-m" row has no "details" key
    And the response does not contain "Stale words."

  Scenario: A private station's text does not reach the row
    Given a private station declares the summary "Private words." for "someone-7b"
    When a consumer GETs /market
    Then the "someone-7b" row's details.summary is "A 7B chat model."
    And the response does not contain "Private words."

  Scenario: A banned station's text does not reach the row
    Given a banned station declares the summary "Banned words." for "mystery-m"
    When a consumer GETs /market
    Then the "mystery-m" row has no "details" key
    And the response does not contain "Banned words."

  Scenario: A node cannot claim the curated source
    Given a station declares the summary "Trust me." for "mystery-m" with details source "rogerai"
    When a consumer GETs /market
    Then the "mystery-m" row's details.source is "declared"
    And the "mystery-m" row's details has no "page"

  Scenario: The text arrives escaped, never as raw markup
    Given a station declares the summary "<b>Fast</b> & small" for "markup-m"
    When a consumer GETs /market
    Then the "markup-m" row's details.summary is "<b>Fast</b> & small"
    And the response does not contain "<b>"

  Scenario: The model filter keeps the block
    When a consumer GETs /market?model=demo-1b-fire-decision
    Then the "demo-1b-fire-decision" row's details.source is "rogerai"

  Scenario: The category filter keeps the block
    Given a station declares the summary "Sorts tickets." for "sorter-m" in category "language/classification"
    When a consumer GETs /market?category=language
    Then the "sorter-m" row's details.summary is "Sorts tickets."
    And every row that has text carries its details block

  Scenario: The text does not change the signal or the order
    Given a consumer GETs /market
    And the same offers with every details field removed
    When a consumer GETs /market again
    Then every row's signal equals its signal with details
    And the rows are in the same signal order

  Scenario: The cached feed is not recomputed per request for details
    When /market is read 3 times within the cache window
    Then the details were computed once per cache fill

  Scenario: A voice model's row can carry text too
    Given a tts station declares the summary "A calm voice." for "calm-voice"
    When a consumer GETs /market
    Then the "calm-voice" row's details.summary is "A calm voice."

  Scenario: The API reference documents the block
    Then the public API reference describes details on /market as the same block as rogerai.details on /v1/models

# --- regression (review 2026-11-23) ---

  # B. The background station declares Apache-2.0, so four licenses tie one each. The
  # lexically smallest wins, on every cache fill and on both feeds.
  Scenario: Facts a curated note leaves out do not depend on station order
    Given RogerAI's curated note for "demo-1b-fire-decision" has no license
    And stations declare the licenses "MIT", "GPL-3.0" and "BSD-3-Clause" for "demo-1b-fire-decision"
    When /market and /v1/models are each read 20 times, each past the cache
    Then on every read the "demo-1b-fire-decision" row's details.license is "Apache-2.0"
    And on every read the row's details equals the entry's rogerai.details

  # C. /v1/models leaves out a station that fails its liveness probes. So does the row.
  Scenario: The text of a station that fails its liveness probes does not reach the row
    Given a station that declares the summary "Dead words." for "mystery-m" has failed its liveness probes
    When a consumer GETs /market
    Then the "mystery-m" row has no "details" key
    And the response does not contain "Dead words."
