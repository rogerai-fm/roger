# APPROVED SPEC - founder approved 2026-11-24 ("approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: PARTIAL. Built: cmd/rogerai-broker/tower.go (brokerOperatorPolicy, towerOwnerBanned,
# keyAccountBanned, mayRunTower, suspendTowersOfBanned, sweepBannedTowersIfDue; checks at token and
# challenge, and a re-check after admission), strikes.go (banOwner suspends), report.go and tunnel.go
# (the ban-sync tick re-runs suspension), toweredgeattach.go (attach uses the same helper),
# internal/towercore/enroll (the one enroll-time ban check; spends a ban-refused token),
# internal/towercore/admit (SpendToken, AllTowers). Not built: moving a suspended Tower's Stations
# under R4, which waits on station_mobility_idempotency.feature R4.
#
# Build notes (2026-11-24): an account that cannot be checked is refused with 503, never granted.
# Expired Towers cannot be suspended and are skipped; any future re-admission path must check bans
# itself. A Tower enrolled under a login the account has since renamed cannot be traced to the
# account: the store keeps no record of earlier logins.
#
# Found 2026-11-23 while fixing the Station attach signer (station_attach_signer.feature).
#
# Bans are recorded under an account KEY: banOwner(acct) in strikes.go, where acct is the owner
# key bound to a node (AccountOfNode) or, from toweraudit.go, the attachment's canonical account
# key. Tower enrollment checks the ban with MayEnroll(owner) twice: at POST /tower/token
# (tower.go) and again at enroll (towerenroll.go, MayEnroll(tok.Owner)). In both, owner is the
# account's LOGIN when it has one (towerOperator). A login never equals a key, so for an account
# with a login, neither check ever sees a ban, and a banned operator can enroll a Tower. For an
# account with no login (Apple, some email accounts) owner is the signing key, so a ban under
# that key is seen, but a ban under the account's canonical key is not. The challenge step checks
# no ban at all. The attach path had the same bug and was fixed by also checking the signer's key
# and the account's canonical key (PR #24, live 2026-11-23).
#
# No code today checks a Tower OPERATOR's ban after enrollment (brokerOwners.Suspended covers
# Station owners only).
#
# A wider gap, left for a ruling: a ban under one device key does not cover an account's other
# device keys anywhere (tunnel.go, toweredge.go and this path each check one key).

Feature: A banned account cannot enroll or keep a Tower

  Rule: Every enrollment step checks the ban under the keys bans are recorded under

    Scenario Outline: A ban refuses an enrollment token
      Given account A "<login>" with canonical key D1 and a second device key D2
      And A is banned under "<banned>"
      When A asks for a Tower enrollment token signed by "<signer>"
      Then the answer is 403 "this account may not run a Tower"
      And no token is issued

      Examples:
        | login            | banned | signer |
        | has login "op"   | D1     | D1     |
        | has login "op"   | D1     | D2     |
        | has login "op"   | D2     | D2     |
        | has no login     | D1     | D2     |
        | has no login     | D2     | D2     |

    Scenario: A ban under a device key that is neither the signer nor canonical
      Given A has login "op", canonical key D1 and device key D2
      And A is banned under D2 only
      When A asks for a token signed by D1
      Then a token is issued
      # Written down on purpose: this is the case the per-account ruling below would change.

    Scenario: An account with no ban enrolls as before
      Given account A is not banned under any of its keys
      When A asks for a token
      Then a token is issued, as today

    Scenario: A renamed login does not escape a ban
      Given A is banned under D1 and then renames its login
      When A asks for a token
      Then the answer is 403

    Scenario Outline: A ban that lands mid-enrollment stops it
      Given A holds an unused enrollment token
      When A is banned "<when>"
      Then enrollment is refused with 403 and no Tower is admitted

      Examples:
        | when                                     |
        | before the challenge is requested        |
        | between the challenge and the enrollment |

    Scenario: A token refused for a ban is spent
      Given A holds an unused token and is banned
      When the enrollment is refused for the ban
      Then that token can never be used again
      # NEW BEHAVIOUR: today a policy refusal leaves the token usable.

    Scenario: A peer instance sees the ban from the store, not only its memory
      Given A was banned on instance 1 less than 15 seconds ago
      When A asks instance 2 for a token
      Then instance 2 refuses it

  Rule: A joined Tower whose owner is banned stops carrying traffic

    Scenario: A ban on a Tower owner suspends their Tower
      Given account A runs joined Tower T and is banned under its signing or canonical key
      Then within 30 seconds T may not take work, on every broker instance
      And T's Stations are moved under station_mobility_idempotency.feature R4 with reason "ineligible:banned"
      # NEW BEHAVIOUR: today no check runs on a Tower operator after enrollment.

  Rule: Refusals keep today's words

    Scenario: A missing account and a banned account get today's answers
      Then a request with no signed-in account gets 401 "running a Tower requires a signed-in account - try `roger tower login`"
      And a banned account gets 403 "this account may not run a Tower"
      And neither says under which key the ban was recorded

  # RULING NEEDED (not part of this spec until ruled): should a ban under ANY device key of an
  # account cover ALL its device keys everywhere (serving, Tower inference, attach, enroll)? If
  # yes, the "neither the signer nor canonical" scenario above flips to 403.
