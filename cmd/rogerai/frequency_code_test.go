package main

import (
	"bytes"
	"strings"
	"testing"
)

// A private share that gets neither a new code nor an existing band back still tells the
// owner where they stand, instead of printing nothing (PR #41 audit follow-up).
func TestPrintFrequencyCodeNoCodeNoBand(t *testing.T) {
	var b bytes.Buffer
	printFrequencyCode(&b, "m", "", "", "how")
	out := b.String()
	for _, want := range []string{"private", "no frequency code", "roger bands"} {
		if !strings.Contains(out, want) {
			t.Fatalf("no-code notice missing %q: %q", want, out)
		}
	}
}

// The two existing branches are unchanged: a new code prints the code; an existing band
// names the band and says the code was shown only once.
func TestPrintFrequencyCodeBranches(t *testing.T) {
	var b bytes.Buffer
	printFrequencyCode(&b, "m", "RX-123", "147.2 MHz", "how-to-line")
	if !strings.Contains(b.String(), "RX-123") || !strings.Contains(b.String(), "how-to-line") {
		t.Fatalf("new code not printed: %q", b.String())
	}
	b.Reset()
	printFrequencyCode(&b, "m", "", "147.2 MHz", "how")
	if !strings.Contains(b.String(), "147.2 MHz") || strings.Contains(b.String(), "no frequency code") {
		t.Fatalf("existing band line wrong: %q", b.String())
	}
}
