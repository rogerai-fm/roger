package main

import (
	"bytes"
	"strings"
	"testing"
)

// --help on a search share or a search check prints its usage, not "flag: help requested".
func TestSearchHelpPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	if err := cmdSearch(config{}, []string{"check", "--help"}, &out); err != nil || !strings.Contains(out.String(), "roger search check --upstream") {
		t.Fatalf("search check --help: %v %q", err, out.String())
	}
	out.Reset()
	if err := shareSearch(config{}, []string{"--search", "native", "--help"}, searchShareIO{out: &out}); err != nil || !strings.Contains(out.String(), "--price-per-1k") {
		t.Fatalf("share --search --help: %v %q", err, out.String())
	}
}
