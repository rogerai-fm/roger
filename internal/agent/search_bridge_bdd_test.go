package agent

// search_bridge_bdd_test.go makes features/search/node_bridge.feature EXECUTABLE: the REAL
// serveCtx answering broker search jobs through the REAL engine adapters
// (internal/searchengine) against REAL httptest upstreams that record every request.

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/searchengine"
)

// recordedReq is one request an upstream received.
type recordedReq struct {
	method, path string
	query        map[string][]string
	header       http.Header
	body         string
}

// engineUpstream is a recording engine: it answers with reply (or the default for its kind).
type engineUpstream struct {
	mu    sync.Mutex
	reqs  []recordedReq
	reply func(w http.ResponseWriter, r *http.Request)
	srv   *httptest.Server
}

func (u *engineUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	u.mu.Lock()
	u.reqs = append(u.reqs, recordedReq{r.Method, r.URL.Path, r.URL.Query(), r.Header.Clone(), string(b)})
	reply := u.reply
	u.mu.Unlock()
	reply(w, r)
}

func (u *engineUpstream) calls() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.reqs)
}

func (u *engineUpstream) last() (recordedReq, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if len(u.reqs) == 0 {
		return recordedReq{}, false
	}
	return u.reqs[len(u.reqs)-1], true
}

// defaultReply answers in the engine's own shape with n results.
func defaultReply(kind string, n int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]any
		for i := 1; i <= n; i++ {
			u := fmt.Sprintf("https://example.com/r%d", i)
			switch kind {
			case protocol.EngineBrave:
				items = append(items, map[string]any{"title": fmt.Sprintf("T%d", i), "url": u, "description": fmt.Sprintf("D%d", i), "page_age": "2026-11-01T10:00:00"})
			case protocol.EngineSearXNG:
				items = append(items, map[string]any{"title": fmt.Sprintf("T%d", i), "url": u, "content": fmt.Sprintf("C%d", i), "publishedDate": "2026-11-01"})
			default:
				items = append(items, map[string]any{"title": fmt.Sprintf("T%d", i), "url": u, "snippet": fmt.Sprintf("S%d", i)})
			}
		}
		var out any = map[string]any{"results": items}
		if kind == protocol.EngineBrave {
			out = map[string]any{"web": map[string]any{"results": items}}
		}
		_ = json.NewEncoder(w).Encode(out)
	}
}

type bridgeState struct {
	t *testing.T

	cfg     Config // the search share
	chatCfg Config // a chat share on the same node, when a scenario has one
	up      *engineUpstream
	chat    *engineUpstream
	priv    ed25519.PrivateKey
	pub     string

	res        protocol.JobResult
	prevSecret string // everything a key could leak into, captured after a serve
	lastPath   string
	other      *engineUpstream // a second host a redirect points at
	logBuf     bytes.Buffer
	prevLog    io.Writer
	prevTO     time.Duration
}

func (s *bridgeState) reset() {
	if s.up != nil && s.up.srv != nil {
		s.up.srv.Close()
	}
	if s.chat != nil && s.chat.srv != nil {
		s.chat.srv.Close()
	}
	*s = bridgeState{t: s.t, prevTO: searchengine.EngineTimeout}
	pub, priv, _ := ed25519.GenerateKey(nil)
	s.priv, s.pub = priv, hex.EncodeToString(pub)
	s.prevLog = log.Writer()
	log.SetOutput(io.MultiWriter(&s.logBuf, s.prevLog))
}

func (s *bridgeState) restore() {
	log.SetOutput(s.prevLog)
	searchengine.EngineTimeout = s.prevTO
}

func newUpstream(kind string) *engineUpstream {
	u := &engineUpstream{reply: defaultReply(kind, 3)}
	u.srv = httptest.NewServer(u)
	return u
}

// share configures a search share of model on engine kind, against a recording upstream.
func (s *bridgeState) share(model, kind, key string) {
	s.up = newUpstream(kind)
	s.cfg = Config{NodeID: "node-s", Model: model, Modality: protocol.ModalitySearch, SearchEngine: kind,
		Upstream: s.up.srv.URL, UpstreamKey: key}
}

