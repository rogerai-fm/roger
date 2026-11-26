# APPROVED SPEC - founder approved 2026-11-24 ("2. yes. approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: BUILT. internal/towercore/dispatch (SignFailureReceipt, ParseReceipt, IsFailureReceipt),
# internal/station/edge_sealed.go (failureReceipt), internal/towerhub/serve.go (a failure receipt rides
# with a failure), cmd/rogerai-broker/toweredge.go (settle writes no outcome or audit for one) and
# edgebridge.go (no StationFault row for a proven upstream failure on the consumer path; the
# canary still records CanaryFail; the consumer's status follows the Station's signed class).
#
# FOUNDER RULING 2026-11-26 ("Split it"), where this file and the other meet: a failure receipt
# proves the upstream failed, not that the prompt was too long. A context-window claim that is not
# believed is a Station fault on the Tower's record whether or not a receipt rode along
# (context_window_class.feature D7). Every other failure the Station proves with a receipt records
# none (tower_failed_attempt_release.feature).
#
# Deploy Core before nodes update: an old Core settles a failure receipt at zero but records it as
# uncorroborated.
#
# Founder ruling 2026-11-24 ("fix"): make Core do what tower_upstream_key.feature already says -
# when the upstream refuses or fails, "the consumer's hold is released, and it is never a strike"
# - promptly, on every Tower path. Today (review of wt/upstream-key, 2026-11-24):
#   - bridge path (edgebridge.go): the hold is released at once, but a non-scoring
#     StationFault row is written for a failure the Tower only forwarded;
#   - direct edge path: Core never hears of the failure, because the hub couriers only
#     receipts (towerhub/server.go). The hold waits for the stale-hold sweep, 10 minutes
#     (holdsweep.go), so the consumer's funds are tied up for no work.
#
# Fix: on an upstream failure the Station signs a FAILURE RECEIPT - zero usage, no response
# bytes, the failure class - and the hub couriers it like any receipt. Settling a zero-usage
# receipt already captures the hold at zero cost, which refunds it in full (toweredge.go,
# captureEdgeCharge). A failure receipt is against the Station's own interest (it claims
# nothing), so it cannot be used to avoid paying for work that was served.

Feature: A failed Tower attempt releases the consumer's hold at once and costs no one

  Background:
    Given a consumer authorized a priced edge attempt A to Station S through joined Tower T
    And Core holds the consumer's funds for A

  Rule: The Station reports a failure with a signed zero receipt

    Scenario Outline: An upstream failure produces a failure receipt
      Given S's upstream answers "<failure>"
      When S serves A
      Then S signs a receipt for A with zero usage, no response bytes and the failure class "<class>"
      And the consumer is told only the class, as today
      And T couriers the receipt to Core like any other

      Examples:
        | failure                  | class            |
        | 401                      | upstream-refused |
        | 403                      | upstream-refused |
        | 429 with Retry-After 30  | throttled        |
        | 500                      | upstream-error   |
        | a timeout                | upstream-error   |
        | an empty body            | upstream-error   |
        | a reply over the limit   | too-large        |

    Scenario: A failure receipt releases the hold at once
      When Core settles A's failure receipt
      Then the consumer's hold is released in full within one courier delivery
      And nothing is charged, nobody is paid, and no lot is minted

    Scenario: A failure is never a strike and never moves a reputation rate
      When Core settles A's failure receipt
      Then no strike is recorded against S, S's owner or T
      And no outcome is recorded that any reputation rate, audit or ban reads
      And the bridge path's StationFault row for a forwarded upstream failure is no longer written

    Scenario: A failure receipt cannot carry usage
      Given a receipt for A marked as a failure but claiming usage above zero
      When Core settles it
      Then it is refused and the hold stays until a valid receipt or the sweep

    Scenario: A failure receipt cannot follow a served receipt
      Given A already settled on a served receipt
      When a failure receipt for A arrives
      Then it is answered as a duplicate and changes nothing

    Scenario: A served receipt cannot follow a failure receipt
      Given A already settled on a failure receipt
      When a served receipt for A arrives
      Then it is answered as a duplicate and the consumer is not charged

    Scenario: Only S can sign A's failure receipt
      When a failure receipt for A is signed by any key but S's assertion key
      Then it is refused and the hold is unchanged

    Scenario: A Tower cannot turn a served attempt into a failure
      Given S served A and signed a served receipt
      Then T cannot produce a failure receipt for A, because only S signs receipts

    Scenario: A lost failure receipt is caught by the sweep
      Given T never delivers A's failure receipt
      Then the stale-hold sweep releases the hold, as today

    Scenario: An old node that sends no failure receipt behaves as today
      Given S runs a node from before this change
      When its upstream fails
      Then the hold is released by the sweep, as today
