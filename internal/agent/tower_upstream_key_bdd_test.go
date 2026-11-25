package agent

// Contract: features/tower/tower_upstream_key.feature.
//
// The runner drives the REAL serving paths a share uses: ServeTower against a Core and a hub
// that mint real grants and seal real envelopes (so the consumer's bytes can be opened and
// read, which is where a leak on the sealed path would land), Start against a classic broker,
// and runLocalJob for the local Tower. The only stand-ins are the two HTTP servers a node talks
// to and the operator's upstream, each recording everything it is sent.
//
// Where a key came FROM (flag, saved config, onboarding prompt) is decided in cmd/rogerai
// before the agent sees it (share_upstream_persist_test.go, onboard_cov_test.go,
// share_stale_key_test.go cover that resolution); here every source arrives the same way, as
// Config.UpstreamKey, which is the property the rows need: the key reaches the upstream on
// every serving path whichever source supplied it.

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/station"
	"rogerai.fm/roger/v7/internal/towercore/dispatch"
	"rogerai.fm/roger/v7/internal/towercore/envelope"
	"rogerai.fm/roger/v7/internal/towercore/link"
	"rogerai.fm/roger/v7/internal/towerhub"
)

const ukKey = "sk-test-0123456789abcdefKEYZ"

// ukSent is one request a node sent to Core, the hub or the classic broker.
type ukSent struct {
	path   string
	header http.Header
	body   []byte
}

type ukState struct {
	t *testing.T

	key         string // what the share was started with
	requireK    bool   // the upstream refuses a request without "Bearer K"
	echo        string // where the upstream puts K
	upAuth      []string
	upMu        sync.Mutex
	upstream    *httptest.Server
	startErr    error
	sentMu      sync.Mutex
	sent        []ukSent
	consumer    []byte // the consumer's opened bytes (joined Tower path)
	receipt     []byte
	failure     string
	notices     []string
	transcript  []byte // the response bytes the audit answer carried
	coreEnvPriv []byte
	// what the receipt is verified against
	assertion ed25519.PublicKey
	attempt   string
	stationID string
	jobs      int    // how many jobs the hub hands the node (default 1)
	served    int    // upstream calls so far, for per-request bodies
	msgRow    string // the control-message row under test
}

func (s *ukState) record(r *http.Request) []byte {
	b, _ := io.ReadAll(r.Body)
	s.sentMu.Lock()
	s.sent = append(s.sent, ukSent{path: r.URL.Path, header: r.Header.Clone(), body: b})
	s.sentMu.Unlock()
	return b
}

func (s *ukState) notice(err error) {
	s.sentMu.Lock()
	s.notices = append(s.notices, err.Error())
	s.sentMu.Unlock()
}

// --- the upstream --------------------------------------------------------------------

func jsonEscapedAll(k string) string {
	var b strings.Builder
	for _, c := range []byte(k) {
		fmt.Fprintf(&b, `\u%04x`, c)
	}
	return b.String()
}

