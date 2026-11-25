# Private stations on their own frequency, for every kind of station.
#
# APPROVED by the founder 2026-11-15. Founder ruling 2026-11-15 (marketplace visibility
# audit): every kind of station can be private, unlisted and reachable only with its band's
# frequency code, and the code works on every relay.
#
# What is broken today: `roger share --private` mints a band and a code for a tts, stt or
# search share, but only the chat relay reads X-Roger-Freq (cmd/rogerai-broker/tunnel.go,
# privateAllow). The voice relays and the search relay pass no privateAllow to pickFor, which
# drops every private node, so a private voice or search station can never be reached.
#
# Ground truth: tunnel.go relay (resolveFreqAllow, privateAllow, the uniform "no station on that
# frequency" 503), band.go resolveFreqAllow / bandOffers, audio.go audioRelayCore, search.go.
#
# DECISIONS PINNED BY THIS SPEC (approved 2026-11-15):
#   PF1 /v1/audio/speech, /v1/audio/transcriptions and /v1/search read X-Roger-Freq exactly as
#       the chat relay does: the same constant-work lookup, the same uniform 503 on any miss.
#   PF2 a code admits ONLY its band's station; it is never spend authorization. A paid private
#       station still needs a signed wallet or a grant; a free one serves anyone with the code.
#   PF3 a code never widens a grant: a grant plus a code reaches the band only when the band's
#       node is inside the grant.
#   PF4 a code is never logged raw on any relay.
#   PF5 a private station of any kind appears on no public listing (/discover, /market,
#       /voices, /search/stations, /v1/models, /categories counts).

