# features/multinode/rc_inbox_delivery.feature
#
# APPROVED by the founder 2026-11-26. Extends rc_inbox.feature R2 ("every inbound reaches the host
# exactly once, in send order").
#
# Problem (found by the TestRCInboxBDD "between polls" flake): the host's poll popped a turn from the
# session inbox, then wrote it to the poll response. When the host's connection was already gone
# (a network drop, a laptop sleeping, `roger` quitting mid-poll), the write went nowhere and the turn
# was lost: the viewer saw it sent, the host never got it. Delivery was at most once.
#
# DECISIONS (approved 2026-11-26):
#   D1  Delivery is acknowledged. A poll from a host that supports acks hands the turn out and parks
#       it in the session's in-flight set, in one atomic step, keyed by the turn's id. The host's
#       NEXT poll carries `ack=<id>` and the broker drops the acked turn from in-flight.
#   D2  An in-flight turn not acked within 30 s goes back to the FRONT of the inbox, so order is
#       kept, and is handed out again. (Implementation note: a new poll from an acking host also
#       returns every turn it did not ack at once. The host polls one at a time, so a turn the new
#       poll does not ack was never received. Each acking poll also fences the polls before it, so a
#       poll whose host has gone can no longer take a turn.)
#   D3  Every inbound carries a stable id. The host ignores an id it has already handled, so a
#       redelivery after a lost ack never runs twice: at least once on the wire, exactly once for
#       the user.
#   D4  A host without ack support keeps today's behavior, plus a guard: if its poll request is
#       already gone after the pop, the turn goes back to the front. A host says it supports acks
#       with the header X-Roger-RC-Ack: 1 on its poll.
#   D5  A turn handed out 5 times without an ack is dropped, and the viewers get a frame saying the
#       host did not receive it, so nothing loops forever.
#   D6  The in-flight set expires with the session's inbox, the same as the inbox does today.
#
# SCOPE: every broker shape. A multi-instance broker keeps the in-flight set in the shared store
# on the inbox plane, in every dispatch mode (queue-only, which production runs, queue and bus); a
# single instance keeps it in the session's hub. The store-TTL scenario (@store) needs a store.
#
# ENFORCED BY: cmd/rogerai-broker/rc_inbox_bdd_test.go (TestRCInboxDeliveryBDD, in queue-only, bus and single-instance mode: real broker instances
# over real HTTP, real miniredis, no mocks) and internal/client/rc_ack_test.go (the host's ack and
# dedupe in the real RCBridge poll loop).

Feature: A remote-control turn is never lost between the broker and the host

  Background:
    Given a multi-instance broker of two instances sharing one store
    And a remote-control session whose host polls instance A
    And a viewer streaming on instance B

  Scenario: A turn handed to a host whose connection is gone is delivered on its next poll
    Given the host's poll connection drops right after the broker hands it turn "late"
    When the host polls again
    Then the host receives "late" exactly once

  Scenario: The host acknowledges a turn on its next poll and it is not sent again
    When the viewer sends turn "one"
    And the host receives "one" and polls again with its ack
    Then the in-flight list for the session is empty
    And the host never receives "one" a second time

  Scenario: A lost ack redelivers the turn and the host runs it once
    When the viewer sends turn "one"
    And the host receives "one" but its next poll never reaches the broker
    And 30 seconds pass
    And the host polls again
    Then the broker hands out "one" again with the same id
    And the host runs "one" exactly once

  Scenario: Redelivery keeps send order
    Given the host received "one" and did not ack it
    And the viewer has since sent "two"
    When 30 seconds pass and the host polls twice
    Then the host receives "one" and then "two"

  Scenario: An old host without ack support loses nothing when it quits mid-poll
    Given the host does not advertise ack support
    And the host's poll request is cancelled after the broker pops turn "late"
    When the host polls again
    Then the host receives "late" exactly once

  Scenario: A turn that is never acked stops after 5 tries and the viewer is told
    Given the host receives turn "one" 5 times and never acks it
    Then the broker drops "one"
    And the viewer receives a frame saying the host did not receive "one"

  @store
  Scenario: The in-flight list expires with the session
    Given the host received "one" and did not ack it
    When the session's inbox TTL passes with no host poll
    Then "one" is gone from the store
