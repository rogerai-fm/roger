package harness

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
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

// plainchat_lan_bdd_test.go - features/agent/plain_chat_fallback_lan.feature, the untagged
// scenarios. The REAL Loop runs on the REAL BrokerCompleterRoute with the REAL LAN relay set
// (client.SetLANRelay), in front of httptest sources that answer the way a model server does:
// this machine's own serve (plain HTTP), an Edge member (mutual TLS, pinned certificate) and a
// LAN Tower (a private address dialed to a local listener). Every assertion is made on what
// reached a source, or on what the relay answered.

// lanSrc is one local source and the way it answers.
type lanSrc struct {
	kind string // "self", "edge" or "tower"
	addr string // tower: the private address the relay names it by
	srv  *httptest.Server
	peer client.EdgePeer
	down bool // this machine's serve is not answering

	mu          sync.Mutex
	reqs        []lanReq
	refuseTools *plainOverride // any request that carries tools
	refuseAll   *plainOverride // every request
	plainFail   *plainOverride // a request without tools
	rawStatus   int            // answer every request with this status and rawBody
	rawBody     string
	toolFirst   bool // the first tools request gets a tool call back
	toolCalls   bool // a success carries a tool_calls array
}

type lanReq struct {
	body   map[string]any
	status int
	fail   string
}

func (c *lanSrc) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	c.mu.Lock()
	defer c.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if c.kind == "tower" {
		w.Header().Set("X-Roger-Local", "1")
	}
	fail := func(o plainOverride) {
		c.reqs = append(c.reqs, lanReq{body: body, status: o.status, fail: o.msg})
		w.WriteHeader(o.status)
		_, _ = w.Write(failBody(o.msg))
	}
	switch {
	case c.rawStatus != 0:
		c.reqs = append(c.reqs, lanReq{body: body, status: c.rawStatus})
		w.WriteHeader(c.rawStatus)
		_, _ = io.WriteString(w, c.rawBody)
		return
	case c.refuseAll != nil:
		fail(*c.refuseAll)
		return
	case c.toolFirst && hasTools(body):
		c.toolFirst = false
		c.reqs = append(c.reqs, lanReq{body: body, status: 200})
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"c1","type":"function","function":{"name":"list_dir","arguments":"{\"path\":\".\"}"}}]},"finish_reason":"tool_calls"}]}`)
		return
	case c.refuseTools != nil && hasTools(body):
		fail(*c.refuseTools)
		return
	case c.plainFail != nil && !hasTools(body):
		fail(*c.plainFail)
		return
	}
	c.reqs = append(c.reqs, lanReq{body: body, status: 200})
	msg := map[string]any{"role": "assistant", "content": plainReply}
	if c.toolCalls {
		msg["tool_calls"] = []any{map[string]any{"id": "c1", "type": "function",
			"function": map[string]any{"name": "read_file", "arguments": `{"path":"notes.md"}`}}}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": msg, "finish_reason": "stop"}}})
}

func (c *lanSrc) seen() []lanReq {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]lanReq(nil), c.reqs...)
}

type lanPlainState struct {
	t *testing.T

	lanMode    bool
	relayOn    bool
	edgePrefer string
	model      string
	routing    client.Routing

	srcs   []*lanSrc
	last   *lanSrc // "that source"
	market *lanSrc // the market broker: must stay silent on the LAN
	cert   *tls.Certificate

	loop       *Loop
	turnFirst  map[*lanSrc]int
	events     []Event
	turnEvts   []Event
	answer     string
	err        error
	cost       float64
	toolsRun   int
	plainCalls int
}

func (s *lanPlainState) reset() {
	s.close()
	*s = lanPlainState{t: s.t, turnFirst: map[*lanSrc]int{}}
	s.market = &lanSrc{kind: "market"}
	s.market.srv = httptest.NewServer(http.HandlerFunc(s.market.serve))
}

func (s *lanPlainState) close() {
	client.SetLANRelay(nil)
	for _, c := range s.srcs {
		if c.srv != nil {
			c.srv.Close()
		}
	}
	if s.market != nil && s.market.srv != nil {
		s.market.srv.Close()
	}
}

// --- sources ---------------------------------------------------------------------------------

func (s *lanPlainState) add(c *lanSrc) *lanSrc {
	switch c.kind {
	case "edge":
		c.srv, c.peer = lanEdgeFace(c)
	default:
		c.srv = httptest.NewServer(http.HandlerFunc(c.serve))
	}
	s.srcs = append(s.srcs, c)
	s.last = c
	return c
}

// lanEdgeFace is an Edge member's face: TLS with a certificate the fleet pinned.
func lanEdgeFace(c *lanSrc) (*httptest.Server, client.EdgePeer) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "jetson"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour)}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(c.serve))
	ts.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}, ClientAuth: tls.RequestClientCert}
	ts.StartTLS()
	return ts, client.EdgePeer{Addr: strings.TrimPrefix(ts.URL, "https://"), Pin: fmt.Sprintf("%x", sha256.Sum256(der)), Node: "jetson", Instance: "serve"}
}

// source finds or makes the source a step names: "this machine", "an Edge member", or
// `a LAN Tower at "<address>"`.
func (s *lanPlainState) source(name string) *lanSrc {
	kind, addr := "self", ""
	switch {
	case strings.HasPrefix(name, "an Edge member"):
		kind = "edge"
	case strings.HasPrefix(name, "a LAN Tower at "):
		kind, addr = "tower", strings.Trim(strings.TrimPrefix(name, "a LAN Tower at "), `"`)
	}
	for _, c := range s.srcs {
		if c.kind == kind && c.addr == addr {
			s.last = c
			return c
		}
	}
	return s.add(&lanSrc{kind: kind, addr: addr})
}

