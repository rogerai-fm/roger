package protocol

// features/details/interaction.feature: the interaction and the example prompt.

import (
	"crypto/ed25519"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalInteraction(t *testing.T) {
	for in, want := range map[string]string{
		"chat": InteractionChat, " Chat ": InteractionChat,
		"single_turn": InteractionSingleTurn, "SINGLE_TURN": InteractionSingleTurn,
		"single-turn": InteractionSingleTurn, "single turn": InteractionSingleTurn,
		"": "", "oneshot": "", "singleturn": "", "single__turn": "", "chat\x1b[31m": "",
	} {
		require.Equal(t, want, CanonicalInteraction(in), "%q", in)
	}
}

func TestNormalizeInteractionAndExample(t *testing.T) {
	long := strings.Repeat("a", ExampleMax+1)
	for _, c := range []struct {
		name            string
		in              Taxonomy
		interaction, ex string
	}{
		{"kept", Taxonomy{Interaction: "single-turn", Example: "Ping."}, InteractionSingleTurn, "Ping."},
		{"unknown dropped", Taxonomy{Interaction: "oneshot", Example: "Ping."}, "", "Ping."},
		{"example cut", Taxonomy{Example: long}, "", long[:ExampleMax]},
		{"example cleaned", Taxonomy{Example: "<b>x</b>  Say\thi\x07‮"}, "", "<b>x</b> Say hi"},
		{"reserved use dropped", Taxonomy{Interaction: "chat", Example: "Give me a diagnosis."}, InteractionChat, ""},
		{"nothing", Taxonomy{}, "", ""},
	} {
		got := c.in.Canonical("")
		require.Equal(t, c.interaction, got.Interaction, c.name)
		require.Equal(t, c.ex, got.Example, c.name)
		// The two fields are not card text: they never label a details author.
		require.Empty(t, got.DetailsSource, c.name)
		require.Empty(t, got.TaxonomySource, c.name)
		require.Equal(t, got, got.Canonical(""), c.name) // idempotent
	}
	// A registered offer goes through the same rules.
	o := ModelOffer{Model: "m", Taxonomy: Taxonomy{Interaction: "SINGLE_TURN", Example: " Ping. "}}
	o.Normalize()
	require.Equal(t, InteractionSingleTurn, o.Interaction)
	require.Equal(t, "Ping.", o.Example)
}

func TestValidateInteractionAndExample(t *testing.T) {
	require.NoError(t, Taxonomy{Interaction: "single-turn", Example: strings.Repeat("a", ExampleMax)}.Validate(""))
	require.NoError(t, Taxonomy{Interaction: "Chat"}.Validate(""))

	err := Taxonomy{Interaction: "oneshot"}.Validate("")
	require.EqualError(t, err, `unknown interaction "oneshot" (valid: chat, single_turn)`)

	err = Taxonomy{Example: strings.Repeat("a", ExampleMax+1)}.Validate("")
	require.EqualError(t, err, "example: at most 400 characters (got 401)")

	err = Taxonomy{Example: "Triage this patient."}.Validate("")
	require.ErrorContains(t, err, "example: reserved use")

	got, err := Taxonomy{Interaction: "single-turn", Example: "Ping."}.Declare("")
	require.NoError(t, err)
	require.Equal(t, Taxonomy{Interaction: InteractionSingleTurn, Example: "Ping."}, got)
	require.True(t, got.InteractionDeclared())
	require.False(t, got.DetailsDeclared())
	require.False(t, Taxonomy{Summary: "s"}.InteractionDeclared())
}

func TestWithDetailsOfCarriesInteraction(t *testing.T) {
	inferred := Taxonomy{Category: "language", Subcategory: "chat", TaxonomySource: SourceInferred}
	got := inferred.WithDetailsOf(Taxonomy{Summary: "s", Interaction: InteractionSingleTurn, Example: "Ping."})
	require.Equal(t, InteractionSingleTurn, got.Interaction)
	require.Equal(t, "Ping.", got.Example)
	require.Equal(t, "language", got.Category)
	got = got.WithDetailsOf(Taxonomy{})
	require.Empty(t, got.Interaction)
	require.Empty(t, got.Example)
}

func TestOnAirInteraction(t *testing.T) {
	inf := Taxonomy{Category: "language", Subcategory: "chat", TaxonomySource: SourceInferred}
	node := Taxonomy{Category: "language", Subcategory: "chat", TaxonomySource: SourceInferred,
		Summary: "node words", DetailsSource: SourceDeclared, Interaction: InteractionChat, Example: "Hello."}
	for _, c := range []struct {
		name                     string
		declared, other          Taxonomy
		interaction, ex, summary string
	}{
		{"interaction alone goes on air", Taxonomy{Interaction: InteractionSingleTurn, Example: "Ping."}, inf, InteractionSingleTurn, "Ping.", ""},
		{"it keeps the other side's words", Taxonomy{Interaction: InteractionSingleTurn}, node, InteractionSingleTurn, "", "node words"},
		{"words alone keep the other side's interaction", Taxonomy{Summary: "mine"}, node, InteractionChat, "Hello.", "mine"},
		{"a category alone keeps both", Taxonomy{Category: "vision", Subcategory: "ocr"}, node, InteractionChat, "Hello.", "node words"},
		{"an example alone replaces the pair", Taxonomy{Example: "Ping."}, node, "", "Ping.", "node words"},
		{"nothing declared", Taxonomy{}, node, InteractionChat, "Hello.", "node words"},
	} {
		got := c.declared.OnAir("", c.other)
		require.Equal(t, c.interaction, got.Interaction, c.name)
		require.Equal(t, c.ex, got.Example, c.name)
		require.Equal(t, c.summary, got.Summary, c.name)
	}
}

func TestFoldInteraction(t *testing.T) {
	on := func(interaction, example string) Ballot {
		return Ballot{T: Taxonomy{Interaction: interaction, Example: example}, Online: true}
	}
	off := func(interaction, example string) Ballot {
		return Ballot{T: Taxonomy{Interaction: interaction, Example: example}}
	}
	const st, chat = InteractionSingleTurn, InteractionChat
	for _, c := range []struct {
		name            string
		ballots         []Ballot
		interaction, ex string
	}{
		{"no ballots", nil, "", ""},
		{"nobody declares", []Ballot{on("", ""), on("", "")}, "", ""},
		{"one station", []Ballot{on(st, "Ping.")}, st, "Ping."},
		{"majority single_turn", []Ballot{on(st, ""), on(st, ""), on(chat, "")}, st, ""},
		{"majority chat", []Ballot{on(chat, ""), on(st, ""), on(chat, "")}, chat, ""},
		{"a tie is chat", []Ballot{on(st, "One job."), on(chat, "Hello.")}, chat, "Hello."},
		{"silent stations do not vote", []Ballot{on(st, ""), on("", ""), on("", "")}, st, ""},
		{"offline stations do not vote", []Ballot{on(chat, ""), off(st, "x"), off(st, "x")}, chat, ""},
		{"only offline stations", []Ballot{off(st, "x")}, "", ""},
		{"a curated proxy votes like any station", []Ballot{{T: Taxonomy{Interaction: st}, Online: true, Curated: true}}, st, ""},
		{"example from the winning side only", []Ballot{on(st, "Read this channel."), on(st, "Read this channel."), on(chat, "Tell me a story.")}, st, "Read this channel."},
		{"the most common example wins", []Ballot{on(st, "B"), on(st, "A"), on(st, "B")}, st, "B"},
		{"an example tie is alphabetical", []Ballot{on(st, "B"), on(st, "A")}, st, "A"},
		{"winners without an example give none", []Ballot{on(st, ""), on(st, ""), on(chat, "Tell me a story.")}, st, ""},
		{"an example with no interaction anywhere", []Ballot{on("", "Ping."), on("", "")}, "", "Ping."},
		{"an undeclared station's example loses to the declared side", []Ballot{on("", "Stray."), on(st, "")}, st, ""},
		{"values are canonicalized again", []Ballot{on("Single-Turn", " Ping.\x07 "), on("oneshot", "")}, st, "Ping."},
		{"an over-long example is cut", []Ballot{on(st, strings.Repeat("a", ExampleMax+50))}, st, strings.Repeat("a", ExampleMax)},
	} {
		interaction, ex := FoldInteraction(c.ballots)
		require.Equal(t, c.interaction, interaction, c.name)
		require.Equal(t, c.ex, ex, c.name)
	}
}

func TestInteractionOutsideSigningBytes(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	reg := func(tx Taxonomy) NodeRegistration {
		return NodeRegistration{NodeID: "n", PubKey: string(pub), TS: 1, Offers: []ModelOffer{{Model: "m", Taxonomy: tx}}}
	}
	plain, with := reg(Taxonomy{}), reg(Taxonomy{Interaction: InteractionSingleTurn, Example: "Ping."})
	require.Equal(t, plain.regSigningBytes(), with.regSigningBytes())
	require.NotContains(t, string(with.regSigningBytes()), "interaction")
}
