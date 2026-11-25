# Approved by the founder 2026-11-23 (D1 to D5 as written).
#
# A band served only by curated stations has a real signal, so a client that breaks ties on
# signal does not rank a working model as dead.
#
# Source of the ask: the iOS app's request for model information (2026-11-23).
#
# GROUND TRUTH (origin/main 8c788286):
#   - A curated station proxies a commercial API. /market counts it in curated_providers,
#     never in providers: "providers" is the market's claim about people's own hardware
#     (features/curated/curated_web.feature, regression of 2026-09-02).
#   - computeSignal returns 0 with an empty terms block when providers is 0
#     ("no providers = dead channel"). It never looks at curated_providers.
#   - So every curated-only band reads providers 0 and signal 0 while it is on air and
#     answering.
#   - The website already reads curated_providers for these rows. The apps read providers
#     and signal.
#
# DECISIONS IN THIS DRAFT (each is a ruling to confirm):
#   D1. providers stays as it is. A curated-only band still reads providers 0 and
#       curated_providers N. A client that wants "is anyone serving this" adds the two.
#   D2. The signal is computed whenever providers + curated_providers is above 0. Only a
#       band with neither is a dead channel with signal 0.
#   D3. The supply term still counts people's hardware only. A curated-only band scores 0
#       on supply and earns its signal from the other terms (speed, latency, verified,
#       success, trust). Curated supply never raises the supply term of any band.
#   D4. Nothing else in the formula changes. A band with at least one person's station
#       scores exactly what it scores today.
#   D5. Routing does not change. This is the published number only.
#
# AMENDED 2026-11-25 by the founder's three rulings ("1. ok" to each; the recommended picks):
#   D6. No busy discount for a band with no live person's station. A hosted service scales on
#       its own, so in-flight requests on curated stations do not lower a curated-only band's
#       signal (terms.congestion 0). A band with a live person's station keeps its discount
#       as today.
#   D7. The wording: the signal is computed whenever either count is above 0 (was "0 only
#       when both are 0").
#   D8. A station failing its health checks adds nothing to the signal or to ctx_max or
#       ctx_max_estimated (the parameter count already left it out), for
#       curated and people's stations alike. "Failing" is the veto that already shows a
#       heartbeating station offline on /discover and keeps it out of /v1/models: a failed
#       probe streak of probeDeadStreak with no recent serving evidence, read by the
#       instance that hosts the station's poll. The published providers and
#       curated_providers counts do not change. A band whose every station fails reads
#       signal 0. The row's other fields (in_flight, best_tps, ttft_ms, quality,
#       success_rate, verified) still read every counted station, as today. D4 now reads: a
#       band whose stations all pass scores what it scores today.

