# APPROVED by the founder 2026-11-24 ("ok fix this, lets make sure we have a similar experience
# throughout our clients"). D5 and the local-server scenarios were added in the same pass, for that
# parity: the iOS app already falls back on a broker 503 and on an Edge member that refuses tools.
# REVISED after review and RE-APPROVED by the founder 2026-11-26 ("approved"): plain chat sticks only once the retry answers (a band
# briefly off air is refused in the same words), the fallback is first-call-of-turn only, a refusal
# naming a routing filter is left as today, and LAN mode is named out of scope.
#
# The agent answers as plain chat on a band that cannot drive tools, as the terminal app
# already tells the user it will.
#
# WHY. Found 2026-11-23 by an outside test of the two Wave Nano fire bands through the public
# broker. The terminal app's agent and the web console's agent send the full tools array on
# every turn. A non-empty tools array implicitly requires the tools capability
# (features/routing/capability_gating.feature, approved), so on a band with no tools-capable
# station the broker answers 503 no_match "no node offers <model> with the tools capability".
# Every agent turn on such a band fails. Meanwhile the terminal app says on entry "this band
# cannot drive tools - the agent will answer as plain chat" (features/answers/tui_display.feature),
# which is the behavior this spec makes true. It affects every band without tools, not only
# the fire bands.
#
# GROUND TRUTH (origin/main 7cf851ce):
#   - internal/harness/loop.go Loop.Send calls l.complete(ctx, l.messages, ToolSchemas(l.tools))
#     on every step. It already retries ONCE per turn on one named error (context overflow,
#     auto-compaction), with a notice in the transcript.
#   - internal/harness/broker.go BrokerCompleterRoute puts "tools" and "tool_choice": "auto"
#     in every request body. A broker refusal comes back as an error carrying the broker's text.
#   - The broker's capability no_match (cmd/rogerai-broker/tunnel.go) is written before any
#     hold is placed, so a refused request is never billed. It names the capability as
#     "with the tools capability" or "with the tools and vision capability" WHENEVER the request
#     needs tools, whatever emptied the pool: a band briefly off air (every broker deploy), a
#     grant owner not serving, or a routing filter (" under <filters>") get the same suffix. So
#     the refusal alone does not prove the band lacks tools; only a retry that succeeds does.
#   - internal/tui/agent.go, internal/webui/agent.go and the Edge delegated turn
#     (cmd/rogerai/edgerunner.go) all run Loop, on BrokerCompleter or LocalCompleter.
#   - The terminal app keeps one loop across a /model switch (bindAgentEndpoint); the web
#     console builds a new loop per model.
#   - The tune-in chat (sendChat) is not affected: it sends no tools. The iOS app retries without
#     tools on a broker 503 and on an Edge member whose error names tools (ChatService.swift).
#
# DECISIONS:
#   D1. The fallback is reactive: the loop retries without tools when the request is refused for
#       the tools capability. It does not guess from published capabilities, because an absent
#       capability set is "undetermined", not "no tools" (tui_display.feature).
#   D2. Once a band has refused tools AND answered without them, the rest of that session on that
#       model is plain chat, so later turns do not pay a refused round trip first. A retry that
#       also fails changes nothing: no notice, and the next turn sends tools again. A new model,
#       or a new session, starts with tools again.
#   D3. An explicit tools requirement from the user's own routing (roger use --require tools, or
#       a profile that requires tools) is never dropped: the refusal is shown as today.
#   D4. Only the tools refusal falls back, and only on the first model call of a turn: a refusal
#       after the turn already ran a tool means the tools station went away mid-turn, not that
#       the band cannot drive tools. A refusal that names a routing filter (" under ...") is the
#       operator's own rule at work and is shown as today. Any other no_match, and every other
#       error, behaves exactly as today.
#   ACCEPTED: if a band's only tools station is briefly off air while a plain station of the same
#       band answers, the retry succeeds and plain chat sticks for that session. The band truly
#       could not take tools then; /clear, /model or (web console) a reload starts with tools again.
#   OUT OF SCOPE (said, not hidden): LAN mode (roger --lan) folds a peer's refusal into
#       "every local source serving M failed (peer: status N)" without the peer's words, so a LAN
#       peer that refuses tools is not recognised. It needs the peer's text carried first.
#   D5. A model server reached directly (a local model, an Edge delegated turn) that refuses the
#       tools field falls back the same way. Recognised refusals: llama.cpp "tools param requires
#       --jinja flag", vLLM "auto tool choice requires --enable-auto-tool-choice", and Ollama
#       "does not support tools".

