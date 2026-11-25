package agent

// Contract: features/tower/tower_upstream_key.feature ("shorter than 8 characters").
//
// The minimum counts CHARACTERS, as the spec says, not bytes: "clé-sec" is 7 characters and 8
// bytes, and must be refused like any other 7-character key.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamKeyMinimumCountsCharacters(t *testing.T) {
	require.ErrorIs(t, CheckUpstreamKey("clé-sec"), ErrUpstreamKeyTooShort, "7 characters, 8 bytes")
	require.ErrorIs(t, CheckUpstreamKey("😀😀"), ErrUpstreamKeyTooShort, "2 characters, 8 bytes")
	require.NoError(t, CheckUpstreamKey("clé-secr"), "8 characters")
	require.NoError(t, CheckUpstreamKey(""), "no key at all is fine")
}
