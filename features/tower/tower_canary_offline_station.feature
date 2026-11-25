# APPROVED SPEC - founder approved 2026-11-24 ("approved, but audit it and review and fix any issues"). Changes to an approved scenario need re-approval.
# REVISED 2026-11-24 after the requested audit (17 findings, all applied; see "Audit changes" below).
# BUILD STATUS: NOT BUILT. Approval is not implementation - this line says which.
#
# Found 2026-11-24. The house Tower (tw-3081d0565b48999cdb34fbce) was suspended by Core and then
# lapsed, after doing nothing wrong:
#   05:16 UTC  a test share attached to it and served one request ("receipt settled")
#   ~05:18     the test share was stopped (Ctrl-C). Its Station stayed attached: stopping a
#              share does not detach it, and an attachment retires only after 7 days unseen.
#   05:31      Core's canary (the sweep runs on every broker instance) had probed that Station 5
#              times; every probe failed, because nothing was serving it. Core suspended the
#              Tower: "lifecycle from=active to=suspended reason=reputation canary_fail=5
#              canary_pass=0"
#   05:32 on   `roger tower serve` exited on every refused request and its service manager
#              restarted it every RestartSec (15 s on the house unit), for 17 hours.
#   16:17      the lease ran out (a suspended Tower cannot renew): "reason=lease_lapsed".
#
# Why: the canary blames the Tower for every failure on purpose, because a hostile Tower could
# fake any of them (towercanary.go). It picks its target from the Stations placement would use.
# Placement drops a node that this instance holds a registration for and has not heard from for
# nodeTTL (45 s) ("the share went home", toweredge.go edgeEligibleM); a node it has never heard
# of falls to Tier B. When no Station is eligible, canaryTargetFor falls back to "everything that
# resolved", which puts the switched-off Station back in. With one Station behind a Tower, every
# probe hit it.
#
# Core does not need the Tower's word to know a node is offline: the node registers with and
# heartbeats Core directly, and a Tower cannot forge those heartbeats. This removes the
# switched-off-Station case, including attaching Stations and switching them off to take an
# honest Tower off the fabric. A Station that keeps heartbeating Core but does not serve still
# costs its Tower, bounded as today by canaryTroubledSlowdown. A Tower whose Stations are all
# offline is still measured, by the Station-free hub check (tower_trust_ramp.feature).
#
# The same thing would happen to any public Tower, Nora's included, the first time a Station
# behind it is switched off.
#
# Audit changes (2026-11-24, applied at the founder's request):
#   - corrected header claims (Tier B, RestartSec, what the fix does and does not close)
#   - one liveness predicate for the canary and for placement, so no Station carries paid traffic
#     unprobed; "unknown to this instance" keeps today's Tier B rule for both
#   - exact boundary: offline at >= nodeTTL (45 s); no attempt opened, no probe budget spent
#   - DELETED "failures recorded while the node was offline do not count": once the fix ships no
#     probe starts against an offline node, and a reading taken after the probe would let a
#     Tower and node line stalled probes up with heartbeat gaps and launder them
#   - the restore scenario is general and names the two admin steps; the Tower id lives in the
#     runbook
#   - roger tower serve: Core returns the refusal state only to the Tower's own signed request;
#     terminal refusals (revoked, unknown Tower, wrong key or mode) exit with a status the
#     service manager does not restart; recoverable ones back off on the node re-attach backoff
#   - the leave message is bound to Station, Tower and epoch, rate limited, sent with a 2 s limit
#   - a leave during a move ends the move
#   - "may probe" became "is in the canary's candidate set", since selection is random
#   - SUPERSEDES TestTheCanaryStillRotatesWhenNothingBehindATowerIsPlaceable, which required the
#     canary to keep probing switched-off nodes. Its spreading property is kept, re-based on
#     heartbeating Stations that placement rejects for another reason.

