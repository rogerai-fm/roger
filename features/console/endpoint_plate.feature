# APPROVED by the founder 2026-11-22, with rulings R1-R7 below accepted as proposed.
# Enforced by: internal/webui/endpoint_plate_bdd_test.go (godog, strict).
#
# THE CONSOLE'S ENDPOINT PLATE ("give me the OpenAI endpoint and key", in a browser).
#
# WHY THIS EXISTS
#
# `roger use <model>` ends on a BASE URL / API KEY / MODEL plate (internal/client/client.go,
# Use). The TUI's tune-in ends on the same plate (internal/tui/view_connect.go, endpointBlock)
# and `/endpoint` copies it. The browser console stops at CHAT: internal/webui/chat.go relays
# a turn through client.ChatTurns and returns a reply, so an operator working in the browser
# can talk to a band but has nothing to point a tool at. This is the same plate, drawn in the
# console, over the same hardened local proxy.
#
# WHAT IT IS, AND WHAT IT IS NOT
#
# The plate is a LOCAL endpoint: http://127.0.0.1:<port>/v1, guarded by a per-session bearer
# key, alive only while this `roger` process runs. It is for a tool on THIS machine. It is not
# a key for a bot on a server. That is an account key (`roger keys mint`, the site's API keys
# page, features/auth/key_guardrails.feature), and the plate says so in one line rather than
# letting someone paste a 127.0.0.1 URL into a hosted app and wonder why it fails.
#
# ONE PROXY, NOT A SECOND IMPLEMENTATION. The channel is client.ProxyHandlerLive over a
# client.ProxyOptionsHolder, the handler `roger use` and the TUI already serve. Everything
# features/proxy/*.feature says about that handler (bearer auth on every route, model
# rewrite, the 4 MiB body cap, OpenAI-shaped errors, Retry-After, stream bounds, the "no band
# tuned" refusal) holds here unchanged and is not restated. This file specifies only what the
# console adds: opening, showing, copying, re-pointing and closing it from the browser.
#
# THE CONSOLE HOLDS ITS OWN CHANNEL [PROPOSED]. The TUI's channel lives in the TUI's model,
# not in the shared *node.Controller, and the console already keeps its own agent session for
# the same reason (internal/webui/server.go, agentSess). So the console opens its own
# listener and mints its own key. With the TUI on 4141 the console's lands on the next free
# port. The two channels are independent: closing one never closes the other.
#
# THE RULE THE WHOLE PLATE OBEYS
#
# THE KEY IS HANDED OVER ONLY WHEN ASKED FOR, AND ONLY TO THE PAGE THAT HOLDS THE RUN TOKEN.
# It is never in /api/state, never in the /api/events stream, never in a GET, never in a URL,
# never in a log line, never in browser storage, and never drawn in full. This is the browser
# twin of the TUI rule "the full key never sits on screen" (view_connect.go, audit P0).
#
# MONEY. A request through the plate spends from the operator's wallet exactly as a console
# chat turn or `roger use` does. Opening a channel is therefore a write: POST only, token
# gated (server.go, action). The operator's per-band caps (Options.ReadLimits, the store the
# TUI's [4] CONFIG edits) bound it, and the account's monthly cap bounds it at the broker.
#
# GROUND TRUTH: internal/webui/server.go (routes, auth, action, localhostOnly, Options),
# internal/webui/chat.go (the open-market-only rule, recordConsoleSession), internal/webui/
# limits.go (SpendLimit), internal/webui/assets/console.{html,js,css}, internal/client/
# client.go (Use, ProxyHandlerLive, ProxyOptionsHolder, NewSessionKey, ConsumerDefaultMaxOut
# = 10, ConsumerConfirmThreshold = 20), internal/tui/tui.go (bindChannel, maskKey),
# internal/tui/view_connect.go (the confirm, the anonymous-on-a-paid-band refusal,
# connectExport), internal/netbind (ListenFree), features/proxy/*.feature.
#
# FOUNDER RULINGS NEEDED (each scenario that depends on one is marked [PROPOSED]):
#   R1  Own channel (proposed) or mirror the TUI's open channel when both run in one process.
#   R2  Session spend budget. Proposed: none, as `roger use` (the per-band caps and the
#       monthly cap apply). The alternative is the guest-operator DefaultSessionBudget.
#   R3  Private bands. Proposed: refused, as console chat is ("the console reaches the open
#       market only", chat.go). The alternative is a frequency-code field on the plate.
#   R4  LAN mode. Proposed: the plate is unavailable and says so, until LAN routing for the
#       console proxy is specified.
#   R5  Where it sits. Proposed: on the CHAT tab, under the composer, for the model in the
#       picker. The alternative is a row action on BROWSE.
#   R6  Whether closing the last console tab closes the channel. Proposed: no. A bot keeps
#       running when the tab is closed, and the channel ends on Close or when roger stops.
#   R7  The confirm line. Proposed: the TUI's, a typed out-price above $20 per 1M out.
#
# Step definitions run the REAL console server (httptest) and the
# REAL proxy handler against a REAL broker stand-in on loopback, read the REAL JSON, and
# assert on the shipped console.html / console.js. No mocks.

