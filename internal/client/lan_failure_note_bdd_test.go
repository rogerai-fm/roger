package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

// lan_failure_note_bdd_test.go - the @relay scenarios of
// features/agent/plain_chat_fallback_lan.feature (L2): the words a LAN Tower or an Edge member
// leaves behind when it answers 5xx. The REAL relayLAN runs against httptest sources; every
// assertion is made on the refusal the relay writes.

type lanNoteState struct {
	t       *testing.T
	srvs    []*httptest.Server
	towers  []string
	dial    dialMap
	peers   []EdgePeer
	stall   chan struct{}
	took    time.Duration
	message string // the long message a scenario sent
	status  int
	refusal string
	code    string
}

func (s *lanNoteState) reset() {
	s.close()
	*s = lanNoteState{t: s.t, dial: dialMap{}}
}

func (s *lanNoteState) close() {
	if s.stall != nil {
		close(s.stall)
		s.stall = nil
	}
	for _, srv := range s.srvs {
		srv.Close()
	}
}

func lanNoteHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	})
}

func (s *lanNoteState) tower(addr string, status int, body string) error {
	srv := httptest.NewServer(lanNoteHandler(status, body))
	s.srvs = append(s.srvs, srv)
	s.towers = append(s.towers, addr)
	s.dial[addr] = srv.Listener.Addr().String()
	return nil
}

func (s *lanNoteState) towerLong(addr string, status, n int) error {
	s.message = strings.Repeat("abcdéfghij", n/10) // é: the cut counts characters, not bytes
	raw, _ := json.Marshal(map[string]any{"error": map[string]any{"message": s.message}})
	return s.tower(addr, status, string(raw))
}

// towerHuge answers valid JSON whose message sits past the part of the body that is read.
func (s *lanNoteState) towerHuge(addr string, status, mb int) error {
	return s.tower(addr, status, `{"pad":"`+strings.Repeat("x", mb<<20)+`","error":{"message":"late"}}`)
}

// edgeStalls answers 5xx, sends the start of a body and then nothing more.
func (s *lanNoteState) edgeStalls(node string, status int) error {
	s.stall = make(chan struct{})
	return s.edgeFace(node, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"error":`)
		w.(http.Flusher).Flush()
		<-s.stall
	}))
}

func (s *lanNoteState) edge(node string, status int, body string) error {
	return s.edgeFace(node, lanNoteHandler(status, body))
}

func (s *lanNoteState) edgeFace(node string, h http.Handler) error {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: node},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour)}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	ts := httptest.NewUnstartedServer(h)
	ts.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}, ClientAuth: tls.RequestClientCert}
	ts.StartTLS()
	s.srvs = append(s.srvs, ts)
	s.peers = append(s.peers, EdgePeer{Addr: strings.TrimPrefix(ts.URL, "https://"), Pin: fmt.Sprintf("%x", sha256.Sum256(der)), Node: node, Instance: "serve"})
	return nil
}

func (s *lanNoteState) relay(band string) error {
	opts := ProxyOptions{LANMode: true, LANKey: lanKey(s.t), LANHTTP: s.dial.client()}
	if s.towers != nil {
		opts.LANTowers = func(string) []string { return s.towers }
	}
	if s.peers != nil {
		opts.EdgePeers = func(string) []EdgePeer { return s.peers }
	}
	rec := httptest.NewRecorder()
	start := time.Now()
	defer func() { s.took = time.Since(start) }()
	body := []byte(`{"model":"` + band + `","messages":[{"role":"user","content":"hi"}],"tools":[{"type":"function"}]}`)
	if served, refused := relayLAN(context.Background(), rec, opts, band, body, func(float64) {}); served || !refused {
		return fmt.Errorf("the relay served a turn every source failed (served=%v refused=%v)", served, refused)
	}
	var out struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		return fmt.Errorf("the relay's refusal is not JSON: %s", rec.Body.String())
	}
	s.status, s.refusal, s.code = rec.Code, out.Error.Message, out.Error.Code
	return nil
}

func (s *lanNoteState) answers503() error {
	if s.status != http.StatusServiceUnavailable || s.code != "lan_only" {
		return fmt.Errorf("the relay answered %d %q, want 503 lan_only: %s", s.status, s.code, s.refusal)
	}
	return nil
}

func (s *lanNoteState) contains(want string) error {
	if !strings.Contains(s.refusal, want) {
		return fmt.Errorf("the refusal does not contain %q:\n%s", want, s.refusal)
	}
	return nil
}

func (s *lanNoteState) endsWith(want string) error {
	if !strings.HasSuffix(s.refusal, want) {
		return fmt.Errorf("the refusal does not end with %q:\n%s", want, s.refusal)
	}
	return nil
}

// tried is the one "tried" entry of a refusal that names a single failed source.
func (s *lanNoteState) tried() (string, error) {
	const open, shut = "failed (", "); LAN mode is on"
	i, j := strings.Index(s.refusal, open), strings.LastIndex(s.refusal, shut)
	if i < 0 || j < i {
		return "", fmt.Errorf("the refusal names no tried source:\n%s", s.refusal)
	}
	return s.refusal[i+len(open) : j], nil
}

func (s *lanNoteState) triedIs(want string) error {
	got, err := s.tried()
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("the tried entry is %q, want %q", got, want)
	}
	return nil
}

func (s *lanNoteState) triedCutTo200() error {
	got, err := s.tried()
	if err != nil {
		return err
	}
	if want := s.towers[0] + ": status 500: " + string([]rune(s.message)[:200]); got != want {
		return fmt.Errorf("the tried entry is %d characters long:\n%q\nwant the first 200 of the message:\n%q", len(got), got, want)
	}
	return nil
}

func TestLANFailureNoteBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &lanNoteState{t: t}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				s.close()
				return ctx, err
			})
			q := `"([^"]*)"`
			// Background: the relay under test is built per scenario, in LAN mode, for the band
			// the When names.
			sc.Step(`^LAN mode is on$`, func() error { return nil })
			sc.Step(`^the agent is on model `+q+`$`, func(string) error { return nil })
			sc.Step(`^a LAN Tower at `+q+` answers (\d+) with body '(.*)'$`, s.tower)
			sc.Step(`^a LAN Tower at `+q+` answers (\d+) with an error message of (\d+) characters$`, s.towerLong)
			sc.Step(`^a LAN Tower at `+q+` answers (\d+) with a (\d+) MB JSON reply whose error message comes after the first 64 KB$`, s.towerHuge)
			sc.Step(`^an Edge member `+q+` answers (\d+) and then never finishes its reply$`, s.edgeStalls)
			sc.Step(`^the relay answers 503 lan_only within (\d+) seconds$`, func(n int) error {
				if s.took > time.Duration(n)*time.Second {
					return fmt.Errorf("the relay took %v", s.took)
				}
				return s.answers503()
			})
			sc.Step(`^an Edge member `+q+` answers (\d+) with body '(.*)'$`, s.edge)
			sc.Step(`^a turn for `+q+` is relayed on the LAN$`, s.relay)
			sc.Step(`^the relay answers 503 lan_only$`, s.answers503)
			sc.Step(`^the refusal contains `+q+`$`, s.contains)
			sc.Step(`^the refusal ends with `+q+`$`, s.endsWith)
			sc.Step(`^the tried entry for that Tower is `+q+`$`, s.triedIs)
			sc.Step(`^the tried entry for that Tower carries the message's first 200 characters and no more$`, s.triedCutTo200)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/agent/plain_chat_fallback_lan.feature"},
			Tags:     "@relay",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback_lan.feature @relay failed (see godog output above)")
	}
}
