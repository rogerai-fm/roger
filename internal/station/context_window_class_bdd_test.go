package station

// context_window_class_bdd_test.go makes the Station and documentation scenarios of
// features/tower/context_window_class.feature EXECUTABLE against the real Station: a real
// HTTP model stand-in (httptest), the real HTTPUpstream, real sealing and a real Core-signed
// grant (sealedRig). Nothing of ours is mocked.
//
// Fixture notes (stated, not hidden):
//   - "a Station whose model replies N with the body ..." is an HTTP server that answers every
//     request with that status and that text.
//   - "the failure that crosses the Tower" is the failure string ServeSealed returns: the one
//     thing a Station hands the Tower in the clear.
//   - "the Station's own log" is the executor's Log writer, which the node wires to the
//     operator's own output.
//   - "the Tower operator documentation" is docs/tower.md.
//
// The broker scenarios of the same feature run in the broker's own package, and the
// self-hosted plane scenario in internal/agent.

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"rogerai.fm/roger/v7/internal/towercore/dispatch"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

const cwFeature = "../../features/tower/context_window_class.feature"

// cwStationScenarios are the scenarios this runner owns, by title.
var cwStationScenarios = []string{
	"The Station names the class for every overflow spelling",
	"The model's words still never cross the Tower",
	"A 400 that is not an overflow keeps the plain class",
	"Only a 400 carries the word",
	"A 429 with a wait is unchanged",
	"The operator's log keeps the model's words",
	"The reference documents the class",
}

type cwState struct {
	t *testing.T

	status     int
	retryAfter string
	body       string

	served  bool
	hits    int
	failure string
	log     bytes.Buffer
}

func (s *cwState) modelReplies(status, body string) error {
	s.status, _ = strconv.Atoi(status)
	s.body = body
	return nil
}

func (s *cwState) modelRepliesWithWait(status, wait, body string) error {
	s.retryAfter = wait
	return s.modelReplies(status, body)
}

func (s *cwState) servesSealed() error {
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.hits++
		if s.retryAfter != "" {
			w.Header().Set("Retry-After", s.retryAfter)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(s.status)
		_, _ = w.Write([]byte(s.body))
	}))
	defer model.Close()

	e, st, g, _, _ := sealedRig(s.t, time.Unix(1_700_000_000, 0), nil)
	e.Upstream = HTTPUpstream{URL: model.URL}
	e.Log = &s.log
	env, receipt, failure := e.ServeSealed(context.Background(), g.Signed, sealFor(s.t, st, g.AttemptID, []byte(`{"prompt":"hello"}`)))
	if s.hits != 1 {
		return fmt.Errorf("the model was asked %d times, want once (failure %q)", s.hits, failure)
	}
	// No envelope, and no receipt that claims work: a failed serve may carry only the Station's
	// signed zero-usage FAILURE receipt (tower_failed_attempt_release.feature), which releases the
	// consumer's hold and bills nobody.
	if len(env) != 0 || (len(receipt) != 0 && !dispatch.IsFailureReceipt(receipt)) {
		return fmt.Errorf("a failed serve returned an envelope (%d bytes) or a receipt that is not a zero-usage failure receipt (%d bytes)", len(env), len(receipt))
	}
	s.served, s.failure = true, failure
	return nil
}

func (s *cwState) crossed() (string, error) {
	if !s.served || s.failure == "" {
		return "", fmt.Errorf("no failure crossed the Tower (served %v)", s.served)
	}
	return s.failure, nil
}

func (s *cwState) failureIs(want string) error {
	got, err := s.crossed()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("the failure that crosses the Tower is %q, want %q", got, want)
	}
	return nil
}

func (s *cwState) failureLacks(text string) error {
	got, err := s.crossed()
	if err != nil {
		return err
	}
	if strings.Contains(got, text) {
		return fmt.Errorf("the failure that crosses the Tower contains %q: %q", text, got)
	}
	return nil
}

func (s *cwState) logContains(text string) error {
	if !s.served {
		return fmt.Errorf("the Station served nothing")
	}
	for _, line := range strings.Split(s.log.String(), "\n") {
		if strings.Contains(line, text) {
			return nil
		}
	}
	return fmt.Errorf("no line of the Station's own log contains %q; the log is %q", text, s.log.String())
}

