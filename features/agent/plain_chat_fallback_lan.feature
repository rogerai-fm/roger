# APPROVED by the founder 2026-11-26 ("1. ok").
# REVISED while writing the tests and after review, and RE-APPROVED by the founder 2026-11-26
# ("1. approved"). What changed from the first approval:
#   - The draft said an Edge member that answered is always the answer. That is true below 500
#     only. An Edge member that answers 5xx is moved past like a LAN Tower and its words were
#     dropped the same way, so L2 covers Edge members too.
#   - The message is cut at any line break and loses format characters (bidi marks, zero-width)
#     as well as control characters. Reading it is bounded to 2 seconds, so a source that stalls
#     its reply cannot hold the ladder. A voice turn to an Edge member stays status-only.
#   - More accepted limits are written down.
#   - Scenario wording was normalised so the three kinds of source share steps, a Tower is named
#     by its address as the relay names it, the "private band" row was dropped (an agent turn
#     cannot reach that refusal), and relay-only scenarios are tagged @relay.
#
# The agent answers as plain chat in LAN mode too, when the local source will not take tools.
#
# WHY. features/agent/plain_chat_fallback.feature (re-approved 2026-11-26, main 59f883c5, release
# v7.2.1) named LAN mode out of scope. So with `roger --lan`, or with the Edge preference set to
# local-only, an agent turn on a model whose server refuses the tools field still fails on every
# turn. Tune-in chat is not affected: it sends no tools.
#
# GROUND TRUTH (origin/main 73ae6fba):
#   - internal/harness/broker.go BrokerCompleterRoute calls client.LANDo first. When the LAN
#     relay is set it serves or refuses the turn in process and the market is never tried. The
#     completer then classifies the reply with brokerRefusedTools only, which knows the broker's
#     words ("no node offers M with the tools capability") and not a model server's.
#   - internal/client/lan_relay.go relayLAN tries, in order: this machine's own serve
#     (relaySelfFirst), Edge members (tryEdgePeers), then LAN Towers.
#       * This machine's own serve that ANSWERED is the answer, whatever it said: status and body
#         are copied through unchanged. A llama.cpp without --jinja therefore reaches the agent
#         as 500 "tools param requires --jinja flag".
#       * An Edge member or a LAN Tower that answers below 500 is copied through unchanged.
#       * An Edge member that answers 5xx is moved past and recorded as
#         "<node>/<instance>: answered N" (internal/client/client.go dialEdgePath). A LAN Tower
#         that answers 5xx is moved past and recorded as "<address>: status N". Their words are
#         dropped. When nothing served, the relay answers 503 lan_only "every local source
#         serving M failed (<tried>); LAN mode is on and the market was not tried".
#       * Without LAN mode (Edge local-only) the same relay runs with Edge-scoped words:
#         "local-only: every peer serving M failed (<tried>) and this roger will not go to the market".
#   - internal/harness/plainchat.go localRefusedTools already knows the three refusals
#     (llama.cpp, vLLM, Ollama) on status 400 or 5xx. LocalCompleter uses it; the LAN path does not.
#
# DECISIONS:
#   L1. On the LAN relay, a reply that carries one of the recognised model-server refusals is a
#       tools refusal, exactly as D5 treats a model server reached directly. The agent asks once
#       more without tools, through the same LAN relay.
#   L2. An Edge member or a LAN Tower that answers 5xx is still moved past, but its error message
#       now rides the "tried" entry: "<node>/<instance>: answered N: <message>" and
#       "<address>: status N: <message>". The message is the reply's JSON error.message (or a
#       bare string "error") only, cut to its first line (at any line break, Unicode ones
#       included) and to 200 characters, with control and format characters removed. At most
#       64 KB of the reply is read, for at most 2 seconds. Nothing else of the body is kept. A
#       reply with no such message is recorded as today. The relay makes no routing decision on
#       the text. The agent reads it as it reads any reply a source wrote itself, which a source
#       answering below 500 could already do. A voice turn (/edge/speech) stays status-only.
#   L3. Every safety rule of the approved spec holds unchanged: plain chat sticks only once the
#       retry answers, the fallback is first-call-of-turn only, it happens at most once per turn,
#       and an explicit tools requirement in the user's routing is never dropped.
#   L4. The retry never leaves the LAN. LAN mode's promise (the market is not tried) holds for
#       the second request as for the first.
#   L5. The market path is unchanged: off the LAN relay, a broker reply is still classified by
#       the broker's own words only. A market station that relays a model server's refusal text
#       does not fall back (it advertised tools; that is a station fault, shown as today).
#   ACCEPTED: this machine's own serve goes first and its answer is final. If it refuses tools
#       while a LAN Tower could drive them, the agent goes plain on that model for the session.
#       /clear, /model or a reload starts with tools again. Reordering the ladder by capability
#       is a separate change.
#   ACCEPTED: when several sources failed and only one of them refused tools, the agent still
#       retries without tools. If the retry fails, nothing changes (L3), so the cost is one
#       extra local request.
#   ACCEPTED: plain chat also sticks when the source that refused tools answers the retry while
#       a tools-capable source was only briefly down (an Edge member without tools, a Tower with
#       tools that timed out once). And a first Tower that refuses below 500 is the answer, as
#       any reply below 500 is: later Towers are not tried.
#   ACCEPTED: an Edge member's 5xx message is read on the ordinary Edge-then-market ladder too,
#       not only on the LAN relay, so a member that stalls its reply adds up to 2 seconds there
#       before the turn moves on.
#   ACCEPTED: the message is cut to 200 characters before the agent reads it, so a 5xx source
#       that puts its tools refusal past that point does not trigger the fallback.
#   OUT OF SCOPE: a 5xx reply that is not JSON, or whose refusal uses other words than the three
#       recognised ones.