Feature: A curated-only band has a real signal

  Background:
    Given a broker with an empty in-memory node registry

  Scenario: A curated-only band that is on air has a signal above zero
    Given one curated station is on air for "vendor/model-a"
    When a consumer GETs /market
    Then the "vendor/model-a" row has providers 0 and curated_providers 1
    And the "vendor/model-a" row's signal is above 0
    And the "vendor/model-a" row's terms.total equals its signal

  Scenario: A curated-only band scores nothing on supply
    Given three curated stations are on air for "vendor/model-a"
    When a consumer GETs /market
    Then the "vendor/model-a" row's terms.supply is 0

  Scenario: More curated stations do not raise the signal through supply
    Given one curated station is on air for "vendor/model-a"
    And three curated stations with the same measurements are on air for "vendor/model-b"
    When a consumer GETs /market
    Then the "vendor/model-a" row's signal equals the "vendor/model-b" row's signal

  Scenario: A band with a person's station scores exactly what it scores today
    Given station "s1" is on air for "m"
    When a consumer GETs /market
    Then the "m" row's signal and terms equal the values computed by today's formula

  Scenario: Curated supply beside a person's station does not change the supply term
    Given station "s1" is on air for "m"
    And one curated station is on air for "m"
    When a consumer GETs /market
    Then the "m" row has providers 1 and curated_providers 1
    And the "m" row's terms.supply equals the supply term for one provider

  Scenario: A measured curated band outranks an unmeasured one
    Given one curated station is on air for "vendor/fast" with a measured speed and a served record
    And one curated station is on air for "vendor/unknown" with no measurements
    When a consumer GETs /market
    Then the "vendor/fast" row's signal is above the "vendor/unknown" row's signal

  Scenario: A band with nobody on air is still a dead channel
    Given the only curated station for "vendor/model-a" has gone off the air
    When a consumer GETs /market
    Then there is no "vendor/model-a" row

  Scenario: A cooling curated band keeps its row and its signal rule
    Given the only curated station for "vendor/model-a" is cooling after an upstream rate limit
    When a consumer GETs /market
    Then the "vendor/model-a" row's signal follows the same rule as any cooling band today

  Scenario: A curated station cannot claim to be a person's station to gain supply
    Given a station registers "vendor/model-a" as curated and also declares providers 5
    When a consumer GETs /market
    Then the "vendor/model-a" row has providers 0
    And the "vendor/model-a" row's terms.supply is 0

  Scenario: A private or banned curated station gives the band no signal
    Given a private curated station and a banned curated station serve "vendor/model-a"
    When a consumer GETs /market
    Then there is no "vendor/model-a" row

  Scenario: Routing is unchanged
    Given one curated station and station "s1" are on air for "m" at the same price
    And "u-1" is a funded consumer
    When "u-1" relays for "m" 20 times
    Then the stations chosen are the ones today's routing chooses

  # --- D6: no busy discount for a curated-only band ---------------------------------------

  Scenario: A busy curated-only band keeps its full signal
    Given one curated station is on air for "vendor/model-a" with 4 requests in flight
    And one curated station is on air for "vendor/model-b"
    When a consumer GETs /market
    Then the "vendor/model-a" row's terms.congestion is 0
    And the "vendor/model-a" row's signal equals the "vendor/model-b" row's signal

  Scenario: A busy band with a person's station keeps today's busy discount
    Given station "s1" is on air for "m" with 4 requests in flight
    When a consumer GETs /market
    Then the "m" row's terms.congestion is 1
    And the "m" row's signal is below 39

  Scenario: Busy curated stations beside a person's station count toward the discount as today
    Given station "s1" is on air for "m"
    And one curated station is on air for "m" with 4 requests in flight
    When a consumer GETs /market
    Then the "m" row's terms.congestion is 1

  # --- D8: a station failing its health checks adds nothing ----------------------------------

  Scenario: A band whose only person's station fails its health checks reads signal 0
    Given station "s1" is on air for "m"
    And station "s1" is failing its health checks
    When a consumer GETs /market
    Then the "m" row has providers 1 and curated_providers 0
    And the "m" row's signal is 0

  Scenario: A curated-only band whose station fails its health checks reads signal 0
    Given one curated station is on air for "vendor/model-a"
    And that curated station is failing its health checks
    When a consumer GETs /market
    Then the "vendor/model-a" row has providers 0 and curated_providers 1
    And the "vendor/model-a" row's signal is 0

  Scenario: A failing station does not add to the supply term
    Given station "s1" is on air for "m"
    And station "s2" is on air for "m"
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row has providers 2 and curated_providers 0
    And the "m" row's signal and terms equal the values computed by today's formula

  Scenario: A failing station's speed does not lift the band
    Given station "s1" is on air for "m"
    And station "s2" is on air for "m" with a measured speed of 300 tok/s
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row's terms.speed is 0

  Scenario: A failing station's load does not count toward the busy discount
    Given station "s1" is on air for "m"
    And station "s2" is on air for "m" with 4 requests in flight
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row's terms.congestion is 0

  Scenario: A failing curated station beside a passing one leaves the curated-only signal as one station gives it
    Given one curated station is on air for "vendor/model-a"
    And one curated station is on air for "vendor/model-a" with a measured speed of 300 tok/s
    And that curated station is failing its health checks
    And one curated station is on air for "vendor/model-b"
    When a consumer GETs /market
    Then the "vendor/model-a" row's signal equals the "vendor/model-b" row's signal

  Scenario: A failing person's station beside a busy curated one leaves no busy discount
    Given station "s1" is on air for "m"
    And station "s1" is failing its health checks
    And one curated station is on air for "m" with 4 requests in flight
    When a consumer GETs /market
    Then the "m" row's terms.congestion is 0
    And the "m" row's terms.supply is 0

  Scenario: A failing station's measurements still show on the row outside the signal
    Given station "s1" is on air for "m"
    And station "s2" is on air for "m" with a measured speed of 300 tok/s
    And station "s2" has 4 requests in flight
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row's in_flight is 4 and best_tps is 300
    And the "m" row's terms.speed is 0
    And the "m" row's terms.congestion is 0

  Scenario: A failing station does not widen ctx_max_estimated
    Given station "s1" is on air for "m" with an estimated window of 8192
    And station "s2" is on air for "m" with an estimated window of 16384
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row's ctx_max_estimated is 8192

  Scenario: A failing station does not widen ctx_max
    Given station "s1" is on air for "m" with a window of 8192
    And station "s2" is on air for "m" with a window of 32768
    And station "s2" is failing its health checks
    When a consumer GETs /market
    Then the "m" row's ctx_max is 8192

  Scenario: A failing streak with recent serving evidence still counts
    Given station "s1" is on air for "m"
    And station "s1" has a failed probe streak but was measured within the heartbeat window
    When a consumer GETs /market
    Then the "m" row's signal is above 0

  Scenario: A failing streak read by an instance that does not host the station's poll still counts
    Given station "s1" is on air for "m"
    And station "s1" is failing its health checks
    And this instance does not host station "s1"'s poll
    When a consumer GETs /market
    Then the "m" row's signal is above 0

  # --- D7 and the reference -------------------------------------------------------------------

  Scenario: The API reference says how to read a curated-only row
    Then the public API reference says providers counts people's own hardware and curated_providers counts proxied stations
    And it says the signal is computed whenever either count is above 0
    And it says a station failing its health checks adds nothing to the signal or to ctx_max
    And it says a band with no live person's station has no busy discount