// cwDoc is the Tower operator documentation with its whitespace folded, so a sentence wrapped
// over two lines still reads as one.
func cwDoc() (string, error) {
	raw, err := os.ReadFile("../../docs/tower.md")
	if err != nil {
		return "", err
	}
	return strings.Join(strings.Fields(string(raw)), " "), nil
}

func cwDocSays(phrases ...string) error {
	doc, err := cwDoc()
	if err != nil {
		return err
	}
	for _, p := range phrases {
		if !strings.Contains(doc, p) {
			return fmt.Errorf("docs/tower.md does not say %q", p)
		}
	}
	return nil
}

func (s *cwState) docListsClasses(word string) error {
	return cwDocSays(
		"Failure classes that cross a Tower",
		"`the model did not answer`",
		"`the model did not answer (status N)`",
		"`the model did not answer (status N, retry-after S)`",
		"`the model did not answer (status 400, "+word+")`",
	)
}

func (s *cwState) docSaysNoText() error {
	return cwDocSays("No text from the model crosses a Tower")
}

func TestTowerContextClassStationBDD(t *testing.T) {
	st := &cwState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
				*st = cwState{t: t}
				return c, nil
			})
			sc.Step(`^a Station whose model replies (\d+) with the body "([^"]*)"$`, st.modelReplies)
			sc.Step(`^a Station whose model replies (\d+) with Retry-After (\d+) and the body "([^"]*)"$`, st.modelRepliesWithWait)
			sc.Step(`^the Station serves a sealed attempt$`, st.servesSealed)
			sc.Step(`^the failure that crosses the Tower is "([^"]*)"$`, st.failureIs)
			sc.Step(`^the failure that crosses the Tower does not contain "([^"]*)"$`, st.failureLacks)
			sc.Step(`^the Station's own log line contains "([^"]*)"$`, st.logContains)
			sc.Step(`^the Tower operator documentation lists the failure classes that cross a Tower, including "([^"]*)"$`, st.docListsClasses)
			sc.Step(`^it says no text from the model crosses$`, st.docSaysNoText)
		},
		Options: &godog.Options{
			Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios(cwFeature, bddrun.Titled(false, cwStationScenarios...)),
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("context_window_class.feature (station) failed")
	}
}

// The class is decided on the model's whole reply, not on the part kept for the log: an
// overflow complaint that starts after the first 200 bytes is still an overflow.
func TestUpstreamOverflowIsReadFromTheWholeBody(t *testing.T) {
	long := strings.Repeat("x", 400) + " context length exceeded"
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(long))
	}))
	defer model.Close()

	_, err := HTTPUpstream{URL: model.URL}.Serve(context.Background(), []byte(`{}`))
	se, ok := err.(*UpstreamStatusError)
	if !ok {
		t.Fatalf("a non-2xx reply is typed, got %v", err)
	}
	if got, want := se.Class(), "the model did not answer (status 400, context-window)"; got != want {
		t.Fatalf("class %q, want %q", got, want)
	}
	if strings.Contains(se.Class(), "xxx") {
		t.Fatalf("the class echoes the model's words: %q", se.Class())
	}
}

// A 400 with a wait and an overflow body is still the context-window class: the wait is not
// what the consumer needs to know.
func TestUpstreamOverflowWinsOverRetryAfter(t *testing.T) {
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("too many tokens"))
	}))
	defer model.Close()

	_, err := HTTPUpstream{URL: model.URL}.Serve(context.Background(), []byte(`{}`))
	se, ok := err.(*UpstreamStatusError)
	if !ok {
		t.Fatalf("a non-2xx reply is typed, got %v", err)
	}
	if got, want := se.Class(), "the model did not answer (status 400, context-window)"; got != want {
		t.Fatalf("class %q, want %q", got, want)
	}
}

// cwServe runs one sealed attempt against a model that answers with status and body, and
// returns the failure that crossed and the operator's log.
func cwServe(t *testing.T, status int, body string) (failure, log string) {
	t.Helper()
	s := &cwState{t: t, status: status, body: body}
	if err := s.servesSealed(); err != nil {
		t.Fatal(err)
	}
	return s.failure, s.log.String()
}