const lanSrcRe = `(this machine|an Edge member|a LAN Tower at "[^"]*")`

func (s *lanPlainState) refusesTools(name, _ string, status int, msg string) error {
	s.source(name).refuseTools = &plainOverride{status, msg}
	return nil
}

func (s *lanPlainState) refusesAll(name, _ string, status int, msg string) error {
	s.source(name).refuseAll = &plainOverride{status, msg}
	return nil
}

func (s *lanPlainState) drivesTools(name, _ string) error { s.source(name); return nil }

func (s *lanPlainState) towerRaw(addr string, status int, body string) error {
	c := s.source(`a LAN Tower at "` + addr + `"`)
	c.rawStatus, c.rawBody = status, body
	return nil
}

func (s *lanPlainState) selfDown(string) error {
	c := s.source("this machine")
	c.down = true
	c.srv.Close()
	return nil
}

func (s *lanPlainState) plainFails(status int, msg string) error {
	s.last.plainFail = &plainOverride{status, msg}
	return nil
}

func (s *lanPlainState) toolFirst() error           { s.last.toolFirst = true; return nil }
func (s *lanPlainState) nextToolCalls() error       { s.last.toolCalls = true; return nil }
func (s *lanPlainState) nothingServes(string) error { return nil }

// --- the relay and the loop ------------------------------------------------------------------

// relay sets the process's LAN relay over the sources declared so far, as `roger --lan` does.
func (s *lanPlainState) relay() {
	if !s.relayOn {
		client.SetLANRelay(nil)
		return
	}
	dial := map[string]string{}
	var self *lanSrc
	var peers []client.EdgePeer
	var towers []string
	for _, c := range s.srcs {
		switch c.kind {
		case "self":
			self = c
		case "edge":
			peers = append(peers, c.peer)
		case "tower":
			towers = append(towers, c.addr)
			dial[c.addr] = c.srv.Listener.Addr().String()
		}
	}
	if s.cert == nil {
		key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		tmpl := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "this-node"},
			NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour)}
		der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
		s.cert = &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	}
	opts := client.ProxyOptions{LANMode: s.lanMode, EdgePrefer: s.edgePrefer, EdgeCert: s.cert,
		LANHTTP: &http.Client{Transport: &http.Transport{DisableKeepAlives: true,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if to, ok := dial[addr]; ok {
					addr = to
				}
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			}}}}
	if self != nil {
		opts.LANSelf = func(string) (string, string) { return self.srv.URL + "/v1/chat/completions", "" }
	}
	if peers != nil {
		opts.EdgePeers = func(string) []client.EdgePeer { return peers }
	}
	if towers != nil {
		opts.LANTowers = func(string) []string { return towers }
	}
	client.SetLANRelay(&opts)
}

