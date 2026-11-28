# CLI market listing: every kind of station in its own section, priced in its own unit.
#
# APPROVED by the founder 2026-11-25 ("Approve spec, include it", for release v7.2.0).
#
# Build notes (no behavior change):
#   - Section order with every kind present: MODELS, SPECIALIZED, VOICES, SEARCH, OTHER.
#   - The K9 surface outline is split into tagged Examples so each surface runs in the package
#     that owns it: @cli and @tui in Go suites, @console in internal/webui, @site in
#     web/test/specialized-off-chat.test.mjs.
#
# Audit (2026-11-24, internal/client/client.go Search, cmd/rogerai/main.go dispatch):
#   `roger search`, `roger discover` and `roger models` all run client.Search, which prints every
#   /discover offer in one table with no kind column. A voice (tts, priced per 1M characters), a
#   listener (stt, priced per 1M audio bytes) and a search station (priced per 1M queries) sit
#   among chat models, their prices under "$/1M in" and "$/1M out", with CTX 0 and no tok/s.
#   Extra words are ignored: `roger search tides` lists the market, it does not search.
#   The TUI already puts voices in their own section and leaves search off the dial. The site
#   shows voices as SPEAKS / LISTENS and drops search. No surface shows search stations except
#   the JSON at GET /search/stations.
#
# DECISIONS PINNED BY THIS SPEC (approved 2026-11-25):
#   K1  The listing has up to three sections, in this order: MODELS (chat), VOICES (tts and stt),
#       SEARCH. An empty section is not printed.
#   K2  Each section prices in its own unit: models $/1M tokens in and out; voices $/1K characters
#       (speaks) or ~$/min (listens, at the TUI's 128 kbps estimate, marked with *); search
#       $/1K queries. The same units the TUI and /search/stations use.
#   K3  Voices are labeled SPEAKS or LISTENS, the words the site uses.
#   K4  Search rows show the engine (brave, searxng, native), FREE or the price, the operator
#       status, and the scope line when the station has one.
#   K5  `--kind models|voices|search` prints one section. `roger models` means `--kind models`.
#   K6  `--json` prints the offers grouped by kind, for scripts.
#   K7  An offer with an unknown modality is listed under OTHER with its raw modality, never as a
#       model, so the next new kind cannot pose as chat again.
#   K9  A model whose published interaction is "single_turn" (features/details/interaction.feature,
#       approved 2026-11-23) is listed under SPECIALIZED, never under MODELS, on every client:
#       CLI, TUI dial and agent picker, node web console, site Playbox chat shelf. Its row shows
#       its summary and its input.example, so a reader sees what to send. It is never offered as
#       a general chat partner or an agent brain. It stays callable through the API exactly as
#       today (interaction D5: display only). A model with no interaction is chat (D4), and no
#       client guesses single_turn from a category.
#       Audit 2026-11-24: RogerAI's curated notes already mark wave-nano-fire-controller,
#       wave-nano-fire-vision and wave-pico-293m single_turn, with examples. Only the iOS app
#       reads it; every other chat picker offers them as chat partners, and their guards
#       refuse free-form messages.
#   K10 `roger share --modality search` is refused with "use --search <engine>": a search
#       station needs an engine, which only --search sets.
#   K8  `roger search <words>` keeps listing the market (founder ruling 2026-11-26). It does not
#       search the web. Extra words print a one-line hint first, pointing to POST /v1/search,
#       and are otherwise ignored.

