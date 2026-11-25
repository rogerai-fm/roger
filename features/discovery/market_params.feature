# Approved by the founder 2026-11-23 (D1 to D4 as written).
#
# GET /market carries each model's parameter count, so a model picker can order by size
# ("Smartest") from the feed it already reads instead of using price as a stand-in.
#
# Source of the ask: the iOS handoff BROKER-MODEL-INFO-FOR-CLIENTS-HANDOFF-2026-11-23 (item 8).
# The founder chose the parameter count over a hand-set tier on 2026-11-23.
#
# GROUND TRUTH (origin/main 8c788286):
#   - Each /discover offer carries params_b (billions) and params_estimated: the station's
#     declaration, else the broker's estimate from the model id. A model whose id names no
#     size and whose stations declare none carries neither key.
#   - /v1/models folds the online public offers into rogerai.params_b and
#     rogerai.params_estimated (models.go): the stations' declarations when any station
#     declares one, else the estimates; the largest value wins; when the values differ the
#     block also carries params_b_min and params_b_max.
#   - /market rows carry none of this today.
#   - On 2026-11-23 the live band had a count for 6 of 19 chat models, all estimated from
#     the id. The hosted commercial models publish no count and have none.
#
# DECISIONS IN THIS DRAFT (each is a ruling to confirm):
#   D1. The /market row gains the SAME keys with the SAME values as the rogerai block on
#       /v1/models: params_b, params_estimated, and params_b_min / params_b_max when the
#       stations differ. One rule, two feeds.
#   D2. No count means no keys. The broker does not invent a number for a model whose size
#       is not published. A client must place those models itself (for example after the
#       sized ones, or by its own list).
#   D3. Display only on /market. The count is not part of the signal and does not change
#       the row order. Routing by size stays where it is (roger.params_b and the @class
#       names).
#   D4. No "capability" or "tier" key. A count is a fact; a tier is a judgment, and nobody
#       owns that judgment yet.

