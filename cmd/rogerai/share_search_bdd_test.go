package main

// share_search_bdd_test.go makes features/search/share_cli.feature and the CLI half of
// features/search/native_provider.feature EXECUTABLE: the REAL `roger share --search` and
// `roger search check` code paths against REAL httptest engines (SearXNG, Brave and native
// shapes) that record every request. The go-live is the agentStart seam: "goes on air" means
// the share reached agentStart with a search config.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/agent"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/searchengine"
)

// countKey carries the count a native request asked for to the reply.
type countKey struct{}

// fakeEngine is a recording engine answering in one engine's shape.
type fakeEngine struct {
	kind  string
	mu    sync.Mutex
	qs    []string
	auths []string
	srv   *httptest.Server
	reply func(w http.ResponseWriter, r *http.Request, q string)
}

func newFakeEngine(kind string) *fakeEngine {
	f := &fakeEngine{kind: kind}
	f.reply = func(w http.ResponseWriter, r *http.Request, q string) {
		n, _ := r.Context().Value(countKey{}).(int)
		f.answer(w, r, n, "https")
	}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		n := 3
		if f.kind == protocol.EngineNative {
			var b struct {
				Query string `json:"query"`
				Count int    `json:"count"`
			}
			raw, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(raw)) // the reply may read it again
			_ = json.Unmarshal(raw, &b)
			q = b.Query
			if b.Count > 0 && b.Count < n {
				n = b.Count // a conforming native service honours count
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), countKey{}, n))
		f.mu.Lock()
		f.qs = append(f.qs, q)
		f.auths = append(f.auths, r.Header.Get("X-Subscription-Token")+r.Header.Get("Authorization"))
		reply := f.reply
		f.mu.Unlock()
		reply(w, r, q)
	}))
	return f
}

// answer writes n results in the engine's own shape, with URLs on scheme.
func (f *fakeEngine) answer(w http.ResponseWriter, r *http.Request, n int, scheme string) {
	var items []map[string]any
	for i := 1; i <= n; i++ {
		u := fmt.Sprintf("%s://example.com/r%d", scheme, i)
		switch f.kind {
		case protocol.EngineBrave:
			items = append(items, map[string]any{"title": "T", "url": u, "description": "D"})
		case protocol.EngineSearXNG:
			items = append(items, map[string]any{"title": "T", "url": u, "content": "C"})
		default:
			items = append(items, map[string]any{"title": "T", "url": u, "snippet": "S"})
		}
	}
	var out any = map[string]any{"results": items}
	if f.kind == protocol.EngineBrave {
		out = map[string]any{"web": map[string]any{"results": items}}
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (f *fakeEngine) queries() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.qs...)
}

type shareSearchState struct {
	t *testing.T

	cfg      config
	env      map[string]string
	answers  string // what the provider types at prompts
	tty      bool
	secretOK bool // readSecret was called

	searx, native, brave *fakeEngine

	out      bytes.Buffer
	err      error
	started  *agent.Config // what reached agentStart (the go-live)
	runs     int
	engineOf string // the engine a Given is about
	chatUp   *httptest.Server
	chatCfg  *agent.Config
	bandCode string // the frequency code the stub broker grants a private band ("" = none)
}

func (s *shareSearchState) reset() {
	for _, f := range []*fakeEngine{s.searx, s.native, s.brave} {
		if f != nil {
			f.srv.Close()
		}
	}
	if s.chatUp != nil {
		s.chatUp.Close()
	}
	// a person at a terminal who answers yes, unless a scenario says otherwise: the
	// terms scenarios set the terminal and the answer themselves.
	*s = shareSearchState{t: s.t, env: map[string]string{}, tty: true, answers: "yes\n"}
	s.t.Setenv("XDG_CONFIG_HOME", s.t.TempDir())
	s.t.Setenv("HOME", s.t.TempDir())
	s.searx = newFakeEngine(protocol.EngineSearXNG)
	s.native = newFakeEngine(protocol.EngineNative)
	s.brave = newFakeEngine(protocol.EngineBrave)
	braveEndpoint = s.brave.srv.URL + "/res/v1/web/search"
	writeAuth(s.t) // a signed-in owner: earning shares are allowed
}