Feature: The agent answers as plain chat on a band that cannot drive tools

  Background:
    Given the agent is on model "solo-m"
    And no station of "solo-m" has the tools capability

  # --- the fallback --------------------------------------------------------------------------

  Scenario: A turn the broker refuses for tools is sent again without tools
    When the user sends "hello" to the agent
    Then the broker receives a request for "solo-m" with a non-empty tools array
    And the broker answers 503 no_match "no node offers solo-m with the tools capability"
    And the broker then receives a second request for "solo-m" with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer

  Scenario: The second request carries the same conversation
    Given the agent's session holds a system persona and two earlier turns
    When the user sends "hello" to the agent and the broker refuses it for tools
    Then the second request's messages equal the first request's messages

  Scenario: The transcript says once that the band is answering as plain chat
    When the user sends "hello" to the agent and the broker refuses it for tools
    Then the transcript shows a notice that "solo-m" cannot drive tools and is answering as plain chat
    When the user sends "and again" to the agent
    Then no second notice is shown

  Scenario: Later turns on the same model go straight to plain chat
    Given the broker refused an earlier turn of this session for tools
    When the user sends "and again" to the agent
    Then the broker receives exactly one request for that turn
    And it has no "tools" key and no "tool_choice" key

  Scenario: A new model starts with tools again
    Given the broker refused an earlier turn of this session for tools
    When the agent moves to model "tool-m"
    And the user sends "hello" to the agent
    Then the broker receives a request for "tool-m" with a non-empty tools array

  Scenario: A new session starts with tools again
    Given the broker refused an earlier turn of an earlier session for tools
    When a new agent session starts on "solo-m"
    And the user sends "hello" to the agent
    Then the broker receives a request for "solo-m" with a non-empty tools array

  Scenario: The fallback happens at most once per turn
    When the user sends "hello" to the agent
    And the broker refuses the first request for tools
    And the broker refuses the second request with 503 no_match "no node offers solo-m"
    Then the turn ends with the error the second refusal gives
    And the broker received exactly two requests for that turn

  Scenario: A retry that also fails changes nothing
    When the user sends "hello" to the agent
    And the broker refuses the first request for tools
    And the broker refuses the second request with 503 no_match "no node offers solo-m with the tools capability"
    Then the turn ends with the error the second refusal gives
    And no plain chat notice is shown
    When the user sends "and again" to the agent
    Then the broker receives a request for "solo-m" with a non-empty tools array
    # A band off air during a deploy refuses with the same words. Plain chat must not stick.

  Scenario: A tools refusal after a tool already ran in the turn behaves as today
    Given the first reply of the turn calls a tool
    And no station of "solo-m" has the tools capability after that reply
    When the user sends "hello" to the agent
    Then the turn ends with the error it gives today
    And the broker received exactly two requests for that turn

  Scenario: The refused request is not billed
    When the user sends "hello" to the agent and the broker refuses it for tools
    Then the turn's cost counts only the second request
    # The broker refuses before it places a hold, and the refusal carries no cost headers.

  Scenario: A streamed turn falls back the same way
    Given the agent streams its turns
    When the user sends "hello" to the agent and the broker refuses it for tools
    Then the broker then receives a second request with no "tools" key and "stream" true
    And the agent shows the streamed second reply as its answer

  Scenario: A band that needs tools and vision is retried without tools
    When the user sends "hello" to the agent
    And the broker answers 503 no_match "no node offers solo-m with the tools and vision capability"
    Then the broker then receives a second request for "solo-m" with no "tools" key and no "tool_choice" key

  # --- what does not fall back ----------------------------------------------------------------

  Scenario: An explicit tools requirement is never dropped
    Given the user's routing requires "tools"
    When the user sends "hello" to the agent
    And the broker answers 503 no_match "no node offers solo-m with the tools capability"
    Then the turn ends with that refusal, as today
    And the broker received exactly one request for that turn

  Scenario Outline: Other refusals behave as today
    When the user sends "hello" to the agent
    And the broker answers <status> with "<message>"
    Then the turn ends with the error it gives today
    And the broker received exactly one request for that turn

    Examples:
      | status | message                                                 |
      | 503    | no node offers solo-m                                   |
      | 503    | no node offers solo-m with the vision capability        |
      | 503    | no node offers solo-m in region eu                      |
      | 402    | insufficient balance                                    |
      | 400    | this request exceeds the context window of solo-m       |
      | 429    | rate limited                                            |
      | 500    | the tools capability is down for maintenance            |
      | 503    | no node offers solo-m with the tools capability under max_price |
      | 503    | no node offers solo-m with the tools capability under region    |

  Scenario: A band that drives tools is unchanged
    Given a station of "solo-m" has the tools capability
    When the user sends "hello" to the agent
    Then the broker receives exactly one request for that turn, with a non-empty tools array

  Scenario: A tool call that arrives on a plain-chat turn is not run
    Given the broker refused an earlier turn of this session for tools
    When the second reply carries a tool_calls array
    Then no tool is run
    And the agent shows the reply's text as its answer
    And the conversation keeps no tool_calls from that reply

  # --- a model server reached directly (D5) --------------------------------------------------

  Scenario Outline: A local model server that refuses tools falls back the same way
    Given the agent is on a local model whose server refuses tools with <status> "<message>"
    When the user sends "hello" to the agent
    Then the local server then receives a second request with no "tools" key and no "tool_choice" key
    And the agent shows the second reply as its answer
    And the transcript shows a notice that the model cannot drive tools and is answering as plain chat

    Examples:
      | status | message                                                                              |
      | 500    | tools param requires --jinja flag                                                    |
      | 400    | "auto" tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set |
      | 400    | registry.ollama.ai/library/gemma:2b does not support tools                           |

  Scenario: Any other local server error behaves as today
    Given the agent is on a local model whose server answers 500 "failed to process speculative batch"
    When the user sends "hello" to the agent
    Then the turn ends with the error it gives today
    And the local server received exactly one request for that turn

  # --- every front end -------------------------------------------------------------------------
  # Untagged scenarios run against the loop and the real completers (internal/harness). The
  # tagged ones drive the front end itself: @tui in internal/tui, @webui in internal/webui.

  @tui
  Scenario: The terminal app's agent falls back
    Given the terminal app's agent is on model "solo-m"
    When the user sends "hello" and the broker refuses it for tools
    Then the terminal app shows the second reply as its answer
    And the terminal app's transcript shows the plain chat notice

  @webui
  Scenario: The web console's agent falls back
    Given the web console's agent is on model "solo-m"
    When the user sends "hello" and the broker refuses it for tools
    Then the web console shows the second reply as its answer
    And the web console streams the plain chat notice

  @tui
  Scenario: The terminal app's /model switch starts with tools again
    Given the terminal app's agent fell back to plain chat on "solo-m"
    When the operator switches the agent to "tool-m" with /model
    Then the next turn sends a non-empty tools array

  @tui
  Scenario: The terminal app's /clear starts with tools again
    Given the terminal app's agent fell back to plain chat on "solo-m"
    When the operator clears the session with /clear
    Then the next turn sends a non-empty tools array

  @tui
  Scenario: The terminal app's entry warning is unchanged
    Given the tuned band declares capabilities that do not include "tools"
    When the user enters the agent
    Then the transcript says this band cannot drive tools
    # The existing tui_display.feature scenario; this spec makes its promise true.

  # --- regression: the fire bands ----------------------------------------------------------------

  Scenario: A fire tick typed into the agent gets the controller's decision
    Given the agent is on model "wave-nano-fire-controller"
    And no station of "wave-nano-fire-controller" has the tools capability
    When the user sends a bare tick to the agent
    Then the broker's second request carries the persona, the tick and no tools
    And the agent shows the station's reply as its answer
