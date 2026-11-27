# APPROVED by the founder, 2026-11-23 ("yes to all, go ahead"): rulings P1 to P16 below.
# Do not edit a scenario or weaken a step without the founder approving the change.
# Changed 2026-11-23 with founder approval: three scenarios whose rule setup also gives frank
# the $0.05 "friends" grant now remove it first, so the scenario shows what it names.
#
# PRICED GRANTS: THE PERSON USING A PRICED GRANT PAYS FOR IT.
#
# WHY. Founder, 2026-11-23: a grant gives someone access to a model you share, free or at a
# price you set just for that grant, which may be lower or higher than your public price.
#   - A free grant works for whoever holds the secret, with no login.
#   - A priced grant works for whoever holds the secret AND has a logged-in account with a
#     balance. That account pays. The grant is not tied to one named person.
#
# WHAT THE CODE DOES TODAY (cmd/rogerai-broker/grant.go, resolvePricing):
#   - A priced grant bills the OWNER's own wallet ("sponsorship"). The friend never pays.
#   - A request carrying a grant secret skips caller identity: one credential per request.
#   - No live grant depends on the owner-pays rule (checked 2026-11-23).
#
# GROUND TRUTH THIS FILE BUILDS ON:
#   - store.Grant and Grant.GrantPrice(): the one price source every settle path reads.
#   - features/grants/grant_price_safety.feature: a grant price is never negative.
#   - registerPriceCeiling (cmd/rogerai-broker/pricesafety.go): the public price ceiling.
#   - Account wallets, account keys (rog-key_...), key limits and the account monthly cap:
#     features/auth/key_guardrails.feature and features/relay/key_limits.feature.
#   - Hold before work, settle after: the same flow as a public request.
#
# RULINGS ALREADY GIVEN (founder, 2026-11-23):
#   P1  The caller pays for a priced grant, from their own account wallet.
#   P2  Any logged-in account with a balance may use a priced grant it holds the secret for.
#       No email or named user is attached to a grant.
#   P3  A friend adds the grant to their account once. After that their login and their
#       account keys get the grant's price on the owner's stations.
#   P4  The owner does not know who the friends are, so a price RAISE replaces the secret.
#       The old secret stops working and everyone adds the new one, which shows them the new
#       price. A price CUT applies at once on the same secret.
#
# RULINGS PROPOSED BY THE AGENT AND APPROVED (founder, 2026-11-23):
#   P5  A priced grant sent as a bare API key, with no account, is refused. It never falls
#       back to the public price and never bills the owner.
#   P6  The price is shown when a grant is added, before anything is charged.
#   P7  (founder, 2026-11-23, replaces the first draft's "lower of the two" rule) The grant
#       price applies as the owner set it, whether it is below or above the station's public
#       price. The owner may be giving something extra for the higher price. The price is
#       shown when the grant is added, and the caller can remove the grant at any time.
#   P7a When the grant price is above the station's public price, the add screen
#       says so next to the price.
#   P7b Two added grants that cover the same station and model give the lower of
#       the two grant prices.
#   P7c Wherever the router compares stations by price, it uses the price this
#       caller would really pay, so a station with a dearer grant ranks as dearer.
#   P8  The caller's own limits still apply: their maximum price, their account key limit and
#       their account monthly cap all count grant spend.
#   P9  The owner's own use of their own priced grant is $0, as self-use is today.
#   P10 The marketplace fee, holds and payout rules apply to grant earnings exactly as they do
#       to public earnings. A priced grant is not a way around the fee.
#   P11 The owner sees how many accounts added a grant, never who they are.
#   P12 (approved by the founder, 2026-11-23) A revoked or expired grant stops giving its price at once. A request from an account
#       that added it is then routed and priced as an ordinary public request, and the reply
#       says the grant no longer applies. Nobody is charged a price they did not see: the
#       caller's maximum price still bounds it.
#   P13 The "owner sponsors a priced grant" rule is removed. A grant stored as not free with a
#       zero price is treated and listed as free.
#       (Header wording only, 2026-11-25: production counts removed from this public file; the
#       rules are unchanged.)
#   P14 A free grant is not added to an account. It is used as the API key, as today.
#   P15 A grant's token caps count every caller's use of it together.
#   P19 (founder, 2026-11-24, "1. yes") A grant at its token cap gives no price. A caller who
#       named the station is refused; any other caller is routed as if they had not added it,
#       at the public price. A cap that cannot be read counts as reached.
#       Note 2026-11-24 (founder, "5. ok"): "named" is judged on the first pick. A pinned
#       station, or a station in the caller's order or only list that the first pick lands on,
#       is refused. A named station reached later by failover is served at the public price.
#       Neither ever bills the grant price.
#   P20 (founder, 2026-11-24, "4. ok go with the public price") When the grants an account
#       added cannot be read, the request is priced as public and the reply carries
#       X-RogerAI-Grant-Notice "your added grants could not be read - you are charged the
#       public price", so the caller is never charged more than they were shown in silence.
#   P21 (founder, 2026-11-26, "approved") A priced grant an account added prices chat only.
#       Voice and search are billed in other units (characters, audio bytes, queries), so
#       those requests pay the public price, and the add never lists a voice or search offer.
#       Per-kind grant prices are a later spec.
#   P17 (founder, 2026-11-23, "do the recommendation") A priced grant never reaches further than
#       what the caller was shown. Adding stations or models to a priced grant is a reprice with
#       a new secret, like a raise. Narrowing applies at once. A free grant may still be widened
#       by an edit, since it shows no price.
#   P18 (founder, 2026-11-23, same answer) A priced grant with no station list covers the
#       owner's stations as they were when the caller added it. A station the owner brings on
#       air later is charged at its public price until the caller adds the grant again, which
#       shows the new list.
#   P16 A grant does not open a hidden-band station. Checked in the code on
#       2026-11-23: the router drops a hidden station unless the request carries its
#       frequency code (pickFor, cmd/rogerai-broker/tunnel.go), and a grant only narrows the
#       station list, it never widens it. This file keeps that. A grant covers the owner's
#       public stations, and a hidden one still needs its frequency code as well.
#
# NOT IN THIS FILE:
#   - What the daily token cap default should be. Still open.
#   - Voice (TTS) grants. Token caps do not count voice use today.
#   - The create and add screens on each surface. Each surface's own spec covers its screen
#     once these rules are approved.