func (s *lanPlainState) completer() Completer {
	return func(ctx context.Context, messages []Message, tools []map[string]any) (Message, error) {
		if len(tools) == 0 {
			s.plainCalls++ // counted here: a refusal with no source up reaches no source
		}
		return BrokerCompleterRoute(BrokerRoute{
			Broker: s.market.srv.URL, User: "u", Model: s.model, Routing: s.routing,
			OnCost:          func(c float64, _, _ int, _ float64) { s.cost += c },
			FallbackTimeout: brokerTimeout,
		})(ctx, messages, tools)
	}
}

func (s *lanPlainState) ensureLoop() {
	if s.loop == nil {
		s.loop = NewLoop(s.t.TempDir(), "You are a test persona.", s.completer(), func(string, map[string]any) bool { return false })
	}
}

func (s *lanPlainState) send(text string) error {
	s.relay()
	s.ensureLoop()
	for _, c := range append([]*lanSrc{s.market}, s.srcs...) {
		s.turnFirst[c] = len(c.seen())
	}
	s.turnEvts, s.cost, s.plainCalls = nil, 0, 0
	s.answer, s.err = s.loop.Send(context.Background(), text, func(e Event) {
		if e.Kind == EventToolResult {
			s.toolsRun++
		}
		s.events = append(s.events, e)
		s.turnEvts = append(s.turnEvts, e)
	})
	return nil
}

func (s *lanPlainState) turn(c *lanSrc) []lanReq { return c.seen()[s.turnFirst[c]:] }

func (s *lanPlainState) earlierRefused() error {
	_ = s.send("earlier")
	if s.err != nil {
		return fmt.Errorf("the earlier turn did not fall back: %v", s.err)
	}
	s.toolsRun = 0
	return nil
}

func (s *lanPlainState) holdsHistory() error {
	s.ensureLoop()
	s.loop.messages = append(s.loop.messages,
		Message{Role: "user", Content: "one"}, Message{Role: "assistant", Content: "first"},
		Message{Role: "user", Content: "two"}, Message{Role: "assistant", Content: "second"})
	return nil
}

// --- assertions ------------------------------------------------------------------------------

func lanNotices(evs []Event) (n int, text string) {
	for _, e := range evs {
		if e.Kind == EventNotice && strings.Contains(e.Text, "plain chat") {
			n++
			text = e.Text
		}
	}
	return
}

func plainKeys(r lanReq) error {
	for _, k := range []string{"tools", "tool_choice"} {
		if _, ok := r.body[k]; ok {
			return fmt.Errorf("the request carried %q: %v", k, r.body[k])
		}
	}
	return nil
}

func (s *lanPlainState) firstHasTools() error {
	rs := s.turn(s.last)
	if len(rs) == 0 || !hasTools(rs[0].body) {
		return fmt.Errorf("that source's first request of the turn did not carry tools: %d requests", len(rs))
	}
	return nil
}

func (s *lanPlainState) secondPlain() error {
	rs := s.turn(s.last)
	if len(rs) != 2 {
		return fmt.Errorf("that source received %d requests for the turn, want 2 (turn error: %v)", len(rs), s.err)
	}
	if !hasTools(rs[0].body) {
		return fmt.Errorf("the first request carried no tools")
	}
	return plainKeys(rs[1])
}

func (s *lanPlainState) showsSecondReply() error {
	if s.err != nil || s.answer != plainReply {
		return fmt.Errorf("the turn gave %q, %v; want %q", s.answer, s.err, plainReply)
	}
	return nil
}

func (s *lanPlainState) showsNotice(model string) error {
	n, text := lanNotices(s.turnEvts)
	if n != 1 || !strings.Contains(text, model) || !strings.Contains(text, "cannot drive tools") {
		return fmt.Errorf("want one notice naming %s as unable to drive tools, got %d: %q", model, n, text)
	}
	return nil
}

func (s *lanPlainState) noNotice() error {
	if n, text := lanNotices(s.turnEvts); n != 0 {
		return fmt.Errorf("a plain chat notice was shown: %q", text)
	}
	return nil
}