@stations @private @security
Feature: Every kind of station can be private and reached only with its frequency code

  Background:
    Given a signed-in owner "@acme"

  Rule: a private station is reachable with its code on every relay (PF1)

    Scenario Outline: A private station answers a request carrying its code
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When a caller sends a <route> request for "<model>" with that band's code in X-Roger-Freq
      Then the private station answers it

      Examples:
        | kind   | model         | route                    |
        | chat   | llama3.2      | /v1/chat/completions     |
        | tts    | front-desk    | /v1/audio/speech         |
        | stt    | whisper-small | /v1/audio/transcriptions |
        | search | go-docs       | /v1/search               |

    Scenario Outline: Without the code, a private station is unreachable
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When a caller sends a <route> request for "<model>" with no code
      Then the response status is 503
      And the private station is never called

      Examples:
        | kind   | model         | route                    |
        | chat   | llama3.2      | /v1/chat/completions     |
        | tts    | front-desk    | /v1/audio/speech         |
        | stt    | whisper-small | /v1/audio/transcriptions |
        | search | go-docs       | /v1/search               |

    Scenario Outline: A wrong, revoked or off-air code reads the same on every relay
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When a caller sends a <route> request with <code>
      Then the response status is 503
      And the error says there is no station on that frequency
      And the body is identical across wrong, revoked and off-air codes, except the request id

      Examples:
        | kind   | model         | route                    | code                       |
        | tts    | front-desk    | /v1/audio/speech         | a code that was never minted |
        | tts    | front-desk    | /v1/audio/speech         | the band's revoked code    |
        | stt    | whisper-small | /v1/audio/transcriptions | a code that was never minted |
        | search | go-docs       | /v1/search               | a code that was never minted |
        | search | go-docs       | /v1/search               | the band's revoked code    |

    Scenario Outline: A namespaced id with the code reaches the private station
      # Regression (pre-push audit 2026-11-15): namespaced resolution skipped private nodes, so
      # "@acme/<model>" with a valid code answered 503.
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When a caller sends a <route> request for "@acme/<model>" with that band's code in X-Roger-Freq
      Then the private station answers it

      Examples:
        | kind   | model      | route            |
        | tts    | front-desk | /v1/audio/speech |
        | search | go-docs    | /v1/search       |

    Scenario: A search with a code and no model goes to the band's station
      Given "@acme" shares a search station "go-docs" privately at price 0
      And a public search station "web-search" is on air
      When a caller sends a /v1/search request with no model and that band's code
      Then "go-docs" answers it

    Scenario: A code with no model reaches only the models its band carries
      # Regression (pre-push audit 2026-11-21): the band's model list was checked only when the
      # request named a model, so a search with no model could be answered by another search
      # offer on the same node.
      Given "@acme" shares search offers "go-docs" and "rust-docs" from one node privately, with a band that carries only "go-docs"
      When a caller sends 40 /v1/search requests with no model and that band's code
      Then every one is answered for "go-docs"

    Scenario: A code with no model for a band whose models are all off the node gets the uniform 503
      Given "@acme" shares search offers "go-docs" and "rust-docs" from one node privately, with a band that carries only "zig-docs"
      When a caller sends a /v1/search request with no model and that band's code
      Then the response status is 503
      And the error says there is no station on that frequency
      And the answer reads the same as for a code that was never minted

    Scenario Outline: A grant's model list and the band's model list both bind a search with no model
      Given "@acme" shares search offers "go-docs" and "rust-docs" from one node privately, with a band that carries only "go-docs"
      And a grant owned by "@acme" limited to "<grant model>"
      When a /v1/search request with no model arrives with that grant key and the band's code
      Then the response status is <status>
      And <reading>

      Examples:
        | grant model | status | reading                                                       |
        | go-docs     | 200    | "go-docs" answered                                            |
        | rust-docs   | 503    | the answer reads the same as for a code that was never minted |

    Scenario: A code for a chat band does not admit a voice request
      Given "@acme" shares a chat station "llama3.2" privately at price 0
      When a caller sends a /v1/audio/speech request with that band's code
      Then the response status is 503
      And no station is called

  Rule: a code admits, it never pays (PF2, PF3)

    Scenario Outline: A paid private station still needs a wallet
      Given "@acme" shares a <kind> station "<model>" privately at a price above 0
      When an unsigned caller from an allowed origin sends a <route> request with the band's code
      Then the response status is 403
      And the private station is never called

      Examples:
        | kind   | model      | route            |
        | tts    | front-desk | /v1/audio/speech |
        | search | go-docs    | /v1/search       |

    Scenario: A signed caller with the code pays the private station's price
      Given "@acme" shares a search station "go-docs" privately at 2000 credits per 1,000,000 queries
      And a consumer with a funded wallet
      When the consumer sends a signed /v1/search request for "go-docs" with the band's code
      Then the consumer is charged 0.002000 credits
      And "@acme" earns it under the standard fee

    Scenario: The owner's own private station costs them nothing
      Given "@acme" shares a search station "go-docs" privately at 2000 credits per 1,000,000 queries
      When "@acme" sends a signed /v1/search request with the band's code
      Then "@acme" is charged 0 credits

    Scenario: A grant plus a code cannot reach a band outside the grant
      Given "@zed" shares a tts station "zed-voice" privately at price 0
      And a grant owned by "@acme"
      When a /v1/audio/speech request arrives with that grant key and "@zed"'s band code
      Then the response status is 503
      And "@zed"'s station is never called

  Rule: codes stay secret and private stations stay unlisted (PF4, PF5)

    Scenario Outline: A code is never logged raw
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When a caller sends a <route> request with that band's code
      Then the broker log does not contain the code

      Examples:
        | kind   | model         | route                    |
        | tts    | front-desk    | /v1/audio/speech         |
        | stt    | whisper-small | /v1/audio/transcriptions |
        | search | go-docs       | /v1/search               |

    Scenario Outline: A private station is on no public listing
      Given "@acme" shares a <kind> station "<model>" privately at price 0
      When the public listings are read
      Then "<model>" is on none of /discover, /market, /voices, /search/stations, /v1/models
      And no category count includes it

      Examples:
        | kind   | model         |
        | chat   | llama3.2      |
        | tts    | front-desk    |
        | stt    | whisper-small |
        | search | go-docs       |
