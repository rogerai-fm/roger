# Search API: who may call it, and which station answers.
#
# APPROVED by the founder 2026-11-15. Auth, rate limits, grant confinement and modality isolation are the
# voice relay's rules applied to search. The only new routing input is that "model" is optional:
# absent means "any search station", chosen by the same market selection chat uses.
#
# Ground truth: cmd/rogerai-broker/audio.go (resolveGrant, identityOf, walletOf, web session +
# originAllowed, grantRL / anonRL / rl, parseNamespacedVoice + resolveNamespacedVoice,
# gc.nodeAllow, gc.modelDenied, pickFor with pickReq.modality).
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   A1  founder ruling 2026-11-15: signed-out callers may use FREE search stations only, under
#       the anonymous per-IP limit. A paid station needs a signed request, a grant key or a
#       signed-in web session.
#   A2  /v1/search routes only to search offers; a search offer is never picked for chat, tts
#       or stt. No eligible station is a 503, never a cross-modality route.
#   A3  a request with no model is routed among ALL search offers the caller may use. When the
#       caller is signed out, that set is the free search offers only.
#   A4  a grant key confines search to its owner's nodes and model list, as for chat and voice.

@search @routing
Feature: Search reaches only search stations, on the same access rules as chat and voice

  Rule: authentication

    Scenario: A signed request is accepted
      Given a search station "web-search" on air at 5000 credits per 1,000,000 queries
      And a consumer with a funded wallet
      When the consumer sends a signed search for "web-search"
      Then the response status is 200

    # Rewritten 2026-11-26 (founder: "approved"): owner-pays ended with
    # features/grants/priced_grants.feature. A free grant is the API key; a priced one is not.
    Scenario: A free grant key is accepted as an OpenAI-style bearer key
      Given a search station "web-search" on air owned by the grant's owner
      When a search arrives with header "Authorization: Bearer <free grant key>"
      Then the response status is 200
      And nobody is charged

    Scenario: A priced grant key sent as a bearer key is refused
      Given a search station "web-search" on air owned by the grant's owner
      When a search arrives with header "Authorization: Bearer <priced grant key>"
      Then the response status is 401 saying "this grant is priced - log in and add it to your account"
      And the station is never called and nobody is charged

    Scenario Outline: A bad credential is refused before anything else
      When a search arrives with <credential>
      Then the response status is 401
      And no hold is placed
      And no moderation call is made

      Examples:
        | credential                         |
        | a revoked grant key                |
        | an unknown grant key               |
        | a signature over a different body  |
        | a signature by an unbound key      |
        | an expired web session cookie      |

    Scenario: A web session counts only from an allowed origin
      Given a valid web session cookie
      When a search arrives with that cookie from an origin that is not allowed
      Then the response status is 401

  Rule: signed-out callers use free stations only (A1, A3)

    Scenario: A signed-out caller searches a free station
      Given a search station "free-search" on air at 0 credits per 1,000,000 queries
      When an unsigned search for "free-search" arrives from an allowed origin with no cookie
      Then the response status is 200
      And the cost is 0 credits

    Scenario: A signed-out caller is refused a paid station
      Given a search station "web-search" on air at 5000 credits per 1,000,000 queries
      When an unsigned search for "web-search" arrives from an allowed origin with no cookie
      Then the response status is 403
      And the error asks the caller to sign in
      And the station is never called

    Scenario: A signed-out caller with no model is routed to a free station only
      Given a search station "web-search" on air at 5000 credits per 1,000,000 queries
      And a search station "free-search" on air at 0 credits per 1,000,000 queries
      When 50 unsigned searches with no model arrive from an allowed origin
      Then every one is answered by "free-search"

    Scenario: A signed-out caller with no model and no free station gets 503, not 403
      Given only paid search stations are on air
      When an unsigned search with no model arrives from an allowed origin
      Then the response status is 503

    # A device key that has never logged in (the phone app before sign-in) has no wallet to
    # spend from, so it is routed like a signed-out caller: free stations only (pre-push audit
    # 2026-11-22: it could be picked onto a paid station and refused while a free one was on air).
    Scenario: A signed device key with no login and no model is routed to a free station only
      Given a search station "web-search" on air at 5000 credits per 1,000,000 queries
      And a search station "free-search" on air at 0 credits per 1,000,000 queries
      When 50 signed searches with no model arrive from a device key with no login
      Then every one is answered by "free-search"

    Scenario: A signed device key with no login and no free station gets 503, not 403
      Given only paid search stations are on air
      When a signed search with no model arrives from a device key with no login
      Then the response status is 503

    Scenario: A signed device key with no login is served by a free station it names
      Given a search station "free-search" on air at 0 credits per 1,000,000 queries
      When a signed search for model "free-search" arrives from a device key with no login
      Then the response status is 200

    Scenario: A signed device key with no login is refused a paid station it names
      Given a search station "web-search" on air at 5000 credits per 1,000,000 queries
      When a signed search for model "web-search" arrives from a device key with no login
      Then the response status is 403
      And the station is never called

    Scenario: An unsigned request from outside a browser is refused
      When an unsigned search arrives with no Origin header
      Then the response status is 401

    Scenario: A spoofed legacy identity header earns no rate bucket of its own
      Given a search station "free-search" on air at 0 credits per 1,000,000 queries
      When unsigned searches arrive from one IP carrying 100 different legacy identity headers
      Then they all draw from the one anonymous bucket for that IP

  Rule: rate limits

    Scenario: The anonymous per-IP limit applies
      Given a search station "free-search" on air at 0 credits per 1,000,000 queries
      When unsigned searches from one IP exceed the anonymous limit
      Then the excess is refused with 429 and a Retry-After header
      And the station is not called for the excess

    Scenario: The per-account limit applies to signed searches
      When one account's signed searches exceed the per-account limit
      Then the excess is refused with 429 and a Retry-After header

    Scenario: A grant key's own RPM applies
      Given a grant key with an RPM of 2
      When three searches arrive with that key inside one minute
      Then the third is refused with 429 and a Retry-After header

    Scenario: Search and chat share the account's limit, they do not double it
      When one account sends chat and search requests together above the per-account limit
      Then the excess is refused with 429 whichever route it came in on

  Rule: modality isolation (A2)

    Scenario: A search goes to a search station
      Given a chat node offering "llama3.2"
      And a search station "web-search" on air
      When a search with no model arrives
      Then it is relayed to the search station

    Scenario Outline: A search never reaches a station of another modality
      Given only a <modality> node offering "web-search" is on air
      When a search for "web-search" arrives
      Then the response status is 503
      And that node is never called

      Examples:
        | modality |
        | chat     |
        | tts      |
        | stt      |

    Scenario Outline: A search station is never picked for another route
      Given only a search station "web-search" is on air
      When a POST <route> for model "web-search" arrives
      Then the search station is never called

      Examples:
        | route                    |
        | /v1/chat/completions     |
        | /v1/audio/speech         |
        | /v1/audio/transcriptions |

    Scenario: The 503 for an unknown model reads the same as one for an offline model
      When a search for "no-such-search" arrives
      And a search for a model whose only station just went off air arrives
      Then both responses are 503 with the same body

  Rule: choosing a station

    Scenario: A raw model id routes among the stations offering it
      Given two stations both offering "web-search"
      When 100 searches for "web-search" arrive
      Then every search is answered by one of those two stations

    Scenario: A namespaced id pins the named operator's station
      Given station "@acme" offers "web-search" and station "@zed" offers "web-search"
      When a search for "@acme/web-search" arrives
      Then it is answered by "@acme" and "@acme" is the one paid

    Scenario: A namespaced id whose station is off air is 503, never another operator
      Given station "@acme" is off air and station "@zed" offers "web-search"
      When a search for "@acme/web-search" arrives
      Then the response status is 503
      And "@zed" is never called

    Scenario: With no model, selection uses the same market rule chat uses
      Given three search stations at different prices and latencies
      When searches with no model arrive
      Then each is routed by the same selection function the chat relay uses, restricted to search offers

    # Found by the push audit of 2026-11-23: with latency equal, price did not weigh at all.
    Scenario: With no model and equal latency, the cheaper station is picked more often
      Given three search stations that differ only in price
      When searches with no model arrive
      Then the cheapest station answers more often than the dearest

    Scenario: A banned node is never picked for search
      Given the only "web-search" station's node is banned
      When a search for "web-search" arrives
      Then the response status is 503

  Rule: grant keys confine search (A4)

    Scenario: A grant reaches only its owner's search stations
      Given a grant owned by "@acme"
      And only "@zed" has a search station on air
      When a search arrives with that grant key
      Then the response status is 503
      And "@zed" is never called and nobody is charged

    Scenario: A grant's model list confines search
      Given a grant owned by "@acme" allowing only model "go-docs"
      And "@acme" offers "go-docs" and "web-search"
      When a search for "web-search" arrives with that grant key
      Then the response status is 503 with the same body as an unknown model

    Scenario: A grant with no model in the request routes within its allowed list only
      Given a grant owned by "@acme" allowing only model "go-docs"
      And "@acme" offers "go-docs" and "web-search"
      When a search with no model arrives with that grant key
      Then it is answered by "go-docs"

    Scenario: A grant with a namespaced id for another operator cannot escape
      Given a grant owned by "@acme"
      When a search for "@zed/web-search" arrives with that grant key
      Then the response status is 503
      And "@zed" is never called