func (s *ukState) startUpstream() {
	s.upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		auth := r.Header.Get("Authorization")
		s.upMu.Lock()
		s.upAuth = append(s.upAuth, auth)
		s.upMu.Unlock()
		if s.requireK && auth != "Bearer "+ukKey {
			http.Error(w, `{"error":"missing or wrong api key"}`, http.StatusUnauthorized)
			return
		}
		ok := `{"choices":[{"message":{"content":"hi"}}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`
		switch s.echo {
		case "a 200 response body":
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"your key is ` + ukKey + `"}}],"usage":{"prompt_tokens":3,"completion_tokens":5}}`))
		case "a streaming chunk":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"" + ukKey + "\"}}]}\n\ndata: [DONE]\n\n"))
		case "two streaming chunks, split in the middle of K":
			// K crosses two network writes: the first ends inside K, the second starts with
			// the rest of it.
			w.Header().Set("Content-Type", "text/event-stream")
			half := len(ukKey) / 2
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"" + ukKey[:half]))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(20 * time.Millisecond)
			_, _ = w.Write([]byte(ukKey[half:] + "\"}}]}\n\ndata: [DONE]\n\n"))
		case "a 401 error body":
			http.Error(w, `{"error":"key `+ukKey+` is not valid"}`, http.StatusUnauthorized)
		case "a 500 error body":
			http.Error(w, `{"error":"internal: auth=`+ukKey+`"}`, http.StatusInternalServerError)
		case "an error body longer than 200 bytes, with K across byte 200":
			pad := strings.Repeat("x", 200-len(`{"error":"`)-len(ukKey)/2)
			http.Error(w, `{"error":"`+pad+ukKey+strings.Repeat("y", 60)+`"}`, http.StatusInternalServerError)
		case `a JSON string, with K's characters escaped as \u00XX`:
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"` + jsonEscapedAll(ukKey) + `"}}]}`))
		case "rejects K":
			// A real provider's refusal differs per request (an id, a timestamp).
			s.upMu.Lock()
			s.served++
			n := s.served
			s.upMu.Unlock()
			http.Error(w, fmt.Sprintf(`{"error":"invalid api key %s","request_id":"req-%d-%d"}`, ukKey, n, time.Now().UnixNano()),
				http.StatusUnauthorized)
		default:
			_, _ = w.Write([]byte(ok))
		}
	}))
	s.t.Cleanup(s.upstream.Close)
}

// --- the joined Tower rig --------------------------------------------------------------

// serveThroughTower runs ServeTower against a Core and hub that hand the node ONE real,
// Core-signed, sealed job, and waits for its completion (and, when wantAudit, for the audit
// answer of that attempt too).
func (s *ukState) serveThroughTower(wantAudit bool) error {
	corePub, corePriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	coreEnvPub, coreEnvPriv, err := envelope.NewKey()
	if err != nil {
		return err
	}
	s.coreEnvPriv = coreEnvPriv
	consumerEnvPub, consumerEnvPriv, err := envelope.NewKey()
	if err != nil {
		return err
	}
	consumerPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	reg := dispatch.NewWithStore(dispatch.Config{Network: link.PublicNetwork, Signer: corePriv, Lifetime: time.Minute}, nil)

	var mu sync.Mutex
	var stationID, assertionHex, sessionHex string
	var attemptID string
	handed, handedN, completions := false, 0, 0
	jobs := s.jobs
	if jobs <= 0 {
		jobs = 1
	}
	completed := make(chan struct{}, 1)
	audited := make(chan struct{}, 1)

	hubMux := http.NewServeMux()
	hubMux.HandleFunc(towerhub.PathPoll, func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		mu.Lock()
		defer mu.Unlock()
		if handed || stationID == "" {
			mu.Unlock()
			time.Sleep(30 * time.Millisecond)
			mu.Lock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		assertion, _ := hex.DecodeString(assertionHex)
		session, _ := hex.DecodeString(sessionHex)
		g, gerr := reg.MintEdge(dispatch.EdgeTarget{
			TowerID: "tw-uk", StationID: stationID, Model: "m", Modality: "chat",
			RelayName: stationID + ".relay.example", MaxIn: 1 << 16, MaxOut: 1 << 16,
			MaxTokIn: 1 << 16, MaxTokOut: 1 << 16,
			AssertionKey: ed25519.PublicKey(assertion), ConsumerKey: consumerPub,
			ConsumerEnvKey: consumerEnvPub,
		})
		if gerr != nil {
			http.Error(w, gerr.Error(), http.StatusInternalServerError)
			return
		}
		sealed, serr := envelope.SealTo(session, []byte(`{"model":"m","messages":[{"role":"user","content":"hi"}]}`), g.AttemptID)
		if serr != nil {
			http.Error(w, serr.Error(), http.StatusInternalServerError)
			return
		}
		raw, _ := sealed.Marshal()
		handedN++
		handed, attemptID = handedN >= jobs, g.AttemptID
		s.assertion, s.attempt, s.stationID = ed25519.PublicKey(assertion), g.AttemptID, stationID
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"attempt_id": g.AttemptID, "station_id": stationID,
			"grant":    base64.StdEncoding.EncodeToString(g.Signed),
			"envelope": base64.StdEncoding.EncodeToString(raw),
		})
	})
	hubMux.HandleFunc(towerhub.PathComplete, func(w http.ResponseWriter, r *http.Request) {
		b := s.record(r)
		var c struct {
			Envelope string `json:"envelope"`
			Receipt  string `json:"receipt"`
			Failure  string `json:"failure"`
		}
		_ = json.Unmarshal(b, &c)
		if c.Envelope != "" {
			raw, _ := base64.StdEncoding.DecodeString(c.Envelope)
			if sealed, perr := envelope.Parse(raw); perr == nil {
				mu.Lock()
				id := attemptID
				mu.Unlock()
				s.consumer, _ = envelope.OpenWith(consumerEnvPriv, sealed, id)
			}
		}
		s.receipt, _ = base64.StdEncoding.DecodeString(c.Receipt)
		s.failure = c.Failure
		w.WriteHeader(http.StatusOK)
		mu.Lock()
		completions++
		all := completions >= jobs
		mu.Unlock()
		if all {
			select {
			case completed <- struct{}{}:
			default:
			}
		}
	})
	hubMux.HandleFunc(towerhub.PathAuditWanted, func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		mu.Lock()
		id := attemptID
		mu.Unlock()
		wanted := []string{}
		if id != "" && wantAudit {
			wanted = append(wanted, id)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"wanted": wanted})
	})
	hubMux.HandleFunc(towerhub.PathAuditTranscript, func(w http.ResponseWriter, r *http.Request) {
		b := s.record(r)
		var reply struct {
			AttemptID    string `json:"attempt_id"`
			SealedBundle string `json:"sealed_bundle"`
		}
		_ = json.Unmarshal(b, &reply)
		raw, _ := base64.StdEncoding.DecodeString(reply.SealedBundle)
		if sealed, perr := envelope.Parse(raw); perr == nil {
			if bundle, oerr := envelope.OpenWith(coreEnvPriv, sealed, reply.AttemptID); oerr == nil {
				var bm map[string]string
				_ = json.Unmarshal(bundle, &bm)
				s.transcript, _ = base64.StdEncoding.DecodeString(bm["response"])
			}
		}
		w.WriteHeader(http.StatusOK)
		select {
		case audited <- struct{}{}:
		default:
		}
	})
	hubMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		w.WriteHeader(http.StatusNoContent)
	})
	hub := httptest.NewServer(hubMux)
	s.t.Cleanup(hub.Close)

	coreMux := http.NewServeMux()
	coreMux.HandleFunc("/tower/edge/attach", func(w http.ResponseWriter, r *http.Request) {
		b := s.record(r)
		var f map[string]any
		_ = json.Unmarshal(b, &f)
		mu.Lock()
		stationID, _ = f["station_id"].(string)
		assertionHex, _ = f["assertion_key"].(string)
		sessionHex, _ = f["session_key"].(string)
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(TowerAttachment{
			StationID: stationID, TowerID: "tw-uk", Endpoint: hub.Listener.Addr().String(),
			HubToken: "hub-token", State: "active", TowerKeyHash: "00",
		})
	})
	coreMux.HandleFunc("/tower/dispatch/key", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"dispatch_key": hex.EncodeToString(corePub),
			"envelope_key": hex.EncodeToString(coreEnvPub),
		})
	})
	coreMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	core := httptest.NewServer(coreMux)
	s.t.Cleanup(core.Close)

	_, nodePriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ServeTower(ctx, Config{
			Broker: core.URL, NodeID: "n-uk", Model: "m", Modality: "chat", Parallel: 1,
			Upstream: s.upstream.URL, UpstreamKey: s.key,
		}, nodePriv, s.t.TempDir(), io.Discard, s.notice)
	}()
	defer func() { cancel(); <-done }()

	select {
	case <-completed:
	case <-time.After(10 * time.Second):
		return errors.New("the node never completed the job it was handed through the Tower")
	}
	if wantAudit {
		select {
		case <-audited:
		case <-time.After(10 * time.Second):
			return errors.New("the node never answered the audit for the attempt it served")
		}
	}
	// Give a notice raised on the completion a moment to land.
	time.Sleep(50 * time.Millisecond)
	return nil
}

// --- classic and local paths -----------------------------------------------------------

func (s *ukState) serveClassic(stream bool) error {
	var once sync.Once
	results := make(chan struct{}, 1)
	broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		switch {
		case strings.HasPrefix(r.URL.Path, "/nodes/register"):
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		case strings.HasPrefix(r.URL.Path, "/agent/poll"):
			handed := false
			once.Do(func() {
				body := `{"model":"m","messages":[]}`
				if stream {
					body = `{"model":"m","messages":[],"stream":true}`
				}
				_ = json.NewEncoder(w).Encode(protocol.Job{ID: "job-uk", User: "u", Body: json.RawMessage(body)})
				handed = true
			})
			if !handed {
				time.Sleep(20 * time.Millisecond)
				w.WriteHeader(http.StatusNoContent)
			}
		case strings.HasPrefix(r.URL.Path, "/agent/result"), strings.HasPrefix(r.URL.Path, "/agent/stream"):
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
			select {
			case results <- struct{}{}:
			default:
			}
		default:
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		}
	}))
	defer broker.Close()
	sess, err := Start(Config{Broker: broker.URL, Upstream: s.upstream.URL, UpstreamKey: s.key,
		NodeID: "n-uk", Model: "m", Parallel: 1})
	if err != nil {
		return err
	}
	defer sess.Stop()
	select {
	case <-results:
	case <-time.After(10 * time.Second):
		return errors.New("the classic job was never answered")
	}
	// A heartbeat or two, so the control-message rows see one.
	time.Sleep(150 * time.Millisecond)
	return nil
}

func (s *ukState) serveLocalTower() error {
	_, _ = runLocalJob(&http.Client{Timeout: 5 * time.Second},
		Config{Upstream: s.upstream.URL, UpstreamKey: s.key, Model: "m"},
		[]byte(`{"model":"m","messages":[]}`))
	return nil
}

// --- steps --------------------------------------------------------------------------------

func (s *ukState) operatorWithKey() error {
	s.key, s.requireK = ukKey, true
	s.startUpstream()
	return nil
}

func (s *ukState) keyLongEnough() error {
	if len(s.key) < 16 {
		return fmt.Errorf("K is %d characters", len(s.key))
	}
	return nil
}

func (s *ukState) tookKeyFrom(source string) error {
	switch source {
	case "the --upstream-key flag", "saved config", "the onboarding prompt":
		s.key = ukKey
		return nil
	}
	return fmt.Errorf("unknown key source %q", source)
}

func (s *ukState) servedOnPath(path string) error {
	switch path {
	case "joined Tower, pull", "joined Tower, hub":
		// From the node's side the two are one path: the Tower got the work by pulling it from
		// Core or by a consumer submitting to its hub, and either way hands it to the node over
		// the hub poll, served by the Station's EdgeExecutor.
		return s.serveThroughTower(false)
	case "classic, non-streaming":
		return s.serveClassic(false)
	case "classic, streaming":
		return s.serveClassic(true)
	case "local Tower":
		return s.serveLocalTower()
	}
	return fmt.Errorf("unknown path %q", path)
}

func (s *ukState) upstreamGotKey() error {
	s.upMu.Lock()
	defer s.upMu.Unlock()
	if len(s.upAuth) == 0 {
		return errors.New("the upstream was never called")
	}
	for _, a := range s.upAuth {
		if a != "Bearer "+ukKey {
			return fmt.Errorf("the upstream received Authorization %q, want the share's key", a)
		}
	}
	return nil
}

func (s *ukState) noKeyShare() error {
	s.key, s.requireK = "", false
	return nil
}

func (s *ukState) servedThroughTower() error { return s.serveThroughTower(false) }

func (s *ukState) upstreamGotNoAuth() error {
	s.upMu.Lock()
	defer s.upMu.Unlock()
	if len(s.upAuth) == 0 {
		return errors.New("the upstream was never called")
	}
	for _, a := range s.upAuth {
		if a != "" {
			return fmt.Errorf("the upstream received Authorization %q from a share with no key", a)
		}
	}
	return nil
}

func (s *ukState) upstreamEchoes(where string) error {
	s.echo = where
	return nil
}

// leakForms are the ways K could be read back out of bytes.
func leakForms() []string {
	return []string{ukKey, jsonEscapedAll(ukKey), strings.ToUpper(jsonEscapedAll(ukKey)),
		base64.StdEncoding.EncodeToString([]byte(ukKey))}
}

func containsLeak(b []byte) (string, bool) {
	for _, f := range leakForms() {
		if strings.Contains(string(b), f) {
			return f, true
		}
	}
	// K's characters escaped as \u00XX, in either case of hex digit.
	if strings.Contains(strings.ToLower(string(b)), strings.ToLower(jsonEscapedAll(ukKey))) {
		return "escaped K", true
	}
	return "", false
}

func (s *ukState) noKeyAnywhere() error {
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	for _, m := range s.sent {
		if f, ok := containsLeak(m.body); ok {
			return fmt.Errorf("K (%q form) was sent to %s in the body", f, m.path)
		}
		for h, vs := range m.header {
			for _, v := range vs {
				if f, ok := containsLeak([]byte(v)); ok {
					return fmt.Errorf("K (%q form) was sent to %s in header %s", f, m.path, h)
				}
			}
		}
	}
	for name, b := range map[string][]byte{"the consumer's bytes": s.consumer, "the receipt": s.receipt,
		"the failure": []byte(s.failure)} {
		if f, ok := containsLeak(b); ok {
			return fmt.Errorf("K (%q form) is in %s: %q", f, name, b)
		}
	}
	return nil
}

func (s *ukState) redactedWhereKeyWas() error {
	if len(s.consumer) > 0 {
		if !strings.Contains(string(s.consumer), "[redacted]") {
			return fmt.Errorf("the consumer's bytes do not show [redacted] where K was: %q", s.consumer)
		}
		return nil
	}
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	for _, n := range s.notices {
		if strings.Contains(n, "[redacted]") {
			return nil
		}
	}
	return fmt.Errorf("no consumer bytes and no operator line shows [redacted] where K was (failure %q, notices %q)",
		s.failure, s.notices)
}

func (s *ukState) bodyContainsKey() error {
	s.echo = "a 200 response body"
	return nil
}

func (s *ukState) serveWithAudit() error { return s.serveThroughTower(true) }

func (s *ukState) sameRedactedBytes() error {
	if len(s.consumer) == 0 {
		return fmt.Errorf("the consumer got nothing (failure %q)", s.failure)
	}
	if f, ok := containsLeak(s.consumer); ok {
		return fmt.Errorf("the consumer's bytes still hold K (%q form)", f)
	}
	signed, err := dispatch.ParseReceipt(s.receipt, s.assertion, link.PublicNetwork, s.attempt, s.stationID)
	if err != nil {
		return fmt.Errorf("the receipt does not verify: %v", err)
	}
	if signed.ResponseDigest != dispatch.DigestOf(s.consumer) {
		return errors.New("the receipt was signed over different bytes than the consumer received")
	}
	if signed.Usage.Out != int64(len(s.consumer)) {
		return fmt.Errorf("the receipt measured %d bytes out; the consumer received %d", signed.Usage.Out, len(s.consumer))
	}
	if signed.TokUsage.Out != 5 {
		return fmt.Errorf("the token count is %d, want the model's own 5", signed.TokUsage.Out)
	}
	if string(s.transcript) != string(s.consumer) {
		return fmt.Errorf("the audit transcript holds different bytes than the consumer received:\n%q\n%q", s.transcript, s.consumer)
	}
	return nil
}

func (s *ukState) keptInTranscripts() error {
	s.echo = "a 200 response body"
	return s.serveWithAudit()
}

func (s *ukState) auditAsks() error { return nil } // the rig's hub asks for the served attempt

func (s *ukState) transcriptHoldsNoKey() error {
	if len(s.transcript) == 0 {
		return errors.New("no transcript was answered")
	}
	if f, ok := containsLeak(s.transcript); ok {
		return fmt.Errorf("the audit transcript holds K (%q form)", f)
	}
	return nil
}

func (s *ukState) stationSends(msg string) error {
	s.msgRow = msg
	switch msg {
	case "registration", "heartbeat":
		return s.serveClassic(false)
	case "attach", "receipt", "hub poll":
		return s.serveThroughTower(false)
	}
	return fmt.Errorf("unknown message %q", msg)
}

// rowPath is the request each control-message row is about; the row proves nothing unless the
// Station really sent one.
var rowPath = map[string]string{
	"registration": "/nodes/register", "heartbeat": "/nodes/heartbeat", "attach": "/tower/edge/attach",
	"receipt": towerhub.PathComplete, "hub poll": towerhub.PathPoll,
}

func (s *ukState) messageHoldsNoKey() error {
	if err := s.noKeyAnywhere(); err != nil {
		return err
	}
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	want, sawRow := rowPath[s.msgRow], false
	for _, m := range s.sent {
		if m.path == want && (s.msgRow != "receipt" || strings.Contains(string(m.body), `"receipt":"`)) {
			sawRow = true
		}
	}
	if !sawRow {
		return fmt.Errorf("the Station never sent a %s (%s), so the row checked nothing", s.msgRow, want)
	}
	for _, m := range s.sent {
		low := strings.ToLower(string(m.body))
		for _, field := range []string{"upstream_key", "upstreamkey", "upstream-key"} {
			if strings.Contains(low, field) {
				return fmt.Errorf("%s carries a field named for an upstream key", m.path)
			}
		}
	}
	return nil
}

