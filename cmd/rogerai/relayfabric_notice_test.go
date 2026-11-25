package main

// The relay plane's notice sink.
//
// `roger share` discards the relay plane's progress on purpose - the on-air line is already
// printed and relay chatter underneath it describes a plane the operator did not opt into. The
// defect was that ONE discard covered both progress and the errors that cost the operator money:
// a completion the hub never couriered, a served result that could not be handed back, audit
// failures, and the failure to pin Core's grant key. These are the rules that keep the second
// kind out of the bin without putting the first kind back in.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"rogerai.fm/roger/v7/internal/agent"
)

func TestRelayNoticesSaysARepeatedThingOnce(t *testing.T) {
	var buf bytes.Buffer
	n := &relayNotices{out: &buf}
	err := errors.New("the hub accepted this completion but did not forward the receipt")
	for i := 0; i < 5; i++ {
		n.report(err)
	}
	if got := strings.Count(buf.String(), "did not forward the receipt"); got != 1 {
		t.Fatalf("a standing condition was printed %d times; these loops retry forever and it would "+
			"scroll the on-air line off the screen.\n%s", got, buf.String())
	}
	if !strings.HasPrefix(buf.String(), "  relay: ") {
		t.Fatalf("a notice must be recognisable as the relay plane's: %q", buf.String())
	}
}

// A distinct condition is a distinct notice - deduping must not collapse two different problems
// into one.
func TestRelayNoticesKeepsDistinctProblemsApart(t *testing.T) {
	var buf bytes.Buffer
	n := &relayNotices{out: &buf}
	n.report(errors.New("relay audit: the audit plane is unavailable"))
	n.report(errors.New("this attempt was served but its result could not be returned to the hub"))
	n.report(errors.New("relay audit: the audit plane is unavailable"))

	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.said) != 2 {
		t.Fatalf("expected two distinct notices, got %d: %v", len(n.said), n.said)
	}
	for msg := range n.said {
		if strings.TrimSpace(msg) == "" {
			t.Fatal("an empty notice was recorded")
		}
	}
}

// Concurrent reporters: the serve workers and the audit loop all report from their own
// goroutines, and a `roger share` runs with -race in CI.
func TestRelayNoticesIsSafeUnderConcurrentReporters(t *testing.T) {
	n := &relayNotices{out: io.Discard}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); n.report(errors.New("the same standing condition")) }()
	}
	wg.Wait()
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.said) != 1 {
		t.Fatalf("expected one recorded notice, got %d", len(n.said))
	}
}

// A nil error is not a notice.
func TestRelayNoticesIgnoresNil(t *testing.T) {
	var buf bytes.Buffer
	n := &relayNotices{out: &buf}
	n.report(nil)
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.said) != 0 || buf.Len() != 0 {
		t.Fatalf("a nil error was reported as a notice")
	}
}

// The model's words for an attempt it refused reach the operator's terminal through the
// config a share's relay join really uses (relayConfig), one line per refused attempt: two
// refusals with the same words are two lines, unlike a standing notice.
// Contract: features/tower/context_window_class.feature ("The operator's log keeps the
// model's words").
func TestAShareShowsEachRefusedAttemptOnTheOperatorsTerminal(t *testing.T) {
	var term bytes.Buffer
	n := &relayNotices{out: &term}
	cfg := relayConfig(agent.Config{Model: "m"}, n)
	if cfg.ModelLog == nil {
		t.Fatal("a share's relay join keeps no line for a refused attempt")
	}
	if cfg.Model != "m" {
		t.Fatalf("relayConfig changed the share's own settings: %+v", cfg.Model)
	}
	line := "the model answered attempt att-1 with status 400: the request exceeds the available context size\n"
	for i := 0; i < 2; i++ {
		if w, err := cfg.ModelLog.Write([]byte(line)); err != nil || w != len(line) {
			t.Fatalf("write %d: %d, %v", i, w, err)
		}
	}
	got := term.String()
	if strings.Count(got, "exceeds the available context size") != 2 || strings.Count(got, "\n") != 2 {
		t.Fatalf("want one terminal line per refused attempt, got %q", got)
	}
	for _, l := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if !strings.HasPrefix(l, "  relay: ") {
			t.Fatalf("a line must be recognisable as the relay plane's: %q", l)
		}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.said) != 0 {
		t.Fatalf("refused attempts are not standing notices and must not be remembered: %v", n.said)
	}
}

// A consumer who keeps provoking refusals must not be able to scroll the operator's terminal
// or fill their journal: at most five lines in ten seconds, then one line saying more were
// held back, then nothing until the ten seconds are over.
func TestRefusedAttemptLinesAreBoundedPerTenSeconds(t *testing.T) {
	var term bytes.Buffer
	at := time.Unix(1_700_000_000, 0)
	n := &relayNotices{out: &term, now: func() time.Time { return at }}
	cfg := relayConfig(agent.Config{}, n)
	write := func(i int) {
		line := fmt.Sprintf("the model answered attempt att-%d with status 400: too many tokens\n", i)
		if w, err := cfg.ModelLog.Write([]byte(line)); err != nil || w != len(line) {
			t.Fatalf("write %d: %d, %v (a held-back line is still a completed write)", i, w, err)
		}
	}
	for i := 0; i < 40; i++ {
		write(i)
		at = at.Add(100 * time.Millisecond) // 40 lines in 4 seconds
	}
	lines := strings.Split(strings.TrimSuffix(term.String(), "\n"), "\n")
	if len(lines) != 6 {
		t.Fatalf("want 5 lines and one notice that more were held back, got %d:\n%s", len(lines), term.String())
	}
	for i, l := range lines[:5] {
		if !strings.Contains(l, fmt.Sprintf("att-%d ", i)) {
			t.Fatalf("line %d is not the attempt's own: %q", i, l)
		}
	}
	if !strings.HasPrefix(lines[5], "  relay: ") || !strings.Contains(lines[5], "suppressed") || strings.Contains(lines[5], "att-") {
		t.Fatalf("the sixth line must say more were suppressed, and carry no attempt: %q", lines[5])
	}
	// Ten seconds after the first line the bound starts over.
	at = time.Unix(1_700_000_000, 0).Add(10 * time.Second)
	term.Reset()
	write(99)
	if got := term.String(); !strings.Contains(got, "att-99 ") {
		t.Fatalf("a line after the ten seconds was held back: %q", got)
	}
}
