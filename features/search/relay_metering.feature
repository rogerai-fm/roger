# Search API: the money contract.
#
# APPROVED by the founder 2026-11-15. A search is metered as ONE query, priced from the chosen offer,
# held before dispatch and settled on the same wallet, ledger, fee and hold rules as chat and
# voice. Only the unit changes. Because the unit count (1) is known up front, the hold equals
# the final charge and there is no recount.
#
# Ground truth: cmd/rogerai-broker/audio.go audioRelayCore (resolvePricing, clampSettleCost,
# monthlyCapCheckCap, ensureSeeded, holdUnderCap, ReleaseHoldFor, settleRequest, receipt
# VerifyNode + BindsTo, X-RogerAI-Provider/Cost/Balance). The search relay is a new spec on
# that spine, not a copy of it.
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   M1  cost = price_per_1M_queries / 1,000,000, one query per request, whatever the count.
#   M2  a well-formed reply with ZERO results is still charged: the provider did the work
#       and paid its upstream. A reply the broker cannot use is a station error and is free.
#   M3  a free offer or a free window is exactly $0: no hold, no money rows, one $0 metering
#       receipt for lineage (as chat and voice do).
#   M4  platform fee, payout hold and reserve are the ones every modality already uses. Search
#       adds no new money rule.
#   M5  every failure before a usable reply refunds the hold in full.

