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

  Scenario: An offline station's text does not reach the row
    Given the only station that declares text for "someone-7b" stops heartbeating
    When a consumer GETs /market
    Then the "someone-7b" row has no "details" key

  Scenario: A private station's text does not reach the row
    Given a private station declares the summary "Private words." for "someone-7b"
    When a consumer GETs /market
    Then the "someone-7b" row's details.summary is "A 7B chat model."
    And the response does not contain "Private words."

  Scenario: A banned station's text does not reach the row
    Given the stations that declare text for "someone-7b" are banned
    When a consumer GETs /market
    Then the "someone-7b" row has no "details" key

  Scenario: A node cannot claim the curated source
    Given a station registers "someone-7b" with details source "rogerai"
    When a consumer GETs /market
    Then the "someone-7b" row's details.source is "declared"

  Scenario: The text arrives escaped, never as raw markup
    Given a station declares the summary "<b>Fast</b> & small" for "markup-m"
    When a consumer GETs /market
    Then the "markup-m" row's details.summary is the sanitized text of text_sanitization.feature
    And the raw response body contains no "<b>"

  Scenario: The model filter keeps the block
    When a consumer GETs /market?model=demo-1b-fire-decision
    Then the "demo-1b-fire-decision" row's details.source is "rogerai"

  Scenario: The category filter keeps the block
    When a consumer GETs /market?category=language
    Then every row that has text carries its details block

  Scenario: The text does not change the signal or the order
    Given the same offers with every details field removed
    When a consumer GETs /market
    Then every row's signal and the row order equal the response with details

  Scenario: The cached feed is not recomputed per request for details
    When a consumer GETs /market twice within the cache window
    Then computeMarket ran once

  Scenario: A voice model's row can carry text too
    Given a tts station declares the summary "A calm voice." for "calm-voice"
    When a consumer GETs /market
    Then the "calm-voice" row's details.summary is "A calm voice."

  Scenario: The API reference documents the block
    Then the public API reference describes details on /market as the same block as rogerai.details on /v1/models
