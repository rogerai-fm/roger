package webui

// Executable spec: features/console/endpoint_plate.feature - the console's endpoint plate.
// The REAL console server over HTTP, the REAL local proxy it binds on loopback, and a broker
// stand-in on loopback that serves /discover, /balance and the relay. The page scenarios EXECUTE
// the shipped console-channel.js with node against that server. No mocks.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/edge"
	"rogerai.fm/roger/v7/internal/netbind"
	"rogerai.fm/roger/v7/internal/node"
	"rogerai.fm/roger/v7/internal/protocol"
)

type plateReply struct {
	code int
	hdr  http.Header
	raw  string
	js   map[string]any
}

type plateRelay struct {
	model  string
	maxOut float64
	user   string
	signed bool
}

type platePage struct {
	Text        string
	Attrs       string
	Buttons     []string
	Rows        []string
	Clipboard   string
	Toasts      []string
	Tabs        []string
	Fetches     []struct{ Method, Path string }
	PollMs      int
	After       string // the plate's text after its timer fired
	TickFetches []struct{ Method, Path string }
}

type plateBDD struct {
	t *testing.T

	mu           sync.Mutex
	offers       map[string]client.Offer
	loggedIn     bool
	balance      float64
	monthlyStop  bool
	costPerCall  string
	relays       []plateRelay
	reqN         int
	binds        int             // endpoint binds this scenario
	brokerHits   int             // every request the stand-in saw
	hidden       map[string]bool // private bands: on air, absent from /discover
	holdStream   chan struct{}   // non-nil: a stream waits on it before finishing
	streamStart  chan struct{}
	streamResult chan string

	broker  *httptest.Server
	console *httptest.Server
	srv     *Server
	tok     string
	limits  map[string]SpendLimit
	ledger  *edge.Sessions
	lan     bool
	noBrk   bool
	served  net.Listener // a console started with Serve(ln), for the stop scenario

	last   plateReply
	second plateReply
	tool   plateReply
	base   string
	key    string
	oldKey string
	oldURL string
	quoted float64

	tuiLn       net.Listener
	tuiKey      string
	hold        net.Listener // a port held busy by the scenario
	logBuf      *bytes.Buffer
	page        platePage
	clipRefuses bool

	restoreLogout func() error
	restoreListen func(string) (net.Listener, error)
}

// ---- the broker stand-in ---------------------------------------------------------------

func (s *plateBDD) startBroker() {
	s.broker = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.brokerHits++
		switch {
		case r.URL.Path == "/discover":
			list := []client.Offer{}
			for m, o := range s.offers {
				if !s.hidden[m] {
					list = append(list, o)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"offers": list})
		case r.URL.Path == "/balance":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"user": "op", "balance": s.balance, "logged_in": s.loggedIn})
		case r.URL.Path == "/v1/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"object":"list","data":[]}`)
		case r.URL.Path == "/v1/chat/completions":
			s.relay(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

// relay is the stand-in's relay. A request with neither a device signature nor a recognised
// bearer cannot spend, as on the real broker. Called with s.mu held; a held stream releases it.
func (s *plateBDD) relay(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	signed := r.Header.Get("X-Roger-Sig") != ""
	bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !signed && !strings.HasPrefix(bearer, "rog-key_") && !strings.HasPrefix(bearer, "rog-grant_") && r.Header.Get("X-Roger-User") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"unsigned request cannot spend","type":"authentication_error"}}`)
		return
	}
	rel := plateRelay{user: r.Header.Get("X-Roger-User"), signed: signed}
	rel.model, _ = m["model"].(string)
	if v, err := strconv.ParseFloat(r.Header.Get("X-Roger-Max-Price-Out"), 64); err == nil {
		rel.maxOut = v
	}
	if pv, _ := m["provider"].(map[string]any); pv != nil {
		if mp, _ := pv["max_price"].(map[string]any); mp != nil {
			if c, ok := mp["completion"].(float64); ok {
				rel.maxOut = c
			}
		}
	}
	s.relays = append(s.relays, rel)
	if s.monthlyStop {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = io.WriteString(w, `{"error":{"message":"monthly spend cap reached: $5.00 of $5.00 this month","type":"insufficient_quota","code":"monthly_cap"}}`)
		return
	}
	o, on := s.offers[rel.model]
	if !on || !o.Online {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error":{"message":"no station on air for `+rel.model+`","type":"api_error","code":"no_station"}}`)
		return
	}
	s.reqN++
	rec := protocol.EncodeReceipt(protocol.UsageReceipt{RequestID: fmt.Sprintf("req-%d", s.reqN), NodeID: o.NodeID, Model: rel.model, CompletionTokens: 3})
	w.Header().Set("X-RogerAI-Provider", o.NodeID)
	if stream, _ := m["stream"].(bool); stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fl, _ := w.(http.Flusher)
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"po\"}}]}\n\n")
		fl.Flush()
		if hold, started := s.holdStream, s.streamStart; hold != nil {
			s.mu.Unlock()
			close(started)
			<-hold
			s.mu.Lock()
		}
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ng\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n: rogerai-cost=0.001\n\n: rogerai-receipt="+rec+"\n\n")
		return
	}
	if s.costPerCall != "" {
		w.Header().Set("X-RogerAI-Cost", s.costPerCall)
	}
	w.Header().Set("X-RogerAI-Receipt", rec)
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"pong from `+rel.model+`"}}]}`)
}

// ---- the console under test -----------------------------------------------------------

func (s *plateBDD) startConsole() {
	if s.console != nil {
		s.console.Close()
	}
	opts := Options{
		User: "op",
		ReadLimits: func() map[string]SpendLimit {
			s.mu.Lock()
			defer s.mu.Unlock()
			out := map[string]SpendLimit{}
			for k, v := range s.limits {
				out[k] = v
			}
			return out
		},
		WriteLimit: func(model string, l SpendLimit) { s.mu.Lock(); s.limits[model] = l; s.mu.Unlock() },
		Edge:       EdgeHooks{Sessions: s.ledger},
	}
	if !s.noBrk {
		opts.Broker = s.broker.URL
	}
	if s.lan {
		opts.LANDial = func() client.LANDial { return client.LANDial{} }
	}
	s.srv = New(node.New(node.Config{Station: "amber-fox"}), opts)
	s.srv.token = s.tok
	s.console = httptest.NewServer(s.srv.Handler())
}

func (s *plateBDD) reset() {
	s.cleanup()
	*s = plateBDD{t: s.t, offers: map[string]client.Offer{}, hidden: map[string]bool{}, limits: map[string]SpendLimit{}, tok: "tok-run",
		ledger: edge.NewSessions("acct-1")}
	s.restoreLogout = logoutReturn
	logoutReturn = func() error { return nil }
	s.restoreListen = channelListen
	listen := channelListen
	channelListen = func(addr string) (net.Listener, error) {
		s.mu.Lock()
		s.binds++
		s.mu.Unlock()
		return listen(addr)
	}
	s.startBroker()
}

func (s *plateBDD) cleanup() {
	if s.holdStream != nil {
		select {
		case <-s.holdStream:
		default:
			close(s.holdStream)
		}
	}
	if s.console != nil {
		_, _ = s.do("POST", "/api/channel/close", "", s.tok)
		s.console.Close()
	}
	if s.served != nil {
		_ = s.served.Close()
	}
	for _, ln := range []net.Listener{s.tuiLn, s.hold} {
		if ln != nil {
			_ = ln.Close()
		}
	}
	if s.broker != nil {
		s.broker.Close()
	}
	if s.restoreLogout != nil {
		logoutReturn = s.restoreLogout
	}
	if s.restoreListen != nil {
		channelListen = s.restoreListen
	}
	if s.logBuf != nil {
		log.SetOutput(os.Stderr)
	}
}

func (s *plateBDD) do(method, path, body, tok string) (plateReply, error) {
	req, err := http.NewRequest(method, s.console.URL+path, strings.NewReader(body))
	if err != nil {
		return plateReply{}, err
	}
	if tok != "" {
		req.Header.Set("X-Roger-Token", tok)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return send(req)
}

func send(req *http.Request) (plateReply, error) {
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return plateReply{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	out := plateReply{code: resp.StatusCode, hdr: resp.Header, raw: string(b)}
	_ = json.Unmarshal(b, &out.js)
	return out, nil
}

func (s *plateBDD) page_(method, path, body string) error {
	r, err := s.do(method, path, body, s.tok)
	s.last = r
	if err == nil && path == "/api/channel/open" && r.js["ok"] == true {
		s.base, _ = r.js["base_url"].(string)
		s.key, _ = r.js["key"].(string)
	}
	return err
}

// open opens a channel on model the way the page would, typing the price back when asked.
func (s *plateBDD) open(model string) error {
	if s.console == nil {
		s.startConsole()
	}
	if err := s.page_("POST", "/api/channel/open", `{"model":"`+model+`"}`); err != nil {
		return err
	}
	if s.last.js["confirm_needed"] == true {
		p, _ := s.last.js["price_out"].(float64)
		if err := s.page_("POST", "/api/channel/open", fmt.Sprintf(`{"model":%q,"confirm_price_out":%g}`, model, p)); err != nil {
			return err
		}
	}
	if s.last.js["ok"] != true || s.base == "" || s.key == "" {
		return fmt.Errorf("the channel on %q did not open: %d %s", model, s.last.code, s.last.raw)
	}
	return nil
}

func (s *plateBDD) toolCall(method, path, bearer, body string) (plateReply, error) {
	req, err := http.NewRequest(method, s.base+path, strings.NewReader(body))
	if err != nil {
		return plateReply{}, err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return send(req)
}

func (s *plateBDD) chat(model string, stream bool) (plateReply, error) {
	body := fmt.Sprintf(`{"model":%q,"stream":%v,"messages":[{"role":"user","content":"ping"}]}`, model, stream)
	return s.toolCall("POST", "/chat/completions", s.key, body)
}

func (s *plateBDD) status() (map[string]any, error) {
	r, err := s.do("GET", "/api/channel", "", s.tok)
	if err != nil {
		return nil, err
	}
	if r.code != 200 {
		return nil, fmt.Errorf("GET /api/channel is %d: %s", r.code, r.raw)
	}
	return r.js, nil
}

func (s *plateBDD) nothingOpen() error {
	st, err := s.status()
	if err != nil {
		return err
	}
	if st["open"] != false {
		return fmt.Errorf("a channel is open: %v", st)
	}
	r, err := s.do("POST", "/api/channel/key", "", s.tok)
	if err != nil {
		return err
	}
	if r.code != http.StatusConflict {
		return fmt.Errorf("POST /api/channel/key is %d, want 409 (a key exists)", r.code)
	}
	return nil
}

func portOf(base string) (int, error) {
	u, err := url.Parse(base)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(u.Port())
}

func refused(base string) error {
	u, err := url.Parse(base)
	if err != nil {
		return err
	}
	c, err := net.DialTimeout("tcp", u.Host, time.Second)
	if err == nil {
		_ = c.Close()
		return fmt.Errorf("%s still accepts connections", u.Host)
	}
	return nil
}

func (s *plateBDD) answers(base, key string) error {
	req, _ := http.NewRequest("GET", base+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	r, err := send(req)
	if err != nil {
		return err
	}
	if r.code != 200 {
		return fmt.Errorf("%s/models with its key is %d: %s", base, r.code, r.raw)
	}
	return nil
}

func (s *plateBDD) lastRelay() (plateRelay, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.relays) == 0 {
		return plateRelay{}, fmt.Errorf("the broker stand-in received no relay")
	}
	return s.relays[len(s.relays)-1], nil
}

func (s *plateBDD) setOffer(model string, in, out float64, free bool) {
	s.mu.Lock()
	s.offers[model] = client.Offer{NodeID: "st-" + model, Model: model, PriceIn: in, PriceOut: out, Online: true, FreeNow: free, TPS: 90, Signal: 80, Modality: "chat"}
	s.mu.Unlock()
}

// ---- the page, executed --------------------------------------------------------------

// runPage executes the shipped console-channel.js with node against the live console. `then`
// names what the operator does; `canned` (optional) replaces the reply to POST /api/channel/open.
func (s *plateBDD) runPage(model, then, canned string, clipboardRefuses bool) error {
	nodeBin, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("node is needed to execute the console's JavaScript: %v", err)
	}
	script := `