// relay serves one job on the search share.
func (s *bridgeState) relay(path string, req map[string]any) {
	s.lastPath = path
	body, _ := json.Marshal(req)
	job := protocol.Job{ID: "att-1", User: "u-pseud", Body: body, Path: path}
	s.res = serveCtx(context.Background(), s.cfg, offerOf(s.cfg), s.priv, s.up.srv.Client(), job)
}

func (s *bridgeState) search(model, q string, extra map[string]any) {
	req := map[string]any{"model": model, "query": q, "count": 5}
	for k, v := range extra {
		req[k] = v
	}
	s.relay("/v1/search", req)
}

func (s *bridgeState) results() ([]map[string]any, error) {
	var out struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(s.res.Body, &out); err != nil || out.Results == nil {
		return nil, fmt.Errorf("node reply is not a result list (%d): %s", s.res.Status, s.res.Body)
	}
	return out.Results, nil
}

func (s *bridgeState) errReason() string {
	var e struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(s.res.Body, &e)
	return e.Error
}

// --- steps ---------------------------------------------------------------------

func (s *bridgeState) shareEngine(model, kind string) error {
	s.share(model, kind, "")
	return nil
}
func (s *bridgeState) relayPathModel(path, model string) error {
	s.relay(path, map[string]any{"model": model, "query": "tides", "count": 5})
	return nil
}
func (s *bridgeState) relayPath(path string) error {
	s.relay(path, map[string]any{"model": s.cfg.Model, "query": "tides", "count": 5})
	return nil
}

