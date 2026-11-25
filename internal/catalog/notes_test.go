package catalog

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/protocol"
)

func TestBuiltinNotesShip(t *testing.T) {
	n, err := Builtin()
	require.NoError(t, err)
	for _, id := range []string{"wave-nano-chat", "wave-nano-fire-controller", "wave-nano-fire-vision", "wave-pico-293m"} {
		require.True(t, n.Has(id), id)
	}
}

func TestNotesApply(t *testing.T) {
	ns, err := ParseNotes([]byte(`[{"id":"Demo-1B","summary":"S","best_at":"B","not_for":"N","base_model":"base","license":"MIT","page":"/p.html"},
		{"id":"bare","summary":"S2"}]`))
	require.NoError(t, err)

	got := ns.Apply("demo-1b\u200b", protocol.Taxonomy{Summary: "station", License: "X", Category: "language"})
	require.Equal(t, protocol.Taxonomy{Summary: "S", BestAt: "B", NotFor: "N", BaseModel: "base", License: "MIT",
		DetailsSource: protocol.SourceRogerAI, Page: "/p.html", Category: "language"}, got)

	got = ns.Apply("bare", protocol.Taxonomy{License: "X", BaseModel: "Y"})
	require.Equal(t, "X", got.License) // a note without a license keeps the station's
	require.Equal(t, "Y", got.BaseModel)

	// no note: a "rogerai" label is demoted and a page dropped
	got = ns.Apply("other", protocol.Taxonomy{Summary: "s", DetailsSource: protocol.SourceRogerAI, Page: "/p.html"})
	require.Equal(t, protocol.SourceDeclared, got.DetailsSource)
	require.Empty(t, got.Page)
	got = ns.Apply("other", protocol.Taxonomy{})
	require.Empty(t, got.DetailsSource)

	var none Notes
	require.False(t, none.Has("demo-1b"))
	require.Equal(t, "s", none.Apply("demo-1b", protocol.Taxonomy{Summary: "s"}).Summary)
}

func TestPageURLAndDisplay(t *testing.T) {
	require.Equal(t, "", PageURL("https://x/", ""))
	require.Equal(t, "https://x/p.html", PageURL("https://x/", "/p.html"))
	require.Equal(t, "rogerai.fm/p.html#f", PageDisplay("/p.html#f"))
}

func TestNoteRejectsStrayWhitespace(t *testing.T) {
	_, err := ParseNotes([]byte(`[{"id":"a","summary":"s ","page":"/p.html"}]`))
	require.ErrorContains(t, err, "control character or stray whitespace")
	_, err = ParseNotes([]byte(`[{"id":"a","summary":"s","license":"MIT\n"}]`))
	require.ErrorContains(t, err, "control character")
}

// The shipped notes are outward copy: they keep the model pages' rules
// (web/test/model-pages.test.mjs BANNED) and carry no build or run identifier.
func TestBuiltinNotesKeepOutwardCopyRules(t *testing.T) {
	banned := []*regexp.Regexp{
		regexp.MustCompile(`(?i)hugging\s?face`),
		regexp.MustCompile(`(?i)\b(first|novel|unprecedented)\b|state[- ]of[- ]the[- ]art`),
		regexp.MustCompile(`(?i):\d{4}\b|localhost|127\.0\.0\.1|systemd|\.service\b|/mnt/|~/|hermes|hf_serve|serve-wave|workstation`),
		regexp.MustCompile(`(?i)FIgLib|HPWREN|pyro-?sdis|PyroNear|SKAB|\bNAB\b|ALERTCalifornia|SigLIP|Idefics|SmolLM`),
		regexp.MustCompile(`(?i)\bcorpus\b|billion tokens|\b90B\b|\btokens of\b|dataset`),
		regexp.MustCompile(`(?i)\bClaude\b|AI assistant|revenue`),
		regexp.MustCompile(`\b[A-Z]\d{2,}\b|\bv\d+[a-z]\b|\bseed \d|\bR\.\d+`), // run ids, build versions, R-numbers
		regexp.MustCompile(`[\x{2013}\x{2014}]`),                               // no en or em dashes
	}
	b, err := os.ReadFile("notes.json")
	require.NoError(t, err)
	var list []Note
	require.NoError(t, json.Unmarshal(b, &list))
	for _, n := range list {
		for _, v := range []string{n.Summary, n.BestAt, n.NotFor, n.BaseModel, n.License} {
			for _, re := range banned {
				require.False(t, re.MatchString(v), "%s: %q matches %s", n.ID, v, re)
			}
		}
	}
}

