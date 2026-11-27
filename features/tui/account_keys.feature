# APPROVED by the founder 2026-11-22, with the rulings below accepted as proposed.
#
# ACCOUNT KEYS AND GRANTS IN THE TUI.
#
# WHY. The founder wants both credential types on every surface. Checked on 2026-11-22:
#   - account keys (rog-key_...) exist in the CLI (`roger keys`, cmd/rogerai/keys.go) and on
#     the site. The TUI has nothing but a help row that sends people to the CLI.
#   - grants (rog-grant_...) exist in the TUI as `/grant`, with three gaps:
#       1. `/grant` says "N grants - see the panel" and no panel draws them
#          (internal/tui/tui.go grantListMsg; m.grantList has no reader).
#       2. `/grant create` puts the whole secret on the status line and copies nothing.
#       3. there is no way to revoke a grant.
#
# WHAT THIS ADDS. One KEYS panel that lists both kinds, a `/keys` command for account keys,
# and the missing `/grant` pieces. Nothing new on the broker: the TUI calls the same signed
# /account/keys and /grants routes the CLI calls (internal/client/acctkeys.go, grant.go).
#
# The broker's own rules (field limits, the 32 key cap, the rate limit, who may manage keys)
# are specified in features/auth/key_guardrails.feature and are not repeated here. This file
# covers what the TUI sends, shows, copies and refuses.
#
# RULINGS PROPOSED (each needs a yes or a change):
#   K1  A mint in the TUI must state a limit. A dollar amount, or the word "none" for no
#       limit. The CLI mints with no limit when --limit is left out; the TUI will not, because
#       a key spends wallet money on the whole network.
#   K2  A new secret is copied to the clipboard and shown ONCE in the KEYS panel. It is never
#       put on the status line or in a transcript, so /save and /copy cannot write it out.
#       It is shown as well as copied because a clipboard write can fail without telling us
#       (ssh, no clipboard tool).
#   K8  (founder, 2026-11-22) A limit is read by the same code as `roger keys --limit`
#       (client.ParseKeyLimit), so the CLI, the TUI and the console agree: digits with an
#       optional "$", or "none" for unlimited.
#   K9  (founder, 2026-11-22: "type 0 meaning no spending probably is more what makes sense")
#       A typed 0 is refused everywhere, the CLI included. To the broker 0 means unlimited,
#       the opposite of what someone typing it expects. "none" is the way to ask for that.
#       This replaces the first version of this spec, where 0 was the same as "none".
#   K10 (founder, 2026-11-22) The word for a key with no limit is "unlimited", as in the
#       CLI's table.
#   K3  The secret leaves the screen as soon as the panel closes or any other /keys or /grant
#       command runs. There is no way to bring it back.
#   K4  Revoking asks first, default no, for both keys and grants.
#   K5  A key or grant is named by its id or by its name. A name that matches more than one
#       is refused and the matches are listed.
#   K6  `/grant create` stays free only, as today. A priced grant is still CLI or site work.
#   K11 (founder, 2026-11-23) The verb to remove a key is "revoke", the same as for a grant.
#       It is the word the help and the usage lines show. "rm" and "delete" keep working.
#   K12 (founder, 2026-11-23) "none" is a window: `/keys limit ci 5 none` puts a key back to a
#       limit that never resets, as the CLI and the console already allow.
#   K13 (founder, 2026-11-23) Removing a limit from a key says the same thing a mint with no
#       limit says: the key can spend the whole wallet.
#   K7  The help row "roger keys (CLI only)" is replaced by `/keys`.