Feature: The console's endpoint plate

  Background:
    Given a console server over a node controller with broker "https://broker.test"
    And the run token is "tok-run"
    And band "qwen3-32b" is on air on the open market at $0.20 in / $0.60 out per 1M
    And band "free-4b" is on air on the open market, free now
    And the operator is logged in with balance $20.00
    And no channel is open

  # --- opening -------------------------------------------------------------------
  Scenario: Opening a channel returns the plate
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response is 200 with "ok" true
    And "base_url" is "http://127.0.0.1:<port>/v1" for the port the listener bound
    And "model" is "qwen3-32b"
    And "key" is a fresh session key
    And "key_hint" is the last 4 characters of that key
    And "station" names the station the channel locked

  Scenario: The endpoint answers with the key the plate returned
    Given a channel is open on "qwen3-32b"
    When a tool GETs <base_url>/models with "Authorization: Bearer <key>"
    Then the response is 200 listing exactly "qwen3-32b"

  Scenario: A chat completion through the plate reaches the band and is receipted
    Given a channel is open on "qwen3-32b"
    When a tool POSTs <base_url>/chat/completions with the key
    Then the broker stand-in receives one relay for "qwen3-32b" as the operator
    And the tool receives the reply

  Scenario: The listener starts at 4141 and scans upward when the port is taken
    Given port 4141 on 127.0.0.1 is already bound
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then "base_url" names a port above 4141
    And the plate shows that port, not 4141

  Scenario: The listener binds loopback only
    Given a channel is open on "qwen3-32b"
    Then the listener's address is 127.0.0.1
    And a connection to the same port on a non-loopback interface is refused

  Scenario: A free band opens without a login
    Given the operator is not logged in
    When the page POSTs /api/channel/open with {"model":"free-4b"}
    Then the response is 200 with "ok" true

  Scenario: A paid band is refused without a login, and nothing is bound
    Given the operator is not logged in
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response has "ok" false and "login_needed" true
    And "message" says the band is paid and to log in on ACCOUNT
    And no listener is bound and no key is minted

  Scenario: A model with no station on air is refused in the broker's own words
    When the page POSTs /api/channel/open with {"model":"nobody-serves-this"}
    Then the response has "ok" false
    And "message" carries the resolver's reason, not "open failed"
    And no listener is bound and no key is minted

  Scenario: No broker configured is a plain refusal
    Given a console server with no broker
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response is 503 and "message" says a channel needs a broker

  Scenario Outline: A malformed open is a 400 and binds nothing
    When the page POSTs /api/channel/open with <body>
    Then the response is 400 and "message" is "<message>"
    And no listener is bound and no key is minted

    Examples:
      | body                 | message              |
      | {"model":""}         | pick a model first   |
      | {"model":"   "}      | pick a model first   |
      | {}                   | pick a model first   |
      | not json             | malformed request    |
      | {"model":["a","b"]}  | malformed request    |

  Scenario: An open body over 4 KiB is refused before parsing
    When the page POSTs /api/channel/open with a 5 KiB body
    Then the response is 413
    And no listener is bound and no key is minted

  Scenario: A bind failure leaves no half-open channel
    Given no loopback port can be bound
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response has "ok" false and "message" says the endpoint could not bind
    And GET /api/channel reports "open" false

  # --- price and caps --------------------------------------------------------------
  Scenario: The plate shows what a reply costs before it opens
    When the page GETs /api/channel/quote?model=qwen3-32b
    Then the response carries the price now in and out, the live range, the station count
    And it carries the operator's cap for the band and whether the price is above it
    And it mints nothing and binds nothing

  Scenario: The operator's band cap rides the channel
    Given the operator's cap for "qwen3-32b" is $1.00 per 1M out
    And a channel is open on "qwen3-32b"
    When a tool POSTs <base_url>/chat/completions with the key
    Then the relay carries a max out price of $1.00

  Scenario: With no cap set the consumer default cap applies
    Given the operator has set no cap for "qwen3-32b"
    And a channel is open on "qwen3-32b"
    When a tool POSTs <base_url>/chat/completions with the key
    Then the relay carries a max out price of $10.00

  Scenario: A band above the operator's cap does not open
    Given the operator's cap for "qwen3-32b" is $0.50 per 1M out
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response has "ok" false and "over_limit" true
    And "message" names the cheapest price on air and the operator's cap
    And no listener is bound and no key is minted

  Scenario: A price exactly at the cap opens
    Given the operator's cap for "qwen3-32b" is $0.60 per 1M out
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response is 200 with "ok" true

  Scenario: A cap changed in SETTINGS binds the next request on an open channel
    Given a channel is open on "qwen3-32b"
    When the operator sets the cap for "qwen3-32b" to $0.30 per 1M out
    And a tool POSTs <base_url>/chat/completions with the key
    Then the relay carries a max out price of $0.30
    And the key and the port are unchanged

  Scenario: An expensive band needs the price typed back [PROPOSED R7]
    Given band "big-405b" is on air at $25.00 out per 1M
    And the operator's cap for "big-405b" is $30.00 per 1M out
    When the page POSTs /api/channel/open with {"model":"big-405b"}
    Then the response has "ok" false and "confirm_needed" true with "price_out" 25.00
    And no listener is bound and no key is minted

  Scenario Outline: The typed price must match the quoted price exactly [PROPOSED R7]
    Given band "big-405b" is on air at $25.00 out per 1M
    And the operator's cap for "big-405b" is $30.00 per 1M out
    When the page POSTs /api/channel/open with {"model":"big-405b","confirm_price_out":<typed>}
    Then the response has "ok" <ok>

    Examples:
      | typed | ok    |
      | 25.00 | true  |
      | 25    | true  |
      | 24.99 | false |
      | 25.01 | false |
      | 0     | false |
      | -25   | false |

  Scenario: A price that moved between the quote and the confirm is quoted again [PROPOSED R7]
    Given band "big-405b" was quoted at $25.00 out per 1M
    And it is now on air at $27.00 out per 1M
    When the page POSTs /api/channel/open with {"model":"big-405b","confirm_price_out":25.00}
    Then the response has "ok" false and "confirm_needed" true with "price_out" 27.00

  Scenario: A band exactly at the $20 line opens without the typed confirm [PROPOSED R7]
    Given band "mid-70b" is on air at $20.00 out per 1M
    And the operator's cap for "mid-70b" is $30.00 per 1M out
    When the page POSTs /api/channel/open with {"model":"mid-70b"}
    Then the response is 200 with "ok" true

  Scenario: The channel carries no session budget [PROPOSED R2]
    Given a channel is open on "qwen3-32b"
    When tools spend more than the guest-operator session budget through it
    Then no request is refused with 402 by the proxy
    And the broker's monthly cap still refuses when it is reached

  Scenario: The monthly cap refusal reaches the tool in the broker's words
    Given the account's monthly cap is reached
    And a channel is open on "qwen3-32b"
    When a tool POSTs <base_url>/chat/completions with the key
    Then the tool receives the broker's refusal as an OpenAI-shaped error

  # --- the key ---------------------------------------------------------------------
  Scenario: The key is not in the state snapshot
    Given a channel is open on "qwen3-32b"
    When the page GETs /api/state
    Then the body does not contain the key

  Scenario: The key is not in the event stream
    Given a channel is open on "qwen3-32b"
    When the page reads three frames of /api/events
    Then no frame contains the key

  Scenario: The channel status carries the hint, never the key
    Given a channel is open on "qwen3-32b"
    When the page GETs /api/channel
    Then the response has "open" true, "base_url", "model" "qwen3-32b", "station" and "key_hint"
    And the body does not contain the key

  Scenario: The key is fetched again only by a POST
    Given a channel is open on "qwen3-32b"
    When the page POSTs /api/channel/key
    Then the response carries the same "key" the open returned

  Scenario: A GET for the key is refused
    Given a channel is open on "qwen3-32b"
    When the page GETs /api/channel/key
    Then the response is 405
    And the body does not contain the key

  Scenario: Asking for the key with no channel open is a plain refusal
    When the page POSTs /api/channel/key
    Then the response is 409 and "message" is "no channel open"

  Scenario Outline: Every reply that carries the key forbids caching
    Given a channel is open on "qwen3-32b"
    When the page POSTs <path>
    Then the response has "Cache-Control: no-store"

    Examples:
      | path              |
      | /api/channel/open |
      | /api/channel/key  |

  Scenario: The key is never written to the console's log
    Given the console logs requests
    When a channel is opened, its key fetched, and closed
    Then no log line contains the key

  Scenario: The key never appears in a URL
    Then console.js builds no URL that contains the key
    And the plate's copy action posts and reads a JSON body

  Scenario: The page draws the key masked
    Given a channel is open on "qwen3-32b"
    Then the plate shows bullets followed by the last 4 characters of the key
    And the full key is in no element's text and no attribute

  Scenario: The page keeps the key out of browser storage
    Then console.js writes the key to neither localStorage nor sessionStorage nor a cookie

  Scenario: A new key is minted for each new channel
    Given a channel was opened and closed
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the "key" differs from the earlier channel's key

  Scenario: The session key is not an account key and does not work at the broker
    Given a channel is open on "qwen3-32b"
    Then the key does not start with "rog-key_" or "rog-grant_"
    And a request to the broker stand-in with that key as its bearer cannot spend

  Scenario Outline: The endpoint refuses a missing or wrong key on every route
    Given a channel is open on "qwen3-32b"
    When a tool calls <route> with <bearer>
    Then the response is 401 with an OpenAI-shaped "authentication_error"
    And the broker stand-in receives nothing

    Examples:
      | route                     | bearer                    |
      | GET /v1/models            | no Authorization header   |
      | GET /v1/models            | Bearer wrong              |
      | POST /v1/chat/completions | no Authorization header   |
      | POST /v1/chat/completions | Bearer wrong              |
      | POST /v1/chat/completions | the console run token     |
      | POST /v1/chat/completions | the key with one char off |
      | POST /v1/chat/completions | the key of a closed channel |

  Scenario: The console run token is not accepted by the endpoint, and the key is not accepted by the console
    Given a channel is open on "qwen3-32b"
    When a request to /api/state carries the key as its token
    Then the response is 403

  # --- who may open, read and close ----------------------------------------------------
  Scenario Outline: Every channel route needs the run token
    When a request to <method> <path> carries <token>
    Then the response is 403
    And nothing is opened, read or closed

    Examples:
      | method | path               | token            |
      | POST   | /api/channel/open  | no token         |
      | POST   | /api/channel/open  | a wrong token    |
      | POST   | /api/channel/key   | no token         |
      | POST   | /api/channel/key   | a wrong token    |
      | POST   | /api/channel/close | no token         |
      | POST   | /api/channel/close | a wrong token    |
      | GET    | /api/channel       | no token         |
      | GET    | /api/channel       | a wrong token    |
      | GET    | /api/channel/quote | no token         |

  Scenario Outline: A write by GET is refused even with the token
    When the page GETs <path> with the run token
    Then the response is 405
    And nothing is opened or closed

    Examples:
      | path               |
      | /api/channel/open  |
      | /api/channel/key   |
      | /api/channel/close |

  Scenario: A non-loopback peer is refused before the token is looked at
    When a request to POST /api/channel/open arrives from a non-loopback address with the run token
    Then the response is 403 "console is localhost-only"

  Scenario: A cross-site form post cannot open a channel
    When another origin's page POSTs /api/channel/open as a form without the run token
    Then the response is 403
    And no listener is bound and no key is minted

  Scenario: The channel routes answer no cross-origin preflight
    When another origin sends OPTIONS /api/channel/key
    Then the response carries no Access-Control-Allow-Origin header

  # --- one channel, re-pointed -----------------------------------------------------------
  Scenario: Opening another model re-points the same endpoint and keeps the key
    Given a channel is open on "qwen3-32b"
    When the page POSTs /api/channel/open with {"model":"free-4b"}
    Then "base_url" and "key" are unchanged
    And "model" is "free-4b"
    And GET <base_url>/models with the key lists exactly "free-4b"

  Scenario: A request in flight during a re-point finishes on the band it started on
    Given a channel is open on "qwen3-32b" and a stream is in flight
    When the page POSTs /api/channel/open with {"model":"free-4b"}
    Then the stream completes from "qwen3-32b" and is receipted to it
    And the next request goes to "free-4b"

  Scenario: Opening the model that is already open changes nothing
    Given a channel is open on "qwen3-32b"
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response is 200 with the same "base_url" and "key"
    And exactly one listener is bound

  Scenario: Two opens at the same moment yield one channel
    When two pages POST /api/channel/open with {"model":"qwen3-32b"} at once
    Then both responses carry the same "base_url" and "key"
    And exactly one listener is bound

  Scenario: Two opens at the same moment for different models end on one of them
    When two pages POST /api/channel/open at once, one for "qwen3-32b" and one for "free-4b"
    Then exactly one listener is bound
    And GET /api/channel names the model the later one asked for
    And both responses carry the same "base_url" and "key"

  Scenario: A failed re-point keeps the open channel as it was
    Given a channel is open on "qwen3-32b"
    When the page POSTs /api/channel/open with {"model":"nobody-serves-this"}
    Then the response has "ok" false
    And GET /api/channel still names "qwen3-32b" with the same "base_url" and "key_hint"

  Scenario: A second tab sees the open channel and can copy the key
    Given a channel was opened from one tab
    When a second tab with the run token GETs /api/channel
    Then it shows the same plate
    And its copy action receives the same key

  Scenario: The console's channel is not the TUI's [PROPOSED R1]
    Given the TUI in the same process has a channel open on port 4141
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the console's "base_url" names another port and its "key" differs from the TUI's
    And the TUI's endpoint still answers with the TUI's key

  Scenario: Closing the console's channel leaves the TUI's open [PROPOSED R1]
    Given the TUI and the console each have a channel open
    When the page POSTs /api/channel/close
    Then the TUI's endpoint still answers with the TUI's key

  # --- closing -----------------------------------------------------------------------
  Scenario: Closing ends the endpoint and the key
    Given a channel is open on "qwen3-32b"
    When the page POSTs /api/channel/close
    Then the response is 200 and GET /api/channel reports "open" false
    And a connection to the old port is refused
    And POST /api/channel/key is 409

  Scenario: Closing twice is not an error
    Given a channel was opened and closed
    When the page POSTs /api/channel/close
    Then the response is 200 and "open" is false

  Scenario: A stream in flight at close completes and is receipted
    Given a channel is open on "qwen3-32b" and a stream is in flight
    When the page POSTs /api/channel/close
    Then the stream completes and its turn is receipted
    And a new request to the old port is refused

  Scenario: Closing the browser tab does not close the channel [PROPOSED R6]
    Given a channel is open on "qwen3-32b"
    When every console tab is closed
    Then the endpoint still answers with the key

  Scenario: The console stopping ends the channel
    Given a channel is open on "qwen3-32b"
    When the console server stops
    Then a connection to the old port is refused

  Scenario: Logging out closes a channel on a paid band
    Given a channel is open on "qwen3-32b"
    When the operator logs out on ACCOUNT
    Then GET /api/channel reports "open" false
    And a connection to the old port is refused

  Scenario: Logging out leaves a channel on a free band open
    Given a channel is open on "free-4b"
    When the operator logs out on ACCOUNT
    Then the endpoint still answers with the key

  # --- what it reaches ---------------------------------------------------------------
  Scenario: A private band is refused [PROPOSED R3]
    Given band "hidden-8b" is private behind a frequency code
    When the page POSTs /api/channel/open with {"model":"hidden-8b"}
    Then the response has "ok" false
    And the request body has no field that could carry a frequency code

  Scenario: An unknown field in the open body is a 400
    When the page POSTs /api/channel/open with {"model":"qwen3-32b","freq":"147.520 MHz 8F3K-9M2Q"}
    Then the response is 400 naming "freq"
    And no listener is bound and no key is minted

  Scenario: LAN mode says the plate is unavailable [PROPOSED R4]
    Given the console runs in LAN mode
    When the page POSTs /api/channel/open with {"model":"qwen3-32b"}
    Then the response has "ok" false and "message" says the plate is not available in LAN mode yet
    And the broker stand-in receives nothing

  Scenario: The tool cannot reach past the band
    Given a channel is open on "qwen3-32b"
    When a tool POSTs <base_url>/chat/completions naming model "free-4b" with the key
    Then the relay goes to "qwen3-32b"

  Scenario: Each receipted turn is a session on this machine's Edge
    Given a channel is open on "qwen3-32b"
    When a tool completes one turn through it
    Then the Edge session ledger gains one session attributed to the console

  Scenario: The station going off air is reported, and the endpoint stays
    Given a channel is open on "qwen3-32b"
    When every station for "qwen3-32b" leaves the air
    And a tool POSTs <base_url>/chat/completions with the key
    Then the tool receives an OpenAI-shaped error saying no station is on air
    And GET /api/channel still reports "open" true with the same "base_url"

  # --- the page ----------------------------------------------------------------------
  Scenario: The plate sits on the CHAT tab for the model in the picker [PROPOSED R5]
    Given the chat model picker is on "qwen3-32b"
    Then the CHAT tab shows an "open an endpoint" control naming "qwen3-32b"
    And it shows the price now and the operator's cap beside it

  Scenario: The open plate reads like the terminal's
    Given a channel is open on "qwen3-32b"
    Then the plate shows three rows labelled "BASE URL", "API KEY" and "MODEL"
    And a line "drop-in, OpenAI-compatible - point any OpenAI tool here."
    And a "copy" control and a "close" control

  Scenario: Copy puts the export block on the clipboard
    Given a channel is open on "qwen3-32b"
    When the operator presses "copy"
    Then the clipboard holds three lines:
      """
      export OPENAI_BASE_URL=<base_url>
      export OPENAI_API_KEY=<key>
      export OPENAI_MODEL=qwen3-32b
      """
    And a toast says it was copied
    And the plate still shows the key masked

  Scenario: A refused clipboard write says so and shows nothing
    Given the browser refuses clipboard writes
    When the operator presses "copy"
    Then a toast says the copy failed
    And the full key is still in no element's text

  Scenario: The plate says what it is not
    Given a channel is open on "qwen3-32b"
    Then the plate says the endpoint works on this machine only, while roger is running
    And it says a bot on a server needs an account key, naming "ACCOUNT tab, Account keys"

  Scenario: A refusal is shown in the server's words
    When an open is refused with "message" "this band is paid - log in on ACCOUNT to use your wallet"
    Then the plate shows that message
    And it offers the ACCOUNT tab when "login_needed" is true

  Scenario: The plate reflects a channel closed elsewhere
    Given two tabs show the open plate
    When one tab closes the channel
    Then the other tab shows no plate within 2 seconds

  Scenario: The plate loads nothing from another origin
    Then console.html and console.js reference no third-party script, style or font for the plate

  Scenario: The plate is usable without colour and by keyboard
    Then the copy and close controls are buttons reachable by Tab with visible labels
    And the open and closed states differ by text, not by colour alone
    And the masked key has an accessible name that does not contain the key
