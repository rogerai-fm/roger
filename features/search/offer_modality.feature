# Search as a marketplace modality: layer 1 (protocol).
#
# APPROVED by the founder 2026-11-15. This file and the rest of features/search/ (except
# failover.feature, slice 2) are the contract the search build follows.
#
# Search joins chat, tts and stt as a fourth offer modality. Anyone can serve it, exactly as
# anyone can serve a model or a voice: a provider runs `roger share` in front of a search
# backend, and the broker relays POST /v1/search to them on the same money path the voice
# relay uses (cmd/rogerai-broker/audio.go, audioRelayCore).
#
# Ground truth: internal/protocol/protocol.go (ModelOffer.Modality, canonicalUnit,
# ValidModality, Normalize) and internal/protocol/taxonomy.go (voice category derivation).
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   S1  the modality is "search", the billing unit is "query", canonicalized from the modality.
#   S2  the price is stored per 1,000,000 queries, the same per-1M arithmetic every other
#       modality uses, and shown to people per 1,000 queries.
#   S3  a search offer declares its ENGINE: "brave", "searxng" or "native". The engine is
#       shown everywhere the station is listed. It is the honesty label: a station reselling
#       a commercial API never looks like someone's own index.
#   S4  a search offer's category is always "retrieval" with subcategory "search" (the
#       broker's derivation, never a claim), the same way a voice offer is always "audio".
#       Founder ruling 2026-11-15: search sits under retrieval, next to embeddings and rerank.
#   S5  a search offer may carry an optional SCOPE: a short plain-text line saying what its
#       index covers ("the open web", "Go module docs"). Bounded, flattened, screened.

@search @protocol
Feature: Search offers carry a modality, a truthful unit and an engine label

  Background:
    Given a node registering an offer with the broker

  Rule: search is a known modality billed per query (S1, S2)

    Scenario: A search offer is billed per query
      Given an offer for model "web-search" with modality "search" and engine "searxng"
      When the broker normalizes the offer
      Then the offer unit is "query"
      And its price is read as credits per 1,000,000 queries

    Scenario Outline: The unit is canonicalized from the modality whatever was sent
      Given an offer with modality "search", engine "brave" and a claimed unit "<sent_unit>"
      When the broker normalizes the offer
      Then the offer unit is "query"

      Examples:
        | sent_unit |
        | query     |
        | token     |
        | char      |
        | byte      |
        |           |

    Scenario Outline: The modality enum stays closed
      Given an offer with modality "<modality>"
      When the broker validates the offer
      Then the offer is "<result>"

      Examples:
        | modality   | result   |
        | search     | accepted |
        | Search     | rejected |
        | SEARCH     | rejected |
        | web_search | rejected |
        | websearch  | rejected |
        | searches   | rejected |

    Scenario: Existing modalities are unchanged by the new one
      Given offers with modality "", "chat", "tts" and "stt"
      When the broker normalizes them
      Then their units are "token", "token", "char" and "byte" exactly as before

  Rule: every search offer names its engine, from a closed set (S3)

    Scenario Outline: Only known engines are accepted
      Given an offer with modality "search" and engine "<engine>"
      When the broker validates the offer
      Then the offer is "<result>"

      Examples:
        | engine  | result   |
        | brave   | accepted |
        | searxng | accepted |
        | native  | accepted |
        |         | rejected |
        | google  | rejected |
        | Brave   | rejected |
        | custom  | rejected |

    Scenario: An engine on a non-search offer is rejected
      Given an offer with modality "chat" and engine "brave"
      When the broker validates the offer
      Then the offer is "rejected"

    Scenario: The engine label is covered by the registration signature
      When a relay changes the engine label of a signed registration in flight
      Then the broker rejects the registration

  Rule: the category is derived, never claimed (S4)

    Scenario Outline: A search offer is always in retrieval/search
      Given an offer with modality "search", engine "native" and declared category "<declared>"
      When the broker normalizes the offer
      Then its category is "retrieval" with subcategory "search"
      And its taxonomy source is not "declared" for the category

      Examples:
        | declared         |
        |                  |
        | retrieval        |
        | retrieval/rerank |
        | chat     |
        | audio    |
        | vision   |

    Scenario: A chat offer cannot claim the search subcategory
      Given an offer with modality "chat" and declared category "retrieval/search"
      When the broker normalizes the offer
      Then its subcategory is not "search"

  Rule: the optional scope line is bounded and plain (S5)

    Scenario: A scope line rides the offer
      Given a search offer declaring scope "Go module documentation and the Go blog"
      When the broker normalizes the offer
      Then the offer scope is "Go module documentation and the Go blog"

    Scenario: A scope line is flattened to one line of plain text
      Given a search offer declaring a scope containing a newline, a tab and an HTML tag
      When the broker normalizes the offer
      Then the scope has no newline, tab or tag and its words are separated by single spaces

    Scenario: A scope line over 120 characters is cut at 120 characters
      Given a search offer declaring a scope of 500 characters
      When the broker normalizes the offer
      Then the offer scope is 120 characters long

    Scenario: A scope the content screen flags is dropped, the offer stays
      Given a search offer declaring a scope the content screen flags
      When the broker registers the offer
      Then the offer is on air with no scope

    Scenario: A scope on a non-search offer is dropped
      Given a chat offer declaring scope "the open web"
      When the broker normalizes the offer
      Then the offer has no scope

  Rule: mixed fleets and round trips

    Scenario: One node registers chat, tts and search offers together
      Given a node registers offers:
        | model      | modality | engine  |
        | llama3.2   | chat     |         |
        | kokoro     | tts      |         |
        | web-search | search   | searxng |
      When the broker normalizes the offers
      Then each offer keeps its own modality and canonical unit
      And the node is discoverable under all three

    Scenario: modality, unit, engine and scope survive a JSON round trip
      Given a search offer with engine "native" and scope "the open web"
      When it is marshalled to JSON and decoded back
      Then the modality, unit, engine and scope are preserved