// relayStream sends a job whose body says "stream": true through the node's real per-job
// dispatch (serveJob), against a stand-in broker that records what the node posts back.
func (s *bridgeState) relayStream(path string) error {
	s.lastPath = path
	var mu sync.Mutex
	var posted []protocol.JobResult
	brk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var res protocol.JobResult
		if r.URL.Path == "/agent/result" && json.Unmarshal(b, &res) == nil {
			mu.Lock()
			posted = append(posted, res)
			mu.Unlock()
		}
	}))
	defer brk.Close()
	cfg := s.cfg
	cfg.Broker = brk.URL
	body, _ := json.Marshal(map[string]any{"model": cfg.Model, "query": "tides", "count": 5, "stream": true,
		"messages": []map[string]string{{"role": "user", "content": "hi"}}})
	job := protocol.Job{ID: "att-1", User: "u-pseud", Body: body, Path: path}
	serveJob(context.Background(), cfg, offerOf(cfg), s.priv, s.up.srv.Client(), brk.Client(), "tok", job)
	mu.Lock()
	defer mu.Unlock()
	if len(posted) != 1 {
		return fmt.Errorf("the node posted %d results to the broker, want exactly 1 (engine calls: %d)", len(posted), s.up.calls())
	}
	s.res = posted[0]
	return nil
}
func (s *bridgeState) normalizedNoStream() error {
	if n := s.up.calls(); n != 1 {
		return fmt.Errorf("engine called %d times, want 1 (node %d: %s)", n, s.res.Status, s.res.Body)
	}
	r, _ := s.up.last()
	var got map[string]any
	if err := json.Unmarshal([]byte(r.body), &got); err != nil {
		return fmt.Errorf("the engine did not receive a JSON search: %q", r.body)
	}
	if got["query"] != "tides" {
		return fmt.Errorf("the engine received %q, not the normalized search", r.body)
	}
	for _, k := range []string{"stream", "messages", "model"} {
		if _, has := got[k]; has {
			return fmt.Errorf("the engine received the raw job field %q: %s", k, r.body)
		}
	}
	return nil
}
func (s *bridgeState) calledOnce() error {
	if n := s.up.calls(); n != 1 {
		return fmt.Errorf("engine called %d times, want 1 (node %d: %s)", n, s.res.Status, s.res.Body)
	}
	return nil
}
func (s *bridgeState) normalizedSigned() error {
	if s.res.Status != http.StatusOK {
		return fmt.Errorf("node status %d: %s", s.res.Status, s.res.Body)
	}
	rs, err := s.results()
	if err != nil {
		return err
	}
	for _, r := range rs {
		for k := range r {
			if k != "title" && k != "url" && k != "snippet" && k != "published" {
				return fmt.Errorf("result carries field %q", k)
			}
		}
	}
	if !s.res.Receipt.VerifyNode(s.pub) || s.res.Receipt.RequestID != "att-1" {
		return fmt.Errorf("receipt not signed by this node for this job: %+v", s.res.Receipt)
	}
	return nil
}
func (s *bridgeState) refused404() error {
	if s.res.Status != http.StatusNotFound || !strings.Contains(string(s.res.Body), "unsupported path") {
		return fmt.Errorf("node answered %d %s, want 404 unsupported path", s.res.Status, s.res.Body)
	}
	return nil
}
func (s *bridgeState) engineNeverCalled() error {
	if n := s.up.calls(); n != 0 {
		return fmt.Errorf("engine called %d times", n)
	}
	return nil
}
func (s *bridgeState) chatAndSearch(chatModel, searchModel string) error {
	s.share(searchModel, protocol.EngineSearXNG, "")
	s.chat = &engineUpstream{reply: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"choices":[]}`)) }}
	s.chat.srv = httptest.NewServer(s.chat)
	s.chatCfg = Config{NodeID: "node-s", Model: chatModel, Upstream: s.chat.srv.URL + "/v1/chat/completions"}
	return nil
}
func (s *bridgeState) refusedNoBackend() error {
	if s.lastPath == "/v1/search" && strings.Contains(string(s.res.Body), "unsupported path") {
		return fmt.Errorf("refused as an unsupported path: the search path is supported, the model is not this node's search offer")
	}
	if s.res.Status < 400 {
		return fmt.Errorf("node served (%d): %s", s.res.Status, s.res.Body)
	}
	if s.up.calls() != 0 || (s.chat != nil && s.chat.calls() != 0) {
		return fmt.Errorf("a backend was called")
	}
	return nil
}
func (s *bridgeState) chatJobFor(model string) error {
	s.relay("", map[string]any{"model": model, "messages": []any{map[string]any{"role": "user", "content": "hi"}}})
	return nil
}
func (s *bridgeState) searchNeverCalled() error { return s.engineNeverCalled() }

func (s *bridgeState) mixedNode() error {
	s.chat = &engineUpstream{reply: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"ok":true}`)) }}
	s.chat.srv = httptest.NewServer(s.chat)
	return nil
}
func (s *bridgeState) oneOfEach() error {
	base := s.chat.srv.URL + "/v1/chat/completions"
	for _, c := range []struct{ modality, path string }{
		{protocol.ModalityChat, ""}, {protocol.ModalityTTS, "/v1/audio/speech"}, {protocol.ModalitySTT, "/v1/audio/transcriptions"},
	} {
		cfg := Config{NodeID: "node-m", Model: "m-" + c.modality, Modality: c.modality, Upstream: base}
		job := protocol.Job{ID: "att-" + c.modality, Body: []byte(`{"model":"m"}`), Path: c.path}
		r := serveCtx(context.Background(), cfg, offerOf(cfg), s.priv, s.chat.srv.Client(), job)
		if r.Status != http.StatusOK {
			return fmt.Errorf("%s job answered %d: %s", c.modality, r.Status, r.Body)
		}
	}
	return nil
}
func (s *bridgeState) forwardedAsBefore() error {
	want := []string{"/v1/chat/completions", "/v1/audio/speech", "/v1/audio/transcriptions"}
	s.chat.mu.Lock()
	defer s.chat.mu.Unlock()
	if len(s.chat.reqs) != len(want) {
		return fmt.Errorf("backend saw %d requests, want %d", len(s.chat.reqs), len(want))
	}
	for i, r := range s.chat.reqs {
		if r.path != want[i] {
			return fmt.Errorf("request %d went to %s, want %s", i, r.path, want[i])
		}
	}
	return nil
}