func (s *ukState) shortKey() error {
	s.key = "short7k"
	return nil
}

func (s *ukState) shareStarts() error {
	broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer broker.Close()
	sess, err := Start(Config{Broker: broker.URL, Upstream: s.upstream.URL, UpstreamKey: s.key,
		NodeID: "n-uk", Model: "m", Parallel: 1})
	if sess != nil {
		sess.Stop()
	}
	s.startErr = err
	return nil
}

func (s *ukState) refusesShortKey() error {
	if s.startErr == nil {
		return errors.New("the share started with a key too short to keep out of replies")
	}
	if !strings.Contains(s.startErr.Error(), "too short to keep out of replies") {
		return fmt.Errorf("the refusal does not say why: %v", s.startErr)
	}
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	if len(s.sent) != 0 {
		return fmt.Errorf("the share contacted the broker before refusing: %s", s.sent[0].path)
	}
	return nil
}

func (s *ukState) plainServeSafe() error {
	// The plain path forwards the upstream's error text; with K echoed into a 500 it must not
	// carry K, and it must carry the redaction.
	s.echo = "a 500 error body"
	s.requireK = true
	return s.plainServeCarriesNoKey()
}

func (s *ukState) upstreamRejectsKey() error {
	s.echo = "rejects K"
	s.jobs = 2 // two refusals with different bodies: the operator still hears it once
	return nil
}