// The operator's line is text from a model on the operator's terminal: one line per refused
// attempt, bounded, and with nothing in it that can move the cursor or start a second line.
func TestTheOperatorsLineIsOneBoundedLineWithNoControlCharacters(t *testing.T) {
	body := "context length\r\nexceeded \x1b[31mRED\x1b[0m \x07bell\ttab " + strings.Repeat("é", 4000)
	failure, log := cwServe(t, http.StatusBadRequest, body)
	if failure != "the model did not answer (status 400, context-window)" {
		t.Fatalf("failure %q", failure)
	}
	if !strings.HasSuffix(log, "\n") || strings.Count(log, "\n") != 1 {
		t.Fatalf("want exactly one line, got %q", log)
	}
	line := strings.TrimSuffix(log, "\n")
	if n := len([]rune(line)); n > 300 {
		t.Fatalf("the line is %d characters, want at most 300", n)
	}
	for _, r := range line {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0) || r == '�' {
			t.Fatalf("the line carries the control or broken character %U: %q", r, line)
		}
	}
	for _, want := range []string{"status 400", "context length", "exceeded", "RED", "bell", "tab"} {
		if !strings.Contains(line, want) {
			t.Fatalf("the line lost %q: %q", want, line)
		}
	}
}

// A serve that fails with no reply from the model has no model words, so it writes no line.
func TestAnUnreachableModelWritesNoOperatorLine(t *testing.T) {
	var log bytes.Buffer
	e, st, g, _, _ := sealedRig(t, time.Unix(1_700_000_000, 0), nil)
	e.Upstream = HTTPUpstream{URL: "http://127.0.0.1:1"}
	e.Log = &log
	_, _, failure := e.ServeSealed(context.Background(), g.Signed, sealFor(t, st, g.AttemptID, []byte(`{"prompt":"hello"}`)))
	if failure != "the model did not answer" {
		t.Fatalf("failure %q", failure)
	}
	if log.Len() != 0 {
		t.Fatalf("no model replied, yet the operator's log has %q", log.String())
	}
}

// Regression (review 2026-11-23). Unicode format and separator characters are not control
// characters, and they rewrite a terminal line just as well: a right-to-left override makes
// the model's words read as the line's own prefix. Nothing that is not printable survives,
// anywhere in the line, the attempt id included.
func TestTheOperatorsLineCarriesNoCharacterThatCanRewriteIt(t *testing.T) {
	hostile := []rune{
		'\u202E',     // right-to-left override
		'\u2066',     // left-to-right isolate
		'\u2028',     // line separator
		'\u200B',     // zero width space
		'\uFEFF',     // byte order mark
		'\U000E0001', // language tag
		'\x1b',       // ESC
		'\u009B',     // C1 control sequence introducer
	}
	words := "context length exceeded"
	for _, r := range hostile {
		words += " " + string(r) + ")004 sutats( 1-tta tpmetta desufer ledom eht"
	}
	failure, log := cwServe(t, http.StatusBadRequest, words)
	if failure != "the model did not answer (status 400, context-window)" {
		t.Fatalf("failure %q", failure)
	}
	for _, line := range []string{
		strings.TrimSuffix(log, "\n"),
		operatorLine("att-\u202E1\u2028x\x1b[2J", &UpstreamStatusError{Status: 400, Detail: words}),
	} {
		if !strings.HasPrefix(line, "the model answered attempt att-") || !strings.Contains(line, " with status 400: context length exceeded") {
			t.Fatalf("the line lost its shape: %q", line)
		}
		if n := len([]rune(line)); n > 300 {
			t.Fatalf("the line is %d characters, want at most 300", n)
		}
		for _, r := range line {
			if !unicode.IsPrint(r) {
				t.Fatalf("the line carries the unprintable character %U: %q", r, line)
			}
		}
	}
}

// The line says what happened, whatever the status: the model answered with an error status.
// It does not call a 429 or a 500 a refusal.
func TestTheOperatorsLineIsNeutralForEveryStatus(t *testing.T) {
	for _, status := range []int{400, 429, 500, 503} {
		_, log := cwServe(t, status, "the upstream said no")
		want := fmt.Sprintf(" with status %d: the upstream said no\n", status)
		if !strings.HasPrefix(log, "the model answered attempt att-") || !strings.HasSuffix(log, want) || strings.Contains(log, "refused") {
			t.Fatalf("status %d: the line is %q", status, log)
		}
	}
}
