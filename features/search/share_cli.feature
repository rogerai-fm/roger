# Search API: becoming a provider with `roger share`.
#
# APPROVED by the founder 2026-11-15. Founder direction: roger stays "super stupid simple". One flag picks
# the engine; roger checks the engine works before the station goes on air; a provider never
# has to know about modalities, units or per-1M arithmetic.
#
# Ground truth: cmd/rogerai/main.go share flags (--model, --upstream, --upstream-key,
# --modality, --price-in), the curated-share validation precedent (main.go ~1343: a named
# provider needs an explicit --upstream), and the voice booth's per-model saved config.
#
# COMMANDS
#   roger share --search brave                                   (key from BRAVE_SEARCH_API_KEY, --upstream-key, or a prompt)
#   roger share --search searxng --upstream http://127.0.0.1:8888
#   roger share --search native  --upstream http://127.0.0.1:9000 [--upstream-key K]
#   common: --model <id>  --scope "<what the index covers>"  --price-per-1k <dollars>
#
# DECISIONS PINNED BY THIS SPEC (await founder approval):
#   C1  --search <engine> is the whole switch. It implies modality "search" and the engine.
#   C2  the price is given per 1,000 queries (--price-per-1k), the unit people compare.
#   C3  brave needs an explicit price (it costs the provider money per query); searxng and
#       native default to free.
#   C4  default model ids: brave and searxng "web-search", native "search". --model overrides.
#   C5  before going on air, roger runs ONE check query through the engine and validates the
#       reply against the contract. A failed check keeps the search offer off the air with a
#       plain reason; any chat or voice offers on the same node are unaffected.
#   C6  brave and searxng resell or relay someone else's results. The first time a provider
#       goes on air with one of them, roger says the provider answers for that engine's terms
#       and asks once. The answer is saved per engine. --accept-engine-terms answers it in
#       scripts. native asks nothing: it is the provider's own index.
#   C7  the engine key is saved only in the provider's local roger config, file mode 0600,
#       and is never printed back (only "key set" / "no key").