func (s *ukState) noReceiptNoStrike() error {
	if len(s.receipt) != 0 {
		return errors.New("a failed attempt produced a receipt")
	}
	if s.failure == "" {
		return errors.New("the failed attempt carried no failure")
	}
	return nil
}

func (s *ukState) operatorToldOnce() error {
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	n := 0
	for _, m := range s.notices {
		if strings.Contains(m, "refused its key") {
			n++
		}
	}
	if n != 1 {
		return fmt.Errorf("the operator was told %d times that the upstream refused its key (notices %q)", n, s.notices)
	}
	return nil
}

func (s *ukState) lineRedactedFirst() error {
	s.sentMu.Lock()
	defer s.sentMu.Unlock()
	for _, m := range s.notices {
		if !strings.Contains(m, "refused its key") {
			continue
		}
		for i := 0; i+8 <= len(ukKey); i++ {
			if strings.Contains(m, ukKey[:i+8]) {
				return fmt.Errorf("the operator line holds a %d-character prefix of K: %q", i+8, m)
			}
		}
		return nil
	}
	return errors.New("there is no operator line about the refused key")
}

func (s *ukState) consumerToldClassOnly() error {
	if !strings.HasPrefix(s.failure, "the model did not answer (status 401") {
		return fmt.Errorf("the consumer was told %q, not only the error class", s.failure)
	}
	return nil
}

