# PROPOSED SPEC - awaiting founder approval. Nothing below is built.
# BUILD STATUS: NOT BUILT. Approval is not implementation - this line says which.
#
# Found 2026-11-24. The house Tower (tw-3081d0565b48999cdb34fbce) was suspended by Core and then
# lapsed, after doing nothing wrong:
#   05:16 UTC  a test share attached to it and served one request ("receipt settled")
#   ~05:18     the test share was stopped (Ctrl-C). Its Station stayed attached: stopping a
#              share does not detach it, and an attachment retires only after 7 days unseen.
#   05:31      Core's canary had probed that Station 5 times; every probe failed, because
#              nothing was serving it. Core suspended the Tower:
#              "lifecycle from=active to=suspended reason=reputation canary_fail=5 canary_pass=0"
#   05:32 on   `roger tower serve` exited on every refused request and systemd restarted it
#              every 15 seconds, for 17 hours.
#   16:17      the lease ran out (it cannot renew while suspended): "reason=lease_lapsed".
#
# Why: the canary blames the Tower for every failure on purpose, because a hostile Tower could
# fake any of them (towercanary.go). It picks its target from the Stations placement would use,
# and placement already drops a node Core has itself stopped hearing from ("the share went
# home; the attachment simply has not noticed yet", toweredge.go edgeEligibleM). But when
# no Station is eligible, canaryTargetFor falls back to "everything that resolved", which puts
# the dead Station back in. With one Station behind a Tower, every probe hit it.
#
# Core does not need the Tower's word to know a node is offline: the node registers with and
# heartbeats Core directly. A Tower cannot make Core see a node as offline, and cannot make a
# heartbeating node look offline, so skipping such nodes hands a hostile Tower nothing. It also
# closes the attack the canary comments already worry about: attaching Stations that do not
# serve, to take an honest Tower off the fabric.
#
# The same thing would happen to any public Tower, Nora's included, the first time a Station
# behind it is switched off.

Feature: A Tower is never judged on a Station that Core itself knows is offline

  Background:
    Given a live joined Tower T that may take work

  Rule: The canary probes only nodes Core is hearing from

    Scenario: The only Station behind a Tower goes offline
      Given Station S is attached to T and is T's only Station
      And S's node stopped heartbeating Core longer ago than the node liveness window
      When the canary sweep reaches T
      Then no probe is sent and no outcome is recorded for T or S
      And T's state does not change

    Scenario: An offline Station is skipped and a live one is probed
      Given Stations S1 and S2 are attached to T
      And S1's node is offline by Core's heartbeat and S2's node is heartbeating
      When the canary sweep reaches T
      Then only S2 is probed

    Scenario: The fallback still probes a heartbeating node that placement rejected
      Given S's node is heartbeating Core but placement rejects S for another reason
      When the canary sweep reaches T and no Station is eligible
      Then S is probed, as today
      And a failure is recorded against T, as today

    Scenario Outline: What counts as "Core is hearing from the node"
      Given S's node is "<state>"
      Then the canary "<does>" probe S

      Examples:
        | state                                                        | does     |
        | registered and heartbeating within the liveness window       | may      |
        | registered and silent past the liveness window               | does not |
        | registered on a peer instance and heartbeating there         | may      |
        | unknown to this instance and to the shared registry          | does not |

    Scenario: A Tower cannot make a live node look offline
      Given S's node is heartbeating Core
      When T reports that nobody is serving S
      Then S is still probed and the failure is recorded against T

  Rule: A Tower already wronged this way is not held to that evidence

    Scenario: Probe failures recorded while the node was offline do not count
      Given T has canary failures recorded against a Station whose node Core saw as offline when each probe ran
      When T's reputation is evaluated
      Then those failures are left out of T's tally
      # Needs the liveness reading stored with each canary outcome. Outcomes recorded before
      # this change carry none and count as today; they age out of the 24-hour window.

    Scenario: The house Tower is restored after the fix ships
      Given the fix is deployed
      And the evidence that suspended tw-3081d0565b48999cdb34fbce has left the 24-hour window
      When an administrator restores it with the founder's approval
      Then it renews, carries a Station, and is not suspended again by the old evidence

  Rule: A refused Tower stays up and says why, instead of restarting in a loop

    Scenario Outline: `roger tower serve` meets a refusal it cannot fix by itself
      Given Core refuses T with "<reason>"
      When `roger tower serve` is running
      Then it prints the state and the reason once, in plain words, with what the operator can do
      And it stays running and asks Core again with backoff, starting at 30 seconds and capped at 10 minutes
      And it does not exit and is not restarted by its service manager

      Examples:
        | reason          |
        | state_suspended |
        | lease_lapsed    |
        | state_revoked   |

    Scenario: A refused Tower that is restored carries traffic again without a restart
      Given `roger tower serve` is backing off after "state_suspended"
      When an administrator restores T and its lease is renewed
      Then the running process takes work again on its next retry

  Rule: Stopping a share takes its Station out of rotation at once

    Scenario: Ctrl-C on `roger share` tells Core the Station is leaving
      Given S is attached to T
      When the operator stops `roger share` with Ctrl-C or SIGTERM
      Then the share tells Core, signed by its node key, that S is leaving
      And S is set dormant, so no consumer and no canary is sent to it
      And the next `roger share` on that machine revives S, as a dormant revival does today

    Scenario: A share killed without warning is caught by the liveness window
      Given S's share is killed with SIGKILL or loses power
      Then S is treated as offline once its node is silent past the liveness window, as in the first rule

    Scenario: A leave message for someone else's Station is refused
      When a leave for S arrives signed by a key that is not S's node key
      Then it is refused and S is unchanged