Feature: Account keys and grants in the TUI
  A person in the TUI can mint, see, change and revoke both kinds of key without leaving it,
  and a secret is never left anywhere it can be saved by accident.

  Background:
    Given the TUI is running against a broker
    And the account is logged in

  # ---------------------------------------------------------------------------------------
  # The KEYS panel
  # ---------------------------------------------------------------------------------------

  Scenario: /keys opens the panel with both kinds listed
    Given the account has the account keys
      | id        | name | hint    | limit_usd | reset  | used | state  |
      | key_aaaa1 | ci   | ...9f3c | 5         | weekly | 1.25 | active |
    And the account has the grants
      | name    | price | status |
      | my-bots | free  | active |
    When the operator runs "/keys"
    Then the KEYS panel is shown
    And the panel has a section "ACCOUNT KEYS" with one line per key
    And the line for "ci" shows "key_aaaa1", "...9f3c", "$5.00 weekly", "$1.25 used" and "active"
    And the panel has a section "GRANTS" with one line per grant
    And the line for "my-bots" shows "free" and "active"

  Scenario: Each section says in one line what the kind is for
    When the operator runs "/keys"
    Then the "ACCOUNT KEYS" section says "for you or your own bot - any public station - pays from your wallet"
    And the "GRANTS" section says "for someone else - your stations only - free by default"

  Scenario: Empty sections say how to make the first one
    Given the account has no account keys and no grants
    When the operator runs "/keys"
    Then the "ACCOUNT KEYS" section says "none yet - /keys mint <name> <limit> [daily|weekly|monthly]"
    And the "GRANTS" section says "none yet - /grant create <name>"

  Scenario Outline: The limit column reads plainly for every window
    Given the account has an account key with limit <limit> and reset "<reset>"
    When the operator runs "/keys"
    Then that key's limit reads "<text>"

    Examples:
      | limit | reset   | text           |
      | 5     | daily   | $5.00 daily    |
      | 5     | weekly  | $5.00 weekly   |
      | 5     | monthly | $5.00 monthly  |
      | 5     | none    | $5.00 total    |
      | 0     | none    | unlimited       |
      | 0     | weekly  | unlimited       |

  Scenario Outline: The used column counts the window the limit counts
    Given the account has an account key with reset "<reset>" that spent 1 today, 2 this week, 3 this month and 4 in all
    When the operator runs "/keys"
    Then that key's used amount reads "<used>"

    Examples:
      | reset   | used  |
      | daily   | $1.00 |
      | weekly  | $2.00 |
      | monthly | $3.00 |
      | none    | $4.00 |

  Scenario Outline: A key's state is one word
    Given the account has an account key that is <condition>
    When the operator runs "/keys"
    Then that key's state reads "<state>"

    Examples:
      | condition            | state    |
      | disabled             | disabled |
      | expired              | expired  |
      | disabled and expired | disabled |
      | neither              | active   |

  Scenario: /grant and /grant list open the same panel
    When the operator runs "/grant"
    Then the KEYS panel is shown
    When the operator closes the panel and runs "/grant list"
    Then the KEYS panel is shown

  Scenario: The panel closes with esc and returns to where the operator was
    Given the operator is in the channel
    When the operator runs "/keys" and then presses esc
    Then the operator is back in the channel
    And the transcript has no new lines

  Scenario: A narrow terminal still shows every key
    Given the terminal is 60 columns wide
    And the account has an account key named "ci"
    When the operator runs "/keys"
    Then the line for "ci" shows its name, its limit and its state
    And no line is wider than the terminal

  Scenario: A long key name is cut, never wrapped into the next column
    Given the account has an account key whose name is 128 characters long
    When the operator runs "/keys"
    Then that key's line fits the panel width
    And the name ends with "…"

  Scenario: A key name with control characters cannot draw on the terminal
    Given the broker returns an account key whose name holds an escape sequence
    When the operator runs "/keys"
    Then the panel shows the name with the control characters removed

  Scenario: The list never holds a secret
    Given the account has account keys and grants
    When the operator runs "/keys"
    Then nothing on screen starts with "rog-key_" or "rog-grant_"

  # ---------------------------------------------------------------------------------------
  # Minting an account key
  # ---------------------------------------------------------------------------------------

  Scenario: Minting with a limit and a window
    When the operator runs "/keys mint ci 5 weekly"
    Then the broker receives a mint with name "ci", limit_usd 5 and reset "weekly"
    And the KEYS panel is shown with the new key in the list
    And the panel shows the secret once under "new key - store it now, it is not shown again"
    And the export block is on the clipboard
    And the status line says "Copied to clipboard"

  Scenario: The export block is paste-ready
    When the operator runs "/keys mint ci 5 weekly"
    Then the clipboard holds exactly these lines, with the broker's address and the new secret
      """
      export OPENAI_BASE_URL=<broker>/v1
      export OPENAI_API_KEY=<secret>
      """

  Scenario: A limit with no window counts over the key's whole life
    When the operator runs "/keys mint ci 5"
    Then the broker receives a mint with name "ci", limit_usd 5 and reset "none"

  Scenario: "none" as the window is the same as leaving it out
    When the operator runs "/keys mint ci 5 none"
    Then the broker receives a mint with name "ci", limit_usd 5 and reset "none"

  Scenario: "none" mints a key with no limit, and says so
    When the operator runs "/keys mint ci none"
    Then the broker receives a mint with name "ci" and limit_usd 0
    And the panel says "unlimited - this key can spend your whole wallet"

  Scenario Outline: A mint that states no usable limit is refused before the broker is asked
    When the operator runs "<command>"
    Then the broker receives no mint
    And the status line says "<message>"

    Examples:
      | command                    | message                                                        |
      | /keys mint                 | usage: /keys mint <name> <limit> [daily\|weekly\|monthly]      |
      | /keys mint ci              | state a limit in dollars, or "none" for unlimited               |
      | /keys mint ci five         | state a limit in dollars, or "none" for unlimited               |
      | /keys mint ci -1           | a limit cannot be negative                                     |
      | /keys mint ci 5 yearly     | the window is daily, weekly, monthly or none                   |
      | /keys mint ci 5 weekly now | usage: /keys mint <name> <limit> [daily\|weekly\|monthly]      |

  Scenario Outline: The limit accepts the ways people write dollars
    When the operator runs "/keys mint ci <typed> weekly"
    Then the broker receives a mint with limit_usd <sent>

    Examples:
      | typed | sent |
      | 5     | 5    |
      | $5    | 5    |
      | 5.50  | 5.5  |
      | 0.01  | 0.01 |

  Scenario Outline: A typed 0 is refused, because it would mean unlimited and not no spending
    When the operator runs "/keys mint ci <typed>"
    Then the broker receives no mint
    And the status line says "0 would mean unlimited, not no spending - use none for an unlimited key"

    Examples:
      | typed |
      | 0     |
      | $0    |
      | 0.00  |
      | -0    |

  Scenario: An amount too small for the broker to keep is refused, because it would round to unlimited
    When the operator runs "/keys mint ci 0.0000001"
    Then the broker receives no mint
    And the status line says "state a limit in dollars, or "none" for unlimited"

  Scenario: "unlimited" is another word for "none"
    When the operator runs "/keys mint ci unlimited"
    Then the broker receives a mint with name "ci" and limit_usd 0
    And the panel says "unlimited - this key can spend your whole wallet"

  Scenario: "new" is the same as "mint"
    When the operator runs "/keys new ci 5 weekly"
    Then the broker receives a mint with name "ci", limit_usd 5 and reset "weekly"

  Scenario: The secret is never on the status line or in a transcript
    Given the operator is in the channel
    When the operator runs "/keys mint ci 5 weekly"
    Then the status line does not contain the secret
    And the transcript does not contain the secret
    And "/copy all" does not put the secret on the clipboard
    And "/save" does not put the secret anywhere

  Scenario: The secret is not written to the log buffer
    When the operator runs "/keys mint ci 5 weekly"
    And the operator runs "/log"
    Then the log does not contain the secret

  Scenario: Closing the panel removes the secret for good
    Given the operator minted a key and the secret is showing
    When the operator closes the panel and runs "/keys"
    Then the new key is in the list with its hint
    And the secret is not on screen

  Scenario Outline: Any other key or grant command removes a showing secret
    Given the operator minted a key and the secret is showing
    When the operator runs "<command>"
    Then the first secret is not on screen

    Examples:
      | command                  |
      | /keys                    |
      | /keys mint two 1 daily   |
      | /keys off ci             |
      | /grant create friends    |

  Scenario: A narrow terminal shows the whole secret on its own lines
    Given the terminal is 60 columns wide
    When the operator runs "/keys mint ci 5 weekly"
    Then every character of the secret is on screen
    And no line is wider than the terminal

  Scenario: The broker's refusal is shown in its own words
    Given the account already has 32 account keys
    When the operator runs "/keys mint ci 5 weekly"
    Then the status line says "at most 32 keys per account - delete one first"
    And no secret is shown

  Scenario: Mint while logged out says how to log in from here
    Given the account is logged out
    When the operator runs "/keys mint ci 5 weekly"
    Then the status line says "log in first - /login (keys are per account)"
    And the status line does not say "roger login"

  Scenario: A mint in flight cannot be sent twice
    Given the broker is slow to answer
    When the operator runs "/keys mint ci 5 weekly" twice before the first answer
    Then the broker receives one mint
    And the status line says "still minting ci"

  Scenario: The broker cannot be reached
    Given the broker is unreachable
    When the operator runs "/keys mint ci 5 weekly"
    Then the status line says the broker could not be reached
    And no secret is shown

  # ---------------------------------------------------------------------------------------
  # Changing an account key
  # ---------------------------------------------------------------------------------------

  Scenario Outline: Changing the limit
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "<command>"
    Then the broker receives a change to "key_aaaa1" with only <fields>
    And the KEYS panel shows the key as the broker returned it

    Examples:
      | command                      | fields                              |
      | /keys limit ci 10            | limit_usd 10                        |
      | /keys limit ci 10 monthly    | limit_usd 10 and reset "monthly"    |
      | /keys limit key_aaaa1 10     | limit_usd 10                        |
      | /keys limit ci none          | limit_usd 0                         |
      | /keys limit ci 5 none        | limit_usd 5 and reset "none"        |

  Scenario Outline: Removing a limit says the key can spend the whole wallet
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "<command>"
    Then the status line says "updated ci - unlimited - this key can spend your whole wallet"

    Examples:
      | command                  |
      | /keys limit ci none      |
      | /keys limit ci unlimited |
      | /keys limit ci none weekly |

  Scenario: A change that keeps a limit carries no warning
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "/keys limit ci 10"
    Then the status line says "updated ci"

  Scenario Outline: Turning a key off and back on
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "<command>"
    Then the broker receives a change to "key_aaaa1" with only disabled <value>

    Examples:
      | command      | value |
      | /keys off ci | true  |
      | /keys on ci  | false |

  Scenario Outline: A change that cannot be read is refused before the broker is asked
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "<command>"
    Then the broker receives no change
    And the status line says "<message>"

    Examples:
      | command                  | message                                                  |
      | /keys limit              | usage: /keys limit <key> <limit> [daily\|weekly\|monthly] |
      | /keys limit ci           | state a limit in dollars, or "none" for unlimited         |
      | /keys limit ci ten       | state a limit in dollars, or "none" for unlimited         |
      | /keys limit ci -1        | a limit cannot be negative                               |
      | /keys limit ci 0         | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | /keys limit ci 5 yearly  | the window is daily, weekly, monthly or none             |
      | /keys off                | usage: /keys off <key>                                   |
      | /keys on                 | usage: /keys on <key>                                    |
      | /keys frobnicate         | /keys: list, mint, limit, on, off, revoke                |

  Scenario: Naming a key that does not exist
    When the operator runs "/keys off nosuch"
    Then the broker receives no change
    And the status line says "no key named nosuch - /keys lists them"

  Scenario: A name shared by two keys is refused and the matches are listed
    Given the account has account keys "key_aaaa1" and "key_bbbb2" both named "ci"
    When the operator runs "/keys off ci"
    Then the broker receives no change
    And the status line says "2 keys are named ci - use the id: key_aaaa1, key_bbbb2"

  Scenario: An id wins over a name that looks the same
    Given the account has an account key "key_aaaa1" named "ci"
    And the account has an account key "key_bbbb2" named "key_aaaa1"
    When the operator runs "/keys off key_aaaa1"
    Then the broker receives a change to "key_aaaa1"

  Scenario: The key list is read fresh before a name is matched
    Given the operator has the KEYS panel open
    And a key named "late" is minted on the site afterwards
    When the operator runs "/keys off late"
    Then the broker receives a change to that key

  # ---------------------------------------------------------------------------------------
  # Revoking an account key
  # ---------------------------------------------------------------------------------------

  Scenario: Revoking asks first
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator runs "/keys revoke ci"
    Then the broker receives no revoke
    And the panel asks "revoke ci (key_aaaa1)? anything using it stops working. y/N"

  Scenario Outline: Only y revokes
    Given the operator was asked to confirm revoking "ci"
    When the operator presses "<key>"
    Then the broker <outcome>

    Examples:
      | key   | outcome                      |
      | y     | receives a revoke of key_aaaa1 |
      | Y     | receives a revoke of key_aaaa1 |
      | n     | receives no revoke           |
      | enter | receives no revoke           |
      | esc   | receives no revoke           |
      | x     | receives no revoke           |

  Scenario: A revoked key leaves the list
    Given the operator confirmed revoking "ci"
    Then the KEYS panel no longer lists "ci"
    And the status line says "revoked ci"

  Scenario Outline: The other words for revoke
    Given the account has an account key named "ci"
    When the operator runs "<command>"
    Then the panel asks to confirm revoking "ci"

    Examples:
      | command          |
      | /keys rm ci      |
      | /keys revoke ci  |
      | /keys delete ci  |

  Scenario: Revoke with no key named
    When the operator runs "/keys revoke"
    Then the status line says "usage: /keys revoke <key>"

  # ---------------------------------------------------------------------------------------
  # Grants: the gaps closed
  # ---------------------------------------------------------------------------------------

  Scenario: A new grant's secret is handled like a new key's secret
    When the operator runs "/grant create friends"
    Then the broker receives a grant create with name "friends" and free true
    And the KEYS panel shows the secret once under "new grant - store it now, it is not shown again"
    And the export block is on the clipboard
    And the status line does not contain the secret

  Scenario: The grant name still defaults to my-bots
    When the operator runs "/grant create"
    Then the broker receives a grant create with name "my-bots" and free true

  Scenario: Revoking a grant asks first
    Given the account has a grant named "friends"
    When the operator runs "/grant revoke friends"
    Then the broker receives no revoke
    And the panel asks "revoke grant friends? anyone using it loses access. y/N"

  Scenario: Confirming revokes the grant
    Given the operator was asked to confirm revoking grant "friends"
    When the operator presses "y"
    Then the broker receives a revoke of grant "friends"
    And the KEYS panel shows "friends" as "revoked"

  Scenario Outline: The other words for a grant revoke
    Given the account has a grant named "friends"
    When the operator runs "<command>"
    Then the panel asks to confirm revoking grant "friends"

    Examples:
      | command                |
      | /grant revoke friends  |
      | /grant rm friends      |

  Scenario: Revoking a grant that does not exist
    When the operator runs "/grant revoke nosuch"
    Then the status line says "no grant named nosuch - /grant lists them"

  Scenario: An account with no station cannot make a grant, and is told what to use instead
    Given the broker refuses grants for this account
    When the operator runs "/grant create friends"
    Then the status line carries the broker's reason
    And the status line says "for your own bot, use /keys mint"

  Scenario: Grants refused does not hide account keys
    Given the broker refuses grants for this account
    And the account has an account key named "ci"
    When the operator runs "/keys"
    Then the "ACCOUNT KEYS" section lists "ci"
    And the "GRANTS" section carries the broker's reason

  # ---------------------------------------------------------------------------------------
  # When things are unavailable
  # ---------------------------------------------------------------------------------------

  Scenario: Logged out, the panel says so and lists nothing
    Given the account is logged out
    When the operator runs "/keys"
    Then the KEYS panel says "log in to manage keys - /login"
    And the broker is asked for no secrets

  Scenario: One list failing does not blank the other
    Given the broker answers the key list and fails the grant list
    When the operator runs "/keys"
    Then the "ACCOUNT KEYS" section lists the keys
    And the "GRANTS" section says the grants could not be read

  Scenario: A build with no key hooks points at the CLI
    Given the TUI was started without the account key hooks
    When the operator runs "/keys"
    Then the status line says "keys unavailable in this build - run `roger keys`"

  Scenario: LAN mode has no account, so no keys
    Given the TUI is in LAN mode
    When the operator runs "/keys"
    Then the status line says "keys need the public network - not available in LAN mode"
    And the broker is not asked

  # ---------------------------------------------------------------------------------------
  # Help and hints
  # ---------------------------------------------------------------------------------------

  Scenario: The help lists /keys and no longer says CLI only
    When the operator opens the help
    Then the help has a row for "/keys" that says "account API keys for your own bot on a server"
    And the help has a row for "/grant" that says "a key for someone else to use your stations"
    And the help does not contain "CLI only"

  Scenario: The command menu offers /keys
    When the operator types "/ke"
    Then the command suggestions include "/keys"

  Scenario: The endpoint plate still does not point at account keys
    Given a channel is open
    Then the endpoint plate does not contain "/keys" or "roger keys"

  # ---------------------------------------------------------------------------------------
  # Answers that arrive late, and what closes the panel
  # (added after review; founder ok 2026-11-22)
  # ---------------------------------------------------------------------------------------

  Scenario: The panel opens as soon as a mint is typed
    Given the broker is slow to answer
    When the operator runs "/keys mint ci 5 weekly" and the answer has not come back
    Then the KEYS panel is shown
    And the status line says "minting ci"

  Scenario Outline: A secret that arrives after the panel was closed is dropped, and the operator is told
    Given the broker is slow to answer
    When the operator runs "<command>", closes the panel, and the answer comes back
    Then no secret is shown
    And the clipboard does not hold a secret
    And the status line says "<told>"

    Examples:
      | command                | told                                                                                                |
      | /keys mint ci 5 weekly | ci was made while the panel was closed, so its secret was dropped - /keys revoke ci, then make another |
      | /grant create friends  | friends was made while the panel was closed, so its secret was dropped - /grant revoke friends, then make another |

  Scenario Outline: Only one secret is made at a time, key or grant
    Given the broker is slow to answer
    When the operator runs "<first>" and then "<second>" before the first answer
    Then the broker receives <received>
    And the status line says "<status>"

    Examples:
      | first                  | second                 | received                                               | status                |
      | /keys mint ci 5 weekly | /grant create friends  | one mint and no grant create                           | still minting ci      |
      | /grant create friends  | /keys mint ci 5 weekly | one grant create and no mint                           | still minting friends |

  Scenario Outline: A key that is not esc or the start of a command leaves the panel open
    Given the operator has the KEYS panel open
    When the operator presses "<key>"
    Then the KEYS panel is shown

    Examples:
      | key   |
      | enter |
      | q     |
      | x     |

  Scenario: Starting a command closes the panel
    Given the operator has the KEYS panel open
    When the operator presses "/"
    Then the KEYS panel is not shown

  Scenario: A revoke question is not asked over other work
    Given the account has an account key "key_aaaa1" named "ci"
    And the broker is slow to answer
    When the operator runs "/keys revoke ci", closes the panel, and the answer comes back
    And the operator presses "y"
    Then the KEYS panel is not shown
    And the broker receives no revoke

  Scenario: A grant that is already revoked is not revoked again
    Given the account has a revoked grant named "friends"
    When the operator runs "/grant revoke friends"
    Then the status line says "grant friends is already revoked"
    And the broker receives no revoke

  # ---------------------------------------------------------------------------------------
  # Scrolling (founder ok 2026-11-22)
  #   S1  The two lists scroll together as one view. Up and down move one row, pgup and
  #       pgdown one screen, home and end jump to the top and the bottom.
  #   S2  A new secret and a revoke question are never scrolled away. They stay at the
  #       bottom and the lists take the rows that are left.
  #   S3  When rows are hidden the panel says how many, above and below.
  #   S4  A new secret or a question does not move the lists. Reopening starts at the top.
  #   S5  The scroll keys never answer a revoke question.
  # ---------------------------------------------------------------------------------------

  Scenario: Lists that fit do not scroll
    Given the account has 3 account keys and the terminal is 50 rows tall
    When the operator runs "/keys"
    And the operator presses "down"
    Then key 1 is on screen
    And key 3 is on screen
    And the panel does not say "more above"
    And the panel does not say "more below"

  Scenario: Long lists say what is hidden and stay inside the terminal
    Given the account has 30 account keys and the terminal is 24 rows tall
    When the operator runs "/keys"
    Then key 1 is on screen
    And key 30 is not on screen
    And the panel says how many rows are below
    And the panel does not say "more above"
    And the screen is no taller than the terminal

  Scenario Outline: The keys that move the lists
    Given the account has 30 account keys and the terminal is 24 rows tall
    And the operator has the KEYS panel open
    When the operator presses "<keys>"
    Then the panel <says>

    Examples:
      | keys        | says                       |
      | down        | says "1 more above"        |
      | down,down   | says "2 more above"        |
      | down,up     | does not say "more above"  |
      | up          | does not say "more above"  |
      | pgdown      | says "more above"          |
      | pgdown,pgup | does not say "more above"  |
      | end         | does not say "more below"  |
      | end,home    | does not say "more above"  |
      | end,down    | does not say "more below"  |

  Scenario: The bottom of the lists is reachable, and the lists stop there
    Given the account has 30 account keys and the terminal is 24 rows tall
    And the operator has the KEYS panel open
    When the operator presses "end,down,down"
    Then the "GRANTS" hint "none yet - /grant create <name>" is on screen
    And the screen is no taller than the terminal
    When the operator presses "up"
    Then the panel says "1 more below"

  Scenario: A new secret stays on screen under long lists
    Given the account has 30 account keys and the terminal is 24 rows tall
    When the operator runs "/keys mint ci 5 weekly"
    Then every character of the secret is on screen
    And the panel says how many rows are below
    And the screen is no taller than the terminal

  Scenario: A revoke question stays on screen, and scrolling does not answer it
    Given the account has 30 account keys and the terminal is 24 rows tall
    When the operator runs "/keys revoke key_0001"
    And the operator presses "down,pgdown,end,home"
    Then the panel says "revoke k01 (key_0001)? anything using it stops working. y/N"
    And the broker receives no revoke
    And the screen is no taller than the terminal

  Scenario: Scrolling keys do nothing when everything fits
    Given the operator has the KEYS panel open
    When the operator presses "down,pgdown,end,up,pgup,home"
    Then the KEYS panel is shown
    And the broker receives no change

  Scenario: Reopening the panel starts at the top
    Given the account has 30 account keys and the terminal is 24 rows tall
    And the operator has the KEYS panel open
    When the operator presses "end"
    And the operator closes the panel and runs "/keys"
    Then key 1 is on screen
    And the panel does not say "more above"

  Scenario: A terminal too short for anything else still shows one list row and the secret
    Given the account has 30 account keys and the terminal is 12 rows tall
    When the operator runs "/keys mint ci 5 weekly"
    Then every character of the secret is on screen