Feature: A Tower is never judged on a Station that Core itself knows is offline

  Background:
    Given a live joined Tower T that may take work
    And nodeTTL is 45 seconds

  Rule: The canary and placement use one liveness predicate

    Scenario: The only Station behind a Tower goes offline
      Given Station S is attached to T and is T's only Station
      And S's node is registered on this instance and has been silent for at least nodeTTL
      When the canary sweep reaches T
      Then S is not in the canary's candidate set for T
      And no attempt is opened, no probe budget is spent, and no outcome is recorded for T or S
      And T's state does not change

    Scenario: An offline Station is left out and a live one is a candidate
      Given Stations S1 and S2 are attached to T
      And S1's node has been silent for at least nodeTTL and S2's node is heartbeating
      When the canary sweep reaches T
      Then only S2 is in the canary's candidate set for T

    Scenario: The fallback still covers heartbeating Stations that placement rejected
      Given three Stations behind T whose nodes are heartbeating Core
      But placement rejects all three for a reason other than liveness
      When the canary sweep reaches T 30 times
      Then every probe finds a candidate, all three Stations are probed, and none absorbs half the probes
      And each failure is recorded against T, as today

    Scenario Outline: The liveness predicate
      Given S's node is "<state>"
      Then S "<canary>" in the canary's candidate set for T
      And S "<placement>" offered to a consumer by edge placement

      Examples:
        | state                                                              | canary | placement |
        | registered here and heard from less than nodeTTL ago               | is     | is        |
        | registered here and silent for exactly nodeTTL                     | is not | is not    |
        | registered here and silent for longer than nodeTTL                 | is not | is not    |
        | registered on a peer, heartbeating there, mirrored here            | is     | is        |
        | unknown to this instance after its registry mirror (Tier B)        | is     | is        |
        | registered here, liveness mirror unreachable, peer heartbeats only | is not | is not    |

    Scenario: A broker that has just restarted does not judge Towers on an empty liveness view
      Given this broker instance started less than nodeTTL ago
      When its canary sweep reaches T
      Then no Station whose node it has not yet heard from is probed by it

    Scenario: A Tower cannot make a live node look offline
      Given S's node is heartbeating Core
      When T reports that nobody is serving S
      Then S stays in the candidate set and the failure is recorded against T

    Scenario: A Station that heartbeats but does not serve still costs its Tower
      Given S's node is heartbeating Core and S never answers through T
      When the canary sweep reaches T repeatedly
      Then failures are recorded against T, slowed by canaryTroubledSlowdown, as today

    Scenario: A Tower with every Station offline is still measured
      Given every Station behind T is offline by the liveness predicate
      Then T is measured by the Station-free hub check, as tower_trust_ramp.feature specifies

  Rule: A Tower suspended on evidence that has aged out can be restored

    Scenario: Restoring a Tower suspended only on switched-off-Station evidence
      Given T was suspended on canary failures that have all left the 24-hour reputation window
      And the fix is deployed
      When an administrator moves T to quarantine and then promotes it to active, with the founder's approval
      Then T holds a fresh lease and a live link and may take work
      And the next reputation evaluation does not suspend it again

  Rule: A refused Tower says why, stays up when it can recover, and stops when it cannot

    Scenario Outline: Core tells a Tower its own refusal state, and no one else
      Given T is refused with "<state>"
      When a request signed by T's own identity key is refused
      Then the refusal names "<state>"
      And a request that is unsigned, or signed by any other key, still gets today's generic refusal

      Examples:
        | state           |
        | state_suspended |
        | lease_lapsed    |
        | state_revoked   |

    Scenario Outline: A recoverable refusal keeps `roger tower serve` running
      Given `roger tower serve` meets "<refusal>" at "<step>"
      Then it prints the state in plain words, with what the operator can do, once per state
      And it stays running and asks Core again on the node re-attach backoff (30 s, capped at 15 min, jittered)
      And it does not exit and is not restarted by its service manager

      Examples:
        | refusal         | step                                  |
        | state_suspended | the first session open                |
        | lease_lapsed    | the first session open                |
        | lease_expired   | the first inventory push              |
        | state_suspended | the re-open after a refused heartbeat |
        | state_suspended | the full push after a re-open         |
        | Core unreachable | the first session open               |

    Scenario Outline: A terminal refusal stops `roger tower serve` with a status that is not restarted
      Given `roger tower serve` meets "<refusal>"
      Then it prints the reason in plain words and what the operator must do
      And it exits with the dedicated terminal status
      And the packaged service unit lists that status in RestartPreventExitStatus

      Examples:
        | refusal                              |
        | state_revoked                        |
        | cert_revoked                         |
        | unknown_tower                        |
        | a request signed by the wrong key    |
        | a joined Tower started in standalone mode, or the reverse |

    Scenario: A hub that cannot start is a terminal local error
      Given the hub cannot bind its address
      Then `roger tower serve` prints why and exits with the terminal status

    Scenario: A refused Tower that is restored carries traffic again without a restart
      Given `roger tower serve` is backing off after "state_suspended"
      When an administrator restores T and it holds a fresh lease
      Then the running process takes work again on its next retry

  Rule: Stopping a share takes its Stations out of rotation at once

    Scenario: Ctrl-C on `roger share` tells Core every Station is leaving
      Given S is attached to T at epoch 4
      When the operator stops `roger share` with Ctrl-C or SIGTERM
      Then before exiting, within 2 seconds, the share sends a leave naming S, T and epoch 4, signed by the node key of the live registration bound to S
      And Core sets S dormant, so no consumer and no canary is sent to it
      And every Station that node holds leaves the same way

    Scenario: A second Ctrl-C exits at once
      Given the leave is still in flight
      When the operator presses Ctrl-C again
      Then the share exits at once, and the liveness predicate covers S

    Scenario: A leave that fails is covered by the liveness predicate
      Given the leave cannot reach Core
      Then the share still exits within 2 seconds
      And S is out of rotation once its node has been silent for nodeTTL

    Scenario: A share killed without warning is covered by the liveness predicate
      Given S's share is killed with SIGKILL or loses power
      Then S is out of the canary's candidate set and out of placement once its node has been silent for nodeTTL

    Scenario Outline: A leave is applied only to the Station, Tower and epoch it names
      Given a leave for S at T and epoch 4
      When "<condition>"
      Then Core does not change S

      Examples:
        | condition                                                  |
        | S has since been revived at epoch 5                        |
        | S has since moved to Tower U                               |
        | the leave is signed by a key that is not S's node key      |
        | the leave is signed by the node key of a different node    |
        | the leave is a replay of one already applied               |

    Scenario: Leaves are rate limited with attaches
      When a node sends leaves beyond the per-account attach bucket
      Then the excess is refused with 429 and Retry-After

    Scenario: A leave during a move ends the move
      Given a move for S is pending and S is dormant waiting to re-attach
      When a valid leave for S arrives
      Then the move is ended and is never undone back to live by the 3-nudge rule

    Scenario: The next share revives the Station, possibly elsewhere
      Given S left with a leave
      When `roger share` starts again on that machine
      Then S is revived as a dormant revival is today, and may be placed on another Tower as station_mobility.feature allows
