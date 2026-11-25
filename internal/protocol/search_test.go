package protocol

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanTextAndFlattenText(t *testing.T) {
	cases := []struct{ in, clean, flat string }{
		{"plain words", "plain words", "plain words"},
		{"a\x1b[31mRED", "a [31mRED", "a [31mRED"},
		{"x‮txt", "xtxt", "xtxt"},
		{"ch​ild", "child", "child"},
		{"line\none\ttwo\r\n", "line one two", "line one two"},
		{"  padded  ", "padded", "padded"},
		// entity-encoded: only FlattenText decodes, and it cleans after decoding
		{"a&#27;[31mRED", "a&#27;[31mRED", "a [31mRED"},
		{"x&#x202E;txt", "x&#x202E;txt", "xtxt"},
		{"ch&#x200B;ild", "ch&#x200B;ild", "child"},
		{"<b>Go</b> &amp; Rust", "<b>Go</b> &amp; Rust", "Go & Rust"},
		{"&lt;script&gt;x", "&lt;script&gt;x", "script x"},
	}
	for _, c := range cases {
		require.Equal(t, c.clean, CleanText(c.in), "CleanText(%q)", c.in)
		require.Equal(t, c.flat, FlattenText(c.in), "FlattenText(%q)", c.in)
	}
}

func TestValidEngineAndOfferError(t *testing.T) {
	for _, e := range []string{EngineBrave, EngineSearXNG, EngineNative} {
		require.True(t, ValidEngine(e), e)
	}
	for _, e := range []string{"", "google", "Brave", "custom"} {
		require.False(t, ValidEngine(e), e)
	}
	cases := []struct {
		offer ModelOffer
		want  string // substring of the refusal; "" = accepted
	}{
		{ModelOffer{Model: "m"}, ""},
		{ModelOffer{Model: "m", Modality: ModalityChat}, ""},
		{ModelOffer{Model: "m", Modality: ModalityTTS}, ""},
		{ModelOffer{Model: "s", Modality: ModalitySearch, Engine: EngineSearXNG}, ""},
		{ModelOffer{Model: "v", Modality: "video"}, `unknown modality "video"`},
		{ModelOffer{Model: "s", Modality: ModalitySearch}, "needs engine brave, searxng or native"},
		{ModelOffer{Model: "s", Modality: ModalitySearch, Engine: "google"}, "needs engine"},
		{ModelOffer{Model: "c", Modality: ModalityChat, Engine: EngineBrave}, "cannot carry engine"},
	}
	for _, c := range cases {
		got := c.offer.OfferError()
		if c.want == "" {
			require.Empty(t, got, "%+v", c.offer)
		} else {
			require.Contains(t, got, c.want, "%+v", c.offer)
		}
	}
}

func TestNormalizeSearchOffer(t *testing.T) {
	o := ModelOffer{Model: "s", Modality: ModalitySearch, Engine: EngineNative, Unit: UnitToken,
		PriceOut: 9, Scope: "Go docs\n<b>and</b> " + strings.Repeat("x", 300),
		Taxonomy: Taxonomy{Category: "language", Subcategory: "chat", TaxonomySource: SourceDeclared}}
	o.Normalize()
	require.Equal(t, UnitQuery, o.Unit)
	require.Zero(t, o.PriceOut, "a query has no output side")
	require.Equal(t, "retrieval", o.Taxonomy.Category)
	require.Equal(t, ModalitySearch, o.Taxonomy.Subcategory)
	require.NotEqual(t, SourceDeclared, o.Taxonomy.TaxonomySource)
	require.Len(t, []rune(o.Scope), ScopeMax)
	require.True(t, strings.HasPrefix(o.Scope, "Go docs and "), o.Scope)

	chat := ModelOffer{Model: "c", Scope: "the open web",
		Taxonomy: Taxonomy{Category: "retrieval", Subcategory: ModalitySearch}}
	chat.Normalize()
	require.Empty(t, chat.Scope)
	require.NotEqual(t, ModalitySearch, chat.Taxonomy.Subcategory)

	require.Equal(t, "Go docs", CanonicalScope("  Go\tdocs  "))
}

func TestTaxonomyValidateSearch(t *testing.T) {
	require.NoError(t, Taxonomy{}.Validate(ModalitySearch))
	require.NoError(t, Taxonomy{Category: "retrieval", Subcategory: "search"}.Validate(ModalitySearch))
	require.Error(t, Taxonomy{Category: "language"}.Validate(ModalitySearch))
	require.Error(t, Taxonomy{Category: "retrieval", Subcategory: "rerank"}.Validate(ModalitySearch))
	require.Error(t, Taxonomy{Category: "retrieval", Subcategory: "search"}.Validate(ModalityChat))
}

// A search offer's time-of-use windows have no output side either: a window out-price would
// make a free station read as paid to the router while every listing calls it free.
func TestNormalizeZeroesASearchWindowOutPrice(t *testing.T) {
	o := ModelOffer{Model: "web-search", Modality: ModalitySearch, Engine: EngineSearXNG,
		Schedule: []PriceWindow{{Start: "00:00", End: "06:00", In: 100, Out: 7}}}
	o.Normalize()
	require.Equal(t, 100.0, o.Schedule[0].In)
	require.Zero(t, o.Schedule[0].Out)

	chat := ModelOffer{Model: "m", Schedule: []PriceWindow{{Start: "00:00", End: "06:00", In: 1, Out: 2}}}
	chat.Normalize()
	require.Equal(t, 2.0, chat.Schedule[0].Out, "a chat window keeps its out-price")
}