func TestTowerUpstreamKeyBDD(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	oldHB := heartbeatInterval.Load()
	heartbeatInterval.Store(int64(30 * time.Millisecond))
	t.Cleanup(func() { heartbeatInterval.Store(oldHB) })
	oldAudit := towerAuditEvery.Load()
	towerAuditEvery.Store(int64(100 * time.Millisecond))
	t.Cleanup(func() { towerAuditEvery.Store(oldAudit) })

	s := &ukState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = ukState{t: t}
				return ctx, nil
			})
			sc.Step(`^a signed-in operator sharing model "m" from an upstream that requires bearer key K$`, s.operatorWithKey)
			sc.Step(`^K is at least 16 characters$`, s.keyLongEnough)
			sc.Step(`^the share took K from "([^"]*)"$`, s.tookKeyFrom)
			sc.Step(`^a request is served on the "([^"]*)" path$`, s.servedOnPath)
			sc.Step(`^the upstream receives "Authorization: Bearer K"$`, s.upstreamGotKey)
			sc.Step(`^a share with no upstream key$`, s.noKeyShare)
			sc.Step(`^a request is served through a joined Tower$`, s.servedThroughTower)
			sc.Step(`^the upstream receives no Authorization header$`, s.upstreamGotNoAuth)
			sc.Step(`^the upstream puts K into "(.*)"$`, s.upstreamEchoes)
			sc.Step(`^the request is served through a joined Tower$`, s.servedThroughTower)
			sc.Step(`^K appears in nothing the Station sends to the Tower, to Core or to the consumer$`, s.noKeyAnywhere)
			sc.Step(`^"\[redacted\]" appears where K was$`, s.redactedWhereKeyWas)
			sc.Step(`^the upstream's 200 body contains K$`, s.bodyContainsKey)
			sc.Step(`^the Station serves it through a joined Tower$`, s.serveWithAudit)
			sc.Step(`^the bytes the receipt is signed over, the token count, the audit transcript and the consumer's bytes are the same redacted bytes$`, s.sameRedactedBytes)
			sc.Step(`^a request served through a joined Tower is kept in the Station's transcripts$`, s.keptInTranscripts)
			sc.Step(`^Core's audit asks for that transcript$`, s.auditAsks)
			sc.Step(`^the bytes sent hold no K$`, s.transcriptHoldsNoKey)
			sc.Step(`^the Station sends "([^"]*)"$`, s.stationSends)
			sc.Step(`^the message holds no K and no field named for an upstream key$`, s.messageHoldsNoKey)
			sc.Step(`^the operator gives an upstream key shorter than 8 characters$`, s.shortKey)
			sc.Step(`^`+"`roger share`"+` starts$`, s.shareStarts)
			sc.Step(`^it refuses to start, saying the key is too short to keep out of replies$`, s.refusesShortKey)
			sc.Step(`^EdgeExecutor\.Serve, which forwards the upstream error text, is removed, or redacts K before it forwards anything$`, s.plainServeSafe)
			sc.Step(`^the upstream rejects K with 401$`, s.upstreamRejectsKey)
			sc.Step(`^the attempt produces no receipt, the consumer's hold is released, and it is never a strike$`, s.noReceiptNoStrike)
			sc.Step(`^`+"`roger share`"+` prints once, on the operator's machine only, that the upstream refused its key$`, s.operatorToldOnce)
			sc.Step(`^that line is redacted before it is shortened, and holds no prefix of K of 8 or more characters$`, s.lineRedactedFirst)
			sc.Step(`^the consumer is told only the error class, as today$`, s.consumerToldClassOnly)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/tower_upstream_key.feature", bddrun.Titled(true))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("tower_upstream_key.feature failed")
	}
}