// mapURLs points the spec's addresses at the recording engines.
func (s *shareSearchState) mapURLs(args string) string {
	args = strings.ReplaceAll(args, "http://127.0.0.1:8888", s.searx.srv.URL)
	return strings.ReplaceAll(args, "http://127.0.0.1:9000", s.native.srv.URL)
}

func (s *shareSearchState) io() searchShareIO {
	return searchShareIO{
		getenv:      func(k string) string { return s.env[k] },
		interactive: s.tty,
		in:          bufio.NewReader(strings.NewReader(s.answers)),
		readSecret: func() (string, error) {
			s.secretOK = true
			line, _, _ := bufio.NewReader(strings.NewReader(s.answers)).ReadLine()
			return string(line), nil
		},
		out: &s.out,
	}
}

// share runs `roger share <args>` down the real search-share path with the go-live seam.
func (s *shareSearchState) share(args string) {
	origStart, origBlock := agentStart, shareBlock
	origBand, origWait := bandOf, searchOnAirWait
	agentStart = func(c agent.Config) (*agent.Session, error) {
		cc := c
		s.started = &cc
		return &agent.Session{}, nil // a session the stub broker accepted
	}
	bandOf = func(*agent.Session) (string, string, string) {
		if s.started != nil && s.started.Private && s.bandCode != "" {
			return "band-1", s.bandCode, "RGR-TEST-****"
		}
		return "", "", ""
	}
	searchOnAirWait = 0
	shareBlock = func() {}
	defer func() { agentStart, shareBlock, bandOf, searchOnAirWait = origStart, origBlock, origBand, origWait }()
	s.runs++
	s.started = nil
	s.out.Reset()
	s.err = shareSearch(s.cfg, splitArgs(s.mapURLs(args)), s.io())
	if errors.Is(s.err, errStubOnAir) {
		s.err = nil // the go-live was reached: that is the success this seam reports
	}
	s.cfg = loadConfig()
}

var errStubOnAir = errors.New("stub: on air")

// --- share_cli steps ---------------------------------------------------------------