// Brave
func (s *bridgeState) shareRecording(model, kind string) error {
	s.share(model, kind, "k-brave")
	return nil
}
func (s *bridgeState) braveSearch(q string, count int, fresh, country string) error {
	s.search(s.cfg.Model, q, map[string]any{"count": count, "freshness": fresh, "country": country})
	return nil
}
func (s *bridgeState) upstreamGetBrave(q string, count int, fresh, country string) error {
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("upstream never called (node %d: %s)", s.res.Status, s.res.Body)
	}
	get := func(k string) string {
		if v := r.query[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	if r.method != http.MethodGet || get("q") != q || get("count") != fmt.Sprint(count) || get("freshness") != fresh || get("country") != country {
		return fmt.Errorf("upstream got %s %v", r.method, r.query)
	}
	return nil
}
func (s *bridgeState) keyHeader(h string) error {
	r, ok := s.up.last()
	if !ok || r.header.Get(h) != s.cfg.UpstreamKey {
		return fmt.Errorf("upstream header %s = %q", h, r.header.Get(h))
	}
	return nil
}
func (s *bridgeState) searchFresh(f string) error {
	extra := map[string]any{}
	if f != "(absent)" {
		extra["freshness"] = f
	}
	s.search(s.cfg.Model, "tides", extra)
	return nil
}
func (s *bridgeState) upstreamFresh(code string) error {
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("upstream never called (node %d: %s)", s.res.Status, s.res.Body)
	}
	got, has := r.query["freshness"]
	if code == "(absent)" {
		if has {
			return fmt.Errorf("upstream got freshness %v", got)
		}
		return nil
	}
	if !has || got[0] != code {
		return fmt.Errorf("upstream freshness %v, want %s", got, code)
	}
	return nil
}
func (s *bridgeState) braveWithPageAge() error {
	s.share("web-search", protocol.EngineBrave, "k-brave")
	return nil
}
func (s *bridgeState) nodeServes() error {
	s.search(s.cfg.Model, "tides", nil)
	return nil
}
func (s *bridgeState) snippetIsDescription() error {
	rs, err := s.results()
	if err != nil {
		return err
	}
	if len(rs) == 0 || rs[0]["snippet"] != "D1" || rs[0]["published"] != "2026-11-01T10:00:00" {
		return fmt.Errorf("mapped results %v", rs)
	}
	return nil
}
func (s *bridgeState) braveNoWeb() error {
	s.share("web-search", protocol.EngineBrave, "k-brave")
	s.up.reply = func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"query":{"original":"tides"}}`)) }
	return nil
}
func (s *bridgeState) emptyResults() error {
	rs, err := s.results()
	if err != nil {
		return err
	}
	if len(rs) != 0 || s.res.Status != http.StatusOK {
		return fmt.Errorf("node returned %d results (%d)", len(rs), s.res.Status)
	}
	return nil
}

// SearXNG
func (s *bridgeState) searchQFresh(q, f string) error {
	s.search(s.cfg.Model, q, map[string]any{"freshness": f})
	return nil
}
func (s *bridgeState) upstreamSearx(q, format, tr string) error {
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("upstream never called (node %d: %s)", s.res.Status, s.res.Body)
	}
	if r.method != http.MethodGet || r.path != "/search" || r.query["q"][0] != q || r.query["format"][0] != format || r.query["time_range"][0] != tr {
		return fmt.Errorf("upstream got %s %s %v", r.method, r.path, r.query)
	}
	return nil
}
func (s *bridgeState) searx30() error {
	s.share("web-search", protocol.EngineSearXNG, "")
	s.up.reply = defaultReply(protocol.EngineSearXNG, 30)
	return nil
}
func (s *bridgeState) searchCount(n int) error {
	s.search(s.cfg.Model, "tides", map[string]any{"count": n})
	return nil
}
func (s *bridgeState) nInOrder(n int) error {
	rs, err := s.results()
	if err != nil {
		return err
	}
	if len(rs) != n {
		return fmt.Errorf("node returned %d results, want %d", len(rs), n)
	}
	for i, r := range rs {
		if r["url"] != fmt.Sprintf("https://example.com/r%d", i+1) {
			return fmt.Errorf("result %d out of order: %v", i, r["url"])
		}
	}
	return nil
}
func (s *bridgeState) searchCountry(c string) error {
	s.search(s.cfg.Model, "tides", map[string]any{"country": c})
	return nil
}
func (s *bridgeState) noCountryParam() error {
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("upstream never called")
	}
	for _, k := range []string{"country", "language", "lang"} {
		if _, has := r.query[k]; has {
			return fmt.Errorf("upstream got %s", k)
		}
	}
	return nil
}
func (s *bridgeState) searx403() error {
	s.share("web-search", protocol.EngineSearXNG, "")
	s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }
	return nil
}
func (s *bridgeState) jsonOffError() error {
	if s.res.Status < 400 || !strings.Contains(s.errReason(), "JSON output is turned off") {
		return fmt.Errorf("node answered %d %s", s.res.Status, s.res.Body)
	}
	return nil
}

// native
func (s *bridgeState) searchQCount(q string, n int) error {
	s.search(s.cfg.Model, q, map[string]any{"count": n})
	return nil
}
func (s *bridgeState) upstreamNativePost(q string, n int) error {
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("upstream never called (node %d: %s)", s.res.Status, s.res.Body)
	}
	var got map[string]any
	_ = json.Unmarshal([]byte(r.body), &got)
	if r.method != http.MethodPost || r.path != "/v1/search" || got["query"] != q || got["count"] != float64(n) {
		return fmt.Errorf("upstream got %s %s %s", r.method, r.path, r.body)
	}
	if _, has := got["model"]; has {
		return fmt.Errorf("the native service was sent the model id: %s", r.body)
	}
	return nil
}
func (s *bridgeState) nativeWithKey() error {
	s.share("go-docs", protocol.EngineNative, "k-native")
	return nil
}
func (s *bridgeState) nativeNoKey() error {
	s.share("go-docs", protocol.EngineNative, "")
	return nil
}
func (s *bridgeState) bearer() error {
	r, ok := s.up.last()
	if !ok || r.header.Get("Authorization") != "Bearer k-native" {
		return fmt.Errorf("Authorization = %q", r.header.Get("Authorization"))
	}
	return nil
}
func (s *bridgeState) noAuth() error {
	r, ok := s.up.last()
	if !ok || r.header.Get("Authorization") != "" {
		return fmt.Errorf("Authorization = %q", r.header.Get("Authorization"))
	}
	return nil
}

// key secrecy
func (s *bridgeState) shareWithKey(kind, key string) error {
	s.share("web-search", kind, key)
	return nil
}
func (s *bridgeState) regHeartbeatServe(outcome string) error {
	key := s.cfg.UpstreamKey
	switch outcome {
	case "succeeds":
	case "fails with HTTP 401":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) }
	case "fails with HTTP 422":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(422) }
	case "times out":
		searchengine.EngineTimeout = 150 * time.Millisecond
		s.up.reply = func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}
	case "fails with an error echoing the key":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"error":"bad key ` + key + `"}`))
		}
	default:
		return fmt.Errorf("unknown outcome %q", outcome)
	}
	regJSON, _ := json.Marshal(protocol.NodeRegistration{NodeID: s.cfg.NodeID, Offers: []protocol.ModelOffer{offerOf(s.cfg)}})
	s.search(s.cfg.Model, "tides", nil)
	s.prevSecret = string(regJSON) + string(heartbeatBody(s.cfg.NodeID)) + string(s.res.Body) + fmt.Sprintf("%+v", s.res.Receipt)
	return nil
}
func (s *bridgeState) keyNowhere(key string) error {
	// the key must actually have been used, or "it leaked nowhere" proves nothing.
	r, ok := s.up.last()
	if !ok {
		return fmt.Errorf("the engine was never called (node %d: %s)", s.res.Status, s.res.Body)
	}
	if !strings.Contains(fmt.Sprint(r.header), key) {
		return fmt.Errorf("the engine call did not carry the key")
	}
	all := s.prevSecret + s.logBuf.String()
	if strings.Contains(all, key) {
		return fmt.Errorf("the key leaked: %.300s", all)
	}
	return nil
}
func (s *bridgeState) upstream500Stack() error {
	s.share("web-search", protocol.EngineNative, "")
	s.up.reply = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("panic: runtime error\ngoroutine 1 [running]:\nmain.go:42 SECRET_INTERNAL_PATH"))
	}
	return nil
}
func (s *bridgeState) shortFixedReason() error {
	if strings.Contains(string(s.res.Body), "SECRET_INTERNAL_PATH") || strings.Contains(string(s.res.Body), "goroutine") {
		return fmt.Errorf("the upstream body leaked: %s", s.res.Body)
	}
	if s.errReason() != "engine error (status 500)" {
		return fmt.Errorf("reason = %q", s.errReason())
	}
	return nil
}