Feature: /market carries each model's parameter count

  Background:
    Given a broker with an empty in-memory node registry

  # --- the count -------------------------------------------------------------------------

  Scenario: A declared count reaches the row
    Given station "s1" is on air for "m" declaring params_b 7
    When a consumer GETs /market
    Then the "m" row carries params_b 7 and params_estimated false

  Scenario: An estimate from the model id reaches the row and says so
    Given station "s1" is on air for "thing-27b" declaring no params_b
    When a consumer GETs /market
    Then the "thing-27b" row carries params_b 27 and params_estimated true

  Scenario: A model with no count carries none of the keys
    Given station "s1" is on air for "mystery-m" declaring no params_b
    When a consumer GETs /market
    Then the "mystery-m" row has no "params_b" key
    And the "mystery-m" row has no "params_estimated" key
    And the "mystery-m" row has no "params_b_min" key
    And the "mystery-m" row has no "params_b_max" key

  Scenario: A count below one billion keeps its fraction
    Given station "s1" is on air for "tiny-m" declaring params_b 0.293
    When a consumer GETs /market
    Then the "tiny-m" row carries params_b 0.293 and params_estimated false

  Scenario: A declaration wins over an estimate
    Given station "s1" is on air for "thing-27b" declaring params_b 30
    And station "s2" is on air for "thing-27b" declaring no params_b
    When a consumer GETs /market
    Then the "thing-27b" row carries params_b 30 and params_estimated false

  Scenario: Stations that declare different counts give a range
    Given station "s1" is on air for "m" declaring params_b 7
    And station "s2" is on air for "m" declaring params_b 8
    When a consumer GETs /market
    Then the "m" row carries params_b 8 and params_estimated false
    And the "m" row's params_b_min is 7 and params_b_max is 8

  Scenario: Stations that agree give no range
    Given station "s1" is on air for "m" declaring params_b 7
    And station "s2" is on air for "m" declaring params_b 7
    When a consumer GETs /market
    Then the "m" row has no "params_b_min" key
    And the "m" row has no "params_b_max" key

  Scenario: /market and /v1/models carry the same count for every model
    Given the stations of this feature's earlier scenarios are all on air
    When a consumer GETs /market and GETs /v1/models
    Then for every model on both, the row's params_b, params_estimated, params_b_min and params_b_max equal the entry's rogerai values

  # --- who does not count ------------------------------------------------------------------

  Scenario: An offline station's declaration does not reach the row
    Given station "s1" is on air for "m" declaring no params_b
    And station "s2" declared params_b 70 for "m" and has gone off the air
    When a consumer GETs /market
    Then the "m" row has no "params_b" key

  Scenario: A private station's declaration does not reach the row
    Given station "s1" is on air for "m" declaring params_b 7
    And a private station is on air for "m" declaring params_b 70
    When a consumer GETs /market
    Then the "m" row carries params_b 7 and params_estimated false
    And the "m" row has no "params_b_max" key

  Scenario: A banned station's declaration does not reach the row
    Given station "s1" is on air for "m" declaring params_b 7
    And a banned station declares params_b 70 for "m"
    When a consumer GETs /market
    Then the "m" row carries params_b 7 and params_estimated false
    And the "m" row has no "params_b_max" key

  # --- adversarial -----------------------------------------------------------------------

  Scenario Outline: A count that is not a sane number is not published
    Given station "s1" registers "mystery-m" with params_b <sent>
    When a consumer GETs /market
    Then the "mystery-m" row has no "params_b" key

    Examples:
      | sent    |
      | 0       |
      | -7      |
      | "seven" |
      | null    |

  Scenario: A station cannot claim an estimate is a declaration
    Given station "s1" registers "thing-27b" with no params_b and params_estimated false
    When a consumer GETs /market
    Then the "thing-27b" row carries params_b 27 and params_estimated true

  Scenario: One station inflating its count shows as a range, not as the only number
    Given stations "s1" and "s2" are on air for "m" declaring params_b 7
    And station "s3" is on air for "m" declaring params_b 700
    When a consumer GETs /market
    Then the "m" row's params_b_min is 7 and params_b_max is 700

  # --- display only ----------------------------------------------------------------------

  Scenario: The count does not change the signal or the order
    Given the offers of this feature with and without declared counts
    When a consumer GETs /market for each
    Then every row's signal and the row order are equal

  Scenario: The count is not a /market filter
    Given station "s1" is on air for "m" declaring params_b 7
    When a consumer GETs /market?params_b=7
    Then the response is 400 with error code "unknown_query_param"

  Scenario: The model filter keeps the count
    Given station "s1" is on air for "m" declaring params_b 7
    When a consumer GETs /market?model=m
    Then the "m" row carries params_b 7 and params_estimated false

  Scenario: Rows without a count are unchanged from today
    Given station "s1" is on air for "mystery-m" declaring no params_b
    When a consumer GETs /market
    Then the "mystery-m" row is byte-identical to the row before this change

  Scenario: The cached feed is not recomputed per request
    Given station "s1" is on air for "m" declaring params_b 7
    When a consumer GETs /market twice within the cache window
    Then computeMarket ran once

  # --- docs ------------------------------------------------------------------------------

  Scenario: The API reference documents the keys
    Then the public API reference describes params_b, params_estimated, params_b_min and params_b_max on /market as the same values as on /v1/models
    And it says a model with no published size carries none of them

  # --- regression (review 2026-11-23) ---

  Scenario Outline: An absurd estimate from the model id is not published on any feed
    Given station "s1" is on air for "<model>" declaring no params_b
    When a consumer GETs /market, /v1/models and /discover
    Then no feed carries a parameter count for "<model>"

    Examples:
      | model                                      |
      | big-99999999b                              |
      | x-9999999999999999999999999999999999999999b |

  Scenario Outline: An absurd stored count is not published on any feed
    Given station "s1" is on air for "mystery-m" with a stored params_b of <stored>
    When a consumer GETs /market, /v1/models and /discover
    Then no feed carries a parameter count for "mystery-m"

    Examples:
      | stored |
      | 20000  |
      | 1e300  |

  Scenario: The largest count the register door accepts is still published
    Given station "s1" is on air for "mystery-m" declaring params_b 10000
    When a consumer GETs /market
    Then the "mystery-m" row carries params_b 10000 and params_estimated false

  Scenario: The API reference says what a filtered read counts
    Then the public API reference says a filtered /market read counts only the stations that pass the filter
