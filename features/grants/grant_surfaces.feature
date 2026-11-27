# APPROVED by the founder 2026-11-24 ("3. ok"), with rulings S1 to S11 as proposed. S6 was
# also approved on its own the same day ("2. yes").
#
# ADDING, LISTING, REMOVING AND REPRICING GRANTS IN THE CLI, THE TUI AND THE LOCAL CONSOLE.
#
# WHY. features/grants/priced_grants.feature (approved 2026-11-23, rulings P1 to P19) made a
# priced grant something a caller ADDS to their own account and pays for. The broker routes
# are built on wt/webui-endpoint-plate:
#   POST   /account/grants            {"secret": "rog-grant_..."}  add
#   GET    /account/grants            list what this account added
#   DELETE /account/grants/{id}       remove one
#   POST   /grants/{id}/reprice       owner: new price (and scope), new secret
# No surface calls them yet. Checked 2026-11-24:
#   - CLI `roger grant` has create, list, show, revoke (cmd/rogerai/grant.go). Nothing to add,
#     list added, remove or reprice. `rm` is an alias of revoke.
#   - The broker's limit message already tells people to run `roger grant remove <name>`,
#     which does not exist.
#   - `roger grant show --secret <name>` rotates a FREE grant (revoke, then a new grant under
#     the same name) and refuses a priced one with "revoke then create", which would drop
#     every account that added it.
#   - TUI `/grant` makes free grants only (ruling K6 in features/tui/account_keys.feature).
#   - The console ACCOUNT tab makes free grants from a name only (features/console/account_keys.feature).
#
# The broker's own rules (who may add, the 10 grant limit, the try limit, what a copy covers,
# the price shown) are in priced_grants.feature and are not repeated here. This file covers
# what each surface sends, shows, copies and refuses.
#
# RULINGS (approved 2026-11-24):
#   S1  CLI verbs. `roger grant add [secret]`, `roger grant remove <name|id>`,
#       `roger grant reprice <name|id> --price-out X [--price-in Y] [--nodes a,b] [--models m]`.
#       `roger grant list` shows two sections: "grants you made" and "grants you added".
#   S12 (founder, 2026-11-25, "1. yes") `roger grant set <name|id> [--price-out X]
#       [--price-in Y] [--nodes a,b] [--models m]` edits a grant you made and keeps its secret:
#       a price cut and a narrowing apply at once. A raise or a widening of a priced grant is
#       refused in the broker's words, which name reprice. A reprice that is only a cut names
#       `roger grant set`.
#   S2  `remove` and `revoke` are different. `revoke` ends a grant you made, for everyone.
#       `remove` drops a grant you added, from your account only. `remove` on the name of a
#       grant you made is refused and names `revoke`. `revoke` on the name of a grant you only
#       added is refused and names `remove`. `rm` stays an alias of `revoke` (ruled 2026-11-23).
#   S3  A secret is never typed on the command line by need. `roger grant add` with no
#       argument asks for it with the input hidden, so it stays out of shell history. A secret
#       given as the argument still works. In the TUI, `/grant add` with no secret opens a
#       hidden input; the secret is never written to the transcript, so /save and /copy cannot
#       write it out (same rule as K2).
#   S4  Adding shows what the broker returned: the grant name, its price, each station with
#       its public price, a line for each station where the grant price is above public, and
#       the models. It asks nothing first: adding charges nothing, and removing is one command.
#   S5  A reprice prints the new secret once, says the old secret has stopped working, and
#       says every account that added it is back on the public price until they add the new
#       one. It asks first, default no, because it cannot be undone. `--yes` skips the question.
#   S6  `roger grant show --secret <name>` on a PRICED grant is refused and names
#       `roger grant reprice <name> --price-out <current price>`, which keeps the name, scope,
#       caps and usage. (A same-price reprice is allowed: it issues a new secret.) A free grant
#       keeps today's rotate.
#   S7  TUI. `/grant add`, `/grant added`, `/grant remove <name|id>`. Making and repricing a
#       priced grant stays CLI and console only, as K6 says for priced create.
#   S8  Console. The ACCOUNT tab gains a "Grants you added" card (add box, list, remove) and,
#       in the grants-you-made card, a price field on create and a Reprice action. The console
#       follows the CLI: it sends what was typed and the console server reads it with the same
#       price parser as `roger grant create --price-out`.
#   S9  Lists never show a secret. Grants you added show name, owner's price, state
#       (active, revoked, expired, replaced) and the date added. A grant that is not active
#       says why in words: "replaced - ask the owner for the new secret", "revoked", "expired".
#   S10 Every surface shows the broker's refusal text as given (the messages in
#       priced_grants.feature), with one addition: the limit message on the TUI and console
#       names their own remove action instead of the CLI command.
#   S11 A surface used while logged out says how to log in on that surface (`roger login`,
#       `/login`, the console's login button) and sends nothing.
#
# OUT OF SCOPE: the site's own account page (web/), iOS and Mac, the manual text.
# Voice (TTS and STT) using added grants is a separate open decision.

