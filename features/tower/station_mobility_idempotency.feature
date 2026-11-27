# APPROVED SPEC - founder approved 2026-11-24 ("approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: BUILT. cmd/rogerai-broker/toweredgeattach.go, towerofferchange.go, towermobility.go,
# towerlink.go, sharedstore.go (incrWithTTL); internal/towercore/attach (UpdateOffer, SetStateFrom,
# AttachedAt kept on revival), internal/towercore/mobility (NoteNudgeOnce), internal/towercore/dispatch
# (OpenUntil); internal/store (OwnerBannedSince); internal/agent (stationdir.go, relaywatch.go,
# reattachWait, nudgeActionable, ReattachDelayFor, moveWhy); internal/towerhub (PollAccepted,
# askCoreAboutUnknown).
#
# AMENDED 2026-11-24 (founder: "yes"): modality limits, stale-mirror scenario dropped, recovery bound 12
# minutes with flat 30 s retry after no data plane. An in-place offer change may only go to a modality the
# Tower path already carries (chat, text); "A stale mirror is not taken as a live Tower" was dropped by
# ruling, so private_hub_placement.feature's behaviour for a report while the Tower is not pulling stands.
#
# Founder request 2026-11-23: "review and audit the behavior of our relay towers and how
# stations will work with them, i want it to be as idempotent as possible and stations to be
# able to freely move around."
#
# Source: the read-only mobility audit of origin/main 8c788286 (2026-11-23), 12 findings, plus
# one seen live on 2026-11-24: a share that had just attached to the house Tower
# (tw-3081d0565b48999cdb34fbce) printed "this node's relay hub refuses its identity (401)" one
# second after attaching, and served through that Tower 45 seconds later with no word that it had
# recovered.
#
# This file ADDS rules to the approved station_mobility.feature (2026-11-13). Four of them change
# what an approved scenario allows. Each is marked "CHANGES APPROVED BEHAVIOUR" and names the
# scenario it changes, so it can be approved or refused on its own:
#   R1  an offer change updates the live row in place (was: 409 "revoke and re-attach" for a
#       model or price change, and a SILENT 200 with the old row for a modality change, because
#       the retry branch compares model and prices only)
#   R3  "being moved" is done inline in the attach answer (was: 503 "try again shortly" and the
#       mobility pass moves it later, private_hub_placement.feature, the left-network and
#       unreachable scenarios)
#   R4  Stations leave a Tower that has been ineligible past the window (was: never moved)
#   R5  a Station stays "busy" until its last receipt's settle deadline (was: until the
#       execution deadline), so a move does not void delivered work
#
# Terms:
#   offer          model, modality, price in, price out, as signed into the attach
#   live row       the Station's attachment in state quarantine or active
#   retry branch   an attach whose assertion key already has a live row for the same owner and
#                  session key (toweredgeattach.go)
#   ineligible     a Tower whose link is up but which may not take work: lease expired,
#                  suspended, draining, or failing admission
#   the window     the existing "gone" window, 2 x 180 seconds, unless a scenario names another
#   settle deadline  the attempt's execution deadline plus the settle grace (toweredge.go)

