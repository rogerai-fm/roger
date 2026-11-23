# Approved by the founder 2026-11-23. Step definitions and code follow in the next change.
#
# A model can say plainly that it does one job per message, and give one example prompt, so a
# client does not have to infer it from category names. Today the iOS app treats every
# category except language/chat, language, vision/vlm and language/code as one-job, and
# wave-pico-293m (no declared category, guessed as language/chat) slips through.
#
# Source of the ask: the iOS handoff BROKER-MODEL-INFO-FOR-CLIENTS-HANDOFF-2026-11-23 (item 6).
#
# GROUND TRUTH (origin/main df8bda4b):
#   - protocol.Taxonomy carries the operator text summary, best_at, not_for (bounded,
#     sanitized, display only, outside the signed registration bytes) and details_source.
#   - RogerAI's curated notes (internal/catalog/notes.json) override declared text per model
#     and are the only writer of source "rogerai".
#   - /v1/models folds the stations' text into rogerai.details by band vote.
#   - Nothing today says how a model expects to be talked to.
#
# DECISIONS:
#   D1. Two new optional fields beside summary/best_at/not_for, with the same rules:
#         interaction   "chat" or "single_turn"
#         example       one example prompt, at most 200 characters
#       Set by the operator (roger share --interaction, --example) or by a curated note.
#   D2. They are published where a client branches, at the top of the model's row or entry,
#       in the shape the app asked for:
#         "interaction": "single_turn",
#         "input": {"accepts": ["image", "text"], "example": "..."}
#       on each /discover offer, each /market row and each /v1/models rogerai block.
#   D3. input.accepts is not declared. It is computed: "text" for a chat model, plus "image"
#       when the model has the vision capability; "text" for tts; "audio" for stt.
#   D4. Absent means unknown, and a client treats unknown as chat. The broker never guesses
#       single_turn from a category.
#   D5. Display only, like the rest of the taxonomy: the broker does not strip history or
#       tools from a request to a single_turn model, does not refuse a multi-turn request to
#       it, and never routes or filters on it. Enforcing it at the broker would be its own
#       change.
#   D6. A curated note wins over declarations. Among declarations the value most online
#       stations declare wins; a tie is "chat" (the safe reading).