Feature: Adding, listing, removing and repricing grants on every surface
  A person who was given a priced grant can add it, see it and remove it from the CLI, the TUI
  or the console, and an owner can reprice one, without a secret ever being left anywhere it
  can be saved by accident.

  Background:
    Given olive made a grant "friends" at $0.05 per 1M output tokens on "olive-box"
    And frank is logged in

  # ---------------------------------------------------------------------------------------
  # CLI: add
  # ---------------------------------------------------------------------------------------

  Rule: `roger grant add` adds a priced grant and shows what it costs

    Scenario: Adding with the secret as an argument shows the price, stations and models
      When frank runs `roger grant add <friends secret>`
      Then the output names "friends" and shows $0.05 per 1M output tokens
      And the output lists "olive-box" with its public price and the model "qwen-small"
      And the output never repeats the secret

    Scenario: Adding with no argument asks for the secret with the input hidden
      When frank runs `roger grant add` and types the "friends" secret at the prompt
      Then the typed secret is not echoed
      And the grant is added and shown as above

    Scenario: Adding from a pipe reads the secret from standard input
      When frank pipes the "friends" secret into `roger grant add`
      Then the grant is added and no prompt is printed

    Scenario: A grant priced above public says so on its station row
      Given olive made a grant "dear" at $0.20 per 1M output tokens on "olive-box"
      And "olive-box" sells "qwen-small" at $0.10 publicly
      When frank adds the "dear" secret
      Then the "olive-box" row says the grant price is above the public price

    Scenario: A grant whose owner has no station on air says it covers none yet
      Given "olive-box" is off air
      When frank adds the "friends" secret
      Then the output says the grant covers no station yet and to add it again later

    Scenario Outline: A refusal from the broker is shown as given and nothing is added
      When frank runs `roger grant add` with <secret>
      Then the command fails with "<message>"
      And frank's added grants are unchanged

      Examples:
        | secret                      | message                                     |
        | an empty value              | paste the grant secret                      |
        | an account key              | that is an account key, not a grant         |
        | a free grant's secret       | this grant is free - use it as your API key |
        | a revoked grant's secret    | grant key invalid or revoked                |
        | olive's own secret as olive | this is your own grant - your own stations are already free to you |

    Scenario: Adding at the limit names the remove command, which now exists
      Given frank has added 10 grants
      When frank adds the "friends" secret
      Then the command fails naming `roger grant remove <name>`
      And `roger grant remove` is a working command

    Scenario: Adding while logged out sends nothing
      Given frank is logged out
      When frank runs `roger grant add <friends secret>`
      Then the command fails with "log in first: roger login"
      And no request reached the broker

  # ---------------------------------------------------------------------------------------
  # CLI: list and remove
  # ---------------------------------------------------------------------------------------

  Rule: `roger grant list` shows grants made and grants added, and `remove` drops one added

    Scenario: The list has a section for each kind
      Given olive made a grant "club" and frank added "friends"
      When frank runs `roger grant list`
      Then it shows a "grants you made" section and a "grants you added" section
      And "friends" is under "grants you added" with $0.05, state "active" and the date added
      And no secret appears

    Scenario: An account with nothing added still shows the section, empty, with how to add
      When frank, who added nothing, runs `roger grant list`
      Then the "grants you added" section says none and names `roger grant add`

    Scenario Outline: A grant that stopped applying says why in words
      Given frank added "friends"
      And <event>
      When frank runs `roger grant list`
      Then "friends" shows "<state>"

      Examples:
        | event                         | state                                     |
        | olive revoked "friends"       | revoked                                   |
        | "friends" expired             | expired                                   |
        | olive repriced "friends"      | replaced - ask the owner for the new secret |

    Scenario: Removing by name drops the grant from frank's account only
      Given frank and gina added "friends"
      When frank runs `roger grant remove friends`
      Then frank no longer has "friends"
      And gina still has "friends"
      And frank's next request to "olive-box" is priced at the public price

    # Setup line added 2026-11-24 (founder, 2026-11-25: "lets fix"): frank must hold the grant to remove it.
    Scenario: Removing by id works too
      Given frank added "friends"
      When frank runs `roger grant remove <friends id>`
      Then frank no longer has "friends"

    Scenario: Removing a name frank added twice from different owners lists the ids
      Given frank added two grants named "friends" from different owners
      When frank runs `roger grant remove friends`
      Then the command fails and lists both ids

    Scenario: Removing a grant you made is refused and names revoke
      When olive runs `roger grant remove friends`
      Then the command fails naming `roger grant revoke friends`
      And "friends" still works for everyone who added it

    Scenario: Revoking a grant you only added is refused and names remove
      Given frank added "friends"
      When frank runs `roger grant revoke friends`
      Then the command fails naming `roger grant remove friends`

    Scenario: Removing a grant frank never added says so
      When frank runs `roger grant remove nothing-here`
      Then the command fails with "no grant named "nothing-here" in your added grants"

  # ---------------------------------------------------------------------------------------
  # CLI: reprice
  # ---------------------------------------------------------------------------------------

  Rule: `roger grant reprice` issues a new secret and says who it drops

    Scenario: A reprice asks first and does nothing on no
      When olive runs `roger grant reprice friends --price-out 0.08` and answers no
      Then "friends" is still $0.05 and its secret still works

    # Setup line added 2026-11-24 (founder, 2026-11-25: "lets fix"): the "1 account" below is frank.
    Scenario: A reprice confirmed prints the new secret once and what it ended
      Given frank added "friends"
      When olive runs `roger grant reprice friends --price-out 0.08 --yes`
      Then the new secret is printed once
      And the output says the old secret has stopped working
      And the output says 1 account that added it is on the public price until they add the new one

    Scenario: A reprice may widen the stations or models
      When olive runs `roger grant reprice friends --price-out 0.05 --nodes olive-box,olive-2 --yes`
      Then "friends" covers "olive-box" and "olive-2" under the new secret

    Scenario: A reprice with no price is refused before anything is sent
      When olive runs `roger grant reprice friends --nodes olive-2`
      Then the command fails with "state a price in dollars per 1M tokens, or "free""
      And no request reached the broker

    Scenario: A reprice of a revoked grant shows the broker's refusal
      Given olive revoked "friends"
      When olive runs `roger grant reprice friends --price-out 0.08 --yes`
      Then the command fails with "this grant is revoked - make a new grant instead"

    Scenario: A price cut is not a reprice and keeps the secret
      When olive runs `roger grant reprice friends --price-out 0.03`
      Then the command says a cut keeps the same secret and names the edit command
      And nothing is changed

    Scenario: Recovering a lost priced secret names reprice, not revoke
      When olive runs `roger grant show --secret friends`
      Then the command fails naming `roger grant reprice friends --price-out 0.05`
      And "friends" is unchanged and nobody who added it is dropped

    Scenario: Recovering a lost free secret still rotates it
      Given olive made a free grant "bots"
      When olive runs `roger grant show --secret bots`
      Then a new secret for "bots" is printed once and the old one stops working

  # ---------------------------------------------------------------------------------------
  # CLI: set (S12, added 2026-11-25)
  # ---------------------------------------------------------------------------------------

  Rule: `roger grant set` edits a grant you made and keeps its secret

    Scenario: A price cut applies at once and keeps the secret
      Given frank added "friends"
      When olive runs `roger grant set friends --price-out 0.03`
      Then "friends" is still $0.03 and its secret still works
      And frank still has "friends"

    Scenario: A raise is refused and names reprice
      When olive runs `roger grant set friends --price-out 0.08`
      Then the command fails with "a higher price needs a new secret - reprice the grant"
      And "friends" is still $0.05 and its secret still works

    Scenario: Narrowing the models applies at once
      When olive runs `roger grant set friends --models qwen-other`
      Then "friends" covers only the model "qwen-other"
      And "friends" is still $0.05 and its secret still works

    Scenario: Widening a priced grant is refused and names reprice
      Given olive ran `roger grant set friends --models qwen-small`
      When olive runs `roger grant set friends --models qwen-small,qwen-other`
      Then the command fails with "a wider grant needs a new secret - reprice the grant"
      And "friends" covers only the model "qwen-small"

    Scenario: A set with nothing to change sends nothing
      When olive runs `roger grant set friends`
      Then the command fails with "nothing to change"
      And no request reached the broker

    Scenario: Setting a grant you only added is refused and names remove
      Given frank added "friends"
      When frank runs `roger grant set friends --price-out 0.03`
      Then the command fails naming `roger grant remove friends`

  # ---------------------------------------------------------------------------------------
  # TUI
  # ---------------------------------------------------------------------------------------

  Rule: The TUI adds, lists and removes added grants, and a secret never reaches the transcript

    Scenario: /grant add with no secret opens a hidden input
      When frank types "/grant add" in the TUI
      Then a hidden input asks for the grant secret
      When frank pastes the "friends" secret and presses enter
      Then the KEYS panel shows "friends" at $0.05 with its stations and models
      And the transcript and status line never contain the secret

    Scenario: /grant add with the secret typed inline is cleared from input history
      When frank types "/grant add <friends secret>" in the TUI
      Then the grant is added
      And pressing up in the input does not bring the secret back

    Scenario: /save after an add writes nothing that holds the secret
      Given frank added "friends" in the TUI
      When frank runs /save
      Then no saved file contains the secret

    Scenario: /grant added lists the grants frank added in the KEYS panel
      Given frank added "friends"
      When frank types "/grant added"
      Then the KEYS panel has a "grants you added" section with "friends", $0.05 and "active"

    Scenario: /grant remove asks first, default no
      Given frank added "friends"
      When frank types "/grant remove friends" and answers no
      Then frank still has "friends"
      When frank types "/grant remove friends" and answers yes
      Then frank no longer has "friends"

    Scenario: The limit refusal names the TUI's remove command
      Given frank has added 10 grants
      When frank adds the "friends" secret in the TUI
      Then the refusal names "/grant remove <name>"

    Scenario: /grant add logged out says to log in and sends nothing
      Given frank is logged out
      When frank types "/grant add"
      Then the TUI says "log in first: /login" and no request reached the broker

    Scenario: /grant create stays free only
      When olive types "/grant create --price-out 0.05 --name x"
      Then the TUI says priced grants are made with `roger grant create` or the console

  # ---------------------------------------------------------------------------------------
  # Console
  # ---------------------------------------------------------------------------------------

  Rule: The console ACCOUNT tab adds, lists and removes added grants, and makes and reprices priced grants

    Scenario: The "Grants you added" card adds a pasted secret and shows the price
      When frank pastes the "friends" secret into the add box and presses Add
      Then the card lists "friends" at $0.05 with "olive-box" and "qwen-small"
      And the add box is cleared
      And the secret is in no response the page received after the add

    Scenario: The console routes for added grants need the console token
      When a request without the console token calls the added-grants routes
      Then it is refused and nothing reaches the broker

    Scenario: Removing from the card asks first
      Given frank added "friends"
      When frank presses Remove on "friends" and confirms
      Then the card no longer lists "friends"

    Scenario: The limit refusal names the card's Remove button
      Given frank has added 10 grants
      When frank adds the "friends" secret in the console
      Then the refusal says to remove one with the Remove button

    Scenario: Making a priced grant from the console reads the price like the CLI
      When olive makes a grant "club" with price "0.05" in the console
      Then "club" is priced at $0.05 per 1M output tokens
      And the new secret is shown once with a copy button

    Scenario Outline: A price the CLI would refuse is refused the same way in the console
      When olive makes a grant "club" with price "<typed>" in the console
      Then the console refuses it with the same message as `roger grant create --price-out <typed>`
      And no grant is made

      Examples:
        | typed |
        | -1    |
        | abc   |
        | 1e9   |

    Scenario: Reprice in the console asks first and shows the new secret once
      When olive presses Reprice on "friends", enters 0.08 and confirms
      Then the new secret is shown once with a copy button
      And the page says the old secret has stopped working and how many accounts it dropped

    Scenario: A stopped grant in the added card says why
      Given frank added "friends" and olive repriced it
      When frank opens the ACCOUNT tab
      Then "friends" reads "replaced - ask the owner for the new secret"

    Scenario: The console logged out offers the login button and sends nothing
      Given frank is logged out
      When frank opens the "Grants you added" card
      Then it shows the login button and no add box