// The two fire models are specialized models behind a market guard, not chat models (lab gate,
// 2026-11-20). Their notes must say so, say what input they accept, say they are decision
// support and not a listed fire alarm or a substitute for emergency services, and carry their
// measured numbers with the date they were read.
func TestFireNotesStateSpecializedInputAndDecisionSupport(t *testing.T) {
	ns, err := Builtin()
	require.NoError(t, err)
	date := regexp.MustCompile(`\b20\d\d-\d\d-\d\d\b`)
	number := regexp.MustCompile(`\d+(\.\d+)?%|\b0\.\d+\b`)
	for id, input := range map[string]*regexp.Regexp{
		"wave-nano-fire-controller": regexp.MustCompile(`(?i)\btick\b`),
		"wave-nano-fire-vision":     regexp.MustCompile(`(?i)\bimage\b.*\byes or no\b`),
	} {
		n := ns[NoteID(id)]
		require.Regexp(t, `(?i)\bnot (a )?chat\b`, n.Summary, id)
		require.Regexp(t, input, n.Summary, id)
		require.Regexp(t, `(?i)decision support`, n.NotFor, id)
		require.Regexp(t, `(?i)fire alarm`, n.NotFor, id)
		require.Regexp(t, `(?i)emergency`, n.NotFor, id)
		if id == "wave-nano-fire-vision" { // the module does not carry over to cameras unlike its training ones yet
			require.Regexp(t, `(?i)cameras unlike`, n.NotFor, id)
			require.Regexp(t, `(?i)license`, n.NotFor, id)
		}
		require.Regexp(t, number, n.BestAt, id)
		require.Regexp(t, date, n.BestAt, id)
	}
}

// features/details/interaction.feature: a note's interaction and example.
func TestNoteInteraction(t *testing.T) {
	ns, err := ParseNotes([]byte(`[{"id":"one-job","summary":"S","interaction":"single_turn","example":"Ping."},
		{"id":"no-example","summary":"S","interaction":"single_turn"},
		{"id":"words-only","summary":"S"}]`))
	require.NoError(t, err)

	// the note wins over what a station declared
	got := ns.Apply("one-job", protocol.Taxonomy{Interaction: protocol.InteractionChat, Example: "Tell me a story."})
	require.Equal(t, protocol.InteractionSingleTurn, got.Interaction)
	require.Equal(t, "Ping.", got.Example)
	// and replaces the pair: a station's example never rides on RogerAI's interaction
	got = ns.Apply("no-example", protocol.Taxonomy{Interaction: protocol.InteractionChat, Example: "Tell me a story."})
	require.Equal(t, protocol.InteractionSingleTurn, got.Interaction)
	require.Empty(t, got.Example)
	// a note that says neither keeps the station's own
	got = ns.Apply("words-only", protocol.Taxonomy{Interaction: protocol.InteractionChat, Example: "Hello."})
	require.Equal(t, protocol.InteractionChat, got.Interaction)
	require.Equal(t, "Hello.", got.Example)
	// no note: untouched
	got = ns.Apply("other", protocol.Taxonomy{Interaction: protocol.InteractionSingleTurn, Example: "Ping."})
	require.Equal(t, protocol.InteractionSingleTurn, got.Interaction)
	require.Equal(t, "Ping.", got.Example)

	for doc, want := range map[string]string{
		`[{"id":"bad","summary":"S","interaction":"oneshot"}]`:                           `notes: "bad" interaction must be "chat" or "single_turn", got "oneshot"`,
		`[{"id":"bad","summary":"S","interaction":"single-turn"}]`:                       `notes: "bad" interaction must be "chat" or "single_turn", got "single-turn"`,
		`[{"id":"bad","summary":"S","interaction":"chat","example":" Ping."}]`:           `notes: "bad" has a control character or stray whitespace in " Ping."`,
		`[{"id":"bad","summary":"S","example":"` + strings.Repeat("a", 401) + `"}]`:      `notes: "bad" example: at most 400 characters (got 401)`,
		`[{"id":"bad","summary":"S","interaction":"chat","example":"Triage this one."}]`: `notes: "bad" example: reserved use "triage"`,
	} {
		_, err := ParseNotes([]byte(doc))
		require.EqualError(t, err, want)
	}
}

func TestBuiltinNotesInteraction(t *testing.T) {
	ns, err := Builtin()
	require.NoError(t, err)
	for id, example := range map[string]string{
		"wave-pico-293m":            `channel=s00808c02 tag=AI_12506 stated_unit=? period_s=9.0 n=96 range=[1.32,4.902] mean=2.671 sd=1.3488 first=2.336 last=1.329 slope_per_min=0.0605 sign_changes=56 hf_energy=0.0449 repeat_frac=0.0 longest_run=1 zero_frac=0.0 at_max_frac=0.01 sd_head=0.0963 sd_tail=1.7299 monotonic_frac=0.358 n_resets=61 max_drop=3.406 gap_frac=0.0`,
		"wave-nano-fire-controller": `{"time_local":"14:05","wind_kmh":30,"wind_from":"NE","wind_toward_structures":true,"rh_pct":34,"temp_c":27,"camera_margin_prev":0.82,"camera_margin_now":0.9,"pm25":95,"co_ppm":11,"voc_idx":40,"thermal_delta_c":3,"sensor_fault":false,"escalation_tried_all":false,"acknowledged":false,"minutes_unacked":0}`,
		"wave-nano-fire-vision":     "Is there smoke in this picture? Answer yes or no.",
	} {
		require.Equal(t, protocol.InteractionSingleTurn, ns[id].Interaction, id)
		require.Equal(t, example, ns[id].Example, id)
	}
	require.Empty(t, ns["wave-nano-chat"].Interaction) // unknown reads as chat; nothing is claimed
}