const api = require(process.argv[1]);
const [base, tok, model, then, canned, refuse] = process.argv.slice(2);
const out = { text: "", attrs: "", buttons: [], rows: [], clipboard: "", toasts: [], tabs: [], fetches: [], pollMs: 0, after: "", tickFetches: [] };
function el(tag) {
  return { tag, children: [], attrs: {}, hidden: false, textContent: "", className: "", onclick: null,
    appendChild(c) { this.children.push(c); return c; },
    setAttribute(k, v) { this.attrs[k] = String(v); }, getAttribute(k) { return this.attrs[k]; },
    addEventListener(t, f) { if (t === "click") this.onclick = f; },
    replaceChildren(...cs) { this.children = cs; } };
}
function walk(e, f) { f(e); (e.children || []).forEach((c) => walk(c, f)); }
function mkDoc() {
  const host = el("div");
  return { host, getElementById: (id) => (id === "channel-plate" ? host : null), createElement: el };
}
function read(host) {
  const r = { text: [], attrs: [], buttons: [], rows: [] };
  walk(host, (e) => {
    if (e.hidden) return;
    if (e.textContent) r.text.push(e.textContent);
    Object.values(e.attrs).forEach((v) => r.attrs.push(v));
    if (e.tag === "button") r.buttons.push(e.textContent);
    if (e.className && String(e.className).split(" ").includes("plate-label")) r.rows.push(e.textContent);
  });
  return r;
}
function actions(ticks) {
  return {
    every: (fn, ms) => { out.pollMs = ms; ticks.push(fn); },
    fetch: (path, opts) => {
      opts = opts || {};
      out.fetches.push({ method: opts.method || "GET", path });
      if (canned && path === "/api/channel/open") {
        return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(JSON.parse(canned)) });
      }
      const h = Object.assign({ "X-Roger-Token": tok }, opts.headers || {});
      return fetch(base + path, Object.assign({}, opts, { headers: h }));
    },
    model: () => model,
    clipboard: { writeText: (t) => (refuse === "1" ? Promise.reject(new Error("denied")) : (out.clipboard = t, Promise.resolve())) },
    toast: (m) => out.toasts.push(m),
    tab: (name) => out.tabs.push(name),
  };
}
(async () => {
  const ticksA = [], ticksB = [];
  const a = mkDoc(), plate = await api.mount(a, actions(ticksA));
  if (then.includes("open")) await plate.open();
  if (then.includes("copy")) await plate.copy();
  let b = null;
  if (then.includes("twotabs")) { b = mkDoc(); await api.mount(b, actions(ticksB)); }
  const r = read(a.host);
  out.text = r.text.join("\n"); out.attrs = r.attrs.join("\n"); out.buttons = r.buttons; out.rows = r.rows;
  // The timer the page was handed, fired as the browser would fire it.
  if (b) { await plate.close(); for (const t of ticksB) await t(); out.after = read(b.host).text.join("\n"); }
  else { const n = out.fetches.length; for (const t of ticksA) await t(); out.tickFetches = out.fetches.slice(n); out.after = read(a.host).text.join("\n"); }
  process.stdout.write(JSON.stringify(out));
  process.exit(0);
})().catch((e) => { console.error(e); process.exit(2); });
`
	ref := "0"
	if clipboardRefuses {
		ref = "1"
	}
	cmd := exec.Command(nodeBin, "-e", script, "./assets/console-channel.js", s.console.URL, s.tok, model, then, canned, ref)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("console-channel.js failed: %v: %s", err, b)
	}
	s.page = platePage{}
	return json.Unmarshal(b, &s.page)
}

func plateAsset(name string) (string, error) {
	b, err := assetsFS.ReadFile("assets/" + name)
	return string(b), err
}

// ---- steps -----------------------------------------------------------------------------

func (s *plateBDD) wantCode(code int) error {
	if s.last.code != code {
		return fmt.Errorf("the response is %d, want %d: %s", s.last.code, code, s.last.raw)
	}
	return nil
}

func (s *plateBDD) wantOK(want bool) error {
	if s.last.js == nil || s.last.js["ok"] != want {
		return fmt.Errorf(`"ok" is %v, want %v: %d %s`, s.last.js["ok"], want, s.last.code, s.last.raw)
	}
	return nil
}

func (s *plateBDD) wantFlag(name string) error {
	if s.last.js[name] != true {
		return fmt.Errorf("%q is %v, want true: %s", name, s.last.js[name], s.last.raw)
	}
	return nil
}

func (s *plateBDD) msgHas(words ...string) error {
	m, _ := s.last.js["message"].(string)
	if m == "" {
		m = s.last.raw
	}
	for _, w := range words {
		if !strings.Contains(strings.ToLower(m), strings.ToLower(w)) {
			return fmt.Errorf("the message %q does not say %q", m, w)
		}
	}
	return nil
}

func (s *plateBDD) noKeyIn(what, body string) error {
	if s.key == "" {
		return fmt.Errorf("no key was issued to compare against")
	}
	if strings.Contains(body, s.key) {
		return fmt.Errorf("%s contains the key", what)
	}
	return nil
}

func (s *plateBDD) startStream() error {
	s.mu.Lock()
	s.holdStream, s.streamStart, s.streamResult = make(chan struct{}), make(chan struct{}), make(chan string, 1)
	started := s.streamStart
	s.mu.Unlock()
	go func() {
		r, err := s.chat("ignored", true)
		if err != nil {
			s.streamResult <- "error: " + err.Error()
			return
		}
		s.streamResult <- r.raw
	}()
	select {
	case <-started:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("the stream never reached the broker stand-in")
	}
}

func (s *plateBDD) finishStream() (string, error) {
	s.mu.Lock()
	hold := s.holdStream
	s.mu.Unlock()
	close(hold)
	select {
	case body := <-s.streamResult:
		if !strings.Contains(body, "[DONE]") {
			return body, fmt.Errorf("the stream did not complete: %q", body)
		}
		return body, nil
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("the stream never finished")
	}
}

func (s *plateBDD) receiptedTo(model string) error {
	for _, sess := range s.ledger.Live() {
		if sess.Band == model && sess.Kind == edge.FromConsole {
			return nil
		}
	}
	return fmt.Errorf("no console session on the Edge ledger for %q (ledger holds %d)", model, s.ledger.Len())
}

func TestConsoleEndpointPlateFeature(t *testing.T) {
	s := &plateBDD{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.cleanup()
				s.console, s.broker, s.served, s.tuiLn, s.hold, s.holdStream, s.restoreLogout, s.logBuf, s.restoreListen = nil, nil, nil, nil, nil, nil, nil, nil, nil
				return ctx, nil
			})

			// --- background and givens
			sc.Step(`^a console server over a node controller with broker "[^"]*"$`, func() error { s.startConsole(); return nil })
			sc.Step(`^the run token is "([^"]*)"$`, func(tok string) error {
				if s.srv.Token() != tok {
					return fmt.Errorf("the run token is %q", s.srv.Token())
				}
				return nil
			})
			sc.Step(`^band "([^"]*)" is on air on the open market at \$([\d.]+) in / \$([\d.]+) out per 1M$`, func(m string, in, out float64) error {
				s.setOffer(m, in, out, false)
				return nil
			})
			sc.Step(`^band "([^"]*)" is on air on the open market, free now$`, func(m string) error { s.setOffer(m, 0, 0, true); return nil })
			sc.Step(`^band "([^"]*)" is on air at \$([\d.]+) out per 1M$`, func(m string, out float64) error { s.setOffer(m, 1, out, false); return nil })
			sc.Step(`^it is now on air at \$([\d.]+) out per 1M$`, func(out float64) error { s.setOffer("big-405b", 1, out, false); return nil })
			sc.Step(`^band "([^"]*)" was quoted at \$([\d.]+) out per 1M$`, func(m string, out float64) error {
				s.setOffer(m, 1, out, false)
				s.mu.Lock()
				s.limits[m] = SpendLimit{MaxOut: 30}
				s.mu.Unlock()
				if err := s.page_("GET", "/api/channel/quote?model="+m, ""); err != nil {
					return err
				}
				if p, _ := s.last.js["price_out"].(float64); p != out {
					return fmt.Errorf("the quote is %v, want %v: %s", s.last.js["price_out"], out, s.last.raw)
				}
				return nil
			})
			sc.Step(`^band "([^"]*)" is private behind a frequency code$`, func(m string) error {
				s.setOffer(m, 0, 0, true) // on air and free: only its being private keeps the console out
				s.mu.Lock()
				s.hidden[m] = true
				s.mu.Unlock()
				return nil
			})
			sc.Step(`^the operator is logged in with balance \$([\d.]+)$`, func(b float64) error {
				s.mu.Lock()
				s.loggedIn, s.balance = true, b
				s.mu.Unlock()
				return nil
			})
			sc.Step(`^the operator is not logged in$`, func() error { s.mu.Lock(); s.loggedIn = false; s.mu.Unlock(); return nil })
			sc.Step(`^no channel is open$`, func() error { return nil })
			sc.Step(`^a console server with no broker$`, func() error { s.noBrk = true; s.startConsole(); return nil })
			sc.Step(`^the console runs in LAN mode$`, func() error { s.lan = true; s.startConsole(); return nil })
			sc.Step(`^the operator's cap for "([^"]*)" is \$([\d.]+) per 1M out$`, func(m string, c float64) error {
				s.mu.Lock()
				s.limits[m] = SpendLimit{MaxOut: c}
				s.mu.Unlock()
				return nil
			})
			sc.Step(`^the operator has set no cap for "([^"]*)"$`, func(m string) error { s.mu.Lock(); delete(s.limits, m); s.mu.Unlock(); return nil })
			sc.Step(`^the account's monthly cap is reached$`, func() error { s.mu.Lock(); s.monthlyStop = true; s.mu.Unlock(); return nil })
			sc.Step(`^port 4141 on 127\.0\.0\.1 is already bound$`, func() error {
				s.hold, _ = net.Listen("tcp", "127.0.0.1:4141") // already bound by something else serves as well
				return nil
			})
			sc.Step(`^no loopback port can be bound$`, func() error {
				channelListen = func(string) (net.Listener, error) { return nil, fmt.Errorf("address already in use") }
				return nil
			})
			sc.Step(`^a channel is open on "([^"]*)"$`, s.open)
			sc.Step(`^a channel was opened and closed$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				s.oldKey, s.oldURL = s.key, s.base
				return s.page_("POST", "/api/channel/close", "")
			})
			sc.Step(`^a channel was opened from one tab$`, func() error { return s.open("qwen3-32b") })
			sc.Step(`^a channel is open on "([^"]*)" and a stream is in flight$`, func(m string) error {
				if err := s.open(m); err != nil {
					return err
				}
				return s.startStream()
			})
			sc.Step(`^the TUI in the same process has a channel open on port 4141$`, func() error {
				ln, err := netbind.ListenFree("127.0.0.1:4141")
				if err != nil {
					return err
				}
				s.tuiLn, s.tuiKey = ln, client.NewSessionKey()
				h := client.ProxyHandler(client.ProxyOptions{Broker: s.broker.URL, User: "op", Model: "free-4b", SessionKey: s.tuiKey})
				go func() { _ = http.Serve(ln, h) }()
				return nil
			})
			sc.Step(`^the TUI and the console each have a channel open$`, func() error {
				ln, err := netbind.ListenFree("127.0.0.1:4141")
				if err != nil {
					return err
				}
				s.tuiLn, s.tuiKey = ln, client.NewSessionKey()
				h := client.ProxyHandler(client.ProxyOptions{Broker: s.broker.URL, User: "op", Model: "free-4b", SessionKey: s.tuiKey})
				go func() { _ = http.Serve(ln, h) }()
				return s.open("qwen3-32b")
			})
			sc.Step(`^the console logs requests$`, func() error {
				s.logBuf = &bytes.Buffer{}
				log.SetOutput(s.logBuf)
				return nil
			})
			sc.Step(`^the browser refuses clipboard writes$`, func() error { s.clipRefuses = true; return s.open("qwen3-32b") })
			sc.Step(`^the chat model picker is on "([^"]*)"$`, func(m string) error {
				s.mu.Lock()
				s.limits[m] = SpendLimit{MaxOut: 1}
				s.mu.Unlock()
				return s.runPage(m, "closed", "", false)
			})
			sc.Step(`^two tabs show the open plate$`, func() error { return nil })

			// --- the page's requests
			sc.Step(`^the page POSTs (/api/\S+) with (\{.*\}|not json)$`, func(path, body string) error { return s.page_("POST", path, body) })
			sc.Step(`^the page POSTs (/api/\S+) with a 5 KiB body$`, func(path string) error {
				return s.page_("POST", path, `{"model":"`+strings.Repeat("a", 5*1024)+`"}`)
			})
			sc.Step(`^the page POSTs (/api/\S+)$`, func(path string) error {
				body := ""
				if path == "/api/channel/open" {
					body = `{"model":"qwen3-32b"}` // an open always names its model
				}
				return s.page_("POST", path, body)
			})
			sc.Step(`^the page GETs (/api/\S+) with the run token$`, func(path string) error {
				if path != "/api/channel/open" { // something to hand over or close
					if err := s.open("qwen3-32b"); err != nil {
						return err
					}
				}
				return s.page_("GET", path, "")
			})
			sc.Step(`^the page GETs (/api/\S+)$`, func(path string) error { return s.page_("GET", path, "") })
			sc.Step(`^the page reads three frames of /api/events$`, func() error {
				req, _ := http.NewRequest("GET", s.console.URL+"/api/events", nil)
				req.Header.Set("X-Roger-Token", s.tok)
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()
				resp, err := http.DefaultClient.Do(req.WithContext(ctx))
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				var got bytes.Buffer
				buf := make([]byte, 4096)
				for strings.Count(got.String(), "data: ") < 3 {
					n, err := resp.Body.Read(buf)
					got.Write(buf[:n])
					if err != nil {
						break
					}
				}
				if strings.Count(got.String(), "data: ") < 3 {
					return fmt.Errorf("read fewer than three frames")
				}
				s.last = plateReply{code: resp.StatusCode, raw: got.String()}
				return nil
			})
			sc.Step(`^two pages POST /api/channel/open with \{"model":"([^"]*)"\} at once$`, func(m string) error { return s.twoAtOnce(m, m) })
			sc.Step(`^two pages POST /api/channel/open at once, one for "([^"]*)" and one for "([^"]*)"$`, s.twoAtOnce)
			sc.Step(`^a second tab with the run token GETs /api/channel$`, func() error { return s.page_("GET", "/api/channel", "") })
			sc.Step(`^a request to (GET|POST) (/api/\S+) carries (no token|a wrong token)$`, func(method, path, tok string) error {
				t := ""
				if tok == "a wrong token" {
					t = "not-the-token"
				}
				if path == "/api/channel/quote" {
					path += "?model=qwen3-32b"
				}
				if path != "/api/channel/open" { // something to read or close
					if err := s.open("qwen3-32b"); err != nil {
						return err
					}
				}
				r, err := s.do(method, path, "", t)
				s.last = r
				return err
			})
			sc.Step(`^a request to /api/state carries the key as its token$`, func() error {
				r, err := s.do("GET", "/api/state", "", s.key)
				s.last = r
				return err
			})
			sc.Step(`^a request to POST /api/channel/open arrives from a non-loopback address with the run token$`, func() error {
				req := httptest.NewRequest("POST", "/api/channel/open", strings.NewReader(`{"model":"qwen3-32b"}`))
				req.RemoteAddr = "10.0.0.5:51000"
				req.Header.Set("X-Roger-Token", s.tok)
				rec := httptest.NewRecorder()
				s.srv.Handler().ServeHTTP(rec, req)
				s.last = plateReply{code: rec.Code, hdr: rec.Header(), raw: rec.Body.String()}
				return nil
			})
			sc.Step(`^another origin's page POSTs /api/channel/open as a form without the run token$`, func() error {
				req, _ := http.NewRequest("POST", s.console.URL+"/api/channel/open", strings.NewReader("model=qwen3-32b"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Origin", "https://evil.example")
				r, err := send(req)
				s.last = r
				return err
			})
			sc.Step(`^another origin sends OPTIONS /api/channel/key$`, func() error {
				req, _ := http.NewRequest("OPTIONS", s.console.URL+"/api/channel/key", nil)
				req.Header.Set("Origin", "https://evil.example")
				req.Header.Set("Access-Control-Request-Method", "POST")
				r, err := send(req)
				s.last = r
				return err
			})
			sc.Step(`^the operator sets the cap for "([^"]*)" to \$([\d.]+) per 1M out$`, func(m string, c float64) error {
				if err := s.page_("POST", "/api/limits", fmt.Sprintf(`{"model":%q,"max_out":%g}`, m, c)); err != nil {
					return err
				}
				return s.wantCode(200)
			})
			sc.Step(`^the operator logs out on ACCOUNT$`, func() error {
				s.mu.Lock()
				s.loggedIn = false
				s.mu.Unlock()
				if err := s.page_("POST", "/api/account/logout", ""); err != nil {
					return err
				}
				return s.wantCode(200)
			})
			// The console keeps no count of tabs: a closed tab is simply a page that stops asking.
			sc.Step(`^every console tab is closed$`, func() error { return nil })
			sc.Step(`^the console server stops$`, func() error {
				// Serve the SAME server on a listener of its own, open through it, then stop it.
				ln, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					return err
				}
				_, _ = s.do("POST", "/api/channel/close", "", s.tok)
				done := make(chan struct{})
				go func() { _ = s.srv.Serve(ln); close(done) }()
				keep := s.console.URL
				s.console.URL = "http://" + ln.Addr().String()
				err = s.open("qwen3-32b")
				s.console.URL = keep
				if err != nil {
					return err
				}
				_ = ln.Close()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					return fmt.Errorf("Serve did not return after its listener closed")
				}
				time.Sleep(200 * time.Millisecond)
				return nil
			})
			sc.Step(`^a channel is opened, its key fetched, and closed$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				if err := s.page_("POST", "/api/channel/key", ""); err != nil {
					return err
				}
				if _, err := s.chat("qwen3-32b", false); err != nil {
					return err
				}
				return s.page_("POST", "/api/channel/close", "")
			})

			// --- tools on the endpoint
			sc.Step(`^a tool GETs <base_url>/models with "Authorization: Bearer <key>"$`, func() error {
				r, err := s.toolCall("GET", "/models", s.key, "")
				s.tool = r
				return err
			})
			sc.Step(`^a tool POSTs <base_url>/chat/completions with the key$`, func() error {
				st, _ := s.status()
				m, _ := st["model"].(string)
				r, err := s.chat(m, false)
				s.tool = r
				return err
			})
			sc.Step(`^a tool POSTs <base_url>/chat/completions naming model "([^"]*)" with the key$`, func(m string) error {
				r, err := s.chat(m, false)
				s.tool = r
				return err
			})
			sc.Step(`^a tool completes one turn through it$`, func() error {
				r, err := s.chat("qwen3-32b", false)
				s.tool = r
				if err == nil && r.code != 200 {
					return fmt.Errorf("the turn is %d: %s", r.code, r.raw)
				}
				return err
			})
			sc.Step(`^a tool calls (GET|POST) (/v1/\S+) with (.+)$`, func(method, route, bearer string) error {
				var b string
				switch bearer {
				case "no Authorization header":
				case "Bearer wrong":
					b = "wrong"
				case "the console run token":
					b = s.tok
				case "the key with one char off":
					b = s.key[:len(s.key)-1] + "x"
					if b == s.key {
						b = s.key[:len(s.key)-1] + "y"
					}
				case "the key of a closed channel":
					old := s.key
					if err := s.page_("POST", "/api/channel/close", ""); err != nil {
						return err
					}
					if err := s.open("qwen3-32b"); err != nil {
						return err
					}
					b = old
				default:
					return fmt.Errorf("unknown bearer %q", bearer)
				}
				s.mu.Lock()
				s.relays = nil
				s.mu.Unlock()
				body := ""
				if method == "POST" {
					body = `{"model":"qwen3-32b","messages":[{"role":"user","content":"ping"}]}`
				}
				r, err := s.toolCall(method, strings.TrimPrefix(route, "/v1"), b, body)
				s.tool = r
				return err
			})
			sc.Step(`^tools spend more than the guest-operator session budget through it$`, func() error {
				s.mu.Lock()
				s.costPerCall = fmt.Sprintf("%g", client.DefaultSessionBudget+1)
				s.mu.Unlock()
				for i := 0; i < 3; i++ {
					r, err := s.chat("qwen3-32b", false)
					if err != nil {
						return err
					}
					s.tool = r
					if r.code == http.StatusPaymentRequired {
						return fmt.Errorf("the proxy refused request %d with 402: %s", i+1, r.raw)
					}
				}
				return nil
			})
			sc.Step(`^every station for "([^"]*)" leaves the air$`, func(m string) error { s.mu.Lock(); delete(s.offers, m); s.mu.Unlock(); return nil })

			// --- what came back
			sc.Step(`^the response is (\d+) with "ok" true$`, func(code int) error {
				if err := s.wantCode(code); err != nil {
					return err
				}
				return s.wantOK(true)
			})
			sc.Step(`^the response is (\d+)$`, s.wantCode)
			sc.Step(`^the response is (\d+) "([^"]*)"$`, func(code int, text string) error {
				if err := s.wantCode(code); err != nil {
					return err
				}
				if !strings.Contains(s.last.raw, text) {
					return fmt.Errorf("the body %q does not say %q", s.last.raw, text)
				}
				return nil
			})
			sc.Step(`^the response is (\d+) and "message" is "([^"]*)"$`, func(code int, msg string) error {
				if err := s.wantCode(code); err != nil {
					return err
				}
				if s.last.js["message"] != msg {
					return fmt.Errorf(`"message" is %v, want %q`, s.last.js["message"], msg)
				}
				return nil
			})
			sc.Step(`^the response is 503 and "message" says a channel needs a broker$`, func() error {
				if err := s.wantCode(503); err != nil {
					return err
				}
				return s.msgHas("broker")
			})
			sc.Step(`^the response is 400 naming "([^"]*)"$`, func(f string) error {
				if err := s.wantCode(400); err != nil {
					return err
				}
				return s.msgHas(f)
			})
			sc.Step(`^the response is 200 and GET /api/channel reports "open" false$`, func() error {
				if err := s.wantCode(200); err != nil {
					return err
				}
				st, err := s.status()
				if err != nil {
					return err
				}
				if st["open"] != false {
					return fmt.Errorf("the channel is still open: %v", st)
				}
				return nil
			})
			sc.Step(`^the response is 200 and "open" is false$`, func() error {
				if err := s.wantCode(200); err != nil {
					return err
				}
				if s.last.js["open"] != false {
					return fmt.Errorf(`"open" is %v`, s.last.js["open"])
				}
				return nil
			})
			sc.Step(`^the response is 200 with the same "base_url" and "key"$`, func() error {
				if err := s.wantCode(200); err != nil {
					return err
				}
				return s.sameEndpoint()
			})
			sc.Step(`^the response has "ok" (true|false)$`, func(v string) error { return s.wantOK(v == "true") })
			sc.Step(`^the response has "ok" false and "(login_needed|over_limit)" true$`, func(flag string) error {
				if err := s.wantOK(false); err != nil {
					return err
				}
				return s.wantFlag(flag)
			})
			sc.Step(`^the response has "ok" false and "confirm_needed" true with "price_out" ([\d.]+)$`, func(p float64) error {
				if err := s.wantOK(false); err != nil {
					return err
				}
				if err := s.wantFlag("confirm_needed"); err != nil {
					return err
				}
				if got, _ := s.last.js["price_out"].(float64); got != p {
					return fmt.Errorf(`"price_out" is %v, want %v`, s.last.js["price_out"], p)
				}
				return nil
			})
			sc.Step(`^the response has "ok" false and "message" says the endpoint could not bind$`, func() error {
				if err := s.wantOK(false); err != nil {
					return err
				}
				return s.msgHas("could not bind")
			})
			sc.Step(`^the response has "ok" false and "message" says the plate is not available in LAN mode yet$`, func() error {
				if err := s.wantOK(false); err != nil {
					return err
				}
				return s.msgHas("LAN mode")
			})
			sc.Step(`^"message" says the band is paid and to log in on ACCOUNT$`, func() error { return s.msgHas("paid", "ACCOUNT") })
			sc.Step(`^"message" carries the resolver's reason, not "open failed"$`, func() error {
				if err := s.msgHas("no station on air"); err != nil {
					return err
				}
				if strings.Contains(s.last.raw, "open failed") {
					return fmt.Errorf("the message is the flattened one: %s", s.last.raw)
				}
				return nil
			})
			sc.Step(`^"message" names the cheapest price on air and the operator's cap$`, func() error { return s.msgHas("0.60", "0.50") })
			sc.Step(`^no listener is bound and no key is minted$`, s.nothingOpen)
			sc.Step(`^nothing is opened, read or closed$`, s.untouched)
			sc.Step(`^nothing is opened or closed$`, s.untouched)
			sc.Step(`^GET /api/channel reports "open" false$`, func() error {
				st, err := s.status()
				if err != nil {
					return err
				}
				if st["open"] != false {
					return fmt.Errorf("the channel is open: %v", st)
				}
				return nil
			})
			sc.Step(`^"base_url" is "http://127\.0\.0\.1:<port>/v1" for the port the listener bound$`, func() error {
				if !regexp.MustCompile(`^http://127\.0\.0\.1:\d+/v1$`).MatchString(s.base) {
					return fmt.Errorf(`"base_url" is %q`, s.base)
				}
				u, _ := url.Parse(s.base)
				c, err := net.DialTimeout("tcp", u.Host, time.Second)
				if err != nil {
					return fmt.Errorf("nothing listens on %s: %v", u.Host, err)
				}
				return c.Close()
			})
			sc.Step(`^"model" is "([^"]*)"$`, func(m string) error {
				if s.last.js["model"] != m {
					return fmt.Errorf(`"model" is %v, want %q`, s.last.js["model"], m)
				}
				return nil
			})
			sc.Step(`^"key" is a fresh session key$`, func() error {
				if len(s.key) < 32 || s.key == s.tok {
					return fmt.Errorf(`"key" is %d characters`, len(s.key))
				}
				return nil
			})
			sc.Step(`^"key_hint" is the last 4 characters of that key$`, func() error {
				if s.key == "" || s.last.js["key_hint"] != s.key[len(s.key)-4:] {
					return fmt.Errorf(`"key_hint" is %v`, s.last.js["key_hint"])
				}
				return nil
			})
			sc.Step(`^"station" names the station the channel locked$`, func() error {
				if s.last.js["station"] != "st-qwen3-32b" {
					return fmt.Errorf(`"station" is %v`, s.last.js["station"])
				}
				return nil
			})
			sc.Step(`^"base_url" names a port above 4141$`, func() error {
				p, err := portOf(s.base)
				if err != nil || p <= 4141 {
					return fmt.Errorf(`"base_url" is %q`, s.base)
				}
				return nil
			})
			sc.Step(`^the plate shows that port, not 4141$`, func() error {
				st, err := s.status()
				if err != nil {
					return err
				}
				if st["base_url"] != s.base || strings.Contains(s.base, ":4141/") {
					return fmt.Errorf("the status shows %v, the open returned %q", st["base_url"], s.base)
				}
				return nil
			})
			sc.Step(`^the listener's address is 127\.0\.0\.1$`, func() error {
				if !strings.HasPrefix(s.base, "http://127.0.0.1:") {
					return fmt.Errorf("the endpoint is %q", s.base)
				}
				return nil
			})
			sc.Step(`^a connection to the same port on a non-loopback interface is refused$`, func() error {
				p, _ := portOf(s.base)
				addrs, _ := net.InterfaceAddrs()
				for _, a := range addrs {
					ipn, ok := a.(*net.IPNet)
					if !ok || ipn.IP.IsLoopback() || ipn.IP.To4() == nil {
						continue
					}
					c, err := net.DialTimeout("tcp", net.JoinHostPort(ipn.IP.String(), strconv.Itoa(p)), time.Second)
					if err == nil {
						_ = c.Close()
						return fmt.Errorf("the endpoint answers on %s", ipn.IP)
					}
				}
				return nil
			})

			// --- the tool's view
			sc.Step(`^the response is 200 listing exactly "([^"]*)"$`, func(m string) error { return s.toolLists(s.tool, m) })
			sc.Step(`^GET <base_url>/models with the key lists exactly "([^"]*)"$`, func(m string) error {
				r, err := s.toolCall("GET", "/models", s.key, "")
				if err != nil {
					return err
				}
				return s.toolLists(r, m)
			})
			sc.Step(`^the broker stand-in receives one relay for "([^"]*)" as the operator$`, func(m string) error {
				s.mu.Lock()
				n := len(s.relays)
				s.mu.Unlock()
				rel, err := s.lastRelay()
				if err != nil {
					return err
				}
				if n != 1 || rel.model != m || (rel.user != "op" && !rel.signed) {
					return fmt.Errorf("%d relays, last %+v", n, rel)
				}
				return nil
			})
			sc.Step(`^the tool receives the reply$`, func() error {
				if s.tool.code != 200 || !strings.Contains(s.tool.raw, "pong") {
					return fmt.Errorf("the tool got %d: %s", s.tool.code, s.tool.raw)
				}
				return nil
			})
			sc.Step(`^the relay carries a max out price of \$([\d.]+)$`, func(want float64) error {
				rel, err := s.lastRelay()
				if err != nil {
					return fmt.Errorf("%v (the tool got %d: %s)", err, s.tool.code, s.tool.raw)
				}
				if rel.maxOut != want {
					return fmt.Errorf("the relay carried a max out price of %g, want %g", rel.maxOut, want)
				}
				return nil
			})
			sc.Step(`^the relay goes to "([^"]*)"$`, func(m string) error {
				rel, err := s.lastRelay()
				if err != nil {
					return fmt.Errorf("%v (the tool got %d: %s)", err, s.tool.code, s.tool.raw)
				}
				if rel.model != m {
					return fmt.Errorf("the relay went to %q", rel.model)
				}
				return nil
			})
			sc.Step(`^the key and the port are unchanged$`, s.statusSameEndpoint)
			sc.Step(`^no request is refused with 402 by the proxy$`, func() error {
				if s.tool.code != 200 {
					return fmt.Errorf("the last request is %d: %s", s.tool.code, s.tool.raw)
				}
				return nil
			})
			sc.Step(`^the broker's monthly cap still refuses when it is reached$`, func() error {
				s.mu.Lock()
				s.monthlyStop = true
				s.mu.Unlock()
				r, err := s.chat("qwen3-32b", false)
				if err != nil {
					return err
				}
				if r.code != http.StatusPaymentRequired || !strings.Contains(r.raw, "monthly") {
					return fmt.Errorf("the tool got %d: %s", r.code, r.raw)
				}
				return nil
			})
			sc.Step(`^the tool receives the broker's refusal as an OpenAI-shaped error$`, func() error {
				e, _ := s.tool.js["error"].(map[string]any)
				msg, _ := e["message"].(string)
				if s.tool.code != http.StatusPaymentRequired || !strings.Contains(msg, "monthly spend cap reached") {
					return fmt.Errorf("the tool got %d: %s", s.tool.code, s.tool.raw)
				}
				return nil
			})
			sc.Step(`^the response is 401 with an OpenAI-shaped "authentication_error"$`, func() error {
				e, _ := s.tool.js["error"].(map[string]any)
				if s.tool.code != 401 || e["type"] != "authentication_error" {
					return fmt.Errorf("the tool got %d: %s", s.tool.code, s.tool.raw)
				}
				return nil
			})
			sc.Step(`^the broker stand-in receives nothing$`, func() error {
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.relays) != 0 {
					return fmt.Errorf("the broker stand-in received %d relays", len(s.relays))
				}
				if s.lan && s.brokerHits != 0 {
					return fmt.Errorf("the broker stand-in was asked %d times in LAN mode", s.brokerHits)
				}
				return nil
			})
			sc.Step(`^the tool receives an OpenAI-shaped error saying no station is on air$`, func() error {
				e, _ := s.tool.js["error"].(map[string]any)
				msg, _ := e["message"].(string)
				// The proxy's own words for it (features/proxy/errors.feature) say "provider".
				if s.tool.code < 400 || e["type"] != "api_error" || !strings.Contains(msg, "no provider available") {
					return fmt.Errorf("the tool got %d: %s", s.tool.code, s.tool.raw)
				}
				return nil
			})
			sc.Step(`^GET /api/channel still reports "open" true with the same "base_url"$`, func() error {
				st, err := s.status()
				if err != nil {
					return err
				}
				if st["open"] != true || st["base_url"] != s.base {
					return fmt.Errorf("the status is %v", st)
				}
				return nil
			})
			sc.Step(`^the Edge session ledger gains one session attributed to the console$`, func() error {
				if s.ledger.Len() != 1 {
					return fmt.Errorf("the ledger holds %d sessions", s.ledger.Len())
				}
				return s.receiptedTo("qwen3-32b")
			})

			// --- the quote
			sc.Step(`^the response carries the price now in and out, the live range, the station count$`, func() error {
				for k, want := range map[string]float64{"price_in": 0.20, "price_out": 0.60, "range_min": 0.60, "range_max": 0.60, "stations": 1} {
					if got, ok := s.last.js[k].(float64); !ok || got != want {
						return fmt.Errorf("%q is %v, want %v: %s", k, s.last.js[k], want, s.last.raw)
					}
				}
				return nil
			})
			sc.Step(`^it carries the operator's cap for the band and whether the price is above it$`, func() error {
				if got, ok := s.last.js["cap_out"].(float64); !ok || got != client.ConsumerDefaultMaxOut {
					return fmt.Errorf(`"cap_out" is %v`, s.last.js["cap_out"])
				}
				if s.last.js["over_limit"] != false {
					return fmt.Errorf(`"over_limit" is %v`, s.last.js["over_limit"])
				}
				return nil
			})
			sc.Step(`^it mints nothing and binds nothing$`, s.nothingOpen)

			// --- the key
			sc.Step(`^the body does not contain the key$`, func() error { return s.noKeyIn("the body", s.last.raw) })
			sc.Step(`^no frame contains the key$`, func() error { return s.noKeyIn("the event stream", s.last.raw) })
			sc.Step(`^the response has "open" true, "base_url", "model" "([^"]*)", "station" and "key_hint"$`, func(m string) error {
				j := s.last.js
				if j["open"] != true || j["base_url"] != s.base || j["model"] != m || j["station"] == "" || j["key_hint"] != s.key[len(s.key)-4:] {
					return fmt.Errorf("the status is %s", s.last.raw)
				}
				return nil
			})
			sc.Step(`^the response carries the same "key" the open returned$`, func() error {
				if s.last.code != 200 || s.last.js["key"] != s.key {
					return fmt.Errorf("the key reply is %d", s.last.code)
				}
				return nil
			})
			sc.Step(`^the response has "Cache-Control: no-store"$`, func() error {
				if s.last.code != 200 || s.last.hdr.Get("Cache-Control") != "no-store" {
					return fmt.Errorf("%d with Cache-Control %q", s.last.code, s.last.hdr.Get("Cache-Control"))
				}
				return nil
			})
			sc.Step(`^no log line contains the key$`, func() error {
				if s.last.code != 200 {
					return fmt.Errorf("the flow did not run: %d %s", s.last.code, s.last.raw)
				}
				return s.noKeyIn("the log", s.logBuf.String())
			})
			sc.Step(`^the "key" differs from the earlier channel's key$`, func() error {
				if err := s.wantOK(true); err != nil {
					return err
				}
				if s.key == "" || s.key == s.oldKey {
					return fmt.Errorf("the key was reused")
				}
				return nil
			})
			sc.Step(`^the key does not start with "rog-key_" or "rog-grant_"$`, func() error {
				if strings.HasPrefix(s.key, "rog-key_") || strings.HasPrefix(s.key, "rog-grant_") {
					return fmt.Errorf("the session key looks like an account credential")
				}
				return nil
			})
			sc.Step(`^a request to the broker stand-in with that key as its bearer cannot spend$`, func() error {
				req, _ := http.NewRequest("POST", s.broker.URL+"/v1/chat/completions", strings.NewReader(`{"model":"qwen3-32b"}`))
				req.Header.Set("Authorization", "Bearer "+s.key)
				r, err := send(req)
				if err != nil {
					return err
				}
				if r.code != 401 {
					return fmt.Errorf("the broker stand-in answered %d", r.code)
				}
				return nil
			})

			// --- one channel
			sc.Step(`^"base_url" and "key" are unchanged$`, s.sameEndpoint)
			sc.Step(`^exactly one listener is bound$`, func() error {
				s.mu.Lock()
				n := s.binds
				s.mu.Unlock()
				if n != 1 {
					return fmt.Errorf("the endpoint was bound %d times", n)
				}
				return s.answers(s.base, s.key)
			})
			sc.Step(`^both responses carry the same "base_url" and "key"$`, func() error {
				a, b := s.last.js, s.second.js
				if a["ok"] != true || b["ok"] != true || a["base_url"] != b["base_url"] || a["key"] != b["key"] || a["key"] == nil {
					return fmt.Errorf("the two opens disagree: %s / %s", s.last.raw, s.second.raw)
				}
				return nil
			})
			sc.Step(`^GET /api/channel names the model the later one asked for$`, func() error {
				st, err := s.status()
				if err != nil {
					return err
				}
				m, _ := st["model"].(string)
				if m != "qwen3-32b" && m != "free-4b" {
					return fmt.Errorf("the status names %q", m)
				}
				r, err := s.toolCall("GET", "/models", s.key, "")
				if err != nil {
					return err
				}
				return s.toolLists(r, m)
			})
			sc.Step(`^GET /api/channel still names "([^"]*)" with the same "base_url" and "key_hint"$`, func(m string) error {
				st, err := s.status()
				if err != nil {
					return err
				}
				if st["model"] != m || st["base_url"] != s.base || st["key_hint"] != s.key[len(s.key)-4:] {
					return fmt.Errorf("the status is %v", st)
				}
				return nil
			})
			sc.Step(`^it shows the same plate$`, func() error {
				j := s.last.js
				if j["open"] != true || j["base_url"] != s.base || j["model"] != "qwen3-32b" {
					return fmt.Errorf("the second tab sees %s", s.last.raw)
				}
				return nil
			})
			sc.Step(`^its copy action receives the same key$`, func() error {
				r, err := s.do("POST", "/api/channel/key", "", s.tok)
				if err != nil {
					return err
				}
				if r.js["key"] != s.key {
					return fmt.Errorf("the second tab's key reply is %d", r.code)
				}
				return nil
			})
			sc.Step(`^the console's "base_url" names another port and its "key" differs from the TUI's$`, func() error {
				if err := s.wantOK(true); err != nil {
					return err
				}
				tui := "http://" + s.tuiLn.Addr().String() + "/v1"
				if s.base == tui || s.key == s.tuiKey {
					return fmt.Errorf("the console shares the TUI's endpoint or key")
				}
				return nil
			})
			sc.Step(`^the TUI's endpoint still answers with the TUI's key$`, func() error {
				if s.last.code != 200 {
					return fmt.Errorf("the console call is %d: %s", s.last.code, s.last.raw)
				}
				return s.answers("http://"+s.tuiLn.Addr().String()+"/v1", s.tuiKey)
			})
			sc.Step(`^the stream completes from "([^"]*)" and is receipted to it$`, func(m string) error {
				if err := s.wantOK(true); err != nil {
					return err
				}
				if _, err := s.finishStream(); err != nil {
					return err
				}
				return s.receiptedTo(m)
			})
			sc.Step(`^the next request goes to "([^"]*)"$`, func(m string) error {
				if _, err := s.chat(m, false); err != nil {
					return err
				}
				rel, err := s.lastRelay()
				if err != nil {
					return err
				}
				if rel.model != m {
					return fmt.Errorf("the next relay went to %q", rel.model)
				}
				return nil
			})

			// --- closing
			sc.Step(`^a connection to the old port is refused$`, func() error { return refused(s.base) })
			sc.Step(`^POST /api/channel/key is 409$`, func() error {
				r, err := s.do("POST", "/api/channel/key", "", s.tok)
				if err != nil {
					return err
				}
				if r.code != 409 {
					return fmt.Errorf("POST /api/channel/key is %d", r.code)
				}
				return nil
			})
			sc.Step(`^the stream completes and its turn is receipted$`, func() error {
				if err := s.wantCode(200); err != nil {
					return err
				}
				if _, err := s.finishStream(); err != nil {
					return err
				}
				return s.receiptedTo("qwen3-32b")
			})
			sc.Step(`^a new request to the old port is refused$`, func() error { return refused(s.base) })
			sc.Step(`^the endpoint still answers with the key$`, func() error { return s.answers(s.base, s.key) })
			sc.Step(`^the request body has no field that could carry a frequency code$`, func() error {
				r, err := s.do("POST", "/api/channel/open", `{"model":"hidden-8b","freq":"147.520 MHz 8F3K-9M2Q"}`, s.tok)
				if err != nil {
					return err
				}
				if r.code != 400 {
					return fmt.Errorf("an open carrying a frequency code is %d, want 400", r.code)
				}
				return nil
			})
			sc.Step(`^the response carries no Access-Control-Allow-Origin header$`, func() error {
				if s.last.hdr.Get("Access-Control-Allow-Origin") != "" {
					return fmt.Errorf("the reply allows origin %q", s.last.hdr.Get("Access-Control-Allow-Origin"))
				}
				if s.last.code == 200 || s.last.code == 204 {
					return fmt.Errorf("the preflight was answered %d", s.last.code)
				}
				return nil
			})

			// --- the page
			sc.Step(`^console\.js builds no URL that contains the key$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				if err := s.runPage("qwen3-32b", "copy", "", false); err != nil {
					return err
				}
				for _, f := range s.page.Fetches {
					if strings.Contains(f.Path, s.key) {
						return fmt.Errorf("the page requested %s", f.Path)
					}
				}
				return nil
			})
			sc.Step(`^the plate's copy action posts and reads a JSON body$`, func() error {
				for _, f := range s.page.Fetches {
					if f.Method == "POST" && f.Path == "/api/channel/key" {
						if !strings.Contains(s.page.Clipboard, s.key) {
							return fmt.Errorf("the copy did not carry the key from the reply body")
						}
						return nil
					}
				}
				return fmt.Errorf("the copy did not POST /api/channel/key: %+v", s.page.Fetches)
			})
			sc.Step(`^the plate shows bullets followed by the last 4 characters of the key$`, func() error {
				// The page opens it itself, so the reply it draws from is the one that carries the key.
				if err := s.runPage("qwen3-32b", "open", "", false); err != nil {
					return err
				}
				if !strings.Contains(s.page.Text, "••••"+s.key[len(s.key)-4:]) {
					return fmt.Errorf("the plate does not show the masked key:\n%s", s.page.Text)
				}
				return nil
			})
			sc.Step(`^the full key is in no element's text and no attribute$`, func() error {
				if err := s.noKeyIn("the plate's text", s.page.Text); err != nil {
					return err
				}
				return s.noKeyIn("the plate's attributes", s.page.Attrs)
			})
			sc.Step(`^console\.js writes the key to neither localStorage nor sessionStorage nor a cookie$`, func() error {
				src, err := plateAsset("console-channel.js")
				if err != nil {
					return err
				}
				for _, bad := range []string{"localStorage", "sessionStorage", "document.cookie", "indexedDB"} {
					if strings.Contains(src, bad) {
						return fmt.Errorf("console-channel.js uses %s", bad)
					}
				}
				return nil
			})
			sc.Step(`^the CHAT tab shows an "open an endpoint" control naming "([^"]*)"$`, func(m string) error {
				html, err := plateAsset("console.html")
				if err != nil {
					return err
				}
				chat := html[strings.Index(html, `id="panel-chat"`):]
				chat = chat[:strings.Index(chat, "</section>")]
				if !strings.Contains(chat, `id="channel-plate"`) {
					return fmt.Errorf("the CHAT panel has no #channel-plate")
				}
				for _, b := range s.page.Buttons {
					if strings.Contains(b, "open an endpoint") && strings.Contains(b, m) {
						return nil
					}
				}
				return fmt.Errorf("no open control names %q: %v", m, s.page.Buttons)
			})
			sc.Step(`^it shows the price now and the operator's cap beside it$`, func() error {
				for _, w := range []string{"0.60", "1.00"} {
					if !strings.Contains(s.page.Text, w) {
						return fmt.Errorf("the closed plate does not show %s:\n%s", w, s.page.Text)
					}
				}
				return nil
			})
			sc.Step(`^the plate shows three rows labelled "BASE URL", "API KEY" and "MODEL"$`, func() error {
				if err := s.runPage("qwen3-32b", "shown", "", false); err != nil {
					return err
				}
				if strings.Join(s.page.Rows, "|") != "BASE URL|API KEY|MODEL" {
					return fmt.Errorf("the rows are %v", s.page.Rows)
				}
				return nil
			})
			sc.Step(`^a line "([^"]*)"$`, func(line string) error {
				if !strings.Contains(s.page.Text, line) {
					return fmt.Errorf("the plate does not say %q:\n%s", line, s.page.Text)
				}
				return nil
			})
			sc.Step(`^a "copy" control and a "close" control$`, func() error {
				if strings.Join(s.page.Buttons, "|") != "copy|close" {
					return fmt.Errorf("the controls are %v", s.page.Buttons)
				}
				return nil
			})
			sc.Step(`^the operator presses "copy"$`, func() error {
				return s.runPage("qwen3-32b", "copy", "", s.clipRefuses)
			})
			sc.Step(`^the clipboard holds three lines:$`, func(doc *godog.DocString) error {
				want := strings.NewReplacer("<base_url>", s.base, "<key>", s.key).Replace(doc.Content)
				if s.page.Clipboard != want {
					return fmt.Errorf("the clipboard holds:\n%s\nwant:\n%s", s.page.Clipboard, want)
				}
				return nil
			})
			sc.Step(`^a toast says it was copied$`, func() error { return s.toastHas("copied") })
			sc.Step(`^a toast says the copy failed$`, func() error { return s.toastHas("copy failed") })
			sc.Step(`^the plate still shows the key masked$`, func() error {
				if !strings.Contains(s.page.Text, "••••"+s.key[len(s.key)-4:]) {
					return fmt.Errorf("the plate lost the masked key:\n%s", s.page.Text)
				}
				return s.noKeyIn("the plate's text", s.page.Text)
			})
			sc.Step(`^the full key is still in no element's text$`, func() error {
				if s.page.Clipboard != "" {
					return fmt.Errorf("the refused clipboard holds something")
				}
				return s.noKeyIn("the plate's text", s.page.Text+s.page.Attrs+strings.Join(s.page.Toasts, " "))
			})
			sc.Step(`^the plate says the endpoint works on this machine only, while roger is running$`, func() error {
				if err := s.runPage("qwen3-32b", "shown", "", false); err != nil {
					return err
				}
				for _, w := range []string{"this machine only", "while roger is running"} {
					if !strings.Contains(s.page.Text, w) {
						return fmt.Errorf("the plate does not say %q:\n%s", w, s.page.Text)
					}
				}
				return nil
			})
			sc.Step(`^it says a bot on a server needs an account key, naming "([^"]*)"$`, func(cmd string) error {
				for _, w := range []string{"bot on a server", "account key", cmd} {
					if !strings.Contains(s.page.Text, w) {
						return fmt.Errorf("the plate does not say %q:\n%s", w, s.page.Text)
					}
				}
				return nil
			})
			sc.Step(`^an open is refused with "message" "([^"]*)"$`, func(msg string) error {
				canned, _ := json.Marshal(map[string]any{"ok": false, "login_needed": true, "message": msg})
				return s.runPage("qwen3-32b", "open", string(canned), false)
			})
			sc.Step(`^the plate shows that message$`, func() error {
				const msg = "this band is paid - log in on ACCOUNT to use your wallet"
				if !strings.Contains(s.page.Text, msg) {
					return fmt.Errorf("the plate does not show the server's message:\n%s", s.page.Text)
				}
				if !strings.Contains(s.page.After, msg) {
					return fmt.Errorf("the plate's timer took the message away:\n%s", s.page.After)
				}
				if len(s.page.TickFetches) == 0 {
					return fmt.Errorf("the plate was handed no timer")
				}
				for _, f := range s.page.TickFetches {
					if f.Path != "/api/channel" {
						return fmt.Errorf("the timer requested %s", f.Path)
					}
				}
				return nil
			})
			sc.Step(`^it offers the ACCOUNT tab when "login_needed" is true$`, func() error {
				for _, b := range s.page.Buttons {
					if strings.Contains(b, "ACCOUNT") {
						return nil
					}
				}
				return fmt.Errorf("no control offers ACCOUNT: %v", s.page.Buttons)
			})
			sc.Step(`^one tab closes the channel$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				return s.runPage("qwen3-32b", "twotabs", "", false)
			})
			sc.Step(`^the other tab shows no plate within 2 seconds$`, func() error {
				if strings.Contains(s.page.After, "BASE URL") || !strings.Contains(s.page.Text, "BASE URL") {
					return fmt.Errorf("first tab:\n%s\nother tab after the close:\n%s", s.page.Text, s.page.After)
				}
				if s.page.PollMs <= 0 || s.page.PollMs > 2000 {
					return fmt.Errorf("the plate re-reads every %d ms", s.page.PollMs)
				}
				src, _ := plateAsset("console.js")
				if !strings.Contains(src, "RogerConsoleChannel.mount") || !strings.Contains(src, "every: function (fn, ms) { setInterval(") {
					return fmt.Errorf("console.js does not hand the plate a timer")
				}
				return nil
			})
			sc.Step(`^console\.html and console\.js reference no third-party script, style or font for the plate$`, func() error {
				src, err := plateAsset("console-channel.js")
				if err != nil {
					return err
				}
				if regexp.MustCompile(`https?://`).MatchString(src) {
					return fmt.Errorf("console-channel.js names an absolute URL")
				}
				html, _ := plateAsset("console.html")
				if !strings.Contains(html, `<script src="/assets/console-channel.js" defer></script>`) {
					return fmt.Errorf("console.html does not load the plate from its own assets")
				}
				return nil
			})
			sc.Step(`^the copy and close controls are buttons reachable by Tab with visible labels$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				if err := s.runPage("qwen3-32b", "shown", "", false); err != nil {
					return err
				}
				if strings.Join(s.page.Buttons, "|") != "copy|close" {
					return fmt.Errorf("the controls are %v", s.page.Buttons)
				}
				if strings.Contains(s.page.Attrs, "-1") {
					return fmt.Errorf("a control is taken out of the tab order")
				}
				return nil
			})
			sc.Step(`^the open and closed states differ by text, not by colour alone$`, func() error {
				open := s.page.Text
				if err := s.page_("POST", "/api/channel/close", ""); err != nil {
					return err
				}
				if err := s.runPage("qwen3-32b", "closed", "", false); err != nil {
					return err
				}
				if !strings.Contains(open, "channel open") || strings.Contains(s.page.Text, "channel open") {
					return fmt.Errorf("open:\n%s\nclosed:\n%s", open, s.page.Text)
				}
				return nil
			})
			sc.Step(`^the masked key has an accessible name that does not contain the key$`, func() error {
				if err := s.open("qwen3-32b"); err != nil {
					return err
				}
				if err := s.runPage("qwen3-32b", "shown", "", false); err != nil {
					return err
				}
				if !strings.Contains(s.page.Attrs, "API key ending in "+s.key[len(s.key)-4:]) {
					return fmt.Errorf("the masked key has no accessible name:\n%s", s.page.Attrs)
				}
				return s.noKeyIn("the accessible name", s.page.Attrs)
			})
		},
		Options: &godog.Options{Format: "pretty", TestingT: t, Strict: true,
			Paths: []string{"../../features/console/endpoint_plate.feature"}},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("console endpoint plate scenarios failed")
	}
}

// untouched: the refused request changed nothing. With a channel open beforehand it is still
// open on the same endpoint and key and the refusal did not carry the key; with none, none.
func (s *plateBDD) untouched() error {
	if s.key == "" {
		return s.nothingOpen()
	}
	if strings.Contains(s.last.raw, s.key) || strings.Contains(s.last.raw, s.base) {
		return fmt.Errorf("the refusal carried the plate: %s", s.last.raw)
	}
	st, err := s.status()
	if err != nil {
		return err
	}
	if st["open"] != true || st["base_url"] != s.base {
		return fmt.Errorf("the channel changed: %v", st)
	}
	return s.answers(s.base, s.key)
}

func (s *plateBDD) toolLists(r plateReply, m string) error {
	data, _ := r.js["data"].([]any)
	if r.code != 200 || len(data) != 1 {
		return fmt.Errorf("the model list is %d: %s", r.code, r.raw)
	}
	if first, _ := data[0].(map[string]any); first["id"] != m {
		return fmt.Errorf("the endpoint lists %v, want %q", first["id"], m)
	}
	return nil
}

func (s *plateBDD) sameEndpoint() error {
	if s.last.js["base_url"] != s.base || s.last.js["key"] != s.key || s.key == "" {
		return fmt.Errorf("the endpoint changed: %s", s.last.raw)
	}
	return nil
}

func (s *plateBDD) statusSameEndpoint() error {
	st, err := s.status()
	if err != nil {
		return err
	}
	if st["base_url"] != s.base {
		return fmt.Errorf("the port changed: %v", st["base_url"])
	}
	return s.answers(s.base, s.key)
}

func (s *plateBDD) toastHas(w string) error {
	for _, t := range s.page.Toasts {
		if strings.Contains(strings.ToLower(t), w) {
			return nil
		}
	}
	return fmt.Errorf("no toast says %q: %v", w, s.page.Toasts)
}

func (s *plateBDD) twoAtOnce(a, b string) error {
	var wg sync.WaitGroup
	var ra, rb plateReply
	wg.Add(2)
	go func() { defer wg.Done(); ra, _ = s.do("POST", "/api/channel/open", `{"model":"`+a+`"}`, s.tok) }()
	go func() { defer wg.Done(); rb, _ = s.do("POST", "/api/channel/open", `{"model":"`+b+`"}`, s.tok) }()
	wg.Wait()
	s.last, s.second = ra, rb
	s.base, _ = ra.js["base_url"].(string)
	s.key, _ = ra.js["key"].(string)
	return nil
}