func (s *lanPlainState) marketSilent() error {
	if n := len(s.market.seen()); n != 0 {
		return fmt.Errorf("the market broker received %d requests", n)
	}
	return nil
}

func (s *lanPlainState) marketGot(n int) error {
	if got := len(s.turn(s.market)); got != n {
		return fmt.Errorf("the market broker received %d requests, want %d", got, n)
	}
	return nil
}

func (s *lanPlainState) sameMessages() error {
	rs := s.turn(s.last)
	if len(rs) != 2 {
		return fmt.Errorf("that source received %d requests, want 2", len(rs))
	}
	if !reflect.DeepEqual(rs[0].body["messages"], rs[1].body["messages"]) {
		return fmt.Errorf("the second request's messages differ:\n%v\n%v", rs[0].body["messages"], rs[1].body["messages"])
	}
	if ms, _ := rs[0].body["messages"].([]any); len(ms) != 6 {
		return fmt.Errorf("want the persona, two earlier turns and the new message (6), got %d", len(ms))
	}
	return nil
}

func (s *lanPlainState) receivedN(word string) error {
	want := map[string]int{"one": 1, "two": 2}[word]
	if got := len(s.turn(s.last)); got != want {
		return fmt.Errorf("that source received %d requests for the turn, want %d (turn error: %v)", got, want, s.err)
	}
	return nil
}

func (s *lanPlainState) itIsPlain() error  { return plainKeys(s.turn(s.last)[0]) }
func (s *lanPlainState) itHasTools() error { return s.firstHasTools() }
func (s *lanPlainState) costIsZero() error {
	if s.cost != 0 {
		return fmt.Errorf("the LAN turn cost %v", s.cost)
	}
	return nil
}

// endsWithReply checks the turn's error is the one the completer gives for that reply without
// the fallback: the same parse every relayed turn goes through.
func (s *lanPlainState) endsWithReply(which string) error {
	src := s.last
	if !s.relayOn {
		src = s.market
	}
	rs := s.turn(src)
	i := map[string]int{"first": 0, "second": 1}[which]
	if len(rs) <= i || rs[i].fail == "" {
		return fmt.Errorf("the %s reply of the turn was not a refusal (%d requests, turn error: %v)", which, len(rs), s.err)
	}
	want := todaysError(plainOverride{rs[i].status, rs[i].fail})
	if s.err == nil || s.err.Error() != want {
		return fmt.Errorf("the turn ended with %v, want %q", s.err, want)
	}
	return nil
}

// endsWithRelayRefusal checks the turn's error is what the relay itself answers for this
// request, asked again directly.
func (s *lanPlainState) endsWithRelayRefusal() error {
	if s.err == nil {
		return fmt.Errorf("the turn did not fail (answer %q)", s.answer)
	}
	body, _ := json.Marshal(map[string]any{"model": s.model, "messages": []any{}, "tools": []any{map[string]any{"type": "function"}}})
	req, _ := http.NewRequest(http.MethodPost, "http://relay/v1/chat/completions", strings.NewReader(string(body)))
	before := map[*lanSrc]int{}
	for _, c := range s.srcs {
		before[c] = len(c.seen())
	}
	resp, in := client.LANDo(req, body)
	for _, c := range s.srcs { // the probe is the test's own request, not the turn's
		c.mu.Lock()
		c.reqs = c.reqs[:before[c]]
		c.mu.Unlock()
	}
	if !in {
		return fmt.Errorf("no LAN relay is set")
	}
	raw, _ := io.ReadAll(resp.Body)
	_, want := parseCompletion(raw, resp.StatusCode)
	if want == nil || s.err.Error() != want.Error() {
		return fmt.Errorf("the turn ended with %q, want the relay's own %v", s.err, want)
	}
	return nil
}

func (s *lanPlainState) noPlainRequest() error {
	if s.plainCalls != 0 {
		return fmt.Errorf("the turn asked %d times without tools", s.plainCalls)
	}
	return nil
}

