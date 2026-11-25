package node

import (
	"testing"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/protocol"
)

// The share editors (TUI form, local console) save a declaration built from their own
// fields, which do not include the interaction or the example prompt. Saving a category
// there must not drop an interaction declared with `roger share --interaction`.
func TestSetMetaKeepsAnInteractionTheEditorDoesNotCarry(t *testing.T) {
	var saved protocol.Taxonomy
	c := New(Config{Broker: "http://broker.local", Station: "kitchen", Meta: map[string]protocol.Taxonomy{
		"m": {Interaction: protocol.InteractionSingleTurn, Example: "Read this channel."},
	}, Hooks: Hooks{SaveMeta: func(_ string, t protocol.Taxonomy) { saved = t }}})
	c.SetRows([]ShareRow{{Model: "m", Upstream: "http://127.0.0.1:8081/v1/chat/completions"}})

	require.NoError(t, c.SetMeta("m", protocol.Taxonomy{Category: "language", Subcategory: "classification"}))
	got, _, ok := c.MetaFor("m")
	require.True(t, ok)
	require.Equal(t, "language", got.Category)
	require.Equal(t, protocol.InteractionSingleTurn, got.Interaction, "the editor's save keeps the interaction")
	require.Equal(t, "Read this channel.", got.Example)
	require.Equal(t, protocol.InteractionSingleTurn, saved.Interaction, "and it is what gets persisted")

	// A save that carries its own pair replaces the old one.
	require.NoError(t, c.SetMeta("m", protocol.Taxonomy{Category: "language", Interaction: protocol.InteractionChat}))
	got, _, _ = c.MetaFor("m")
	require.Equal(t, protocol.InteractionChat, got.Interaction)
	require.Empty(t, got.Example, "a declared pair replaces both values")

	// Clearing removes everything, the interaction included.
	require.NoError(t, c.ClearMeta("m"))
	got, _, _ = c.MetaFor("m")
	require.False(t, got.InteractionDeclared())
}

// A declaration that is only an interaction is a declaration: it is kept, not deleted.
func TestAnInteractionAloneIsStored(t *testing.T) {
	c := New(Config{Broker: "http://broker.local", Station: "kitchen"})
	c.SetRows([]ShareRow{{Model: "m", Upstream: "http://127.0.0.1:8081/v1/chat/completions"}})
	require.NoError(t, c.SetMeta("m", protocol.Taxonomy{Interaction: "single-turn"}))
	got, _, _ := c.MetaFor("m")
	require.Equal(t, protocol.InteractionSingleTurn, got.Interaction)
}