func (s *shareSearchState) providerRuns(args string) error {
	s.share(strings.ReplaceAll(args, `\"`, `"`))
	return nil
}
func (s *shareSearchState) registersOffer(model, engine string) error {
	if s.started == nil {
		return fmt.Errorf("nothing went on air: %v\n%s", s.err, s.out.String())
	}
	if s.started.Modality != protocol.ModalitySearch || s.started.Model != model || s.started.SearchEngine != engine {
		return fmt.Errorf("on air as %s/%s/%s, want search/%s/%s", s.started.Modality, s.started.Model, s.started.SearchEngine, model, engine)
	}
	return nil
}
func (s *shareSearchState) exitsNonZero() error {
	if s.err == nil {
		return fmt.Errorf("the command succeeded:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) messageSays(quoted, bare string) error {
	want := quoted + bare
	if s.err == nil || !strings.Contains(s.err.Error(), want) {
		return fmt.Errorf("message %v does not say %q", s.err, want)
	}
	return nil
}
func (s *shareSearchState) nothingRegistered() error {
	if s.started != nil {
		return fmt.Errorf("a station went on air")
	}
	return nil
}
func (s *shareSearchState) priceIs(p float64) error {
	if s.started == nil {
		return fmt.Errorf("nothing went on air: %v", s.err)
	}
	if s.started.PriceIn != p || s.started.PriceOut != 0 {
		return fmt.Errorf("price %v/%v, want %v per 1,000,000 queries", s.started.PriceIn, s.started.PriceOut, p)
	}
	return nil
}
func (s *shareSearchState) priceZero() error { return s.priceIs(0) }
func (s *shareSearchState) scopeIs(scope string) error {
	if s.started == nil || s.started.Scope != scope {
		return fmt.Errorf("scope %v, want %q (%v)", s.started, scope, s.err)
	}
	return nil
}

func (s *shareSearchState) searxAnswers() error { return nil }

func (s *shareSearchState) nativeEmpty() error {
	s.native.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { _, _ = w.Write([]byte(`{"results":[]}`)) }
	return nil
}

func (s *shareSearchState) grantsBand(code string) error { s.bandCode = code; return nil }

func (s *shareSearchState) printsCodeOnce(code string) error {
	if n := strings.Count(s.out.String(), code); n != 1 {
		return fmt.Errorf("the code is printed %d times, want once:\n%s", n, s.out.String())
	}
	return nil
}

func (s *shareSearchState) saysPrivate() error {
	out := strings.ToLower(s.out.String())
	if !strings.Contains(out, "private") || !strings.Contains(out, "only with this code") {
		return fmt.Errorf("output does not say the band is private and reachable only with the code:\n%s", s.out.String())
	}
	return nil
}

func (s *shareSearchState) noCode() error {
	if strings.Contains(s.out.String(), "FREQUENCY CODE") {
		return fmt.Errorf("a public share printed a frequency code:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) exactlyOneCheck() error {
	if n := len(s.searx.queries()); n != 1 {
		return fmt.Errorf("the engine got %d queries, want exactly one check", n)
	}
	return nil
}
func (s *shareSearchState) goesOnAir() error {
	if s.started == nil {
		return fmt.Errorf("the search offer did not go on air: %v\n%s", s.err, s.out.String())
	}
	return nil
}
func (s *shareSearchState) checkPassedCount() error {
	if !regexp.MustCompile(`check passed: \d+ results?`).MatchString(s.out.String()) {
		return fmt.Errorf("output lacks the passed check line:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) engineFails(how string) error {
	switch {
	case strings.Contains(how, "refuses the connection"):
		s.engineOf = "searxng-dead"
	case strings.Contains(how, "403"):
		s.engineOf = protocol.EngineSearXNG
		s.searx.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { w.WriteHeader(403) }
	case strings.Contains(how, "401"):
		s.engineOf = protocol.EngineBrave
		s.brave.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { w.WriteHeader(401) }
	case strings.Contains(how, "without a results array"):
		s.engineOf = protocol.EngineNative
		s.native.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { _, _ = w.Write([]byte(`{"hits":[]}`)) }
	case strings.Contains(how, "only non-http"):
		s.engineOf = protocol.EngineNative
		s.native.reply = func(w http.ResponseWriter, r *http.Request, _ string) { s.native.answer(w, r, 3, "ftp") }
	case strings.Contains(how, "longer than 10 seconds"):
		s.engineOf = protocol.EngineNative
		searchengine.EngineTimeout = 150 * time.Millisecond
		s.native.reply = func(w http.ResponseWriter, r *http.Request, _ string) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}
	default:
		return fmt.Errorf("unknown failure %q", how)
	}
	return nil
}
func (s *shareSearchState) sharesIt() error {
	switch s.engineOf {
	case "searxng-dead":
		s.share("--search searxng --upstream http://127.0.0.1:1")
	case protocol.EngineSearXNG:
		s.share("--search searxng --upstream http://127.0.0.1:8888")
	case protocol.EngineBrave:
		s.share("--search brave --upstream-key k --price-per-1k 6 --accept-engine-terms")
	default:
		s.share("--search native --upstream http://127.0.0.1:9000")
	}
	return nil
}
func (s *shareSearchState) notOnAir() error {
	if s.started != nil {
		return fmt.Errorf("the search offer went on air")
	}
	return nil
}
func (s *shareSearchState) printsReason(reason string) error {
	reason = strings.ReplaceAll(reason, `\"`, `"`)
	all := s.out.String()
	if s.err != nil {
		all += s.err.Error()
	}
	if !strings.Contains(all, reason) {
		return fmt.Errorf("output does not say %q:\n%s", reason, all)
	}
	return nil
}
func (s *shareSearchState) chatAndDeadSearx() error {
	s.chatUp = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"llama3.2"}]}`))
	}))
	return nil
}
func (s *shareSearchState) sharesBoth() error {
	s.share("--search searxng --upstream http://127.0.0.1:1")
	searchStarted := s.started
	origStart, origBlock := agentStart, shareBlock
	agentStart = func(c agent.Config) (*agent.Session, error) { cc := c; s.chatCfg = &cc; return nil, errStubOnAir }
	shareBlock = func() {}
	defer func() { agentStart, shareBlock = origStart, origBlock }()
	_ = cmdShare(config{Broker: "https://b", User: "u"}, []string{"llama3.2", "--upstream", s.chatUp.URL + "/v1/chat/completions"})
	s.started = searchStarted
	return nil
}
func (s *shareSearchState) chatOnAir() error {
	if s.chatCfg == nil || s.chatCfg.Model != "llama3.2" {
		return fmt.Errorf("the chat offer did not go on air")
	}
	return nil
}
func (s *shareSearchState) searchNot() error { return s.notOnAir() }
func (s *shareSearchState) runsCheck() error {
	s.share("--search searxng --upstream http://127.0.0.1:8888")
	return nil
}
func (s *shareSearchState) fixedWeather(q string) error {
	qs := s.searx.queries()
	if len(qs) != 1 || qs[0] != q {
		return fmt.Errorf("check queries %q, want [%q]", qs, q)
	}
	return nil
}

// terms
func (s *shareSearchState) neverShared(engine string) error { s.engineOf = engine; return nil }
func (s *shareSearchState) sharesInteractively() error {
	s.tty, s.answers = true, "yes\n"
	return s.shareEngine()
}
func (s *shareSearchState) shareEngine() error {
	if s.engineOf == protocol.EngineBrave {
		s.share("--search brave --upstream-key k --price-per-1k 6")
	} else {
		s.share("--search searxng --upstream http://127.0.0.1:8888")
	}
	return nil
}
func (s *shareSearchState) saysResponsible(engine string) error {
	if !strings.Contains(s.out.String(), "you are responsible for following "+engine+"'s terms for serving results to others") {
		return fmt.Errorf("no terms line for %s:\n%s", engine, s.out.String())
	}
	return nil
}
func (s *shareSearchState) asksContinue() error {
	if !strings.Contains(s.out.String(), "continue? [y/N]") {
		return fmt.Errorf("no continue prompt:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) yesSaved(engine string) error {
	if !s.cfg.acceptedSearchTerms(engine) {
		return fmt.Errorf("the yes for %s was not saved", engine)
	}
	s.out.Reset()
	s.answers = ""
	s.shareEngine()
	if strings.Contains(s.out.String(), "continue? [y/N]") {
		return fmt.Errorf("the question was asked again:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) answersNo() error {
	s.engineOf, s.tty, s.answers = protocol.EngineSearXNG, true, "no\n"
	return s.shareEngine()
}
func (s *shareSearchState) askedAgain() error {
	if s.cfg.acceptedSearchTerms(protocol.EngineSearXNG) {
		return fmt.Errorf("a no was saved as a yes")
	}
	s.answers = "no\n"
	s.shareEngine()
	if !strings.Contains(s.out.String(), "continue? [y/N]") {
		return fmt.Errorf("not asked again:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) noTerminal() error {
	s.tty = false
	return s.shareEngine()
}
func (s *shareSearchState) namesFlag(flag string) error {
	if s.err == nil || !strings.Contains(s.err.Error(), flag) {
		return fmt.Errorf("message %v does not name %s", s.err, flag)
	}
	return nil
}
func (s *shareSearchState) noTermsQuestion() error {
	if strings.Contains(s.out.String(), "continue? [y/N]") || strings.Contains(s.out.String(), "responsible for following") {
		return fmt.Errorf("a terms question was asked:\n%s", s.out.String())
	}
	return s.goesOnAir()
}

// keys
func (s *shareSearchState) braveEnv(v string) error {
	if v == "" {
		v = "env-key"
	}
	s.env["BRAVE_SEARCH_API_KEY"] = v
	return nil
}
func (s *shareSearchState) braveUsesKey(want string) error {
	if want == "" {
		want = s.env["BRAVE_SEARCH_API_KEY"]
	}
	s.brave.mu.Lock()
	defer s.brave.mu.Unlock()
	if len(s.brave.auths) == 0 || s.brave.auths[len(s.brave.auths)-1] != want {
		return fmt.Errorf("brave saw key %q, want %q (%v)", s.brave.auths, want, s.err)
	}
	return nil
}
func (s *shareSearchState) noBraveKey() error { return nil }
func (s *shareSearchState) sharesBraveInteractively() error {
	s.tty, s.answers = true, "typed-key\n"
	s.cfg.SearchTerms = []string{protocol.EngineBrave}
	s.share("--search brave --price-per-1k 6")
	return nil
}
func (s *shareSearchState) askedHidden() error {
	if !s.secretOK {
		return fmt.Errorf("the key was not asked for through the hidden reader")
	}
	return nil
}
func (s *shareSearchState) savesBraveKey() error {
	s.tty, s.answers = true, "typed-key\n"
	s.cfg.SearchTerms = []string{protocol.EngineBrave}
	s.share("--search brave --price-per-1k 6")
	return nil
}
func (s *shareSearchState) mode0600() error {
	if s.cfg.SearchKeys[protocol.EngineBrave] == "" {
		return fmt.Errorf("the key was not saved (%v)", s.err)
	}
	fi, err := os.Stat(configPath())
	if err != nil {
		return err
	}
	if fi.Mode().Perm() != 0o600 {
		return fmt.Errorf("config mode %v, want 0600", fi.Mode().Perm())
	}
	return nil
}
func (s *shareSearchState) savedKey(k string) error {
	s.cfg.SearchKeys = map[string]string{protocol.EngineBrave: k}
	s.cfg.SearchTerms = []string{protocol.EngineBrave}
	if err := saveConfig(s.cfg); err != nil {
		return err
	}
	return nil
}
func (s *shareSearchState) outputLacks(k string) error {
	all := s.out.String()
	if s.err != nil {
		all += s.err.Error()
	}
	if strings.Contains(all, k) {
		return fmt.Errorf("the key was printed:\n%s", all)
	}
	return nil
}
func (s *shareSearchState) saysKeySet() error {
	if !strings.Contains(s.out.String(), "key set") {
		return fmt.Errorf("output does not say key set:\n%s", s.out.String())
	}
	return nil
}

// --- native_provider steps -----------------------------------------------------------

func (s *shareSearchState) conforming() error { return nil }
func (s *shareSearchState) runsCheckCmd(args string) error {
	origStart := agentStart
	agentStart = func(c agent.Config) (*agent.Session, error) { cc := c; s.started = &cc; return nil, errStubOnAir }
	defer func() { agentStart = origStart }()
	s.out.Reset()
	s.err = cmdSearch(s.cfg, splitArgs(s.mapURLs(args)), &s.out)
	return nil
}
func (s *shareSearchState) everyCheckPasses() error {
	if strings.Contains(s.out.String(), "FAIL") {
		return fmt.Errorf("a check failed:\n%s", s.out.String())
	}
	if !strings.Contains(s.out.String(), "ok") {
		return fmt.Errorf("no checks reported:\n%s", s.out.String())
	}
	return nil
}
func (s *shareSearchState) exitsZero() error {
	if s.err != nil {
		return fmt.Errorf("exited with %v:\n%s", s.err, s.out.String())
	}
	return nil
}
func (s *shareSearchState) nothingWithBroker() error { return s.nothingRegistered() }
func (s *shareSearchState) serviceBreaks(how string) error {
	f := s.native
	answer := func(w http.ResponseWriter, results []map[string]any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
	}
	good := func(n int) []map[string]any {
		out := []map[string]any{}
		for i := 1; i <= n; i++ {
			out = append(out, map[string]any{"title": "T", "url": fmt.Sprintf("https://e.com/%d", i), "snippet": "S"})
		}
		return out
	}
	switch how {
	case "answers 404 on POST /v1/search":
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { w.WriteHeader(404) }
	case "returns HTML instead of JSON":
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { _, _ = w.Write([]byte("<html></html>")) }
	case `returns {"hits": [...]}`:
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { _, _ = w.Write([]byte(`{"hits":[]}`)) }
	case "returns more results than the count asked for":
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) { answer(w, good(25)) }
	case `returns a result with a "javascript:" url`:
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) {
			answer(w, append(good(1), map[string]any{"title": "x", "url": "javascript:alert(1)"}))
		}
	case "returns a result with no title":
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) {
			answer(w, []map[string]any{{"url": "https://e.com/1", "snippet": "S"}})
		}
	case `returns published "last tuesday"`:
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) {
			answer(w, []map[string]any{{"title": "T", "url": "https://e.com/1", "published": "last tuesday"}})
		}
	case "takes 12 seconds to answer":
		searchengine.EngineTimeout = 150 * time.Millisecond
		f.reply = func(w http.ResponseWriter, r *http.Request, _ string) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}
	case "returns a 600 KiB reply":
		f.reply = func(w http.ResponseWriter, _ *http.Request, _ string) {
			answer(w, []map[string]any{{"title": strings.Repeat("x", 600*1024), "url": "https://e.com/1"}})
		}
	case `answers 500 to a query with freshness "week"`:
		f.reply = func(w http.ResponseWriter, r *http.Request, _ string) {
			var b map[string]any
			_ = json.NewDecoder(r.Body).Decode(&b)
			if b["freshness"] == "week" {
				w.WriteHeader(500)
				return
			}
			answer(w, good(2))
		}
	case "answers 500 to a query with Unicode text":
		f.reply = func(w http.ResponseWriter, r *http.Request, q string) {
			for _, c := range q {
				if c > 127 {
					w.WriteHeader(500)
					return
				}
			}
			answer(w, good(2))
		}
	default:
		return fmt.Errorf("unknown break %q", how)
	}
	return nil
}
func (s *shareSearchState) failingCheckNamed(name string) error {
	if s.err == nil {
		return fmt.Errorf("the check passed:\n%s", s.out.String())
	}
	for _, line := range strings.Split(s.out.String(), "\n") {
		if strings.Contains(line, "FAIL") && strings.Contains(line, name) {
			return nil
		}
	}
	return fmt.Errorf("no failing check named %q:\n%s", name, s.out.String())
}
func (s *shareSearchState) ignoresFreshCountry() error {
	s.native.reply = func(w http.ResponseWriter, r *http.Request, _ string) { s.native.answer(w, r, 2, "https") }
	return nil
}
func (s *shareSearchState) warnsIgnored() error {
	out := s.out.String()
	if !strings.Contains(out, "warn") || !strings.Contains(out, "freshness") || !strings.Contains(out, "country") {
		return fmt.Errorf("no ignored-filters warning:\n%s", out)
	}
	return nil
}
func (s *shareSearchState) fixedStrings() error {
	qs := s.native.queries()
	if len(qs) == 0 {
		return fmt.Errorf("no queries were sent")
	}
	fixed := map[string]bool{}
	for _, q := range searchengine.CheckQueries {
		fixed[q] = true
	}
	for _, q := range qs {
		if !fixed[q] {
			return fmt.Errorf("query %q is not one of the fixed test strings", q)
		}
	}
	return nil
}
func (s *shareSearchState) contractHeader() error {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Search-Contract")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()
	_, _ = searchengine.Engine{Kind: protocol.EngineNative, Upstream: srv.URL}.Search(context.Background(), searchengine.Request{Query: "q", Count: 1})
	if got != "search/1" {
		return fmt.Errorf("X-Search-Contract = %q, want search/1", got)
	}
	return nil
}
func (s *shareSearchState) nodeSendsNative() error { return nil }

// the guide
func (s *shareSearchState) readerOpensGuide() error { return nil }
func (s *shareSearchState) guide() (string, error) {
	b, err := os.ReadFile("../../docs/search-providers.md")
	return strings.ToLower(string(b)), err
}
func (s *shareSearchState) guideCovers(what string) error {
	g, err := s.guide()
	if err != nil {
		return fmt.Errorf("public provider guide missing: %w", err)
	}
	needles := map[string][]string{
		"routes":   {"brave", "searxng", "native"},
		"contract": {"search/1", "post /v1/search", "512 kib", "10 s"},
		"example":  {"```"},
		"privacy":  {"do not log", "sell", "identify"},
		"terms":    {"terms"},
		"pricing":  {"--price-per-1k", "payout"},
		"check":    {"roger search check"},
	}[what]
	for _, n := range needles {
		if !strings.Contains(g, n) {
			return fmt.Errorf("the guide does not cover %s (%q missing)", what, n)
		}
	}
	return nil
}
func (s *shareSearchState) guideMatchesChecker() error {
	g, err := s.guide()
	if err != nil {
		return err
	}
	// every rule the guide lists under its contract checks is a check roger runs
	re := regexp.MustCompile("(?m)^- check `([^`]+)`")
	listed := re.FindAllStringSubmatch(g, -1)
	if len(listed) == 0 {
		return fmt.Errorf("the guide lists no contract checks")
	}
	known := map[string]bool{}
	for _, n := range searchengine.CheckNames {
		known[strings.ToLower(n)] = true
	}
	for _, m := range listed {
		if !known[m[1]] {
			return fmt.Errorf("the guide lists check %q that roger search check does not run", m[1])
		}
	}
	if len(listed) != len(searchengine.CheckNames) {
		return fmt.Errorf("the guide lists %d checks, roger runs %d", len(listed), len(searchengine.CheckNames))
	}
	return nil
}

func searchCLISteps(sc *godog.ScenarioContext, s *shareSearchState) {
	sc.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
		s.reset()
		return c, nil
	})
	sc.After(func(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		searchengine.EngineTimeout = searchengine.DefaultEngineTimeout
		braveEndpoint = searchengine.BraveEndpoint
		return c, nil
	})
	sc.Step(`^the provider runs "roger share (.*)"$`, s.providerRuns)
	sc.Step(`^the node registers a search offer "([^"]*)" with engine "([^"]*)"$`, s.registersOffer)
	sc.Step(`^the command exits non-zero$`, s.exitsNonZero)
	sc.Step(`^the message says (?:"(.*)"|([^"].*))$`, s.messageSays)
	sc.Step(`^nothing is registered$`, s.nothingRegistered)
	sc.Step(`^the registered offer price is (\d+) credits per 1,000,000 queries$`, s.priceIs)
	sc.Step(`^the registered offer price is 0$`, s.priceZero)
	sc.Step(`^the registered offer scope is "([^"]*)"$`, s.scopeIs)
	sc.Step(`^a SearXNG at http://127.0.0.1:8888 that answers searches$`, s.searxAnswers)
	sc.Step(`^a native service that answers every search with an empty results array$`, s.nativeEmpty)
	sc.Step(`^the broker grants a private band with frequency code "([^"]*)"$`, s.grantsBand)
	sc.Step(`^roger prints the frequency code "([^"]*)" exactly once$`, s.printsCodeOnce)
	sc.Step(`^the output says the band is private and reachable only with the code$`, s.saysPrivate)
	sc.Step(`^the output has no frequency code$`, s.noCode)
	sc.Step(`^roger sends exactly one check query to it$`, s.exactlyOneCheck)
	sc.Step(`^the search offer goes on air$`, s.goesOnAir)
	sc.Step(`^roger prints that the check passed with the number of results returned$`, s.checkPassedCount)
	sc.Step(`^an engine that (.+)$`, s.engineFails)
	sc.Step(`^the provider shares it as a search station$`, s.sharesIt)
	sc.Step(`^the search offer does not go on air$`, s.notOnAir)
	sc.Step(`^roger prints "(.*)"$`, s.printsReason)
	sc.Step(`^a node with a chat model and a SearXNG that refuses connections$`, s.chatAndDeadSearx)
	sc.Step(`^the provider shares both$`, s.sharesBoth)
	sc.Step(`^the chat offer goes on air$`, s.chatOnAir)
	sc.Step(`^the search offer does not$`, s.searchNot)
	sc.Step(`^roger runs the check query$`, s.runsCheck)
	sc.Step(`^the query is the fixed string "([^"]*)"$`, s.fixedWeather)
	sc.Step(`^the provider has never shared engine "([^"]*)"$`, s.neverShared)
	sc.Step(`^the provider shares it interactively$`, s.sharesInteractively)
	sc.Step(`^roger says the provider is responsible for following ([a-z]+)'s terms for serving results to others$`, s.saysResponsible)
	sc.Step(`^roger asks to continue$`, s.asksContinue)
	sc.Step(`^a "yes" is saved so the question is not asked again for "([^"]*)"$`, s.yesSaved)
	sc.Step(`^the provider answers "no" to the engine terms question$`, s.answersNo)
	sc.Step(`^the question is asked again next time$`, s.askedAgain)
	sc.Step(`^the provider shares it with no terminal attached and without --accept-engine-terms$`, s.noTerminal)
	sc.Step(`^the message names (--[a-z-]+)$`, s.namesFlag)
	sc.Step(`^no terms question is asked$`, s.noTermsQuestion)
	sc.Step(`^BRAVE_SEARCH_API_KEY is set$`, func() error { return s.braveEnv("") })
	sc.Step(`^BRAVE_SEARCH_API_KEY is set to "([^"]*)"$`, s.braveEnv)
	sc.Step(`^the Brave adapter uses that key$`, func() error { return s.braveUsesKey("") })
	sc.Step(`^the Brave adapter uses key "([^"]*)"$`, s.braveUsesKey)
	sc.Step(`^no Brave key is configured$`, s.noBraveKey)
	sc.Step(`^the provider shares Brave interactively$`, s.sharesBraveInteractively)
	sc.Step(`^roger asks for the key with input hidden$`, s.askedHidden)
	sc.Step(`^the provider saves a Brave key$`, s.savesBraveKey)
	sc.Step(`^the config file holding it has mode 0600$`, s.mode0600)
	sc.Step(`^a saved Brave key "([^"]*)"$`, s.savedKey)
	sc.Step(`^the output does not contain "([^"]*)"$`, s.outputLacks)
	sc.Step(`^it says "key set"$`, s.saysKeySet)

	sc.Step(`^a local service that implements search/1 correctly$`, s.conforming)
	sc.Step(`^the provider runs "roger search (.*)"$`, s.runsCheckCmd)
	sc.Step(`^every check passes$`, s.everyCheckPasses)
	sc.Step(`^the command exits zero$`, s.exitsZero)
	sc.Step(`^nothing is registered with the broker$`, s.nothingWithBroker)
	sc.Step(`^a local service that (answers 404.*|returns HTML.*|returns \{"hits".*|returns more results.*|returns a result with.*|returns published.*|takes 12 seconds.*|returns a 600 KiB.*|answers 500 to.*)$`, s.serviceBreaks)
	sc.Step(`^the failing check is named "([^"]*)"$`, s.failingCheckNamed)
	sc.Step(`^a local service that ignores freshness and country$`, s.ignoresFreshCountry)
	sc.Step(`^it warns that freshness and country are ignored$`, s.warnsIgnored)
	sc.Step(`^every query sent is one of the check's fixed test strings$`, s.fixedStrings)
	sc.Step(`^the node sends a search to a native service$`, s.nodeSendsNative)
	sc.Step(`^the request carries header "X-Search-Contract: search/1"$`, s.contractHeader)
	sc.Step(`^a reader opens the public search provider guide$`, s.readerOpensGuide)
	sc.Step(`^it explains the three routes: Brave, SearXNG and native$`, func() error { return s.guideCovers("routes") })
	sc.Step(`^it gives the search/1 contract exactly as roger checks it$`, func() error { return s.guideCovers("contract") })
	sc.Step(`^it shows a short native example service$`, func() error { return s.guideCovers("example") })
	sc.Step(`^it states the provider's privacy duties: do not log or sell queries, do not try to identify callers$`, func() error { return s.guideCovers("privacy") })
	sc.Step(`^it states that Brave and SearXNG providers answer for those engines' terms$`, func() error { return s.guideCovers("terms") })
	sc.Step(`^it explains how to price per 1,000 queries and how payouts work$`, func() error { return s.guideCovers("pricing") })
	sc.Step("^it lists `roger search check` as the first step for a native provider$", func() error { return s.guideCovers("check") })
	sc.Step(`^every rule the guide lists as part of search/1 has a check of the same name in roger search check$`, s.guideMatchesChecker)
	sc.Step(`^roger is not logged in$`, func() error {
		return os.Remove(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "rogerai", "auth.json"))
	})
	sc.Step(`^no check query is sent$`, func() error {
		if n := len(s.searx.queries()); n != 0 {
			return fmt.Errorf("the engine got %d queries before the login check", n)
		}
		return nil
	})
}

func TestSearchShareCLIBDD(t *testing.T) {
	s := &shareSearchState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { searchCLISteps(sc, s) },
		Options: &godog.Options{Format: "pretty", TestingT: t, Strict: true,
			Paths: []string{"../../features/search/share_cli.feature"}},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("search share CLI scenarios failed")
	}
}

func TestSearchNativeProviderBDD(t *testing.T) {
	s := &shareSearchState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { searchCLISteps(sc, s) },
		Options: &godog.Options{Format: "pretty", TestingT: t, Strict: true, Tags: "~@broker",
			Paths: []string{"../../features/search/native_provider.feature"}},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("native provider scenarios failed")
	}
}

// TestSearchOwnerRequiredCLIBDD runs the @cli scenario of owner_required.feature (the broker
// scenarios run in cmd/rogerai-broker).
func TestSearchOwnerRequiredCLIBDD(t *testing.T) {
	s := &shareSearchState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { searchCLISteps(sc, s) },
		Options: &godog.Options{Format: "pretty", TestingT: t, Strict: true, Tags: "@cli",
			Paths: []string{"../../features/search/owner_required.feature"}},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("search owner-required CLI scenarios failed")
	}
}