func (s *lanPlainState) noToolRun() error {
	if s.toolsRun != 0 {
		return fmt.Errorf("%d tools ran on a plain chat turn", s.toolsRun)
	}
	for _, m := range s.loop.messages {
		if len(m.ToolCalls) != 0 {
			return fmt.Errorf("the conversation kept tool_calls from the reply")
		}
	}
	return nil
}

func TestPlainChatFallbackLANBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &lanPlainState{t: t}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				s.close()
				return ctx, err
			})
			q := `"([^"]*)"`
			sc.Step(`^LAN mode is on$`, func() error { s.lanMode, s.relayOn = true, true; return nil })
			sc.Step(`^LAN mode is off$`, func() error { s.lanMode, s.relayOn = false, false; return nil })
			sc.Step(`^LAN mode is off and the Edge preference is local-only$`, func() error {
				s.lanMode, s.relayOn, s.edgePrefer = false, true, "local-only"
				return nil
			})
			sc.Step(`^the agent is on model `+q+`$`, func(m string) error { s.model = m; return nil })
			sc.Step(`^the user's routing requires `+q+`$`, func(c string) error { s.routing.Require = []string{c}; return nil })
			sc.Step(`^the agent's session holds a system persona and two earlier turns$`, s.holdsHistory)

			sc.Step(`^`+lanSrcRe+` serves `+q+` and refuses tools with (\d+) `+q+`$`, s.refusesTools)
			sc.Step(`^`+lanSrcRe+` serves `+q+` and refuses every request with (\d+) `+q+`$`, s.refusesAll)
			sc.Step(`^`+lanSrcRe+` serves `+q+` and drives tools$`, s.drivesTools)
			sc.Step(`^a LAN Tower at `+q+` answers (\d+) with body '(.*)'$`, s.towerRaw)
			sc.Step(`^this machine's serve of `+q+` is down$`, s.selfDown)
			sc.Step(`^nothing on the LAN serves `+q+`$`, s.nothingServes)
			sc.Step(`^that source answers a request without tools with (\d+) `+q+`$`, s.plainFails)
			sc.Step(`^that source's first reply of the turn calls a tool before it refuses$`, s.toolFirst)
			sc.Step(`^that source's next reply carries a tool_calls array$`, s.nextToolCalls)
			sc.Step(`^that source refused an earlier turn of this session for tools$`, s.earlierRefused)
			sc.Step(`^the market broker answers every request with (\d+) `+q+`$`, func(status int, msg string) error {
				s.market.refuseAll = &plainOverride{status, msg}
				return nil
			})

			sc.Step(`^the user sends `+q+` to the agent$`, s.send)

			sc.Step(`^that source receives a request with a non-empty tools array$`, s.firstHasTools)
			sc.Step(`^that source then receives a second request with no "tools" key and no "tool_choice" key$`, s.secondPlain)
			sc.Step(`^that source receives exactly (one|two) requests? for that turn$`, s.receivedN)
			sc.Step(`^it has no "tools" key and no "tool_choice" key$`, s.itIsPlain)
			sc.Step(`^it has a non-empty tools array$`, s.itHasTools)
			sc.Step(`^the agent shows the second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the transcript shows a notice that `+q+` cannot drive tools and is answering as plain chat$`, s.showsNotice)
			sc.Step(`^no plain chat notice is shown$`, s.noNotice)
			sc.Step(`^no second notice is shown$`, s.noNotice)
			sc.Step(`^the market broker received no request$`, s.marketSilent)
			sc.Step(`^the market broker received exactly one request$`, func() error { return s.marketGot(1) })
			sc.Step(`^the second request's messages equal the first request's messages$`, s.sameMessages)
			sc.Step(`^the turn's cost is 0$`, s.costIsZero)
			sc.Step(`^the turn ends with the error the (first|second) reply gives$`, s.endsWithReply)
			sc.Step(`^the turn ends with the relay's own refusal$`, s.endsWithRelayRefusal)
			sc.Step(`^no request without tools was sent$`, s.noPlainRequest)
			sc.Step(`^no tool is run$`, s.noToolRun)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/agent/plain_chat_fallback_lan.feature"},
			Tags:     "~@tui && ~@webui && ~@relay",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback_lan.feature failed (see godog output above)")
	}
}