// plainServeCarriesNoKey drives the unsealed EdgeExecutor.Serve with a real grant against an
// upstream that echoes K into a 500, and checks what Serve hands back.
func (s *ukState) plainServeCarriesNoKey() error {
	s.startUpstream()
	st, err := station.InitOrOpen(s.t.TempDir())
	if err != nil {
		return err
	}
	corePub, corePriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	consumerPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	reg := dispatch.NewWithStore(dispatch.Config{Network: link.PublicNetwork, Signer: corePriv, Lifetime: time.Minute}, nil)
	body := []byte(`{"model":"m","messages":[]}`)
	g, err := reg.MintEdge(dispatch.EdgeTarget{
		TowerID: "tw-uk", StationID: st.StationID, Model: "m", Modality: "chat",
		RelayName: "st.relay.example", MaxIn: 1 << 16, MaxOut: 1 << 16, MaxTokIn: 1 << 16, MaxTokOut: 1 << 16,
		AssertionKey: st.AssertionPub(), ConsumerKey: consumerPub,
	})
	if err != nil {
		return err
	}
	exec := station.EdgeExecutor{Station: st, CoreKey: corePub, Network: link.PublicNetwork,
		Upstream: station.HTTPUpstream{URL: s.upstream.URL, Key: ukKey}}
	res := exec.Serve(context.Background(), station.EdgeRequest{Grant: base64.StdEncoding.EncodeToString(g.Signed), Body: body})
	if res.Status == 200 {
		return errors.New("the upstream's 500 was served as a success")
	}
	if f, ok := containsLeak([]byte(res.Failure)); ok {
		return fmt.Errorf("the plain Serve path forwarded K (%q form): %q", f, res.Failure)
	}
	if !strings.Contains(res.Failure, "[redacted]") {
		return fmt.Errorf("the plain Serve path's failure does not show [redacted] where K was: %q", res.Failure)
	}
	return nil
}