Feature: Priced grants are paid by the person using them
  An owner can give friends or customers their own price on the owner's stations. Whoever
  uses that price pays it from their own wallet, and sees it before they pay.

  Background:
    Given an owner "olive" with a station "olive-box" serving "qwen-small"
    And "olive-box" is public at $0.10 per 1M output tokens
    And a caller "frank" who is logged in with a balance of $5.00

  # ---------------------------------------------------------------------------------------
  # Making a priced grant
  # ---------------------------------------------------------------------------------------

  Rule: An owner sets a grant's price when making it, and free stays the default

    Scenario: A grant made with no price is free
      When olive makes a grant named "family"
      Then the grant is free
      And the list shows "family" as "free"

    Scenario: A grant made with a price is priced
      When olive makes a grant named "friends" at $0.05 per 1M output tokens
      Then the grant is priced at $0.05 per 1M output tokens
      And the secret is shown once

    Scenario Outline: A price of zero is a free grant
      When olive makes a grant named "friends" with output price "<typed>"
      Then the grant is free
      And the list shows "friends" as "free"

      Examples:
        | typed |
        | 0     |
        | $0    |
        | 0.00  |
        | free  |

    Scenario Outline: A price that cannot be read is refused and nothing is made
      When olive makes a grant named "friends" with output price "<typed>"
      Then the request is refused with "<message>"
      And no grant is created

      Examples:
        | typed  | message                              |
        | -1     | price cannot be negative             |
        | cheap  | state a price in dollars per 1M tokens, or "free" |
        | 1e9    | state a price in dollars per 1M tokens, or "free" |
        | NaN    | state a price in dollars per 1M tokens, or "free" |

    Scenario: A price above the public ceiling is refused
      When olive makes a grant named "friends" above the public output price ceiling
      Then the request is refused naming the ceiling
      And no grant is created

    Scenario: Asking for both free and a price is refused
      When olive makes a grant named "friends" marked free with a price of $0.05
      Then the request is refused with "a grant is free or priced, not both"
      And no grant is created

  # ---------------------------------------------------------------------------------------
  # Adding a grant to an account
  # ---------------------------------------------------------------------------------------

  Rule: A caller adds a priced grant to their account once, and sees the price first

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens

    Scenario: Adding shows the price, the stations and the models before anything is charged
      When frank adds the "friends" secret to his account
      Then the reply names the grant "friends"
      And the reply shows the price $0.05 per 1M output tokens
      And the reply lists the station "olive-box" and the model "qwen-small"
      And frank's balance is still $5.00

    Scenario: Adding needs a logged-in account
      When someone who is not logged in tries to add the "friends" secret
      Then the request is refused with "log in to add a grant"

    Scenario: Adding does not need a balance, using does
      Given a caller "nina" who is logged in with a balance of $0.00
      When nina adds the "friends" secret to her account
      Then the grant is added
      When nina asks "olive-box" for "qwen-small"
      Then the request is refused with "insufficient balance - add funds"

    Scenario: Any number of accounts can add the same grant
      Given a caller "gina" who is logged in with a balance of $5.00
      When frank adds the "friends" secret to his account
      And gina adds the "friends" secret to her account
      Then both accounts have the grant
      And olive sees that 2 accounts added "friends"
      And olive is shown no account name, email or id for either

    Scenario: Adding the same grant twice changes nothing
      Given frank added the "friends" secret to his account
      When frank adds the "friends" secret to his account
      Then frank has the grant once
      And olive sees that 1 account added "friends"

    Scenario Outline: A secret that cannot be added is refused the same way
      When frank adds <secret> to his account
      Then the request is refused with "<message>"
      And frank's grants are unchanged

      Examples:
        | secret                         | message                                        |
        | a secret that never existed    | grant key invalid or revoked                   |
        | a revoked grant's secret       | grant key invalid or revoked                   |
        | an expired grant's secret      | grant key expired                              |
        | a free grant's secret          | this grant is free - use it as your API key    |
        | an account key                 | that is an account key, not a grant            |
        | an empty value                 | paste the grant secret                         |

    Scenario: Guessing secrets is slowed down
      When frank tries to add 20 secrets that do not exist within a minute
      Then further tries are refused for a while with a retry time
      And a real secret tried after the wait is accepted

    Scenario: An account can hold only so many grants
      Given frank has added the most grants an account may hold
      When frank adds the "friends" secret to his account
      Then the request is refused naming the limit and how to remove one

    Scenario: A caller can see and remove the grants they added
      Given frank added the "friends" secret to his account
      When frank lists the grants he added
      Then the list shows "friends" at $0.05 per 1M output tokens on "olive-box"
      And the list never shows the secret
      When frank removes "friends"
      Then frank's next request to "olive-box" is priced at the public price

    Scenario: The owner adding their own grant is refused
      When olive adds the "friends" secret to her own account
      Then the request is refused with "this is your own grant - your own stations are already free to you"

  # ---------------------------------------------------------------------------------------
  # Using a priced grant
  # ---------------------------------------------------------------------------------------

  Rule: The caller who added a priced grant pays its price on the owner's stations

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account

    Scenario Outline: The grant price applies however frank signs in
      When frank asks "olive-box" for "qwen-small" using <credential>
      Then the request is served
      And frank is charged at $0.05 per 1M output tokens
      And the reply's cost matches what frank was charged

      Examples:
        | credential               |
        | his login                |
        | his browser session      |
        | one of his account keys  |

    Scenario: The owner is paid, less the marketplace fee
      When frank asks "olive-box" for "qwen-small" and 1,000,000 output tokens are served
      Then frank's wallet is debited $0.05
      And olive is credited $0.05 less the marketplace fee
      And olive's wallet is not debited

    Scenario: The owner's wallet is never charged for a friend's use
      Given olive's balance is $0.00
      When frank asks "olive-box" for "qwen-small"
      Then the request is served
      And olive's balance is still $0.00 before her earnings land

    Scenario: A hold is placed before any work
      Given frank's balance covers less than the most this request could cost
      When frank asks "olive-box" for "qwen-small"
      Then the request is refused with "insufficient balance - add funds"
      And "olive-box" is not asked

    Scenario: Many requests at once cannot overdraw the caller
      Given frank's balance covers 2 requests
      When frank sends 10 requests at once to "olive-box"
      Then at most 2 are served
      And frank's balance never goes below $0.00

    Scenario: The grant price applies only inside the grant's scope
      Given olive made a grant "one-model" at $0.05 limited to the model "qwen-small"
      And frank added the "one-model" secret to his account
      And frank removes "friends" from his account
      When frank asks "olive-box" for a different model it serves
      Then frank is charged the public price for that model

    Scenario: The grant gives no price on another owner's station
      Given another owner's station "pat-box" serves "qwen-small" in public
      When frank asks "pat-box" for "qwen-small"
      Then frank is charged "pat-box"'s public price

    Scenario: A grant alone does not open a hidden-band station
      Given olive has a station "olive-hidden" on a hidden band
      And the grant "friends" covers every station olive owns
      When frank asks "olive-hidden" for "qwen-small" without its frequency code
      Then the request is refused as the station cannot be reached
      And frank is charged nothing

    # 2026-11-24 (founder: "ok"; hidden stations are not listed, but the frequency code reaches
    # them): under P18 a station that came on air after the add is not covered, so frank adds
    # the grant again once the hidden station is on air.
    Scenario: A hidden-band station gives the grant price to a caller who also has its code
      Given olive has a station "olive-hidden" on a hidden band
      And the grant "friends" covers every station olive owns
      And frank adds the "friends" secret to his account
      When frank asks "olive-hidden" for "qwen-small" with its frequency code
      Then the request is served at the grant price

  Rule: The grant price applies as the owner set it, below or above the public price

    Scenario: A grant below the public price gives the grant price
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When frank asks "olive-box" for "qwen-small"
      Then frank is charged at $0.05 per 1M output tokens

    Scenario: A grant above the public price gives the grant price
      Given olive made a grant "premium" at $0.20 per 1M output tokens
      And frank added the "premium" secret to his account
      When frank asks "olive-box" for "qwen-small"
      Then frank is charged at $0.20 per 1M output tokens

    Scenario: A grant equal to the public price gives that price
      Given olive made a grant "same" at $0.10 per 1M output tokens
      And frank added the "same" secret to his account
      When frank asks "olive-box" for "qwen-small"
      Then frank is charged at $0.10 per 1M output tokens

    Scenario: Adding a grant priced above the public price says so
      Given olive made a grant "premium" at $0.20 per 1M output tokens
      When frank adds the "premium" secret to his account
      Then the reply shows the grant price of $0.20 per 1M output tokens
      And the reply says this is above the public price of $0.10 on "olive-box"
      And frank is charged nothing

    Scenario: Removing a dearer grant returns the caller to the public price
      Given olive made a grant "premium" at $0.20 per 1M output tokens
      And frank added the "premium" secret to his account
      When frank removes "premium" from his account
      And frank asks "olive-box" for "qwen-small"
      Then frank is charged at $0.10 per 1M output tokens

    Scenario: A dearer grant makes the station rank as dearer for that caller
      Given another owner's station "pat-box" serves "qwen-small" in public at $0.15 per 1M output tokens
      And olive made a grant "premium" at $0.20 per 1M output tokens
      And frank added the "premium" secret to his account
      When frank asks for "qwen-small" sorted by price without naming a station
      Then "pat-box" is ranked ahead of "olive-box"

    Scenario: A cheaper grant makes the station rank as cheaper for that caller
      Given another owner's station "pat-box" serves "qwen-small" in public at $0.08 per 1M output tokens
      And olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When frank asks for "qwen-small" sorted by price without naming a station
      Then "olive-box" is ranked ahead of "pat-box"

    Scenario: Two grants on the same station give the lower price
      Given olive made a grant "friends" at $0.05 and a grant "club" at $0.03
      And frank added both secrets to his account
      When frank asks "olive-box" for "qwen-small"
      Then frank is charged at $0.03 per 1M output tokens

  Rule: The caller's own limits bound grant spend

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account

    Scenario: A price above the caller's maximum is not used
      Given olive made a grant "premium" at $0.20 per 1M output tokens
      And frank added the "premium" secret to his account
      And frank removes "friends" from his account
      And frank's maximum output price is $0.10 per 1M
      When frank asks "olive-box" for "qwen-small"
      Then the request is refused naming frank's maximum price
      And frank is charged nothing

    Scenario: An account key's limit counts grant spend
      Given frank has an account key with a limit of $1.00 weekly that has spent $1.00
      When frank asks "olive-box" for "qwen-small" using that key
      Then the request is refused as the key's limit is reached
      And frank is charged nothing

    Scenario: The account monthly cap counts grant spend
      Given frank's account monthly cap is reached
      When frank asks "olive-box" for "qwen-small"
      Then the request is refused as the monthly cap is reached

    Scenario: The grant's token caps count every caller together
      Given the grant "friends" has a daily cap of 1,000 tokens
      And another account that added "friends" has used 1,000 tokens today
      When frank asks "olive-box" for "qwen-small"
      Then the request is refused with "grant daily token cap reached"
      And frank is charged nothing

    # Added 2026-11-24 (founder: "1. yes", ruling P19). The scenario above names the station.
    Scenario: A grant at its cap gives the public price to a caller who named no station
      Given the grant "friends" has a daily cap of 1,000 tokens
      And another account that added "friends" has used 1,000 tokens today
      When frank asks for "qwen-small" without naming a station
      Then frank is charged "olive-box"'s public price

  Rule: A priced grant never works as a bare API key

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens

    Scenario: A priced grant sent as the API key is refused
      When a request arrives with the "friends" secret as its API key and no account
      Then the request is refused with "this grant is priced - log in and add it to your account"
      And nobody is charged
      And "olive-box" is not asked

    Scenario: A priced grant sent as the API key by a logged-in caller is refused too
      When frank sends the "friends" secret as his API key alongside his login
      Then the request is refused with "this grant is priced - log in and add it to your account"
      And frank is charged nothing

    Scenario: A free grant still works as a bare API key
      Given olive made a free grant "family"
      When a request arrives with the "family" secret as its API key and no account
      Then the request is served
      And nobody is charged

  Rule: The owner's own use is free

    Scenario: The owner using a station through their own priced grant pays nothing
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      When olive asks "olive-box" for "qwen-small" while logged in
      Then the request is served
      And olive is charged nothing

  # ---------------------------------------------------------------------------------------
  # Changing a price
  # ---------------------------------------------------------------------------------------

  Rule: A price cut applies at once on the same secret

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account

    Scenario: A cut reaches the caller's next request
      When olive lowers "friends" to $0.02 per 1M output tokens
      Then the secret is unchanged
      And frank's next request to "olive-box" is charged at $0.02 per 1M output tokens

    Scenario: A request already running settles at the price it started at
      Given frank has a request to "olive-box" in flight at $0.05
      When olive lowers "friends" to $0.02 per 1M output tokens
      Then the request in flight is charged at $0.05 per 1M output tokens

    Scenario: Making a priced grant free is a cut
      When olive makes "friends" free
      Then the secret is unchanged
      And frank's next request to "olive-box" is charged nothing
      And the "friends" secret now works as a bare API key

  Rule: A price raise replaces the secret

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account

    Scenario: A raise is refused as an edit and names the way to do it
      When olive edits "friends" to $0.08 per 1M output tokens
      Then the request is refused with "a higher price needs a new secret - reprice the grant"
      And the price of "friends" is still $0.05

    Scenario: Repricing issues a new secret and ends the old one
      When olive reprices "friends" to $0.08 per 1M output tokens
      Then a new secret is shown once
      And the old secret can no longer be added
      And frank no longer has the grant
      And olive sees that 0 accounts added "friends"

    Scenario: A caller is never charged the raised price without adding the new secret
      When olive reprices "friends" to $0.08 per 1M output tokens
      And frank asks "olive-box" for "qwen-small"
      Then frank is charged at the public price of $0.10, bounded by his maximum price
      And the reply says the grant "friends" no longer applies

    Scenario: Adding the new secret shows the new price
      Given olive repriced "friends" to $0.08 per 1M output tokens
      When frank adds the new "friends" secret to his account
      Then the reply shows the price $0.08 per 1M output tokens
      And frank's next request to "olive-box" is charged at $0.08 per 1M output tokens

    Scenario Outline: What counts as a raise
      Given the grant "friends" is priced at <before>
      When olive edits it to <after>
      Then the edit is <outcome>

      Examples:
        | before                 | after                  | outcome                      |
        | $0.05 out              | $0.02 out              | applied at once              |
        | $0.05 out              | $0.05 out              | applied at once              |
        | $0.05 out              | $0.08 out              | refused, needs a new secret  |
        | $0.05 out and $0.01 in | $0.02 out and $0.02 in | refused, needs a new secret  |
        | free                   | $0.01 out              | refused, needs a new secret  |
        | $0.05 out              | free                   | applied at once              |

    Scenario: The grant keeps its name, scope, caps and usage across a reprice
      Given the grant "friends" has a daily cap, a model list and recorded usage
      When olive reprices "friends" to $0.08 per 1M output tokens
      Then the name, the caps, the model list and the usage are unchanged

  # ---------------------------------------------------------------------------------------
  # Scope: what a priced grant reaches (P17, P18)
  # ---------------------------------------------------------------------------------------

  Rule: A priced grant never reaches a station or model the caller was not shown

    Scenario: A station olive brings on air after frank added the grant is charged at its public price
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When olive brings a new station "olive-two" on air serving "qwen-small"
      And frank asks "olive-two" for "qwen-small"
      Then frank is charged at $0.10 per 1M output tokens
      And frank's next request to "olive-box" is charged at $0.05 per 1M output tokens

    Scenario: Adding the grant again brings in the stations olive added since, and shows them
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      And olive brings a new station "olive-two" on air serving "qwen-small"
      When frank adds the "friends" secret to his account
      Then the reply lists the station "olive-two" and the model "qwen-small"
      And frank's next request to "olive-two" is charged at $0.05 per 1M output tokens

    Scenario: A caller who adds the grant after a new station came on air gets the grant price there
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And olive brings a new station "olive-two" on air serving "qwen-small"
      When frank adds the "friends" secret to his account
      Then the reply lists the station "olive-two" and the model "qwen-small"
      And frank's next request to "olive-two" is charged at $0.05 per 1M output tokens

    Scenario: A model olive starts serving after frank added the grant is charged at its public price
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When frank asks "olive-box" for a different model it serves
      Then frank is charged the public price for that model

    Scenario: A station that was off air when frank added the grant is charged at its public price
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And "olive-two" goes off air
      And olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When "olive-two" comes back on air
      And frank asks "olive-two" for "qwen-small"
      Then frank is charged at $0.10 per 1M output tokens

    Scenario: Narrowing a priced grant applies to frank's next request at once
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And olive made a grant "scoped" at $0.05 per 1M output tokens covering the stations "olive-box" and "olive-two"
      And frank added the "scoped" secret to his account
      When olive edits "scoped" to cover the station "olive-box"
      Then the edit is applied at once
      And frank's next request to "olive-two" is charged at $0.10 per 1M output tokens
      And frank's next request to "olive-box" is charged at $0.05 per 1M output tokens

    Scenario: A reprice must state the new price
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      When olive reprices "friends" naming only the stations "olive-box"
      Then the request is refused with "state a price in dollars per 1M tokens, or "free""
      And the price of "friends" is still $0.05

    Scenario: Adding a station to a priced grant is refused as an edit and names the way to do it
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And olive made a grant "scoped" at $0.05 per 1M output tokens covering the station "olive-box"
      When olive edits "scoped" to cover the stations "olive-box" and "olive-two"
      Then the request is refused with "a wider grant needs a new secret - reprice the grant"
      And "scoped" still covers only the station "olive-box"

    Scenario Outline: What counts as widening a priced grant
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And olive made a grant "scoped" at $0.05 per 1M output tokens covering <before>
      When olive edits "scoped" to cover <after>
      Then the edit is <outcome>

      Examples:
        | before                                     | after                                      | outcome                     |
        | the station "olive-box"                    | the stations "olive-box" and "olive-two"   | refused, needs a new secret |
        | the stations "olive-box" and "olive-two"   | the station "olive-box"                    | applied at once             |
        | the station "olive-box"                    | every station                              | refused, needs a new secret |
        | every station                              | the station "olive-box"                    | applied at once             |
        | the model "qwen-small"                     | the models "qwen-small" and "qwen-big"     | refused, needs a new secret |
        | the models "qwen-small" and "qwen-big"     | the model "qwen-small"                     | applied at once             |
        | the model "qwen-small"                     | every model                                | refused, needs a new secret |
        | every model                                | the model "qwen-small"                     | applied at once             |

    Scenario: Repricing can widen a priced grant under a new secret
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And olive made a grant "scoped" at $0.05 per 1M output tokens covering the station "olive-box"
      And frank added the "scoped" secret to his account
      When olive reprices "scoped" to $0.05 per 1M output tokens covering the stations "olive-box" and "olive-two"
      Then a new secret is shown once
      And frank no longer has the grant
      And "scoped" covers the stations "olive-box" and "olive-two"

    Scenario: A free grant can still be widened by an edit
      Given olive brings a new station "olive-two" on air serving "qwen-small"
      And olive made a free grant "open" covering the station "olive-box"
      When olive edits "open" to cover the stations "olive-box" and "olive-two"
      Then the edit is applied at once

  # ---------------------------------------------------------------------------------------
  # Revoke and expiry
  # ---------------------------------------------------------------------------------------

  Rule: An added grant prices chat only (P21)

    Scenario: Adding a priced grant lists chat offers only
      Given olive's "olive-box" also serves the voice "kokoro" and the search offer "web-search"
      And olive made a grant "friends" at $0.05 per 1M output tokens
      When frank adds the "friends" secret to his account
      Then the reply lists the station "olive-box" and the model "qwen-small"
      And the reply lists neither "kokoro" nor "web-search"

    Scenario Outline: An added priced grant does not change the price of <kind>
      Given olive's "olive-box" also serves <offer> at a public price
      And frank added olive's grant "friends" at $0.05
      When frank makes a <kind> request to "olive-box"
      Then frank pays the public price from his own wallet
      And the grant "friends" records no use
      And olive's wallet is not charged

      Examples:
        | kind           | offer                         |
        | text-to-speech | the voice "kokoro"            |
        | speech-to-text | the voice model "whisper"     |
        | search         | the search offer "web-search" |

  Rule: A grant that ends stops giving its price at once

    Background:
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account

    Scenario Outline: An ended grant no longer prices a request
      When the grant "friends" is <ended>
      And frank asks "olive-box" for "qwen-small"
      Then frank is charged at the public price of $0.10
      And the reply says the grant "friends" no longer applies
      And frank's list of added grants shows "friends" as <state>

      Examples:
        | ended   | state   |
        | revoked | revoked |
        | expired | expired |

    Scenario: An ended dearer grant drops the caller back to the public price
      Given olive made a grant "premium" at $0.20 per 1M output tokens
      And frank added the "premium" secret to his account
      And frank removes "friends" from his account
      When olive revokes "premium"
      And frank asks "olive-box" for "qwen-small"
      Then frank is charged at the public price of $0.10
      And the reply says the grant "premium" no longer applies

    # Added 2026-11-24 (founder: "4. ok go with the public price", ruling P20).
    Scenario: A caller whose added grants cannot be read is told they pay the public price
      Given the grants frank added cannot be read
      When frank asks "olive-box" for "qwen-small"
      Then frank is charged at the public price of $0.10
      And the reply says frank's added grants could not be read

  # ---------------------------------------------------------------------------------------
  # Money safety
  # ---------------------------------------------------------------------------------------

  Rule: A priced grant can never create money or dodge the fee

    Scenario: A grant price is never negative, at create, at edit or at settle
      Then every scenario in features/grants/grant_price_safety.feature still holds

    Scenario: Grant earnings carry the same fee and hold as public earnings
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      And frank added the "friends" secret to his account
      When frank's use of "friends" settles
      Then the fee taken is the marketplace fee rate on what frank paid
      And olive's earnings are held and paid out under the same rules as public earnings

    Scenario: The ledger balances for every priced grant request
      When any priced grant request settles
      Then what the caller paid equals what the owner earned plus the fee
      And no wallet other than the caller's is debited

    Scenario: A request that fails is refunded in full
      Given frank added a priced grant
      When his request to the owner's station fails before any output
      Then the hold is released and frank is charged nothing

    Scenario: No request is billed to the owner as a sponsor any more
      Given olive made a grant "friends" at $0.05 per 1M output tokens
      When any request uses "friends"
      Then olive's wallet is never the payer

  # ---------------------------------------------------------------------------------------
  # Grants that exist before this change
  # ---------------------------------------------------------------------------------------

  Rule: Existing grants keep working

    Scenario: An existing free grant is unchanged
      Given a free grant made before this change
      When a request arrives with its secret as the API key
      Then the request is served and nobody is charged

    Scenario: A grant stored as not free with a zero price is free
      Given a grant made before this change that is not marked free and has a zero price
      Then the list shows it as "free"
      And a request with its secret as the API key is served and nobody is charged

    Scenario: A self grant is unchanged
      Given a self grant made before this change
      When the owner's own box uses it
      Then the request is served and nobody is charged