Feature: A model declares how it expects to be talked to

  Background:
    Given a broker with an empty in-memory node registry

  # --- declaring -------------------------------------------------------------------------

  Scenario: An operator declares a single-turn model with an example
    Given station "s1" shares "smoke-check" with interaction "single_turn" and example "Is there smoke in this picture? Answer yes or no."
    When a consumer GETs /discover
    Then the "smoke-check" offer has interaction "single_turn"
    And the "smoke-check" offer's input.example is "Is there smoke in this picture? Answer yes or no."

  Scenario: An offer that declares nothing carries none of the keys
    Given station "s1" shares "plain-m" with no interaction and no example
    When a consumer GETs /discover
    Then the "plain-m" offer has no "interaction" key
    And the "plain-m" offer's input has no "example" key

  Scenario Outline: Only the two known values are accepted
    Given station "s1" registers "m" with interaction "<sent>"
    When a consumer GETs /discover
    Then the "m" offer's interaction is <published>

    Examples:
      | sent         | published     |
      | chat         | "chat"        |
      | single_turn  | "single_turn" |
      | SINGLE_TURN  | "single_turn" |
      | single-turn  | "single_turn" |
      | oneshot      | absent        |
      |              | absent        |

  Scenario: The example is bounded
    Given station "s1" registers "m" with an example of 201 characters
    When a consumer GETs /discover
    Then the "m" offer's input.example is at most 200 characters

  Scenario: The example is sanitized like the rest of the text
    Given station "s1" registers "m" with example "<script>x</script> Say hi\u0007"
    When a consumer GETs /discover
    Then the "m" offer's input.example is the sanitized text of text_sanitization.feature

  Scenario: The fields never enter the signed registration bytes
    Given an offer for "m" with interaction "single_turn" and example "Ping."
    When the registration signing bytes are computed
    Then they equal the bytes for the same offer with neither field

  Scenario: A node that predates the fields registers unchanged
    Given a registration body produced by a node without the fields
    When the broker accepts the registration
    Then the offer for that model has no "interaction" key

  # --- what the model accepts ------------------------------------------------------------

  Scenario Outline: input.accepts is computed from the modality and the vision capability
    Given station "s1" is on air for "m" with modality "<modality>" and capabilities <caps>
    When a consumer GETs /discover
    Then the "m" offer's input.accepts is <accepts>

    Examples:
      | modality | caps       | accepts           |
      | chat     | []         | ["text"]          |
      | chat     | ["vision"] | ["image", "text"] |
      |          | []         | ["text"]          |
      | tts      | []         | ["text"]          |
      | stt      | []         | ["audio"]         |

  Scenario: A node cannot declare input.accepts
    Given station "s1" registers "m" with input.accepts ["video"]
    When a consumer GETs /discover
    Then the "m" offer's input.accepts is ["text"]

  # --- the per-model row -----------------------------------------------------------------

  Scenario: /market carries the model's interaction and input
    Given station "s1" shares "smoke-check" with vision, interaction "single_turn" and example "Is there smoke in this picture? Answer yes or no."
    When a consumer GETs /market
    Then the "smoke-check" row has interaction "single_turn"
    And the "smoke-check" row's input is {"accepts": ["image", "text"], "example": "Is there smoke in this picture? Answer yes or no."}

  Scenario: /v1/models carries the same values
    Given station "s1" shares "smoke-check" with vision, interaction "single_turn" and example "Is there smoke in this picture? Answer yes or no."
    When a consumer GETs /market and GETs /v1/models
    Then the "smoke-check" entry's rogerai.interaction and rogerai.input equal the "smoke-check" row's

  Scenario: The value most online stations declare wins
    Given stations "s1" and "s2" share "m" with interaction "single_turn"
    And station "s3" shares "m" with interaction "chat"
    When a consumer GETs /market
    Then the "m" row has interaction "single_turn"

  Scenario: A tie reads as chat
    Given station "s1" shares "m" with interaction "single_turn"
    And station "s2" shares "m" with interaction "chat"
    When a consumer GETs /market
    Then the "m" row has interaction "chat"

  Scenario: Stations that declare nothing do not vote
    Given station "s1" shares "m" with interaction "single_turn"
    And stations "s2" and "s3" share "m" with no interaction and no example
    When a consumer GETs /market
    Then the "m" row has interaction "single_turn"

  Scenario: The example comes from the stations that declared the winning interaction
    Given station "s1" shares "m" with interaction "single_turn" and example "Read this channel."
    And station "s2" shares "m" with interaction "single_turn" and example "Read this channel."
    And station "s3" shares "m" with interaction "chat" and example "Tell me a story."
    When a consumer GETs /market
    Then the "m" row's input.example is "Read this channel."

  Scenario: An offline, private or banned station does not vote
    Given station "s1" shares "m" with interaction "chat"
    And an offline station, a private station and a banned station each share "m" with interaction "single_turn"
    When a consumer GETs /market
    Then the "m" row has interaction "chat"

  Scenario: A model nobody describes has no interaction on its row
    Given station "s1" shares "plain-m" with no interaction and no example
    When a consumer GETs /market
    Then the "plain-m" row has no "interaction" key
    And the "plain-m" row's input is {"accepts": ["text"]}

  # --- curated notes ---------------------------------------------------------------------

  Scenario: A curated note sets the interaction and example for its model
    Given RogerAI's curated notes give "wave-pico-293m" interaction "single_turn" and an example
    And station "s1" shares "wave-pico-293m" with no interaction and no example
    When a consumer GETs /market
    Then the "wave-pico-293m" row has interaction "single_turn"
    And the "wave-pico-293m" row's input.example is the note's example

  Scenario: A curated note wins over what stations declare
    Given RogerAI's curated notes give "wave-pico-293m" interaction "single_turn" and an example
    And stations "s1" and "s2" share "wave-pico-293m" with interaction "chat"
    When a consumer GETs /market
    Then the "wave-pico-293m" row has interaction "single_turn"

  Scenario: A curated note with an unknown interaction is rejected when the notes load
    Given a curated notes file whose note has interaction "oneshot"
    When the notes are parsed
    Then parsing fails naming the note and the field

  Scenario: The shipped notes mark every one-job Wave model single_turn with an example
    Then the built-in notes for "wave-pico-293m", "wave-nano-fire-controller" and "wave-nano-fire-vision" each have interaction "single_turn" and a non-empty example

  # --- display only ----------------------------------------------------------------------

  Scenario: The broker does not change a request to a single-turn model
    Given station "s1" shares "smoke-check" with interaction "single_turn" and example "Ping."
    And "u-1" is a funded consumer
    When "u-1" relays for "smoke-check" with three turns of history and a non-empty tools array
    Then the body "s1" received carries all three turns and the tools array

  Scenario: Interaction is not a routing key or a filter
    Given station "s1" shares "smoke-check" with interaction "single_turn" and example "Ping."
    When a consumer GETs /market?interaction=single_turn
    Then the response is 400 with error code "unknown_query_param"

  Scenario: The fields do not change the signal or the order
    Given the offers of this feature with and without the two fields
    When a consumer GETs /market for each
    Then every row's signal and the row order are equal

  # --- CLI and docs ----------------------------------------------------------------------

  @cli
  Scenario: roger share takes the two flags
    When the operator runs roger share with --interaction single-turn and --example "Ping."
    Then the registration carries interaction "single_turn" and example "Ping."

  @cli
  Scenario: roger share refuses an unknown interaction
    When the operator runs roger share with --interaction oneshot
    Then the command fails and names the two accepted values

  @cli
  Scenario: roger models shows the interaction and the example
    When a user runs roger models smoke-check
    Then the output has a line for the interaction and a line for the example

  Scenario: The API reference documents the fields
    Then the public API reference describes interaction and input on /discover, /market and /v1/models
    And it says absent means unknown and that the broker does not enforce it
