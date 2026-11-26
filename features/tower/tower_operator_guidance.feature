# APPROVED SPEC - founder approved 2026-11-24 ("2. yes. approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: PARTIAL. Built: internal/towercli/guidance.go (init, register, help, serve and status
# words), cmd/rogerai-broker/admintowers.go (re-approval mark; state since and reason on /tower/status),
# towernotify.go (suspension email), web/app/routes/tower and manual (approval text). Not in this repo:
# the admin dashboard that renders the mark (private admin repo reads the "approval" field).
#
# Founder request 2026-11-24: "don't i have to approve each tower? make sure instructions are
# clear on the client side."
#
# Yes. Every joined Tower enrolls into QUARANTINE and carries no public traffic until a RogerAI
# administrator explicitly promotes it (public_enrollment.feature: quarantine -> active needs
# "completed admission probes and explicit promotion"). A suspended Tower returns to service only
# through quarantine and a second promotion (admit: suspended -> quarantine -> active), so it is
# approved again. A lapsed lease cannot be renewed and needs the same re-admission.
#
# approval_visibility.feature (approved 2026-08-23, PARTIAL) already makes serve announce the
# state and the approval. What an operator still cannot learn from the client:
#   - that approval is by RogerAI, per Tower, and again after a suspension or a lapse;
#   - what to do while waiting, and how to reach RogerAI with the right id;
#   - why a Tower was suspended, and that a suspended or lapsed Tower needs RogerAI, not a
#     restart (the house Tower restart-looped for 17 hours on 2026-11-24);
#   - that a standalone Tower needs no approval at all.
# Nora (the first outside operator) had to write in to ask most of these.
#
# Wording rules: plain English, short sentences, no em dashes, no pressure words. The tower id
# shown is always Core's id (tw-...), the one to quote. Contact is the Tower page,
# https://rogerai.fm/tower, so no mailbox is written into the binary.

Feature: A Tower operator always knows whether RogerAI must act, and what to do next

  Rule: The operator learns about approval before enrolling

    Scenario: Init and register say a joined Tower must be approved by RogerAI
      When the operator runs `roger tower init` for a joined Tower, and then `roger tower register`
      Then each says that RogerAI reviews and approves every joined Tower before it carries public traffic
      And register says the Tower is now waiting for that approval, names the tower id to quote, and how to check: `roger tower status`

    Scenario: A standalone Tower is told it needs no approval
      When the operator initializes a standalone Tower
      Then it says a standalone Tower serves its own network and needs no approval from RogerAI

    Scenario: The help text says it too
      When the operator runs `roger tower help`
      Then the joined-Tower lines say approval by RogerAI is required, per Tower

  Rule: Every state says who must act and what the operator can do

    Scenario Outline: serve and status describe each state the same way
      Given a joined Tower in state "<state>"
      When the operator runs `roger tower serve` or `roger tower status`
      Then the state line says "<meaning>"
      And the next-step line says "<next step>"
      And both commands use the same words

      Examples:
        | state        | meaning                                                   | next step                                                                                   |
        | quarantine   | waiting for RogerAI to approve this Tower                 | keep this running; approval switches it on by itself; questions: https://rogerai.fm/tower, quote the tower id |
        | active       | approved and carrying traffic                             | nothing to do                                                                               |
        | draining     | finishing current work, taking no new work                | nothing to do                                                                               |
        | suspended    | stopped by RogerAI, with the reason Core gives            | keep this running; RogerAI must review and approve it again; quote the tower id             |
        | lease_lapsed | its approval lapsed and cannot be renewed from here       | keep this running; RogerAI must re-admit and approve it again; quote the tower id           |
        | revoked      | permanently retired                                       | enroll a new Tower if you want to run one again                                             |

    Scenario: The suspension reason is shown when Core gives it
      Given Core suspended the Tower with reason "reputation"
      Then the suspended line names that reason in plain words
      And it does not show internal counters or evidence ids

    Scenario: The quarantine line is shown again after a restore
      Given an administrator moved a suspended Tower back to quarantine
      Then serve announces quarantine with the "waiting for RogerAI to approve" words, not as a fault

    Scenario: Status always says how long the Tower has been in its state
      Given a Tower has been in quarantine for 3 days
      When the operator runs status
      Then it says "in quarantine for 3 days"

  Rule: The admin side matches

    Scenario: Approving a restored Tower is the same one action as approving a new one
      Given a Tower was restored from suspended to quarantine
      Then it is listed with the pending Towers, marked "re-approval after suspension"

    Scenario: A suspension tells the admin why, so re-approval is an informed decision
      Given a Tower was suspended automatically
      Then the admin email and the Tower detail name the reason and the evidence counts

  Rule: The public docs say the same thing

    Scenario: The Tower page and the manual describe approval
      Then https://rogerai.fm/tower and the manual's Tower section say that every joined Tower is reviewed and approved by RogerAI, again after a suspension or a lapse, and that a standalone Tower needs no approval
