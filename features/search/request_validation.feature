# Search API: the request contract (broker side).
#
# APPROVED by the founder 2026-11-15. POST /v1/search takes a small JSON body. Everything the broker
# refuses is refused BEFORE any hold, any moderation call and any dispatch: no charge for a
# malformed request, and nothing malformed reaches a provider.
#
# Ground truth to build: cmd/rogerai-broker (a searchSpec on the audioRelayCore spine), the
# limits mirror internal/harness/search.go (searchMaxQuery 400) and the handoff's count cap.
#
# REQUEST SHAPE
#   { "query": "...", "count": 5, "freshness": "week", "country": "US", "model": "web-search" }
#   query      required. Trimmed. 1 to 400 Unicode characters.
#   count      optional. Absent or 0 means 5. 1 to 20. Above 20 is lowered to 20.
#   freshness  optional. "day" | "week" | "month" | "year".
#   country    optional. Two ASCII letters (ISO 3166-1 alpha-2), upper-cased by the broker.
#   model      optional. A search offer id, raw ("web-search") or namespaced
#              ("@station/web-search"). Absent means any search station.
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   R1  the broker forwards a REBUILT body holding only the five known fields, never the
#       caller's raw bytes. Unknown fields are ignored, not errors, and never reach a provider.
#   R2  an over-long query is 413, the same status the TTS character cap uses.
#   R3  a count above 20 is lowered, not refused. A negative or non-integer count is 400.
#   R4  the request body is capped at 16 KiB.

@search @api
Feature: POST /v1/search accepts a small, strict request and refuses the rest before any money moves

  Background:
    Given a consumer with a funded wallet
    And a search station "web-search" on air with engine "searxng" at 5000 credits per 1,000,000 queries

  Rule: the method and the body shape

    Scenario Outline: Only POST is accepted
      When a <method> /v1/search arrives
      Then the response status is 405
      And no hold is placed

      Examples:
        | method |
        | GET    |
        | PUT    |
        | DELETE |
        | PATCH  |

    Scenario: A browser preflight from an allowed origin is answered without a charge
      When an OPTIONS /v1/search preflight arrives from an allowed first-party origin
      Then the preflight is answered with that origin, never "*"
      And no hold is placed

    Scenario Outline: A body that is not a JSON object is refused
      When a POST /v1/search arrives with body <body>
      Then the response status is 400
      And no hold is placed
      And no moderation call is made
      And the station is never called

      Examples:
        | body             |
        | ""               |
        | "not json"       |
        | "[]"             |
        | "\"a string\""   |
        | "{"              |
        | "null"           |

    Scenario: A body over 16 KiB is refused (R4)
      When a POST /v1/search arrives with a 17 KiB body
      Then the response status is 413
      And no hold is placed

  Rule: the query (R2)

    Scenario Outline: An empty or blank query is refused before any hold
      When a search arrives with query <query>
      Then the response status is 400
      And the error says the query is empty
      And no hold is placed
      And the station is never called

      Examples:
        | query      |
        | ""         |
        | "   "      |
        | "\t\n"     |
        | (absent)   |
        | (null)     |
        | (a number) |

    Scenario Outline: The query length is counted in Unicode characters after trimming
      When a search arrives with a query of <chars> characters made of "<glyph>"
      Then the response status is <status>

      Examples:
        | chars | glyph | status |
        | 1     | a     | 200    |
        | 400   | a     | 200    |
        | 401   | a     | 413    |
        | 400   | 日    | 200    |
        | 401   | 日    | 413    |
        | 400   | 🌍    | 200    |
        | 401   | 🌍    | 413    |

    Scenario: Surrounding whitespace does not count toward the cap
      When a search arrives with 400 "a" characters padded by 50 spaces on each side
      Then the response status is 200
      And the station receives the 400-character query with no padding

    Scenario: Control characters inside the query are replaced with spaces before it is forwarded
      When a search arrives with query "go\u0000 generics\u001b[31m"
      Then the station receives the query "go generics [31m"

    Scenario: A query that is invalid UTF-8 is refused
      When a search arrives whose query holds an invalid UTF-8 sequence
      Then the response status is 400
      And no hold is placed

  Rule: the count (R3)

    Scenario Outline: The count is defaulted, accepted, lowered or refused
      When a search arrives with count <count>
      Then the response status is <status>
      And the station is asked for <forwarded> results

      Examples:
        | count    | status | forwarded |
        | (absent) | 200    | 5         |
        | 0        | 200    | 5         |
        | 1        | 200    | 1         |
        | 20       | 200    | 20        |
        | 21       | 200    | 20        |
        | 1000000  | 200    | 20        |

    Scenario Outline: A count that is not a non-negative integer is refused
      When a search arrives with count <count>
      Then the response status is 400
      And no hold is placed

      Examples:
        | count  |
        | -1     |
        | 2.5    |
        | "5"    |
        | true   |
        | 1e400  |

  Rule: freshness and country

    Scenario Outline: Freshness is a closed set
      When a search arrives with freshness <freshness>
      Then the response status is <status>

      Examples:
        | freshness | status |
        | (absent)  | 200    |
        | "day"     | 200    |
        | "week"    | 200    |
        | "month"   | 200    |
        | "year"    | 200    |
        | "Week"    | 400    |
        | "hour"    | 400    |
        | ""        | 400    |
        | 7         | 400    |

    Scenario Outline: Country is two ASCII letters, upper-cased
      When a search arrives with country <country>
      Then the response status is <status>
      And the station receives country <forwarded>

      Examples:
        | country  | status | forwarded |
        | (absent) | 200    | (absent)  |
        | "US"     | 200    | "US"      |
        | "us"     | 200    | "US"      |
        | "USA"    | 400    | (none)    |
        | "U"      | 400    | (none)    |
        | "Ü1"     | 400    | (none)    |
        | ""       | 400    | (none)    |

  Rule: the broker forwards a rebuilt body, never the caller's bytes (R1)

    Scenario: Unknown fields are dropped before the station sees the request
      When a search arrives with query "rust async" and extra fields "api_key", "endpoint" and "user"
      Then the response status is 200
      And the body the station receives has only the fields model, query, count, freshness and country

    Scenario: Duplicate keys cannot smuggle a second query past validation
      When a search arrives whose JSON names "query" twice, first blank then 5000 characters long
      Then the request is judged on the value the decoder keeps
      And the station receives at most a 400-character query

    Scenario: The station receives the raw offer id it registered, so it can pick its backend
      Given the station "@acme" also offers "go-docs" with engine "native"
      When a search arrives with query "generics" and model "@acme/go-docs"
      Then the body the station receives has model "go-docs"

    Scenario: A request with no model tells the station which of its offers was chosen
      When a search arrives with query "rust async" and no model
      Then the body the station receives has model "web-search"
