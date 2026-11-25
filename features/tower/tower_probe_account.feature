# APPROVED SPEC - founder approved 2026-11-24 ("approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: NOT BUILT. Approval is not implementation - this line says which.
#
# Found 2026-11-24 while finishing the house Tower test (tw-3081d0565b48999cdb34fbce):
# `roger tower probe --model wave-pico-293m --broker https://broker.rogerai.fm` answered
#   error: authorize failed: roger core answered 403: tower inference requires a signed-in
#   account that has accepted the terms of service
#
# Why: the probe signs with a fresh throwaway key on purpose (internal/towercli/probe.go: "a
# fresh key per run keeps a probe from being mistaken for one"), while Core serves Tower
# inference only to a key bound to a live, non-anonymized account (toweredge.go, the consent gate
# for charging money). The two were written to different rules, so the probe can never pass
# against a real Core. Its unit tests use a stub Core that does not check. Same class as the
# attach signer bug: client and broker disagreeing about which key signs.
#
# The same request signed with the operator's login key went through (attempt
# att-00d704ddb7e052dbf688283a: served 200, "receipt settled", recorded as self-dealing).
#
# Proposal: the probe signs with the operator's login key, the identity `roger` uses for every
# other consumer request. Core's consent gate is unchanged. Money facts this spec respects: the
# consumer pays in full even when Core records self-dealing (only the Station's share is
# withheld), and authorize is where the hold is placed.

Feature: roger tower probe runs as the operator's signed-in account

  Rule: The probe signs with the login key

    Scenario: A signed-in operator's probe goes through a live Tower
      Given an operator signed in with `roger login` or `roger tower login`
      And a live joined Tower carrying a Station for free model "m"
      When the operator runs `roger tower probe --model m`
      Then the probe is authorized, served sealed through a Tower's hub and acknowledged
      And it prints the Tower that served it and "edge path OK"

    Scenario: The probe uses the same key every other consumer request uses
      Then the probe's request signature is made with the key in user.key
      And it is never made with node.key or the Tower's identity key

    Scenario: An operator who is not signed in is told so before any network call
      Given no account is signed in on this machine
      When the operator runs `roger tower probe --model m`
      Then it exits non-zero saying "sign in first: run `roger login`"
      And no request reaches Core
      And no user.key is created as a side effect

    Scenario Outline: Core's refusals are shown in plain words
      Given "<state>"
      When the probe runs
      Then it prints "<plain words>" and does not print "edge path OK"

      Examples:
        | state                                            | plain words                                        |
        | the key in user.key is bound to no account       | this account cannot be served through a Tower      |
        | the account is banned                            | this account cannot be served through a Tower      |
        | the account is at its standing cap (429)         | too many Tower requests open; try again in N s     |
        | the account is rate limited (429)                | rate limited; try again in N s                     |

  Rule: The probe says what it costs, and never spends without asking

    Scenario: A probe of a free model costs nothing
      Given model "m" is free on every Station that carries it
      When the probe runs
      Then it authorizes with a price ceiling of 0 and prints that the attempt was free

    Scenario: A probe never spends money unless told to
      Given model "m" is priced
      When the operator runs the probe without --allow-charge
      Then it authorizes with a price ceiling of 0, so no priced Station is chosen and no hold is placed
      And it prints that "m" is priced and that --allow-charge lets the probe pay for one attempt

    Scenario: --allow-charge bounds what one probe can cost
      Given the operator runs the probe with --allow-charge and a --body file
      Then the probe prints the price per 1M tokens and the largest charge its body and output ceiling allow before it sends anything
      And the charge can never exceed that printed figure

    Scenario: A probe that lands on the operator's own Station says who pays
      Given the probe was served by a Station the operator's account owns
      Then the probe prints that the operator is charged as any consumer is, and that Core withholds the Station's share as self-dealing
      # The probe cannot read the settlement verdict itself; it states the rule.

  Rule: The probe can test one Tower

    Scenario: --tower restricts the probe to one Tower
      When the operator runs the probe with --tower tw-1
      Then the attempt is placed only on Tower tw-1, or refused saying tw-1 cannot serve "m" now

  Rule: The broker default matches the rest of roger

    Scenario: The probe uses the same broker roger does when none is given
      Given neither --broker nor $ROGER_BROKER is set
      When the operator runs the probe
      Then it uses the broker `roger` itself is configured for
      And it does not exit with "--broker or $ROGER_BROKER is required"

  Rule: The real client and the real broker agree

    Scenario: The probe passes end to end against a real broker
      Given a broker built from this tree with a live joined Tower and a Station
      And an operator signed in through the real login code
      When the real probe code runs
      Then it is authorized, served and acknowledged
      And on the previous probe code it fails with 403 "tower inference requires a signed-in account"