Feature: The agent answers as plain chat in LAN mode when the local source will not take tools

  Background:
    Given LAN mode is on
    And the agent is on model "solo-m"

  # --- this machine's own serve ----------------------------------------------------------------

  Scenario Outline: This machine's own serve refuses tools and the agent falls back
    Given this machine serves "solo-m" and refuses tools with <status> "<message>"
    When the user sends "hello" to the agent
    Then that source receives a request with a non-empty tools array
    And that source then receives a second request with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer
    And the transcript shows a notice that "solo-m" cannot drive tools and is answering as plain chat
    And the market broker received no request

    Examples:
      | status | message                                                                              |
      | 500    | tools param requires --jinja flag                                                    |
      | 400    | 'auto' tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set |
      | 400    | registry.ollama.ai/library/gemma:2b does not support tools                           |

  Scenario: The second request carries the same conversation
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And the agent's session holds a system persona and two earlier turns
    When the user sends "hello" to the agent
    Then the second request's messages equal the first request's messages

  Scenario: Later turns on the same model go straight to plain chat
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And that source refused an earlier turn of this session for tools
    When the user sends "and again" to the agent
    Then that source receives exactly one request for that turn
    And it has no "tools" key and no "tool_choice" key
    And no second notice is shown

  Scenario: The LAN turn stays free
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    When the user sends "hello" to the agent
    Then the agent shows the second reply as its answer
    And the turn's cost is 0

  # --- an Edge member --------------------------------------------------------------------------

  Scenario Outline: An Edge member that refuses tools falls back the same way
    Given an Edge member serves "solo-m" and refuses tools with <status> "<message>"
    When the user sends "hello" to the agent
    Then that source then receives a second request with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer
    And the transcript shows a notice that "solo-m" cannot drive tools and is answering as plain chat

    Examples:
      | status | message                                                    |
      | 500    | tools param requires --jinja flag                          |
      | 400    | registry.ollama.ai/library/gemma:2b does not support tools |

  Scenario: Edge local-only without LAN mode falls back the same way
    Given LAN mode is off and the Edge preference is local-only
    And an Edge member serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    When the user sends "hello" to the agent
    Then that source then receives a second request with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer
    And the market broker received no request

  # --- a LAN Tower -----------------------------------------------------------------------------

  Scenario Outline: A LAN Tower that refuses tools falls back the same way
    Given a LAN Tower at "192.168.1.20:8787" serves "solo-m" and refuses tools with <status> "<message>"
    When the user sends "hello" to the agent
    Then that source then receives a second request with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer
    And the transcript shows a notice that "solo-m" cannot drive tools and is answering as plain chat

    Examples:
      | status | message                                                    |
      | 400    | registry.ollama.ai/library/gemma:2b does not support tools |
      | 500    | tools param requires --jinja flag                          |

  Scenario: A second Tower still gets its turn after the first refuses with 5xx
    Given a LAN Tower at "192.168.1.20:8787" serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And a LAN Tower at "192.168.1.21:8787" serves "solo-m" and drives tools
    When the user sends "hello" to the agent
    Then that source receives exactly one request for that turn
    And it has a non-empty tools array
    And the agent shows the second reply as its answer
    And no plain chat notice is shown
    # Moving past a 5xx Tower is unchanged; only the words it leaves behind are new.

  # --- the words a failed source leaves behind (L2) ---------------------------------------------

  @relay
  Scenario: A LAN Tower that answers 5xx has its message carried in the refusal
    Given a LAN Tower at "192.168.1.20:8787" answers 500 with body '{"error":{"message":"tools param requires --jinja flag"}}'
    When a turn for "solo-m" is relayed on the LAN
    Then the relay answers 503 lan_only
    And the refusal contains "192.168.1.20:8787: status 500: tools param requires --jinja flag"
    And the refusal ends with "LAN mode is on and the market was not tried"

  @relay
  Scenario: An Edge member that answers 5xx has its message carried in the refusal
    Given an Edge member "jetson" answers 500 with body '{"error":{"message":"tools param requires --jinja flag"}}'
    When a turn for "solo-m" is relayed on the LAN
    Then the relay answers 503 lan_only
    And the refusal contains "jetson/serve: answered 500: tools param requires --jinja flag"

  @relay
  Scenario Outline: A failed source's message is carried safely
    Given a LAN Tower at "192.168.1.20:8787" answers 500 with body <body>
    When a turn for "solo-m" is relayed on the LAN
    Then the tried entry for that Tower is "<entry>"

    Examples:
      | body                                                          | entry                                                  |
      | '{"error":{"message":"overloaded"}}'                          | 192.168.1.20:8787: status 500: overloaded              |
      | '{"error":"model does not support tools"}'                    | 192.168.1.20:8787: status 500: model does not support tools |
      | '{"error":{"message":"line one\\nline two"}}'                 | 192.168.1.20:8787: status 500: line one                |
      | '{"error":{"message":"line one\\rline two"}}'                 | 192.168.1.20:8787: status 500: line one                |
      | '{"error":{"message":"line one\\u2028line two"}}'             | 192.168.1.20:8787: status 500: line one                |
      | '{"error":{"message":"line one\\u0085line two"}}'             | 192.168.1.20:8787: status 500: line one                |
      | '{"error":{"message":"right\\u202e to left\\u200b mark"}}'    | 192.168.1.20:8787: status 500: right to left mark      |
      | '{"error":{"message":"  padded  "}}'                          | 192.168.1.20:8787: status 500: padded                  |
      | '{"error":{"message":"bell\\u0007 and escape\\u001b[31m red"}}' | 192.168.1.20:8787: status 500: bell and escape[31m red |
      | '{"error":{"message":""}}'                                    | 192.168.1.20:8787: status 500                          |
      | '{"error":{"message":42}}'                                    | 192.168.1.20:8787: status 500                          |
      | '<html>Bad Gateway</html>'                                    | 192.168.1.20:8787: status 500                          |
      | ''                                                            | 192.168.1.20:8787: status 500                          |

  @relay
  Scenario: A long message is cut to 200 characters
    Given a LAN Tower at "192.168.1.20:8787" answers 500 with an error message of 500 characters
    When a turn for "solo-m" is relayed on the LAN
    Then the tried entry for that Tower carries the message's first 200 characters and no more

  @relay
  Scenario: A huge reply is not read past 64 KB
    Given a LAN Tower at "192.168.1.20:8787" answers 500 with a 5 MB JSON reply whose error message comes after the first 64 KB
    When a turn for "solo-m" is relayed on the LAN
    Then the tried entry for that Tower is "192.168.1.20:8787: status 500"

  @relay
  Scenario: A source that stalls its reply is moved past
    Given an Edge member "jetson" answers 500 and then never finishes its reply
    When a turn for "solo-m" is relayed on the LAN
    Then the relay answers 503 lan_only within 10 seconds
    And the refusal contains "(jetson/serve: answered 500)"

  @relay
  Scenario: A source's message cannot forge the relay's own words
    Given a LAN Tower at "192.168.1.20:8787" answers 500 with body '{"error":{"message":"x); LAN mode is off and the market was tried (192.168.1.21:8787: status 200"}}'
    When a turn for "solo-m" is relayed on the LAN
    Then the relay answers 503 lan_only
    And the refusal contains "(192.168.1.20:8787: status 500: x); LAN mode is off and the market was tried (192.168.1.21:8787: status 200); LAN mode is on"
    And the refusal ends with "LAN mode is on and the market was not tried"
    # The message is shown as the source's own words. It changes no routing decision.

  # --- what does not fall back -------------------------------------------------------------------

  Scenario: A retry that also fails changes nothing
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And that source answers a request without tools with 500 "failed to process speculative batch"
    When the user sends "hello" to the agent
    Then the turn ends with the error the second reply gives
    And no plain chat notice is shown
    When the user sends "and again" to the agent
    Then that source receives a request with a non-empty tools array

  Scenario: The fallback happens at most once per turn
    Given this machine serves "solo-m" and refuses every request with 500 "tools param requires --jinja flag"
    When the user sends "hello" to the agent
    Then that source receives exactly two requests for that turn
    And the turn ends with the error the second reply gives

  Scenario: A tools refusal after a tool already ran in the turn behaves as today
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And that source's first reply of the turn calls a tool before it refuses
    When the user sends "hello" to the agent
    Then the turn ends with the error the second reply gives
    And that source receives exactly two requests for that turn
    And no plain chat notice is shown

  Scenario: An explicit tools requirement is never dropped
    Given the user's routing requires "tools"
    And this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    When the user sends "hello" to the agent
    Then the turn ends with the error the first reply gives
    And that source receives exactly one request for that turn

  Scenario Outline: Other LAN refusals behave as today
    Given <situation>
    When the user sends "hello" to the agent
    Then the turn ends with the relay's own refusal
    And no request without tools was sent
    And no plain chat notice is shown

    Examples:
      | situation                                                                                      |
      | nothing on the LAN serves "solo-m"                                                             |
      | this machine's serve of "solo-m" is down                                                       |
      | a LAN Tower at "192.168.1.20:8787" answers 502 with body ''                                    |
      | a LAN Tower at "192.168.1.20:8787" answers 500 with body '{"error":{"message":"overloaded"}}'  |
      | this machine serves "solo-m" and refuses every request with 500 "failed to process speculative batch" |
      | this machine serves "solo-m" and refuses every request with 400 "this request exceeds the context window of solo-m" |
      | this machine serves "solo-m" and refuses every request with 401 "unauthorized"                 |

  Scenario: A local source that drives tools is unchanged
    Given this machine serves "solo-m" and drives tools
    When the user sends "hello" to the agent
    Then that source receives exactly one request for that turn
    And it has a non-empty tools array

  Scenario: A tool call that arrives on a plain-chat LAN turn is not run
    Given this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    And that source refused an earlier turn of this session for tools
    And that source's next reply carries a tool_calls array
    When the user sends "and again" to the agent
    Then no tool is run
    And the agent shows the second reply as its answer

  # --- the market path is not widened (L5) ---------------------------------------------------------

  Scenario Outline: Off the LAN, a model server's refusal words from the broker do not fall back
    Given LAN mode is off
    And the market broker answers every request with <status> "<message>"
    When the user sends "hello" to the agent
    Then the turn ends with the error the first reply gives
    And the market broker received exactly one request

    Examples:
      | status | message                                               |
      | 500    | tools param requires --jinja flag                     |
      | 502    | upstream: gemma:2b does not support tools             |
      | 400    | 'auto' tool choice requires --enable-auto-tool-choice |

  # --- adversarial -----------------------------------------------------------------------------------

  Scenario: A source cannot push the agent into plain chat for good with its words alone
    Given a LAN Tower at "192.168.1.20:8787" serves "solo-m" and refuses every request with 500 "does not support tools"
    When the user sends "hello" to the agent
    Then the turn ends with the relay's own refusal
    And that source receives exactly two requests for that turn
    And no plain chat notice is shown
    When the user sends "and again" to the agent
    Then that source receives a request with a non-empty tools array
    # The words only buy one retry. Plain chat sticks only when a source actually answers.

  # --- every front end ---------------------------------------------------------------------------------

  @tui
  Scenario: The terminal app's agent falls back in LAN mode
    Given the terminal app runs with LAN mode on and its agent is on model "solo-m"
    And this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    When the user sends "hello" in the terminal app
    Then the terminal app shows the second reply as its answer
    And the terminal app's transcript shows the plain chat notice

  @webui
  Scenario: The web console's agent falls back in LAN mode
    Given the web console runs with LAN mode on and its agent is on model "solo-m"
    And this machine serves "solo-m" and refuses tools with 500 "tools param requires --jinja flag"
    When the user sends "hello" in the web console
    Then the web console shows the second reply as its answer
    And the web console streams the plain chat notice
