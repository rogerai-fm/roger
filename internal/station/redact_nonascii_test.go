package station

// Contract: features/tower/tower_upstream_key.feature.
//
// An upstream key is any string an operator's server accepts, and nothing makes it ASCII. The
// pre-push audit (2026-11-24) found the byte-by-byte redaction pattern panicked on the first
// reply for any key byte >= 0x80, taking down every serving path. A key must be redacted in
// every form, non-ASCII or not, and redaction must never panic.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactKeyHandlesNonASCIIKeys(t *testing.T) {
	cases := []struct {
		name, key, echoed string
	}{
		{"accented, verbatim", "clé-secrète-123", "clé-secrète-123"},
		{"accented, fully escaped", "clé-secrète-123", `clé-secrète-123`},
		{"accented, upper-case escape", "clé-secrète-123", `cl\u00E9-secr\u00E8te-123`},
		{"emoji, verbatim", "key-😀-secret", "key-😀-secret"},
		{"emoji, surrogate pair", "key-😀-secret", `key-😀-secret`},
		{"emoji, upper-case surrogate pair", "key-😀-secret", `key-😀-secret`},
		{"cjk, escaped", "鍵鍵鍵-secret", `鍵鍵鍵-secret`},
		{"invalid UTF-8 byte, verbatim", "abc\xffdefgh", "abc\xffdefgh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := []byte(`{"error":"bad key ` + c.echoed + ` here"}`)
			var out []byte
			require.NotPanics(t, func() { out = RedactKey(body, c.key) })
			require.Contains(t, string(out), "[redacted]")
			require.False(t, strings.Contains(string(out), c.echoed), "the key survived: %s", out)
		})
	}
}

func TestRedactKeyNonASCIILeavesOtherTextAlone(t *testing.T) {
	body := []byte(`{"text":"café au lait, 😀"}`)
	require.Equal(t, string(body), string(RedactKey(body, "clé-secrète-123")))
}