@cli @discovery
Feature: The CLI lists models, voices and search stations apart

  Background:
    Given the broker lists a chat model "llama3.2" at $0.10 in and $0.30 out per 1M tokens
    And a voice "roger-operator" that speaks at $1 per 1M characters
    And a voice "whisper-1" that listens at $0.50 per 1M audio bytes
    And a search station "@acme/web-search" on engine "searxng", free, scope "the open web"

  Rule: sections (K1, K3, K4, K7)

    Scenario: Each kind gets its own section, in order
      When I run "roger search"
      Then the output has sections "MODELS", "VOICES" and "SEARCH" in that order
      And "llama3.2" is only under "MODELS"
      And "roger-operator" is only under "VOICES", labeled "SPEAKS"
      And "whisper-1" is only under "VOICES", labeled "LISTENS"
      And "@acme/web-search" is only under "SEARCH", with engine "searxng" and "FREE"

    Scenario: An empty section is not printed
      Given no voices are on the broker
      When I run "roger search"
      Then the output has no "VOICES" section

    Scenario: A broker with only search stations prints only SEARCH
      Given only the search station is on the broker
      When I run "roger search"
      Then the output has only the "SEARCH" section
      And it does not say "no offers yet"

    Scenario: An unknown kind is never listed as a model
      Given an offer "mystery" with modality "video"
      When I run "roger search"
      Then "mystery" is under "OTHER" with kind "video"
      And "mystery" is not under "MODELS"

    Scenario: An offer from an older broker with no modality is a model
      Given an offer "old-model" with no modality field
      When I run "roger search"
      Then "old-model" is under "MODELS"

  Rule: units (K2)

    Scenario Outline: Each section prices in its own unit
      When I run "roger search"
      Then the "<section>" header shows "<unit>"
      And "<row>" shows the price "<price>"

      Examples:
        | section | unit           | row              | price |
        | MODELS  | $/1M in        | llama3.2         | 0.10  |
        | MODELS  | $/1M out       | llama3.2         | 0.30  |
        | VOICES  | $/1K chars     | roger-operator   | 0.001 |
        | VOICES  | ~$/min*        | whisper-1        | 0.48  |
        | SEARCH  | $/1K queries   | @acme/web-search | FREE  |

    Scenario: No voice or search price appears under a token column
      When I run "roger search"
      Then no row under "MODELS" belongs to a voice or a search station

  Rule: filters and scripts (K5, K6)

    Scenario Outline: --kind prints one section
      When I run "roger search --kind <kind>"
      Then the output has only the "<section>" section

      Examples:
        | kind   | section |
        | models | MODELS  |
        | voices | VOICES  |
        | search | SEARCH  |

    Scenario: roger models lists models only
      When I run "roger models"
      Then the output has only the "MODELS" section

    Scenario: An unknown --kind is refused
      When I run "roger search --kind pictures"
      Then the command fails with "--kind must be models, voices or search"

    Scenario: --json groups offers by kind
      When I run "roger search --json"
      Then the JSON has keys "models", "voices", "search" and "other"
      And "@acme/web-search" is in "search" with "engine" "searxng"

  Rule: one-job models are specialized, not chat partners (K9)

    Background:
      Given a chat-kind model "wave-nano-fire-controller" with interaction "single_turn" and example "{\"tick\": 1}"
      And a chat-kind model "wave-pico-293m" with interaction "single_turn"

    Scenario: One-job models get their own section with their example shown
      When I run "roger search"
      Then "wave-nano-fire-controller" is only under "SPECIALIZED", showing its example
      And "wave-pico-293m" is only under "SPECIALIZED"
      And "llama3.2" is still under "MODELS"

    Scenario Outline: No chat surface offers a one-job model as a chat partner
      When the <surface> builds its chat list
      Then "wave-nano-fire-controller" is not in it
      And "llama3.2" is in it

      @tui
      Examples:
        | surface                  |
        | TUI dial chat section    |
        | TUI agent picker         |

      @console
      Examples:
        | surface                  |
        | node web console picker  |

      @site
      Examples:
        | surface                  |
        | site Playbox chat shelf  |

    Scenario: A one-job model stays callable through the API
      When a signed caller sends a chat completion to "wave-nano-fire-controller"
      Then it is routed to the station exactly as before

    Scenario: A model with no interaction stays a chat partner
      Given a chat-kind model "mystery-7b" with no interaction
      When I run "roger search"
      Then "mystery-7b" is under "MODELS"

    Scenario: A task category alone never makes a model specialized
      Given a chat-kind model "extract-helper-8b" with category "language/extraction" and no interaction
      When I run "roger search"
      Then "extract-helper-8b" is under "MODELS"

  Rule: share flags (K10)

    Scenario: --modality search is refused with a pointer to --search
      When I run "roger share --upstream http://127.0.0.1:9000 --modality search"
      Then the command fails with "use --search <engine>"

  Rule: words after search list the market, after a hint (K8, ruled 2026-11-26)

    Scenario: Extra words are not silently ignored
      When I run "roger search tides"
      Then the output starts with a one-line hint about searching the web
      And the market is listed after it