Feature: Stations move freely between Towers, and every attach, move and retry is idempotent

  Background:
    Given a signed-in operator whose node is registered and co-signed by their account
    And two live joined Towers, A and B, that may take that node's Station
    And the node's Station S is attached to Tower A at epoch 4

  # ---------------------------------------------------------------------------------------
  Rule: R1 - an offer change from the proven key holder updates the live row in place
  # CHANGES APPROVED BEHAVIOUR: the 409 "these keys are attached with a different offer" in the
  # retry branch. No approved .feature specifies the 409; only a unit test does
  # (TestSelfAttachRetryWithADifferentOfferIsRefused), and that test is replaced.

    Scenario Outline: A restart with a new offer keeps the Station on the relay
      Given S is live on Tower A with model "m1", modality "chat", price in 100, price out 200
      When the same node re-attaches with the same assertion and session keys and "<change>"
      Then the attach is accepted on Tower A at the same epoch
      And S's row carries the new offer
      And the published relay row for S shows the new offer within one publish cycle
      And no consumer is routed to S at the old offer after the reply is written
      And a "station_offer_changed" event names S, the old offer and the new offer
      And `roger share` prints one line saying the offer changed on the Tower

      Examples:
        | change                     |
        | price in 150               |
        | price out 250              |
        | model "m2"                 |
        | modality "text"            |
        | model "m2" and price in 0  |

    Scenario Outline: An offer change is refused from anyone but the key holder
      Given an attach naming S's assertion key and a different offer, but "<flaw>"
      Then it is refused exactly as that flaw is refused today
      And S's row and offer are unchanged

      Examples:
        | flaw                                               |
        | carrying a different session key                   |
        | not co-signed by the assertion key                 |
        | naming a different node_id than S's row            |

    Scenario: An offer change under another account is refused
      Given S is owned by account A
      When an attach with S's keys arrives resolving to account B with a different offer
      Then it is refused with 403
      And S's row and offer are unchanged

    Scenario: The new offer is held to the same limits as a fresh attach
      When S re-attaches with an offer a fresh attach would refuse
      Then the re-attach is refused with the same status and words a fresh attach gets
      And S keeps its old offer and stays live

    Scenario: A model change does not carry the old model's evidence
      Given S has health, probe, category and trust evidence for model "m1"
      When S's offer changes in place to model "m2"
      Then S is treated for "m2" exactly as a new Station would be: probed before it is served
      And none of "m1"'s evidence is shown or used for "m2"
      And a price-only or modality-only change keeps S's evidence

    Scenario: An attempt granted at the old offer settles at the old offer
      Given an edge attempt was granted to S at price in 100 before the change
      When S's offer changes to price in 150 and the attempt then settles
      Then the attempt is billed and paid at price in 100

    Scenario: Re-sending the same offer change is a no-op
      Given S's offer was just changed to price in 150
      When the same re-attach arrives again
      Then the answer is identical to the first
      And exactly one "station_offer_changed" event exists for that change

    Scenario: Two concurrent offer changes leave exactly one offer
      When two re-attaches for S with offers X and Y land on two broker instances at once
      Then S's row carries exactly one of X or Y
      And the published row, the event and the reply each name the offer that row carries

    Scenario: A dormant revival and a live update agree
      Given S is dormant after a move
      When it re-attaches with a new offer
      Then it is revived with the new offer, as today
      And the event is the same "station_offer_changed" a live update writes

    Scenario: The modality is part of the offer
      Given S is live with modality "chat"
      When it re-attaches with modality "text" and everything else the same
      Then the row is updated to "text" and the event is written
      And a re-attach is never answered with the old row while its modality differs

    Scenario: A change to a modality a Tower does not carry is refused like a fresh attach
      When S re-attaches with modality "tts"
      Then it is refused with the same status and words a fresh tts attach gets
      And S keeps its old offer and stays live

    Scenario: Two shares on one machine do not fight over one Station
      Given one machine runs two `roger share` processes with different models
      When both attach
      Then each gets its own Station identity, or the second is told plainly why it cannot
      And neither silently rewrites the other's offer

  # ---------------------------------------------------------------------------------------
  Rule: A Station whose Tower is gone is re-placed on its next attach, not kept dark

    Scenario: The retry branch re-places a Station whose Tower is gone fleet-wide
      Given Tower A has been absent fleet-wide past the window
      When S re-attaches
      Then S's row is set dormant on Tower A and S is placed on Tower B in the same request
      And the reply names Tower B
      And it is one move, recorded once, under the same pacing and claim every move takes

    Scenario: A Tower absent for less than the window is not abandoned
      Given Tower A has been absent fleet-wide for less than the window
      When S re-attaches
      Then the answer is 503 "the tower this node is attached to has no data plane right now - try again shortly", as today

    Scenario: Time from a Tower dying to its idle Stations serving elsewhere is bounded
      Given Tower A dies with 100 idle Stations attached
      When every Station retries every 30 seconds while Core answers "no data plane" for its Tower, instead of backing off
      Then every one of them is serving through another Tower within 12 minutes of A's death
      And moves keep the approved pacing
      # RULED 2026-11-24: flat 30 s retry after no data plane; 12-minute bound for 100 Stations.

    Scenario: The stale comment that said RelayPlane and Live see one instance is corrected
      Then the code comment at the retry branch describes the shared-mirror reads it really makes

  # ---------------------------------------------------------------------------------------
  Rule: R3 - a nudge or a "being moved" answer is acted on at once, even while the node backs off
  # CHANGES APPROVED BEHAVIOUR: private_hub_placement.feature tells a node that left its Tower's
  # network, or cannot reach its hub, to "try again shortly" and leaves the move to the mobility
  # pass. Here the attach answer itself places the node, under the same claim and pacing.

    Scenario: A nudge cuts a backoff wait short
      Given S's node is waiting out a re-attach backoff of 4 minutes
      When a nudge for S arrives
      Then the node re-attaches within 5 seconds instead of finishing the wait

    Scenario: "Being moved" is done inline for an unreachable hub
      Given S reports Tower A's hub unreachable and another Tower can take it
      When Core decides to move S
      Then the same answer places S on Tower B
      And the node does not wait out a backoff to learn where it went

    Scenario: "Being moved" is done inline for a node that left the Tower's network
      Given S's measured network no longer matches private Tower A's
      When S re-attaches
      Then the same answer places S on a Tower it may use, or says plainly that none can take it now

    Scenario: A node that hears its nudge re-attaches inside the undo window
      Given a move for S is pending and S's node is in a 15-minute backoff
      When the first nudge reaches it
      Then it re-attaches before the second nudge is due, so the approved 3-nudge undo is not reached
      And a node that has truly gone silent is still undone after 3 nudges, as approved

    Scenario: The left-network branch honours the one-hour skip
      Given a left-network move of S was undone in the last hour
      When S re-attaches still on the wrong network
      Then S is not put through another dormant-and-back cycle in that hour

    Scenario: A nudge older than the current attach is dropped
      Given a nudge for S was buffered before S's last successful attach
      When S's tenancy starts
      Then that nudge does not cause another re-attach

  # ---------------------------------------------------------------------------------------
  Rule: R4 - Stations leave a Tower that has been ineligible past the window
  # CHANGES APPROVED BEHAVIOUR: today a linked but ineligible Tower is "not gone" on purpose and
  # its Stations are never moved (towermobility.go). No approved scenario says they must stay.

    Scenario Outline: An ineligible Tower's idle Stations are moved after the window
      Given Tower A's link is up but "<reason>" for longer than the window
      And S is idle by the approved idle gate
      When the mobility pass runs
      Then S's move to Tower B starts, under the approved pacing
      And the move's reason is recorded as "ineligible:<reason>"

      Examples:
        | reason            |
        | lease_expired     |
        | suspended         |
        | draining          |
        | failing_admission |
        | banned            |

    Scenario: A briefly ineligible Tower keeps its Stations
      Given Tower A has been ineligible for less than the window
      When the mobility pass runs
      Then no Station is moved off Tower A

    Scenario: A busy Station on an ineligible Tower waits for idle
      Given Tower A has been ineligible past the window and S has work in flight
      When the mobility pass runs
      Then S is not moved this pass and is checked again next pass

    Scenario: The retry branch stops handing out an ineligible Tower
      Given Tower A has been ineligible past the window
      When S re-attaches
      Then S is placed on another eligible Tower in that answer, or told none can take it now
      And S is never handed Tower A's hub while A is ineligible

    Scenario: A Tower that becomes eligible again can be chosen again
      Given Tower A was ineligible and its lease is renewed
      Then Tower A is a placement candidate again by the approved ranking

  # ---------------------------------------------------------------------------------------
  Rule: R5 - a Station counts as busy until its last receipt can no longer settle
  # CHANGES APPROVED BEHAVIOUR: "An attempt caught by a move is void" stays true for attempts
  # that really are caught. What changes is the idle gate: it no longer calls a Station idle
  # while a receipt for delivered work may still be in its Tower's courier spool.

    Scenario: A receipt waiting in the spool keeps the Station busy
      Given S finished an attempt through Tower A at its execution deadline
      And the receipt is still in Tower A's courier spool
      When the mobility pass runs before that attempt's settle deadline
      Then S is not idle and the move does not start

    Scenario: After the settle deadline the Station is idle again
      Given S's last attempt's settle deadline has passed
      When the mobility pass runs
      Then the idle gate treats S by its approved rules

    Scenario: An attempt that settles early frees the Station early
      Given S's last attempt settled before its settle deadline
      When the mobility pass runs
      Then S is idle by the approved rules without waiting for the deadline

    Scenario: The settle deadline is known to every instance
      Given S's last attempt was granted on instance 1
      When instance 2 runs the mobility pass
      Then instance 2 sees the same settle deadline, from shared state, and fails closed if it cannot read it

    Scenario: A Station in steady use is not moved by the soft triggers, and that is accepted
      Given S is granted a new attempt before each previous settle deadline passes
      Then the failing, better and stage-cap triggers never move S while that lasts
      And a gone or ineligible Tower still moves it by its own rules
      # Accepted cost: a busy Station stays put. The settle grace bounds each wait at 10 minutes.

    Scenario: A gone-Tower move still happens, and its spooled receipts are named
      Given Tower A is gone and S has receipts that have not settled
      When S is moved anyway, as approved
      Then the move record names how many unsettled attempts it may void
      And each 410 at settlement is logged as "edge fence MOVED", as approved

  # ---------------------------------------------------------------------------------------
  Rule: A node is never stranded by a 401 or an endless empty poll

    Scenario: A hub that says "no node is registered" counts as a hub failure
      Given S's node polls Tower A's hub and gets 401 "no node is registered" past the grace below
      When the standing window passes
      Then the node re-attaches through Core with backoff, as on a transport failure

    Scenario: A new attachment gets a short grace for the hub to learn it
      Given S's node attached to Tower B less than 60 seconds ago
      When Tower B's hub answers 401 "no node is registered"
      Then the node retries quietly and prints nothing
      # PROPOSED NUMBER: 60 seconds, twice the hub's 30-second refresh.

    Scenario: The share's warning says what is true, and says when it recovers
      Given Tower B's hub has refused S's node past the grace
      Then `roger share` prints one line saying the relay does not know this node yet
      And it does not blame an old Tower, a revoked attachment or the clock unless that is the cause
      And when polls succeed again it prints one line saying the node is serving through the relay

    Scenario: The hub learns a new Station on a poll, not only on a submit
      Given S just attached to Tower B
      When S's node polls Tower B's hub before B's next 30-second refresh
      Then the hub asks Core about the unknown Station before answering 401
      And a Station Core confirms is served on that poll

    Scenario: Polls naming random Station ids cannot make the hub flood Core
      When a caller polls Tower B's hub with many unknown Station ids
      Then the hub asks Core no more often than its existing debounce allows
      And a Station Core does not confirm gets the same 401 words as today

    Scenario: Endless empty polls after a nudge lead to a re-attach
      Given S's node has heard a nudge and then gets only empty polls for 5 minutes
      When the 5 minutes pass
      Then the node re-attaches through Core
      # PROPOSED NUMBER: 5 minutes.

    Scenario: Endless empty polls with no nudge are just a quiet Station
      Given S's node gets only empty polls and has heard no nudge
      Then the node does not re-attach for that reason

  # ---------------------------------------------------------------------------------------
  Rule: Lost responses and duplicates get the same answer, never a 409

    Scenario: Two identical fresh attaches both get the winner's record
      Given S's node sends a fresh attach and, its reply lost, sends it again
      When both land on two broker instances at once
      Then exactly one attachment exists
      And both replies carry that attachment's record, token, pin and fingerprint

    Scenario: The loser of a two-instance revival race gets the winner's answer
      Given S is dormant and two re-attaches for S land on two instances at once
      Then exactly one revival happens and both replies name it

    Scenario: A store outage during admission is a 503, not a 409
      Given admission cannot reach the Station store
      When S attaches
      Then the answer is 503 with Retry-After
      And no 409 is ever answered for a store outage

  # ---------------------------------------------------------------------------------------
  Rule: State writes are guarded so a later writer cannot undo an earlier one

    Scenario: Promote never revives a row a move or revoke just changed
      Given S is in quarantine and is being promoted
      When a move sets S dormant, or a revoke lands, between the read and the write
      Then the promotion does not change S's state
      And S stays dormant or revoked

    Scenario: The mobility lease is renewed while a pass runs
      Given a mobility pass takes longer than the 14-second lease
      Then the lease is renewed before it expires
      And no second instance runs the same pending step

    Scenario: A nudge is counted once even if two instances send it
      Given two instances send the same nudge for S in one window
      Then S's nudge count goes up by one

  # ---------------------------------------------------------------------------------------
  Rule: The owner is recognised and told the truth on every path

    Scenario: A revoke from any of the owner's device keys is recognised
      Given S was attached under account A's canonical key
      When the owner revokes S signed by another device key of account A
      Then the revoke is accepted, not answered 404

    Scenario Outline: Every move reason has its own words in `roger share`
      Given S was moved for reason "<reason>"
      Then `roger share` names that reason in plain words, not generic text

      Examples:
        | reason            |
        | gone              |
        | failing           |
        | better            |
        | stage_cap         |
        | left_network      |
        | unreachable       |
        | ineligible        |

    Scenario: A revival does not reset how long the Station has been attached
      Given S has been attached for 20 days and is moved
      When it is revived on Tower B
      Then its attached-since time for eviction order and for revoke waits is still 20 days ago

  # ---------------------------------------------------------------------------------------
  Rule: Everything above holds across restarts and two instances

    Scenario: A broker restart in the middle of a move finishes or undoes it, never both
      Given a move for S is between its dormant write and its re-attach
      When the instance that started it restarts
      Then the move completes on S's re-attach, or is undone by the approved rule, exactly once

    Scenario: Every check above reads shared state where the approved spec requires it
      Then no rule in this file depends on state held only in one instance's memory, except the streaks and absent-since times the approved spec already allows