@search @cli
Feature: Anyone can become a search provider with one roger share command

  Rule: the engine flag (C1, C4)

    Scenario Outline: --search picks the engine and implies the search modality
      When the provider runs "roger share --search <engine> <extra>"
      Then the node registers a search offer "<model>" with engine "<engine>"

      Examples:
        | engine  | extra                                                    | model      |
        | searxng | --upstream http://127.0.0.1:8888                         | web-search |
        | native  | --upstream http://127.0.0.1:9000                         | search     |
        | brave   | --upstream-key k --price-per-1k 6 --accept-engine-terms  | web-search |

    Scenario: --model names the offer
      When the provider runs "roger share --search native --upstream http://127.0.0.1:9000 --model go-docs"
      Then the node registers a search offer "go-docs" with engine "native"

    Scenario Outline: Invalid combinations are refused with a plain reason
      When the provider runs "roger share <args>"
      Then the command exits non-zero
      And the message says "<reason>"
      And nothing is registered

      Examples:
        | args                                         | reason                                          |
        | --search google                              | unknown search engine "google": use brave, searxng or native |
        | --search searxng                             | --search searxng needs --upstream, the address of your SearXNG |
        | --search native                              | --search native needs --upstream, the address of your search service |
        | --search searxng --upstream http://x --modality tts | --search already sets the modality; drop --modality |
        | --search brave --curated openrouter          | a search station cannot be curated; the engine label already says what it is |
        | --search searxng --upstream not-a-url        | --upstream must be an http or https address     |

  Rule: prices (C2, C3)

    Scenario: --price-per-1k is stored per 1,000,000 queries
      When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888 --price-per-1k 4"
      Then the registered offer price is 4000 credits per 1,000,000 queries

    Scenario: SearXNG and native default to free
      When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888"
      Then the registered offer price is 0

    Scenario: Brave without a price is refused
      When the provider runs "roger share --search brave --upstream-key k --accept-engine-terms"
      Then the command exits non-zero
      And the message says Brave charges you per query, so set --price-per-1k (0 is allowed if you mean it)

    Scenario: Brave at an explicit price of 0 is allowed
      When the provider runs "roger share --search brave --upstream-key k --price-per-1k 0 --accept-engine-terms"
      Then the registered offer price is 0

    Scenario Outline: A negative or non-finite price is refused
      When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888 --price-per-1k <price>"
      Then the command exits non-zero

      Examples:
        | price |
        | -1    |
        | NaN   |
        | Inf   |

    Scenario: --price-in and --price-out are refused alongside --search
      When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888 --price-in 5"
      Then the message says search is priced with --price-per-1k

  Rule: the check before going on air (C5)

    Scenario: A working engine goes on air after one check query
      Given a SearXNG at http://127.0.0.1:8888 that answers searches
      When the provider runs "roger share --search searxng --upstream http://127.0.0.1:8888"
      Then roger sends exactly one check query to it
      And the search offer goes on air
      And roger prints that the check passed with the number of results returned

    Scenario Outline: A failed check keeps the search offer off the air and says why
      Given an engine that <fails>
      When the provider shares it as a search station
      Then the search offer does not go on air
      And roger prints "<reason>"

      Examples:
        | fails                                     | reason                                                       |
        | refuses the connection                    | could not reach your search engine at the --upstream address |
        | answers 403 to format=json (SearXNG)      | SearXNG's JSON output is off: add json to search.formats in settings.yml |
        | answers 401 (Brave)                       | Brave refused the key                                        |
        | returns a reply without a results array   | the engine's reply does not match the search contract        |
        | returns only non-http result URLs         | the engine returned no usable results for the check query    |
        | takes longer than 10 seconds              | the engine did not answer within 10 seconds                  |

    Scenario: A failed search check does not stop the node's other offers
      Given a node with a chat model and a SearXNG that refuses connections
      When the provider shares both
      Then the chat offer goes on air
      And the search offer does not

    Scenario: The check query is fixed text with no personal data
      When roger runs the check query
      Then the query is the fixed string "weather"

    # Founder ruling 2026-11-25 (option a): a focused index (a shop's products, industrial parts)
    # has nothing for "weather". A well-formed reply with no results passes; the contract allows
    # an empty reply. A reply whose results are all unusable still fails (the outline above).
    Scenario Outline: A well-formed reply with no results passes the check
      Given a native service that answers every search with an empty results array
      When the provider runs "roger share --search native --upstream <upstream>"
      Then the search offer goes on air
      And roger prints "check passed: 0 results for \"weather\" (fine for a focused index)"

      Examples:
        | upstream               |
        | http://127.0.0.1:9000  |

  Rule: a private search band (founder ruling 2026-11-25)

    # A private band is never listed; its frequency code is the only way in. The owner sees the
    # code once, on their own screen, the same way a private chat share prints it.
    Scenario: A private search share prints its frequency code once
      Given the broker grants a private band with frequency code "RGR-TEST-CODE"
      When the provider runs "roger share --search native --upstream http://127.0.0.1:9000 --private"
      Then the search offer goes on air
      And roger prints the frequency code "RGR-TEST-CODE" exactly once
      And the output says the band is private and reachable only with the code

    Scenario: A public search share prints no frequency code
      When the provider runs "roger share --search native --upstream http://127.0.0.1:9000"
      Then the output has no frequency code

  Rule: engine terms (C6)

    Scenario Outline: The first share of a relayed engine asks once
      Given the provider has never shared engine "<engine>"
      When the provider shares it interactively
      Then roger says the provider is responsible for following <engine>'s terms for serving results to others
      And roger asks to continue
      And a "yes" is saved so the question is not asked again for "<engine>"

      Examples:
        | engine  |
        | brave   |
        | searxng |

    Scenario: A "no" stops the search offer and saves nothing
      When the provider answers "no" to the engine terms question
      Then the search offer does not go on air
      And the question is asked again next time

    Scenario: A script must pass --accept-engine-terms
      Given the provider has never shared engine "brave"
      When the provider shares it with no terminal attached and without --accept-engine-terms
      Then the command exits non-zero
      And the message names --accept-engine-terms

    Scenario: A native engine is never asked about terms
      When the provider runs "roger share --search native --upstream http://127.0.0.1:9000"
      Then no terms question is asked

  Rule: the key stays local (C7)

    Scenario: The Brave key is taken from the environment
      Given BRAVE_SEARCH_API_KEY is set
      When the provider runs "roger share --search brave --price-per-1k 6 --accept-engine-terms"
      Then the Brave adapter uses that key

    Scenario: A key from --upstream-key wins over the environment
      Given BRAVE_SEARCH_API_KEY is set to "a"
      When the provider runs "roger share --search brave --upstream-key b --price-per-1k 6 --accept-engine-terms"
      Then the Brave adapter uses key "b"

    Scenario: With no key and a terminal, roger asks for it without echo
      Given no Brave key is configured
      When the provider shares Brave interactively
      Then roger asks for the key with input hidden

    Scenario: A saved key is stored with owner-only permissions
      When the provider saves a Brave key
      Then the config file holding it has mode 0600

    Scenario: roger never prints the key back
      Given a saved Brave key "sk-test-5e3c"
      When the provider runs "roger share --search brave --price-per-1k 6"
      Then the output does not contain "sk-test-5e3c"
      And it says "key set"

  Rule: scope

    Scenario: --scope is sent with the offer
      When the provider runs "roger share --search native --upstream http://127.0.0.1:9000 --scope \"Go module docs\""
      Then the registered offer scope is "Go module docs"