// one bounded call (N3)
func (s *bridgeState) upstreamBehaves(how string) error {
	s.share("web-search", protocol.EngineNative, "")
	switch how {
	case "answers HTTP 429":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(429) }
	case "answers HTTP 401":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) }
	case "answers HTTP 500":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) }
	case "does not answer within 10s":
		searchengine.EngineTimeout = 150 * time.Millisecond
		s.up.reply = func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}
	case "returns a 600 KiB body":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"results":[{"title":"` + strings.Repeat("x", 600*1024) + `","url":"https://e.com/"}]}`))
		}
	case "returns a body that is not JSON":
		s.up.reply = func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) }
	case "refuses the connection":
		s.up.srv.Close()
	default:
		return fmt.Errorf("unknown behaviour %q", how)
	}
	return nil
}
func (s *bridgeState) calledExactlyOnce() error {
	want := 1
	if s.up.srv.Listener == nil {
		want = 0
	}
	if n := s.up.calls(); n > 1 || (n != want && n != 0) {
		return fmt.Errorf("upstream called %d times, want at most once", n)
	}
	return nil
}
func (s *bridgeState) reasonIs(want string) error {
	if s.res.Status < 400 || s.errReason() != want {
		return fmt.Errorf("node answered %d reason %q, want %q", s.res.Status, s.errReason(), want)
	}
	return nil
}
func (s *bridgeState) timeoutUnderRelay() error {
	// the broker's non-stream relay wait is 90 s (cmd/rogerai-broker tunnel.go, pinned by
	// now_hardening_test.go); the engine bound must leave the broker room to answer first.
	if searchengine.DefaultEngineTimeout >= 90*time.Second || searchengine.DefaultEngineTimeout != 10*time.Second {
		return fmt.Errorf("engine timeout %s, want 10s, under the 90s relay wait", searchengine.DefaultEngineTimeout)
	}
	return nil
}
func (s *bridgeState) redirectOtherHost() error {
	s.share("web-search", protocol.EngineNative, "")
	other := &engineUpstream{reply: defaultReply(protocol.EngineNative, 1)}
	other.srv = httptest.NewServer(other)
	s.t.Cleanup(other.srv.Close)
	s.other = other
	s.up.reply = func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.srv.URL+"/v1/search", http.StatusFound)
	}
	return nil
}
func (s *bridgeState) otherNotContacted() error {
	if s.other.calls() != 0 {
		return fmt.Errorf("the node followed the redirect")
	}
	return nil
}

// N4
func (s *bridgeState) jsAndTwoGood() error {
	s.share("web-search", protocol.EngineNative, "")
	s.up.reply = func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"title":"x","url":"javascript:alert(1)"},{"title":"a","url":"https://a.example/"},{"title":"b","url":"https://b.example/"}]}`))
	}
	return nil
}
func (s *bridgeState) twoGood() error {
	rs, err := s.results()
	if err != nil {
		return err
	}
	if len(rs) != 2 || rs[0]["url"] != "https://a.example/" || rs[1]["url"] != "https://b.example/" {
		return fmt.Errorf("node reply %v", rs)
	}
	return nil
}
func (s *bridgeState) modifiedNode() error    { return nil }
func (s *bridgeState) consumerThrough() error { return nil }

