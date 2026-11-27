# APPROVED by the founder 2026-11-22, with the rulings below accepted as proposed.
#
# ACCOUNT KEYS AND GRANTS IN THE LOCAL WEB CONSOLE.
#
# WHY. The founder wants both credential types on every surface. Checked on 2026-11-22:
#   - account keys (rog-key_...) are not in the console at all. The endpoint plate tells
#     people to run `roger keys mint` in a terminal.
#   - grants (rog-grant_...) are on the ACCOUNT tab, with two gaps:
#       1. every row reads "grant". The page reads `name`, `free` and `balance`, and
#          /api/grants answers `Name`, `Price`, `Status` (client.GrantInfo has no json tags).
#       2. there is no way to revoke one.
#
# WHAT THIS ADDS. An "Account keys" card on the ACCOUNT tab, a working grant list, and grant
# revoke. Nothing new on the broker: the console calls the same signed /account/keys and
# /grants routes the CLI calls (internal/client/acctkeys.go, grant.go).
#
# The broker's own rules are specified in features/auth/key_guardrails.feature and are not
# repeated here. This file covers the console's routes and what the page shows.
#
# ROUTES (all need the console's per-run token, like every other console route):
#   GET  /api/keys                  -> {"keys":[...]}          never a secret
#   POST /api/keys/mint             {name, limit_usd, reset}   -> {"ok":true,"key":{...},"secret":"..."}
#   POST /api/keys/set              {id, limit_usd?, reset?, disabled?}  -> {"ok":true,"key":{...}}
#   POST /api/keys/revoke           {id}                       -> {"ok":true}
#   GET  /api/grants                -> [{"name","price","status"}]   (field names fixed)
#   POST /api/grants                {name}                     -> {"ok":true,"secret":"...","export":"..."}
#   POST /api/grants/revoke         {name}                     -> {"ok":true}
#
# RULINGS PROPOSED (each needs a yes or a change):
#   C1  A mint must state a limit: a dollar amount, or the "unlimited" box ticked. Same reason
#       as the TUI: a key spends wallet money on the whole network.
#   C7  (founder, 2026-11-22: "follow what we have working already on the cli") A limit is
#       read by the same code as `roger keys --limit`. The page sends what was typed and the
#       console reads it. "none" is unlimited and carries the warning.
#   C8  (founder, 2026-11-22: "type 0 meaning no spending probably is more what makes sense")
#       A typed 0 is refused on every surface, the CLI included. To the broker 0 means
#       unlimited, the opposite of what someone typing it expects.
#   C9  (founder, 2026-11-22) The word for a key with no limit is "unlimited", as in the
#       CLI's table.
#   C2  A new secret is shown once on the page with a copy button. It is held only in the
#       page's memory. It is never in /api/state, the event stream, the URL, browser storage
#       or the node log.
#   C3  The secret leaves the page when the operator presses "done", leaves the ACCOUNT tab,
#       or reloads. There is no way to bring it back.
#   C4  Revoking asks first, for both keys and grants.
#   C5  The endpoint plate's note changes from "roger keys mint" to pointing at this card.
#       That is a wording change to the approved plate spec
#       (features/console/endpoint_plate.feature) and needs its own yes.
#   C6  (founder, 2026-11-23, replacing "a name and a free box") The console makes a free
#       grant from a name. There is no "free" box: left unticked, the old one made a grant
#       the list showed as "$0/$0". A priced grant is its own spec
#       (features/grants/priced_grants.feature) and gets its own field here when that lands.
#   C10 (founder, 2026-11-23) The grant routes are guarded like the key routes: LAN mode,
#       logged out, an oversized body and an unknown field are refused before the broker is
#       asked, and every refusal is JSON with a message.
#   C11 (founder, 2026-11-23) Removing a limit from a key's row says the same thing a mint
#       with no limit says: the key can spend the whole wallet.

