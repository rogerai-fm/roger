# APPROVED SPEC - founder approved 2026-11-24 ("approved"). Changes to an approved scenario need re-approval.
# BUILD STATUS: BUILT. internal/station/execute.go (HTTPUpstream.Key, RedactKey), internal/station/edge.go
# and edge_sealed.go (OnUpstreamError), internal/agent/tower.go (the joined-Tower executor carries the
# key; upstreamNotice), internal/agent/agent.go (checkUpstreamKey, also in ServeTower and ServeLocalTower).
#
# Found 2026-11-24 in a security check of the house shares. A share whose local server needs an
# API key (`roger share --upstream-key`, a saved config key, or the onboarding prompt; every
# curated house station uses one) sends it as "Authorization: Bearer" on the classic path
# (internal/agent/agent.go) and on the local Tower path (localtower.go, localtower_stream.go),
# and redacts it from what it sends back.
#
# The two JOINED-Tower paths, pull (internal/agent/tower.go) and hub (internal/towerhub/serve.go),
# both serve through station.EdgeExecutor, built with station.HTTPUpstream{URL: cfg.Upstream}.
# HTTPUpstream has no key field (internal/station/execute.go). So:
#   - nothing leaks today: no key is sent, so none can be echoed (checked 2026-11-24, including
#     the local journal, the broker logs and every repo and scratch file);
#   - but every request a key-protected Station serves through a joined Tower reaches its
#     provider with no key and fails. Until 2026-11-23 no signed-in share could attach, so this
#     never showed. Now that attaching works, a house curated station attaches when it restarts.
#
# Adding the key creates the leak surface this file guards. On the sealed path a Tower sees
# only ciphertext, so the party a leak would reach is the CONSUMER: every leak check below names
# the consumer as well as the Tower and Core.
#
# Also found: on the sealed path an upstream error is reduced to its class for the wire
# (edge_sealed.go), and the detail is logged nowhere, although a comment there says "the
# operator still has the full error in their own logs". The operator line below is new.

Feature: A key-protected upstream works through a joined Tower, and its key never leaves the machine

  Background:
    Given a signed-in operator sharing model "m" from an upstream that requires bearer key K
    And K is at least 16 characters

  Rule: Every serving path authenticates to the upstream the same way

    Scenario Outline: The key reaches the upstream on every path
      Given the share took K from "<source>"
      When a request is served on the "<path>" path
      Then the upstream receives "Authorization: Bearer K"

      Examples:
        | source                 | path                   |
        | the --upstream-key flag | joined Tower, pull     |
        | the --upstream-key flag | joined Tower, hub      |
        | saved config           | joined Tower, pull     |
        | the onboarding prompt  | joined Tower, hub      |
        | the --upstream-key flag | classic, non-streaming |
        | the --upstream-key flag | classic, streaming     |
        | the --upstream-key flag | local Tower            |
      # The classic and local Tower rows already pass; they guard against regression.

    Scenario: A share with no upstream key sends no Authorization header
      Given a share with no upstream key
      When a request is served through a joined Tower
      Then the upstream receives no Authorization header

  Rule: K never reaches a Tower, a relay, Core or the consumer

    Scenario Outline: An upstream that echoes the key cannot leak it
      Given the upstream puts K into "<where>"
      When the request is served through a joined Tower
      Then K appears in nothing the Station sends to the Tower, to Core or to the consumer
      And "[redacted]" appears where K was

      Examples:
        | where                                                     |
        | a 200 response body                                       |
        | a streaming chunk                                         |
        | two streaming chunks, split in the middle of K            |
        | a 401 error body                                          |
        | a 500 error body                                          |
        | an error body longer than 200 bytes, with K across byte 200 |
        | a JSON string, with K's characters escaped as \u00XX      |

    Scenario: Redaction happens before anything is measured, signed or kept
      Given the upstream's 200 body contains K
      When the Station serves it through a joined Tower
      Then the bytes the receipt is signed over, the token count, the audit transcript and the consumer's bytes are the same redacted bytes

    Scenario: Transcripts kept for audit never hold K
      Given a request served through a joined Tower is kept in the Station's transcripts
      When Core's audit asks for that transcript
      Then the bytes sent hold no K

    Scenario Outline: K is never sent in a control message
      When the Station sends "<message>"
      Then the message holds no K and no field named for an upstream key

      Examples:
        | message      |
        | registration |
        | heartbeat    |
        | attach       |
        | receipt      |
        | hub poll     |

    Scenario: A key too short to redact safely is refused at start
      Given the operator gives an upstream key shorter than 8 characters
      When `roger share` starts
      Then it refuses to start, saying the key is too short to keep out of replies

    Scenario: The unused plain Serve path cannot become a leak
      Then EdgeExecutor.Serve, which forwards the upstream error text, is removed, or redacts K before it forwards anything
      # It has no production caller today.

  Rule: An upstream that refuses its key is reported plainly, to the operator only

    Scenario: The upstream answers 401 through a joined Tower
      Given the upstream rejects K with 401
      When a request is served through a joined Tower
      Then the attempt produces no receipt, the consumer's hold is released, and it is never a strike
      And `roger share` prints once, on the operator's machine only, that the upstream refused its key
      And that line is redacted before it is shortened, and holds no prefix of K of 8 or more characters
      And the consumer is told only the error class, as today