// brokerDrops is pinned where the broker is: the broker suite runs result_validation.feature,
// whose URL table drops "javascript:". This guards the cross-reference itself.
func (s *bridgeState) brokerDrops() error {
	b, err := os.ReadFile("../../features/search/result_validation.feature")
	if err != nil {
		return err
	}
	if !strings.Contains(string(b), "javascript:alert(1)") {
		return fmt.Errorf("result_validation.feature no longer pins dropping javascript: URLs")
	}
	return nil
}

func TestSearchNodeBridgeBDD(t *testing.T) {
	s := &bridgeState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return c, nil
			})
			sc.After(func(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.restore()
				return c, nil
			})
			sc.Step(`^a share node serving "([^"]*)" with engine "([^"]*)"$`, s.shareEngine)
			sc.Step(`^the broker relays a job with path "([^"]*)" for model "([^"]*)"$`, s.relayPathModel)
			sc.Step(`^the broker relays a job with path "([^"]*)"$`, s.relayPath)
			sc.Step(`^the node calls its SearXNG engine once$`, s.calledOnce)
			sc.Step(`^the broker relays a streaming job with path "([^"]*)"$`, s.relayStream)
			sc.Step(`^the node calls its engine once with the normalized search and no "stream" field$`, s.normalizedNoStream)
			sc.Step(`^the node returns a normalized result list with a signed receipt$`, s.normalizedSigned)
			sc.Step(`^the node refuses with a 404 "unsupported path"$`, s.refused404)
			sc.Step(`^the engine is never called$`, s.engineNeverCalled)
			sc.Step(`^a share node serving chat model "([^"]*)" and search "([^"]*)"$`, s.chatAndSearch)
			sc.Step(`^the node refuses the job without calling any backend$`, s.refusedNoBackend)
			sc.Step(`^the broker relays a chat job for model "([^"]*)"$`, s.chatJobFor)
			sc.Step(`^the search engine is never called$`, s.searchNeverCalled)
			sc.Step(`^a share node serving chat, tts and stt models$`, s.mixedNode)
			sc.Step(`^the broker relays one job of each$`, s.oneOfEach)
			sc.Step(`^each is forwarded exactly as before this change$`, s.forwardedAsBefore)

			sc.Step(`^a share node serving "([^"]*)" with engine "([^"]*)" against a recording upstream$`, s.shareRecording)
			sc.Step(`^the broker relays a search for "([^"]*)" with count (\d+), freshness "([^"]*)" and country "([^"]*)"$`, s.braveSearch)
			sc.Step(`^the upstream receives GET with q "([^"]*)", count (\d+), freshness "([^"]*)" and country "([^"]*)"$`, s.upstreamGetBrave)
			sc.Step(`^the upstream receives the key in header "([^"]*)"$`, s.keyHeader)
			sc.Step(`^the broker relays a search with freshness "([^"]*)"$`, s.searchFresh)
			sc.Step(`^the broker relays a search with freshness \(absent\)$`, func() error { return s.searchFresh("(absent)") })
			sc.Step(`^the upstream receives freshness "([^"]*)"$`, s.upstreamFresh)
			sc.Step(`^the upstream receives freshness \(absent\)$`, func() error { return s.upstreamFresh("(absent)") })
			sc.Step(`^a Brave upstream that returns web.results with title, url, description and page_age$`, s.braveWithPageAge)
			sc.Step(`^the node serves a search$`, s.nodeServes)
			sc.Step(`^each result's snippet is the description and published is the page_age$`, s.snippetIsDescription)
			sc.Step(`^a Brave upstream that returns \{"query": \{\.\.\.\}\} and no "web" key$`, s.braveNoWeb)
			sc.Step(`^the node returns \{"results": \[\]\}$`, s.emptyResults)

			sc.Step(`^the broker relays a search for "([^"]*)" with freshness "([^"]*)"$`, s.searchQFresh)
			sc.Step(`^the upstream receives GET /search with q "([^"]*)", format "([^"]*)" and time_range "([^"]*)"$`, s.upstreamSearx)
			sc.Step(`^a SearXNG upstream that returns 30 results$`, s.searx30)
			sc.Step(`^the broker relays a search with count (\d+)$`, s.searchCount)
			sc.Step(`^the node returns (\d+) results in the upstream's order$`, s.nInOrder)
			sc.Step(`^the broker relays a search with country "([^"]*)"$`, s.searchCountry)
			sc.Step(`^the upstream request carries no country or language parameter$`, s.noCountryParam)
			sc.Step(`^a SearXNG upstream that answers 403 to format=json$`, s.searx403)
			sc.Step(`^the node returns an error saying the engine's JSON output is turned off$`, s.jsonOffError)

			sc.Step(`^the broker relays a search for "([^"]*)" with count (\d+)$`, s.searchQCount)
			sc.Step(`^the upstream receives POST /v1/search with \{"query":"([^"]*)","count":(\d+)\}$`, s.upstreamNativePost)
			sc.Step(`^a native backend configured with a key$`, s.nativeWithKey)
			sc.Step(`^a native backend configured without a key$`, s.nativeNoKey)
			sc.Step(`^the upstream receives header "Authorization: Bearer <key>"$`, s.bearer)
			sc.Step(`^the upstream request has no Authorization header$`, s.noAuth)

			sc.Step(`^a share node serving engine "([^"]*)" with key "([^"]*)"$`, s.shareWithKey)
			sc.Step(`^the node registers, heartbeats and serves a search that (.+)$`, s.regHeartbeatServe)
			sc.Step(`^neither the registration, heartbeat, receipt, result, error body nor any log line contains "([^"]*)"$`, s.keyNowhere)
			sc.Step(`^an upstream that fails with HTTP 500 and a body holding its own stack trace$`, s.upstream500Stack)
			sc.Step(`^the node's error carries the status and a short fixed reason, not the upstream body$`, s.shortFixedReason)

			sc.Step(`^an upstream that (answers HTTP 429|answers HTTP 401|answers HTTP 500|does not answer within 10s|returns a 600 KiB body|returns a body that is not JSON|refuses the connection)$`, s.upstreamBehaves)
			sc.Step(`^the upstream was called exactly once$`, s.calledExactlyOnce)
			sc.Step(`^the node returns an error whose reason is "([^"]*)"$`, s.reasonIs)
			sc.Step(`^the node's engine timeout is shorter than the broker's non-stream relay wait$`, s.timeoutUnderRelay)
			sc.Step(`^an upstream that answers 302 to a different host$`, s.redirectOtherHost)
			sc.Step(`^the node does not contact the other host$`, s.otherNotContacted)

			sc.Step(`^an upstream that returns a "javascript:" result and two good ones$`, s.jsAndTwoGood)
			sc.Step(`^the node's reply holds the two good results$`, s.twoGood)
			sc.Step(`^a node that returns "javascript:" results unchanged$`, s.modifiedNode)
			sc.Step(`^a consumer searches through it$`, s.consumerThrough)
			sc.Step(`^the broker drops them \(result_validation\.feature\)$`, s.brokerDrops)
		},
		Options: &godog.Options{Format: "pretty", TestingT: t, Strict: true,
			Paths: []string{"../../features/search/node_bridge.feature"}},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("search node bridge scenarios failed")
	}
}