Feature: Account keys and grants in the local web console
  A person in the browser console can mint, see, change and revoke both kinds of key, and a
  secret is shown once and kept nowhere.

  Background:
    Given the console is running against a broker
    And the account is logged in
    And the operator has the ACCOUNT tab open

  # ---------------------------------------------------------------------------------------
  # The Account keys card
  # ---------------------------------------------------------------------------------------

  Scenario: The card lists the account's keys
    Given the account has the account keys
      | id        | name | hint    | limit_usd | reset  | used | state  |
      | key_aaaa1 | ci   | ...9f3c | 5         | weekly | 1.25 | active |
    When the ACCOUNT tab loads
    Then the "Account keys" card has one row for "ci"
    And the row shows "...9f3c", "$5.00 weekly", "$1.25 used" and "active"

  Scenario: Each card says in one line what the kind is for
    When the ACCOUNT tab loads
    Then the "Account keys" card says "For you or your own bot. Works on any public station. Pays from your wallet."
    And the "Grants" card says "For someone else. Works on your stations only. Free by default."

  Scenario: No keys yet
    Given the account has no account keys
    When the ACCOUNT tab loads
    Then the "Account keys" card says "none yet"

  Scenario Outline: The limit reads plainly for every window
    Given the account has an account key with limit <limit> and reset "<reset>"
    When the ACCOUNT tab loads
    Then that key's limit reads "<text>"

    Examples:
      | limit | reset   | text          |
      | 5     | daily   | $5.00 daily   |
      | 5     | weekly  | $5.00 weekly  |
      | 5     | monthly | $5.00 monthly |
      | 5     | none    | $5.00 total   |
      | 0     | none    | unlimited      |
      | 0     | weekly  | unlimited      |

  Scenario Outline: The used amount counts the window the limit counts
    Given the account has an account key with reset "<reset>" that spent 1 today, 2 this week, 3 this month and 4 in all
    When the ACCOUNT tab loads
    Then that key's used amount reads "<used>"

    Examples:
      | reset   | used  |
      | daily   | $1.00 |
      | weekly  | $2.00 |
      | monthly | $3.00 |
      | none    | $4.00 |

  Scenario Outline: A key's state is one word
    Given the account has an account key that is <condition>
    When the ACCOUNT tab loads
    Then that key's state reads "<state>"

    Examples:
      | condition            | state    |
      | disabled             | disabled |
      | expired              | expired  |
      | disabled and expired | disabled |
      | neither              | active   |

  Scenario: A key name cannot inject markup
    Given the broker returns an account key named "<img src=x onerror=alert(1)>"
    When the ACCOUNT tab loads
    Then the row shows that name as text
    And the page has no new image element

  Scenario: The list reply never holds a secret
    Given the account has account keys
    When the page reads "/api/keys"
    Then the reply holds no field named "secret"
    And the reply holds no text starting with "rog-key_"

  # ---------------------------------------------------------------------------------------
  # Minting
  # ---------------------------------------------------------------------------------------

  Scenario: Minting with a limit and a window
    When the operator enters the name "ci", the limit "5" and the window "weekly" and presses "Mint key"
    Then the broker receives a mint with name "ci", limit_usd 5 and reset "weekly"
    And the card shows the secret under "New key. Store it now. It is not shown again."
    And the new key is in the list with its hint

  Scenario: The copy button copies a paste-ready block
    Given the operator minted a key and the secret is showing
    When the operator presses "copy"
    Then the clipboard holds exactly these lines, with the broker's address and the new secret
      """
      export OPENAI_BASE_URL=<broker>/v1
      export OPENAI_API_KEY=<secret>
      """
    And the page says "copied"

  Scenario: The clipboard refusing does not lose the secret
    Given the browser refuses clipboard writes
    And the operator minted a key and the secret is showing
    When the operator presses "copy"
    Then the page says "copy failed - select the key and copy it by hand"
    And the secret is still showing

  Scenario: The window defaults to "none" and the limit box starts empty
    When the ACCOUNT tab loads
    Then the limit box is empty
    And the window choice is "no reset"
    And the "unlimited" box is not ticked

  Scenario: A mint with no name is refused on the page
    When the operator enters the name "" and the limit "5" and presses "Mint key"
    Then the console receives no mint request
    And the card says "give the key a name"

  Scenario Outline: A mint that states no usable limit is refused before the broker is asked
    When the operator enters the name "ci" and the limit "<limit>" and presses "Mint key"
    Then the broker receives no mint
    And the card says "<message>"

    Examples:
      | limit | message                                         |
      |       | state a limit in dollars, or tick "unlimited"    |
      | five  | state a limit in dollars, or tick "unlimited"    |
      | -1    | a limit cannot be negative                      |
      | 1e9   | state a limit in dollars, or tick "unlimited"    |
      | 0x10  | state a limit in dollars, or tick "unlimited"    |
      | 0     | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | $0    | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | 0.00  | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | -0    | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | 0.0000001 | state a limit in dollars, or tick "unlimited" |

  Scenario Outline: The limit accepts the ways people write dollars, as the CLI does
    When the operator enters the name "ci" and the limit "<typed>" and presses "Mint key"
    Then the broker receives a mint with name "ci" and limit_usd <sent>

    Examples:
      | typed | sent |
      | 5     | 5    |
      | $5    | 5    |
      | 5.50  | 5.5  |
      | 0.01  | 0.01 |

  Scenario Outline: A typed "none" is the same as ticking "unlimited" and carries the same warning
    When the operator enters the name "ci" and the limit "<typed>" and presses "Mint key"
    Then the broker receives a mint with name "ci" and limit_usd 0
    And the card says "Unlimited. This key can spend your whole wallet."

    Examples:
      | typed     |
      | none      |
      | unlimited |

  Scenario: "unlimited" mints with no limit and says what that means
    When the operator enters the name "ci", ticks "unlimited" and presses "Mint key"
    Then the broker receives a mint with name "ci" and limit_usd 0
    And the card says "Unlimited. This key can spend your whole wallet."

  Scenario: Ticking "unlimited" clears and locks the limit box
    When the operator enters the limit "5" and ticks "unlimited"
    Then the limit box is empty and cannot be typed in

  Scenario: The button is off while a mint is in flight
    Given the broker is slow to answer
    When the operator presses "Mint key" twice
    Then the broker receives one mint

  Scenario: A grant create in flight cannot be sent twice
    Given the broker is slow to answer
    When the operator presses "Create" twice
    Then the broker receives one grant create

  Scenario: The broker's refusal is shown in its own words
    Given the account already has 32 account keys
    When the operator mints a key
    Then the card says "at most 32 keys per account - delete one first"
    And no secret is shown

  # ---------------------------------------------------------------------------------------
  # Where the secret is, and is not
  # ---------------------------------------------------------------------------------------

  Scenario Outline: The secret leaves the page and cannot be brought back
    Given the operator minted a key and the secret is showing
    When the operator <action>
    Then the secret is not on the page
    And no console route returns the secret

    Examples:
      | action                              |
      | presses "done"                      |
      | opens another tab and comes back    |
      | reloads the page                    |
      | mints a second key                  |

  Scenario Outline: A secret that arrives after the operator left is dropped, and the page says so
    Given the broker is slow to answer
    When the operator presses "Mint key" and <action> before the answer
    Then the broker receives one mint
    And no secret is shown
    And the page says "Key ci was minted. It was made while you were away, so its secret was dropped. Revoke it and make another."

    Examples:
      | action                           |
      | opens another tab and comes back |
      | logs out                         |

  Scenario: A grant secret that arrives after the operator left is dropped too
    Given the broker is slow to answer
    When the operator presses "Create" and opens another tab and comes back before the answer
    Then no secret is shown
    And the page says "Grant friends was created. It was made while you were away, so its secret was dropped. Revoke it and make another."

  Scenario Outline: The secret is in no place that outlives the page
    Given the operator minted a key
    Then the secret is not in <place>

    Examples:
      | place                               |
      | the reply to GET /api/state         |
      | the reply to GET /api/keys          |
      | the event stream                    |
      | the page's address                  |
      | localStorage or sessionStorage      |
      | the node log                        |
      | the console's standard output       |

  Scenario Outline: Replies that touch keys are never cached
    When the page calls "<route>"
    Then the reply carries "Cache-Control: no-store"

    Examples:
      | route             |
      | /api/keys         |
      | /api/keys/mint    |
      | /api/keys/set     |
      | /api/keys/revoke  |
      | /api/grants       |

  # ---------------------------------------------------------------------------------------
  # Changing a key
  # ---------------------------------------------------------------------------------------

  Scenario: Changing the limit from the row
    Given the account has an account key "key_aaaa1" named "ci" with limit 5 weekly
    When the operator edits the row's limit to "10" and saves
    Then the broker receives a change to "key_aaaa1" with only limit_usd 10
    And the row reads "$10.00 weekly"

  Scenario: Changing the window from the row
    Given the account has an account key "key_aaaa1" named "ci" with limit 5 weekly
    When the operator changes the row's window to "monthly" and saves
    Then the broker receives a change to "key_aaaa1" with only reset "monthly"

  Scenario Outline: Turning a key off and back on
    Given the account has an account key "key_aaaa1" that is <before>
    When the operator presses "<button>" on its row
    Then the broker receives a change to "key_aaaa1" with only disabled <value>
    And the row's state reads "<after>"

    Examples:
      | before   | button   | value | after    |
      | active   | turn off | true  | disabled |
      | disabled | turn on  | false | active   |

  Scenario Outline: A bad limit on a row is refused before the broker is asked
    Given the account has an account key named "ci"
    When the operator edits the row's limit to "<typed>" and saves
    Then the broker receives no change
    And the row says "<message>"

    Examples:
      | typed | message                    |
      | ten       | state a limit in dollars, or "none" for unlimited |
      | -1        | a limit cannot be negative |
      | 1e9       | state a limit in dollars, or "none" for unlimited |
      | 0.0000001 | state a limit in dollars, or "none" for unlimited |
      | 0         | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | 0.00      | 0 would mean unlimited, not no spending - use none for an unlimited key |

  Scenario: Emptying a row's limit says how to remove it
    Given the account has an account key "key_aaaa1" named "ci" with limit 5 weekly
    When the operator edits the row's limit to "" and saves
    Then the broker receives no change
    And the row says "state a limit in dollars, or "none" for unlimited"

  Scenario Outline: Typing "none" on a row removes the limit
    Given the account has an account key "key_aaaa1" named "ci" with limit 5 weekly
    When the operator edits the row's limit to "<typed>" and saves
    Then the broker receives a change to "key_aaaa1" with only limit_usd 0
    And the row reads "unlimited"
    And the row says "Unlimited. This key can spend your whole wallet."

    Examples:
      | typed     |
      | none      |
      | unlimited |

  Scenario: Saving a row nobody edited sends nothing
    Given the account has an account key "key_aaaa1" named "ci" with limit 5 weekly
    When the operator edits the row's limit to "5" and saves
    Then the console receives no change request

  Scenario: A failed change leaves the row as the broker has it
    Given the broker refuses the change
    When the operator edits a row's limit and saves
    Then the row shows the broker's reason
    And the row's limit reads what it read before

  # ---------------------------------------------------------------------------------------
  # Revoking a key
  # ---------------------------------------------------------------------------------------

  Scenario: Revoking asks first
    Given the account has an account key "key_aaaa1" named "ci"
    When the operator presses "revoke" on its row
    Then the console receives no revoke request
    And the row asks "Revoke ci? Anything using it stops working." with "revoke" and "keep"

  Scenario: "keep" changes nothing
    Given the operator was asked to confirm revoking "ci"
    When the operator presses "keep"
    Then the console receives no revoke request
    And the row is back to normal

  Scenario: Confirming revokes and the row leaves the list
    Given the operator was asked to confirm revoking "ci"
    When the operator presses "revoke"
    Then the broker receives a revoke of "key_aaaa1"
    And the card no longer lists "ci"

  # ---------------------------------------------------------------------------------------
  # Grants: the gaps closed
  # ---------------------------------------------------------------------------------------

  Scenario: The grant list shows each grant's name, price and status
    Given the account has the grants
      | name    | price | status  |
      | my-bots | free  | active  |
      | old     | free  | revoked |
    When the ACCOUNT tab loads
    Then the "Grants" card has a row "my-bots" showing "free" and "active"
    And the "Grants" card has a row "old" showing "free" and "revoked"
    And no row reads "grant"

  Scenario: The grant list reply uses lower-case field names
    When the page reads "/api/grants"
    Then each item has the fields "name", "price" and "status"

  Scenario: Revoking a grant asks first
    Given the account has a grant named "my-bots"
    When the operator presses "revoke" on its row
    Then the console receives no revoke request
    And the row asks "Revoke my-bots? Anyone using it loses access." with "revoke" and "keep"

  Scenario: Confirming revokes the grant
    Given the operator was asked to confirm revoking grant "my-bots"
    When the operator presses "revoke"
    Then the broker receives a revoke of grant "my-bots"
    And the row's status reads "revoked"

  Scenario: Revoking a grant whose name was used before revokes the live one
    Given the account has a grant named "old" that is revoked
    And the account has a grant named "old"
    When the operator presses "revoke" on the live "old" row and confirms
    Then the broker receives a revoke of the live grant "old"

  Scenario: A revoked grant has no revoke button
    Given the account has a grant named "old" that is revoked
    When the ACCOUNT tab loads
    Then the row for "old" has no "revoke" button

  Scenario: A grant made here is free, and the form has only a name
    When the operator creates a grant
    Then the broker receives a grant create with name "friends" and free true
    And the "Grants" card has no "free" box
    And the "Grants" card has a row "friends" showing "free" and "active"

  Scenario: A new grant's copy button copies a paste-ready block
    Given the operator created a grant and the secret is showing
    When the operator presses "copy"
    Then the clipboard holds exactly these lines, with the broker's address and the new secret
      """
      export OPENAI_BASE_URL=<broker>/v1
      export OPENAI_API_KEY=<secret>
      """

  Scenario: An account that cannot make grants is told what to use instead
    Given the broker refuses grants for this account
    When the operator creates a grant
    Then the console receives a grant create request
    And the "Grants" card carries the broker's reason
    And the "Grants" card says "For your own bot, mint an account key above."

  Scenario: Grants refused does not hide account keys
    Given the broker refuses grants for this account
    And the account has an account key named "ci"
    When the ACCOUNT tab loads
    Then the "Account keys" card lists "ci"

  # ---------------------------------------------------------------------------------------
  # The routes, called directly
  # ---------------------------------------------------------------------------------------

  Scenario Outline: Every key route needs the console token
    When "<method> <route>" is called with <token>
    Then the reply is 403
    And the broker is not asked

    Examples:
      | method | route               | token         |
      | GET    | /api/keys           | no token      |
      | GET    | /api/keys           | a wrong token |
      | POST   | /api/keys/mint      | no token      |
      | POST   | /api/keys/mint      | a wrong token |
      | POST   | /api/keys/set       | no token      |
      | POST   | /api/keys/revoke    | no token      |
      | POST   | /api/grants/revoke  | no token      |
      | GET    | /api/grants         | no token      |
      | POST   | /api/grants         | no token      |
      | POST   | /api/grants         | a wrong token |

  Scenario Outline: Routes that change something are POST only
    When "<method> <route>" is called with the console token
    Then the reply is 405
    And the broker is not asked

    Examples:
      | method | route               |
      | GET    | /api/keys/mint      |
      | GET    | /api/keys/set       |
      | GET    | /api/keys/revoke    |
      | GET    | /api/grants/revoke  |
      | PUT    | /api/keys/mint      |
      | DELETE | /api/keys/revoke    |

  Scenario Outline: The list route is GET only
    When "<method> /api/keys" is called with the console token
    Then the reply is 405

    Examples:
      | method |
      | POST   |
      | PUT    |
      | DELETE |

  Scenario Outline: A malformed body is refused before the broker is asked
    When "POST <route>" is called with the body <body>
    Then the reply is 400 with "<message>"
    And the broker is not asked

    Examples:
      | route               | body                                          | message                    |
      | /api/keys/mint      | not JSON                                      | bad request                |
      | /api/keys/mint      | {}                                            | name required              |
      | /api/keys/mint      | {"name":"ci"}                                 | limit_usd required         |
      | /api/keys/mint      | {"name":"ci","limit_usd":-1}                  | a limit cannot be negative |
      | /api/keys/mint      | {"name":"ci","limit_usd":"-1"}                | a limit cannot be negative |
      | /api/keys/mint      | {"name":"ci","limit_usd":"five"}              | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":""}                  | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":1e9}                 | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":true}                | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":null}                | limit_usd required         |
      | /api/keys/set       | {"id":"key_aaaa1","limit_usd":"ten"}          | state a limit in dollars, or "none" for unlimited |
      | /api/keys/set       | {"id":"key_aaaa1","limit_usd":""}             | state a limit in dollars, or "none" for unlimited |
      | /api/keys/mint      | {"name":"ci","limit_usd":0.0000001}           | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":0.0}                 | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | /api/keys/mint      | {"name":"ci","limit_usd":{"a":1}}             | state a limit in dollars, or tick "unlimited" |
      | /api/keys/mint      | {"name":"ci","limit_usd":[5]}                 | state a limit in dollars, or tick "unlimited" |
      | /api/keys/set       | {"id":"key_aaaa1","limit_usd":-1}             | a limit cannot be negative |
      | /api/keys/mint      | {"name":"ci","limit_usd":0}                   | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | /api/keys/mint      | {"name":"ci","limit_usd":"0"}                 | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | /api/keys/set       | {"id":"key_aaaa1","limit_usd":0}              | 0 would mean unlimited, not no spending - use none for an unlimited key |
      | /api/keys/mint      | {"name":"ci","limit_usd":5,"reset":"yearly"}  | reset is daily, weekly, monthly or none |
      | /api/keys/mint      | {"name":"ci","limit_usd":5,"extra":1}         | unknown field "extra"      |
      | /api/keys/set       | {}                                            | id required                |
      | /api/keys/set       | {"id":"key_aaaa1"}                            | nothing to change          |
      | /api/keys/set       | {"id":"key_aaaa1","secret":"x"}               | unknown field "secret"     |
      | /api/keys/revoke    | {}                                            | id required                |
      | /api/grants/revoke  | {}                                            | name required              |
      | /api/grants         | not JSON                                      | bad request                |
      | /api/grants         | {}                                            | name required              |
      | /api/grants         | {"name":"   "}                                | name required              |
      | /api/grants         | {"name":"friends","free":false}               | unknown field "free"       |
      # Row removed 2026-11-25 (founder, "3. yes"): {"name":"friends","price_out":5} was refused
      # as an unknown field until a priced grant got its own field (C9). Ruling S8 in
      # features/grants/grant_surfaces.feature now adds that field to create.

  Scenario: A limit of "none" is allowed and means unlimited, as on the CLI
    When "POST /api/keys/mint" is called with the body {"name":"ci","limit_usd":"none"}
    Then the broker receives a mint with name "ci" and limit_usd 0

  Scenario: A limit sent as typed text is read like a number
    When "POST /api/keys/mint" is called with the body {"name":"ci","limit_usd":"$5.50"}
    Then the broker receives a mint with name "ci" and limit_usd 5.5

  Scenario: An oversized body is refused
    When "POST /api/keys/mint" is called with a body larger than 4 KiB
    Then the reply is 413
    And the broker is not asked

  Scenario: An id cannot reach another broker path
    When "POST /api/keys/revoke" is called with the body {"id":"../../grants/x"}
    Then the reply is 400 with "bad key id"
    And the broker is not asked

  Scenario: The console is not reachable from another origin
    When a page on another origin calls "/api/keys" with the console token
    Then the browser is given no permission to read the reply

  Scenario: Another origin's preflight is not answered
    When another origin sends a preflight for "/api/keys/mint"
    Then the browser is given no permission to send the request

  Scenario: Grant revoke is refused in LAN mode
    Given the console is in LAN mode
    When "POST /api/grants/revoke" is called with the body {"name":"friends"}
    Then the reply is 409 with "Keys need the public network. Not available in LAN mode."
    And the broker is not asked

  Scenario Outline: The grant routes are refused in LAN mode too
    Given the console is in LAN mode
    When "<method> /api/grants" is called with the console token
    Then the reply is 409 with "Keys need the public network. Not available in LAN mode."
    And the broker is not asked

    Examples:
      | method |
      | GET    |
      | POST   |

  Scenario Outline: The grant routes take GET and POST only
    When "<method> /api/grants" is called with the console token
    Then the reply is 405
    And the broker is not asked

    Examples:
      | method |
      | PUT    |
      | DELETE |
      | PATCH  |

  Scenario: An oversized grant create is refused
    When "POST /api/grants" is called with a body larger than 4 KiB
    Then the reply is 413
    And the broker is not asked

  Scenario Outline: Logged out, the grant routes say so without the terminal wording
    Given the account is logged out
    When "<method> /api/grants" is called with the console token
    Then the reply is 401 with "log in to manage keys"
    And the reply does not contain "roger login"

    Examples:
      | method |
      | GET    |
      | POST   |

  Scenario: The broker's logged-out answer never shows its terminal wording
    Given the broker says the account is logged out only when keys are asked for
    When "GET /api/keys" is called with the console token
    Then the reply is 401 with "log in to manage keys"
    And the reply does not contain "roger login"

  Scenario: The console is not reachable from another machine
    When "/api/keys" is called from a non-loopback address
    Then the request is refused

  # ---------------------------------------------------------------------------------------
  # When things are unavailable
  # ---------------------------------------------------------------------------------------

  Scenario: Logged out, the card says so
    Given the account is logged out
    When the ACCOUNT tab loads
    Then the "Account keys" card says "Log in to manage keys."
    And the mint form is hidden

  Scenario: The broker's "run roger login" wording is not shown in the browser
    Given the account is logged out
    When "GET /api/keys" is called with the console token
    Then the reply is 401 with "log in to manage keys"
    And the reply does not contain "roger login"

  Scenario: The broker cannot be reached
    Given the broker is unreachable
    When the ACCOUNT tab loads
    Then the "Account keys" card says the broker could not be reached
    And the "Grants" card does not say "mint an account key above"
    And the rest of the ACCOUNT tab still loads

  Scenario: LAN mode has no account, so no keys
    Given the console is in LAN mode
    When the ACCOUNT tab loads
    Then the "Account keys" card says "Keys need the public network. Not available in LAN mode."
    And the broker is not asked

  Scenario: Logging out clears a showing secret
    Given the operator minted a key and the secret is showing
    When the operator logs out
    Then the secret is not on the page

  # ---------------------------------------------------------------------------------------
  # The endpoint plate points here
  # ---------------------------------------------------------------------------------------

  Scenario: The plate's server note names the card, not the terminal
    Given the operator has a channel open from the CHAT tab on a market model
    Then the endpoint plate says "A bot on a server needs an account key instead: ACCOUNT tab, Account keys"
    And the endpoint plate does not contain "roger keys mint"