@money @search @metering
Feature: A search costs one query, held up front and settled on the same wallet

  Background:
    Given a consumer with a funded wallet of 10.00 credits
    And a search station "web-search" on air with engine "searxng" at 5000 credits per 1,000,000 queries

  Rule: one request is one query (M1)

    Scenario Outline: The per-query price arithmetic is exact
      Given the station's price is <price> credits per 1,000,000 queries
      When the consumer makes one search
      Then the cost is <cost> credits
      And the ledger row records unit "query" and 1 unit

      Examples:
        | price   | cost     |
        | 0       | 0.000000 |
        | 1       | 0.000001 |
        | 3000    | 0.003000 |
        | 5000    | 0.005000 |
        | 7       | 0.000007 |
        | 12345   | 0.012345 |
        | 49999   | 0.049999 |
        | 50000   | 0.050000 |

    Scenario: Many searches add up exactly, with no rounding drift
      Given the station's price is 3 credits per 1,000,000 queries
      When the consumer makes 1000 searches
      Then the consumer's balance is 9.997000 credits

    Scenario: A price change applies to the next search
      Given the consumer makes one search
      And the station's price is 3000 credits per 1,000,000 queries
      When the consumer makes one search
      Then the cost is 0.003000 credits

    Scenario: A priced window covering now sets the charge
      Given the station "web-search" has a window priced at 20000 covering now
      When the consumer makes one search
      Then the cost is 0.020000 credits

    Scenario: A price below zero that reached the market never moves money
      # Registration refuses a negative price (price_limits.feature). This is the relay's own
      # floor, for a price that got past registration some other way.
      Given the station's price is -5000 credits per 1,000,000 queries
      When the consumer makes one search
      Then the consumer's balance is 10.00 credits
      And the station earns nothing

    Scenario Outline: The result count does not change the price
      When the consumer makes one search with count <count>
      Then the cost is 0.005000 credits

      Examples:
        | count |
        | 1     |
        | 5     |
        | 20    |

    Scenario: The station cannot raise the charge by claiming more work
      Given the station's receipt claims 50 queries
      When the consumer makes one search
      Then the cost is 0.005000 credits

    Scenario: The response carries the meter headers
      When the consumer makes one search
      Then the response has header "X-RogerAI-Cost" for 0.005000
      And the response has header "X-RogerAI-Balance" for 9.995000
      And the response has header "X-RogerAI-Provider" naming the station's node

  Rule: the hold equals the charge and every failure refunds it (M5)

    Scenario: The exact cost is held before the station is called
      When the consumer makes one search
      Then a hold of 0.005000 credits exists while the station works
      And after the reply the hold is consumed by exactly one settlement

    Scenario Outline: A failed search refunds the hold in full
      Given the station <failure>
      When the consumer makes one search
      Then the response status is <status>
      And the consumer's balance is 10.00 credits
      And no settlement row exists for the request
      And the station earns nothing

      Examples:
        | failure                                       | status |
        | returns HTTP 502                              | 500    |
        | returns HTTP 429                              | 500    |
        | returns an empty body                         | 500    |
        | returns a body that is not JSON               | 500    |
        | returns JSON without a results array          | 500    |
        | returns a body over the size cap              | 500    |
        | does not answer before the relay timeout      | 504    |
        | returns a receipt signed by another key       | 500    |
        | returns a receipt bound to another request    | 500    |

    Scenario: A receipt bound to another request also earns the station a strike
      Given the station returns a valid signature on a receipt for a different request id
      When the consumer makes one search
      Then the station's node receives an unbound-receipt strike

    Scenario: The station never sees a hold it could settle twice
      Given the station replies twice to the same job
      When the consumer makes one search
      Then exactly one settlement row exists for the request

  Rule: zero results is a real answer (M2)

    Scenario: A well-formed reply with no results is charged
      Given the station returns {"results": []}
      When the consumer makes one search
      Then the response status is 200
      And the response lists no results
      And the cost is 0.005000 credits

    Scenario: A reply whose every result the broker drops is a station error, not a charge
      Given the station returns 5 results that all have "javascript:" URLs
      When the consumer makes one search
      Then the response status is 500
      And the consumer's balance is 10.00 credits

  Rule: wallets, caps and sign-in

    Scenario: An empty wallet gets 402 before the station is called
      Given the consumer's wallet holds 0.001 credits
      When the consumer makes one search
      Then the response status is 402
      And the station is never called

    Scenario: A wallet holding exactly the cost is served and left at zero
      Given the consumer's wallet holds 0.005 credits
      When the consumer makes one search
      Then the response status is 200
      And the consumer's balance is 0.00 credits

    Scenario: A wallet one micro-credit short gets 402 before the station is called
      Given the consumer's wallet holds 0.004999 credits
      When the consumer makes one search
      Then the response status is 402
      And the station is never called

    Scenario: The monthly spend cap applies to search like any other spend
      Given the consumer's monthly cap leaves 0.004 credits
      When the consumer makes one search
      Then the search is refused by the monthly cap with the same status chat uses
      And the station is never called

    Scenario: A grant key's spend cap applies to search
      Given a grant key whose spend cap is reached
      When a search arrives with that grant key
      Then it is refused with the grant cap status
      And the station is never called

    Scenario: An unseeded wallet is seeded, never misread as 402
      Given a signed-in consumer whose wallet has not been seeded
      When the consumer makes one search
      Then the wallet is seeded before the hold

  Rule: free is exactly free (M3)

    Scenario: A price-0 station costs nothing and writes one $0 receipt
      Given a search station "free-search" on air at 0 credits per 1,000,000 queries
      When the consumer makes one search on "free-search"
      Then the cost is 0 credits
      And no hold is placed
      And exactly one $0 metering receipt is recorded for lineage

    Scenario: A station inside its free window costs nothing
      Given the station "web-search" has a free window covering now
      When the consumer makes one search
      Then the cost is 0 credits
      And no hold is placed

  Rule: the provider is paid on the shared rules (M4)

    Scenario: A paid search pays the provider under the standard fee
      When the consumer makes one search
      Then the provider's accrued earnings rise by the cost less the standard platform fee
      And the earnings follow the standard payout hold and reserve

    Scenario: Search spend shows up in the consumer's usage by model
      When the consumer makes three searches
      Then GET /usage lists "web-search" with 3 requests and 0.015000 credits

    Scenario: Search earnings show up in the provider's per-model serve metrics
      When the consumer makes three searches
      Then the provider's /metrics/provider lists "web-search" with 3 requests
