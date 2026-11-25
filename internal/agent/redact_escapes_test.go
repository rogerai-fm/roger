package agent

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The classic and local Tower paths redact with the same implementation as the joined-Tower
// path, escape forms included (features/tower/tower_upstream_key.feature).
func TestClassicRedactionCoversEscapedKeys(t *testing.T) {
	key := "sk-classic-0123456789"
	var esc strings.Builder
	for _, c := range []byte(key) {
		fmt.Fprintf(&esc, `\u%04x`, c)
	}
	out := string(redactUpstreamKey([]byte(`{"x":"`+esc.String()+`"}`), key))
	require.Contains(t, out, "[redacted]")
	require.NotContains(t, out, esc.String())
}
