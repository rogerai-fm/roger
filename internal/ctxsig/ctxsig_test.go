package ctxsig

// The every-spelling table lives beside the ONE list it pins (moved from the
// harness with the matcher, 2026-09-05). A spelling we miss is a turn the client
// fails to compact AND an honest operator the broker wrongly strikes.

import "testing"

func TestOverflowRecognizedInEverySpelling(t *testing.T) {
	for _, s := range []string{
		"Exceeded model context window size",          // Apple foundation
		"context length exceeded",                     // OpenAI-compatible
		"context_length_exceeded",                     // the error code form
		"This model's maximum context length is 8192", // verbatim server text
		"too many tokens in prompt",
		"prompt is too long: 9000 tokens > 8192 maximum",
		"failed to allocate kv cache",
		// llama-server verbatim (live 2026-09-05: a real overflow the list missed)
		"request (13073 tokens) exceeds the available context size (8192 tokens), try increasing it",
		// the broker's own pick-time refusal speaks the same vocabulary
		"request exceeds the context window: ~13073 prompt tokens, but the widest window advertised on gpt-oss-20b right now is 8192",
		// the byte-measured wall
		"Maximum request body size 1048576 exceeded, actual body size 1050714",
		"413 Payload Too Large",
		"Request Entity Too Large",
	} {
		if !IsOverflow(s) {
			t.Errorf("must recognize %q as an overflow", s)
		}
	}
	for _, s := range []string{"no node offers that model", "connection refused", "", "HTTP 413"} {
		if IsOverflow(s) {
			t.Errorf("%q is not an overflow and must not match", s)
		}
	}
}

// A context-window refusal is a 400 in the overflow vocabulary and nothing else: the same
// words under any other status are that status alone.
func TestRefusalIsOnlyA400InTheOverflowVocabulary(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   bool
	}{
		{400, "context length exceeded", true},
		{400, "the request exceeds the available context size", true},
		{400, "unknown parameter foo", false},
		{400, "", false},
		{413, "context length exceeded", false},
		{422, "context length exceeded", false},
		{429, "too many tokens per minute", false},
		{500, "context length exceeded", false},
		{200, "context length exceeded", false},
	} {
		if got := IsRefusal(tc.status, tc.body); got != tc.want {
			t.Errorf("IsRefusal(%d, %q) = %v, want %v", tc.status, tc.body, got, tc.want)
		}
	}
	if !IsOverflow(RefusalMessage) {
		t.Error("the refusal message must read as an overflow to a client that compacts on it")
	}
}
