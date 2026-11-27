package webui

// Executable spec: features/console/account_keys.feature - account keys and grants on the
// console's ACCOUNT tab. The REAL console server over HTTP, a broker stand-in on loopback that
// serves /balance, /account/keys and /grants, and the shipped console-keys.js EXECUTED with
// node against that server. No mocks. Every secret here is made up.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/node"
)

type keysGrant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Free   bool   `json:"free"`
	Price  string `json:"price"`
	Status string `json:"status"`
}

type keysChange struct {
	id     string
	fields map[string]any
}

type keysRow struct {
	Name    string
	Text    string
	Buttons []string
}

type keysInput struct {
	Value    string
	Checked  bool
	Disabled bool
	Hidden   bool
}

type keysPage struct {
	Keys      string
	Grants    string
	KeyRows   []keysRow
	GrantRows []keysRow
	Inputs    map[string]keysInput
	Clipboard string
	Fetches   []struct{ Method, Path, Body string }
	Tags      []string
}

type keysBDD struct {
	t *testing.T

	mu            sync.Mutex
	loggedIn      bool
	keys          []client.AccountKey
	grants        []keysGrant
	mints         []map[string]any
	changes       []keysChange
	revokes       []string
	grantCreates  []map[string]any
	grantRevokes  []string
	secrets       []string
	hits          int
	slow          bool
	refuseChange  bool
	grantsRefused bool
	keysLoggedOut bool

	broker  *httptest.Server
	console *httptest.Server
	srv     *Server
	tok     string
	lan     bool

	ops       [][]string
	usePage   bool
	ran       bool
	page      keysPage
	clipFails bool

	last     plateReply
	hitsWas  int
	logBuf   *bytes.Buffer
	stdout   *os.File
	realOut  *os.File
	plateTxt string

	restoreLogout func() error
}

const keysGrantReason = "grants require a linked operator account - sign in (GitHub or Apple) first"

// ---- the broker stand-in ---------------------------------------------------------------

func keysErr(w http.ResponseWriter, code int, slug, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": slug, "message": msg}})
}

func (s *keysBDD) startBroker() {
	s.broker = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fields map[string]any
		_ = json.Unmarshal(body, &fields)
		// Who is asking is settled when the request arrives, as on the real broker; a slow
		// answer does not change it.
		s.mu.Lock()
		slow, loggedIn := s.slow, s.loggedIn
		s.mu.Unlock()
		if slow && r.Method == http.MethodPost {
			time.Sleep(300 * time.Millisecond)
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.hits++
		w.Header().Set("Content-Type", "application/json")
		id := ""
		for _, pre := range []string{"/account/keys/", "/grants/"} {
			if rest, ok := strings.CutPrefix(r.URL.Path, pre); ok {
				id = rest
			}
		}
		switch {
		case r.URL.Path == "/balance":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": "op", "balance": 4.0, "logged_in": s.loggedIn})
		case strings.HasPrefix(r.URL.Path, "/account/keys") && (!loggedIn || s.keysLoggedOut):
			keysErr(w, http.StatusUnauthorized, "", "log in to manage keys - run `roger login` (keys are per account)")
		case r.URL.Path == "/account/keys" && r.Method == http.MethodGet:
			// A broker that wrongly listed a secret: the console must not pass it on.
			var list []map[string]any
			raw, _ := json.Marshal(s.keys)
			_ = json.Unmarshal(raw, &list)
			for _, k := range list {
				k["secret"] = "rog-key_TESTONLYleaked"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": list})
		case r.URL.Path == "/account/keys" && r.Method == http.MethodPost:
			if len(s.keys) >= 32 {
				keysErr(w, http.StatusBadRequest, "key_limit_count", "at most 32 keys per account - delete one first")
				return
			}
			s.mints = append(s.mints, fields)
			secret := fmt.Sprintf("rog-key_TESTONLY%04dzz%02d", len(s.secrets)+1, len(s.secrets)+1)
			s.secrets = append(s.secrets, secret)
			k := client.AccountKey{ID: fmt.Sprintf("key_new%d", len(s.secrets)), Hint: "..." + secret[len(secret)-4:], Reset: "none"}
			k.Name, _ = fields["name"].(string)
			k.LimitUSD, _ = fields["limit_usd"].(float64)
			if rs, _ := fields["reset"].(string); rs != "" {
				k.Reset = rs
			}
			s.keys = append([]client.AccountKey{k}, s.keys...)
			w.WriteHeader(http.StatusCreated)
			out := map[string]any{}
			raw, _ := json.Marshal(k)
			_ = json.Unmarshal(raw, &out)
			out["secret"] = secret
			_ = json.NewEncoder(w).Encode(out)
		case id != "" && strings.HasPrefix(r.URL.Path, "/account/keys/"):
			at := -1
			for i, k := range s.keys {
				if k.ID == id {
					at = i
				}
			}
			if at < 0 {
				keysErr(w, http.StatusNotFound, "not_found", "no such key")
				return
			}
			if r.Method == http.MethodDelete {
				s.revokes = append(s.revokes, id)
				s.keys = append(s.keys[:at], s.keys[at+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			s.changes = append(s.changes, keysChange{id: id, fields: fields})
			if s.refuseChange {
				keysErr(w, http.StatusBadRequest, "bad_limit", "the limit is above what this account may set")
				return
			}
			k := &s.keys[at]
			if v, ok := fields["limit_usd"].(float64); ok {
				k.LimitUSD = v
			}
			if v, ok := fields["reset"].(string); ok {
				k.Reset = v
			}
			if v, ok := fields["disabled"].(bool); ok {
				k.Disabled = v
			}
			_ = json.NewEncoder(w).Encode(k)
		case strings.HasPrefix(r.URL.Path, "/grants") && s.grantsRefused:
			keysErr(w, http.StatusForbidden, "", keysGrantReason)
		case r.URL.Path == "/grants" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"grants": s.grants})
		case r.URL.Path == "/grants" && r.Method == http.MethodPost:
			s.grantCreates = append(s.grantCreates, fields)
			name, _ := fields["name"].(string)
			secret := fmt.Sprintf("rog-grant_TESTONLY%04dqq%02d", len(s.secrets)+1, len(s.secrets)+1)
			s.secrets = append(s.secrets, secret)
			s.grants = append(s.grants, keysGrant{ID: fmt.Sprintf("gr_%d", len(s.grants)), Name: name, Free: true, Price: "free", Status: "active"})
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "secret": secret})
		case id != "" && r.Method == http.MethodDelete:
			for i := range s.grants {
				if s.grants[i].ID == id {
					if s.grants[i].Status != "revoked" {
						s.grantRevokes = append(s.grantRevokes, s.grants[i].Name)
					}
					s.grants[i].Status = "revoked"
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "revoked": true})
		default:
			http.NotFound(w, r)
		}
	}))
}

func (s *keysBDD) startConsole() {
	if s.console != nil {
		s.console.Close()
	}
	opts := Options{User: "op", Broker: s.broker.URL}
	if s.lan {
		opts.LANDial = func() client.LANDial { return client.LANDial{} }
	}
	s.srv = New(node.New(node.Config{Station: "amber-fox"}), opts)
	s.srv.token = s.tok
	s.console = httptest.NewServer(s.srv.Handler())
}

func (s *keysBDD) reset() {
	s.cleanup()
	*s = keysBDD{t: s.t, tok: "tok-run", loggedIn: true, realOut: os.Stdout}
	// Logging out on the console is what logs the account out on the stand-in.
	s.restoreLogout = logoutReturn
	logoutReturn = func() error { s.mu.Lock(); s.loggedIn = false; s.mu.Unlock(); return nil }
	s.logBuf = &bytes.Buffer{}
	log.SetOutput(s.logBuf)
	if f, err := os.CreateTemp(s.t.TempDir(), "stdout"); err == nil {
		s.stdout = f
		os.Stdout = f
	}
	s.startBroker()
}

func (s *keysBDD) cleanup() {
	if s.realOut != nil {
		os.Stdout = s.realOut
	}
	if s.restoreLogout != nil {
		logoutReturn = s.restoreLogout
	}
	if s.logBuf != nil {
		log.SetOutput(os.Stderr)
	}
	if s.console != nil {
		s.console.Close()
	}
	if s.broker != nil {
		s.broker.Close()
	}
}

func (s *keysBDD) do(method, path, body, tok string) (plateReply, error) {
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

// call makes one direct route call and remembers how many requests the stand-in had seen.
func (s *keysBDD) call(method, path, body, tok string) error {
	s.mu.Lock()
	s.hitsWas = s.hits
	s.mu.Unlock()
	r, err := s.do(method, path, body, tok)
	s.last = r
	return err
}

func (s *keysBDD) addKey(k client.AccountKey) {
	if k.Hint == "" {
		k.Hint = "...0000"
	}
	if k.Reset == "" {
		k.Reset = "none"
	}
	s.mu.Lock()
	s.keys = append(s.keys, k)
	s.mu.Unlock()
}

// ---- the page, executed --------------------------------------------------------------

const keysPageScript = `
const api = require(process.argv[1]);
const [base, tok, opsJSON, refuse] = process.argv.slice(2);
const ops = JSON.parse(opsJSON);
const out = { keys: "", grants: "", keyRows: [], grantRows: [], inputs: {}, clipboard: "", fetches: [], tags: [] };
let inflight = 0;
function el(tag) {
  out.tags.push(tag);
  return { tag, children: [], attrs: {}, hidden: false, textContent: "", className: "", value: "", checked: false, disabled: false, type: "", listeners: {},
    appendChild(c) { this.children.push(c); return c; },
    setAttribute(k, v) { this.attrs[k] = String(v); }, getAttribute(k) { return this.attrs[k]; },
    addEventListener(t, f) { (this.listeners[t] = this.listeners[t] || []).push(f); },
    replaceChildren(...cs) { this.children = cs; } };
}
function fire(e, t) { (e.listeners[t] || []).forEach((f) => f({ target: e })); }
function walk(e, f) { if (e.hidden) return; f(e); (e.children || []).forEach((c) => walk(c, f)); }
function has(e, cls) { return String(e.className || "").split(" ").includes(cls); }
function find(root, pred) { let hit = null; walk(root, (e) => { if (!hit && pred(e)) hit = e; }); return hit; }
function mkDoc() {
  const hosts = { "card-keys-body": el("div"), "card-grants-body": el("div") };
  return { hosts, getElementById: (id) => hosts[id] || null, createElement: el };
}
function text(root) { const t = []; walk(root, (e) => { if (e.textContent) t.push(e.textContent); }); return t.join("\n"); }
function rows(root, cls) {
  const r = [];
  walk(root, (e) => {
    if (!has(e, cls)) return;
    const b = []; walk(e, (c) => { if (c.tag === "button") b.push(c.textContent); });
    r.push({ name: e.attrs["data-name"] || "", text: text(e), buttons: b });
  });
  return r;
}
function row(doc, name, withButton) {
  for (const h of Object.values(doc.hosts)) {
    const r = find(h, (e) => (has(e, "key-row") || has(e, "grant-row")) && e.attrs["data-name"] === name &&
      (!withButton || find(e, (c) => c.tag === "button" && c.textContent === withButton)));
    if (r) return r;
  }
  throw new Error("no row named " + name);
}
function button(root, label) {
  const b = find(root, (e) => e.tag === "button" && e.textContent === label);
  if (!b) throw new Error("no button " + label + " in: " + text(root));
  return b;
}
function anyButton(doc, label) {
  for (const h of Object.values(doc.hosts)) {
    const b = find(h, (e) => e.tag === "button" && e.textContent === label);
    if (b) return b;
  }
  throw new Error("no button " + label);
}
function field(root, cls) {
  const f = find(root, (e) => has(e, cls));
  if (!f) throw new Error("no field " + cls);
  return f;
}
function anyField(doc, cls) {
  for (const h of Object.values(doc.hosts)) { const f = find(h, (e) => has(e, cls)); if (f) return f; }
  throw new Error("no field " + cls);
}
function press(b) { if (!b.disabled) fire(b, "click"); }
function type(f, v) { if (f.disabled) return; f.value = v; fire(f, "input"); fire(f, "change"); }
const actions = {
  fetch: (path, opts) => {
    opts = opts || {};
    out.fetches.push({ method: opts.method || "GET", path, body: opts.body || "" });
    const h = Object.assign({ "X-Roger-Token": tok }, opts.headers || {});
    inflight++;
    return fetch(base + path, Object.assign({}, opts, { headers: h })).finally(() => { inflight--; });
  },
  clipboard: { writeText: (t) => (refuse === "1" ? Promise.reject(new Error("denied")) : (out.clipboard = t, Promise.resolve())) },
};
async function settle() {
  for (let i = 0; i < 400; i++) {
    await new Promise((r) => setTimeout(r, 5));
    if (inflight === 0) { await new Promise((r) => setTimeout(r, 5)); if (inflight === 0) return; }
  }
  throw new Error("the page never settled");
}
(async () => {
  // The console mounts the cards at boot and loads them when ACCOUNT is opened.
  let doc = mkDoc(), card = await api.mount(doc, actions);
  if (out.fetches.length) throw new Error("mount asked the console before the tab was opened");
  await card.load();
  await settle();
  for (const op of ops) {
    const [what, a, b, c] = op;
    if (what === "set") type(anyField(doc, a), b);
    else if (what === "tick") { const f = anyField(doc, a); f.checked = true; fire(f, "change"); }
    else if (what === "click") press(anyButton(doc, a));
    else if (what === "click2") { const x = anyButton(doc, a); press(x); press(x); }
    else if (what === "rowset") type(field(row(doc, a), b), c);
    else if (what === "rowclick") press(button(row(doc, a, b), b));
    else if (what === "race") {
      // Press, then leave while the answer is still on its way.
      press(anyButton(doc, a));
      await new Promise((r) => setTimeout(r, 100));
      if (b === "logout") await actions.fetch("/api/account/logout", { method: "POST" });
      card.leave(); await card.load();
    }
    else if (what === "leave") card.leave();
    else if (what === "load") await card.load();
    else if (what === "remount") { doc = mkDoc(); card = await api.mount(doc, actions); await card.load(); }
    else if (what === "logout") { await actions.fetch("/api/account/logout", { method: "POST" }); card.leave(); await card.load(); }
    else throw new Error("unknown op " + what);
    await settle();
  }
  const k = doc.hosts["card-keys-body"], g = doc.hosts["card-grants-body"];
  out.keys = text(k); out.grants = text(g);
  out.keyRows = rows(k, "key-row"); out.grantRows = rows(g, "grant-row");
  for (const h of [k, g]) (function all(e, hid) {
    hid = hid || e.hidden;
    String(e.className || "").split(" ").filter(Boolean).forEach((cls) => {
      if ((e.tag === "input" || e.tag === "select") && !(cls in out.inputs)) out.inputs[cls] = { value: e.value, checked: e.checked, disabled: e.disabled, hidden: hid };
    });
    (e.children || []).forEach((ch) => all(ch, hid));
  })(h, false);
  process.stdout.write(JSON.stringify(out), () => process.exit(0));
})().catch((e) => { console.error(e); process.exit(2); });
`

// run executes the page once with every step the operator took. Later calls reuse the result.
func (s *keysBDD) run() error {
	if s.ran {
		return nil
	}
	nodeBin, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("node is needed to execute the console's JavaScript: %v", err)
	}
	ops := s.ops
	if ops == nil {
		ops = [][]string{}
	}
	raw, _ := json.Marshal(ops)
	ref := "0"
	if s.clipFails {
		ref = "1"
	}
	cmd := exec.Command(nodeBin, "-e", keysPageScript, "./assets/console-keys.js", s.console.URL, s.tok, string(raw), ref)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("console-keys.js failed: %v: %s", err, b)
	}
	s.ran = true
	s.page = keysPage{}
	return json.Unmarshal(b, &s.page)
}

func (s *keysBDD) op(parts ...string) { s.ops = append(s.ops, parts); s.ran = false }

func (s *keysBDD) keyRow(name string) (keysRow, error) {
	if err := s.run(); err != nil {
		return keysRow{}, err
	}
	for _, r := range s.page.KeyRows {
		if r.Name == name {
			return r, nil
		}
	}
	return keysRow{}, fmt.Errorf("no key row %q; the card reads:\n%s", name, s.page.Keys)
}

func (s *keysBDD) grantRow(name string) (keysRow, error) {
	if err := s.run(); err != nil {
		return keysRow{}, err
	}
	for _, r := range s.page.GrantRows {
		if r.Name == name {
			return r, nil
		}
	}
	return keysRow{}, fmt.Errorf("no grant row %q; the card reads:\n%s", name, s.page.Grants)
}

// onlyKey is the single key row a scenario set up.
func (s *keysBDD) onlyKey() (keysRow, error) {
	if err := s.run(); err != nil {
		return keysRow{}, err
	}
	if len(s.page.KeyRows) != 1 {
		return keysRow{}, fmt.Errorf("%d key rows, want 1:\n%s", len(s.page.KeyRows), s.page.Keys)
	}
	return s.page.KeyRows[0], nil
}

func (s *keysBDD) card(name string) (string, error) {
	if err := s.run(); err != nil {
		return "", err
	}
	if name == "Grants" {
		return s.page.Grants, nil
	}
	return s.page.Keys, nil
}

func (s *keysBDD) cardSays(name, want string) error {
	txt, err := s.card(name)
	if err != nil {
		return err
	}
	if !strings.Contains(txt, want) {
		return fmt.Errorf("the %q card does not say %q:\n%s", name, want, txt)
	}
	return nil
}

func (s *keysBDD) secret() (string, error) {
	if err := s.run(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.secrets) == 0 {
		return "", fmt.Errorf("the stand-in issued no secret")
	}
	return s.secrets[len(s.secrets)-1], nil
}

func (s *keysBDD) lastMint() (map[string]any, error) {
	if s.usePage {
		if err := s.run(); err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.mints) == 0 {
		return nil, fmt.Errorf("the broker received no mint (last reply %d %s)", s.last.code, s.last.raw)
	}
	return s.mints[len(s.mints)-1], nil
}

func (s *keysBDD) changeOnly(id string, want map[string]any) error {
	if err := s.run(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.changes) != 1 {
		return fmt.Errorf("the broker received %d changes, want 1; the card reads:\n%s", len(s.changes), s.page.Keys)
	}
	c := s.changes[0]
	if c.id != id {
		return fmt.Errorf("the change went to %q, want %q", c.id, id)
	}
	if fmt.Sprint(c.fields) != fmt.Sprint(want) {
		return fmt.Errorf("the change carried %v, want only %v", c.fields, want)
	}
	return nil
}

func (s *keysBDD) pageSent(method, path string) int {
	n := 0
	for _, f := range s.page.Fetches {
		if f.Method == method && f.Path == path {
			n++
		}
	}
	return n
}

func (s *keysBDD) brokerNotAsked() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hits != s.hitsWas {
		return fmt.Errorf("the broker stand-in saw %d requests", s.hits-s.hitsWas)
	}
	return nil
}

func (s *keysBDD) exportBlock(lines string) error {
	sec, err := s.secret()
	if err != nil {
		return err
	}
	want := strings.NewReplacer("<broker>", s.broker.URL, "<secret>", sec).Replace(lines)
	if s.page.Clipboard != want {
		return fmt.Errorf("the clipboard holds %q, want %q", s.page.Clipboard, want)
	}
	return nil
}

func TestConsoleAccountKeysFeature(t *testing.T) {
	s := &keysBDD{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.cleanup()
				s.console, s.broker, s.logBuf, s.realOut, s.restoreLogout = nil, nil, nil, nil, nil
				return ctx, nil
			})

			// --- background
			sc.Step(`^the console is running against a broker$`, func() error { s.startConsole(); return nil })
			sc.Step(`^the account is logged in$`, func() error { s.loggedIn = true; return nil })
			sc.Step(`^the operator has the ACCOUNT tab open$`, func() error { s.usePage = true; return nil })
			sc.Step(`^the ACCOUNT tab loads$`, func() error { return s.run() })

			// --- givens: keys
			sc.Step(`^the account has the account keys$`, func(tb *godog.Table) error {
				for _, r := range tb.Rows[1:] {
					limit, _ := strconv.ParseFloat(r.Cells[3].Value, 64)
					used, _ := strconv.ParseFloat(r.Cells[5].Value, 64)
					s.addKey(client.AccountKey{ID: r.Cells[0].Value, Name: r.Cells[1].Value, Hint: r.Cells[2].Value, LimitUSD: limit, Reset: r.Cells[4].Value,
						Usage: used, UsageDaily: used, UsageWeekly: used, UsageMonthly: used})
				}
				return nil
			})
			sc.Step(`^the account has no account keys$`, func() error { return nil })
			sc.Step(`^the account has account keys$`, func() error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci", LimitUSD: 5, Reset: "weekly"})
				s.addKey(client.AccountKey{ID: "key_bbbb2", Name: "bot"})
				return nil
			})
			sc.Step(`^the account has an account key with limit ([\d.]+) and reset "([^"]*)"$`, func(limit float64, reset string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci", LimitUSD: limit, Reset: reset})
				return nil
			})
			sc.Step(`^the account has an account key with reset "([^"]*)" that spent 1 today, 2 this week, 3 this month and 4 in all$`, func(reset string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci", LimitUSD: 9, Reset: reset, UsageDaily: 1, UsageWeekly: 2, UsageMonthly: 3, Usage: 4})
				return nil
			})
			sc.Step(`^the account has an account key that is (disabled|expired|disabled and expired|neither)$`, func(cond string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci", Disabled: strings.Contains(cond, "disabled"), Expired: strings.Contains(cond, "expired")})
				return nil
			})
			sc.Step(`^the broker returns an account key named "(.*)"$`, func(name string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: name})
				return nil
			})
			sc.Step(`^the account already has 32 account keys$`, func() error {
				for i := 0; i < 32; i++ {
					s.addKey(client.AccountKey{ID: fmt.Sprintf("key_full%02d", i), Name: fmt.Sprintf("k%02d", i)})
				}
				return nil
			})
			sc.Step(`^the account has an account key "([^"]*)" named "([^"]*)" with limit ([\d.]+) (daily|weekly|monthly)$`, func(id, name string, limit float64, reset string) error {
				s.addKey(client.AccountKey{ID: id, Name: name, LimitUSD: limit, Reset: reset})
				return nil
			})
			sc.Step(`^the account has an account key "([^"]*)" named "([^"]*)"$`, func(id, name string) error {
				s.addKey(client.AccountKey{ID: id, Name: name, LimitUSD: 5, Reset: "weekly"})
				return nil
			})
			sc.Step(`^the account has an account key "([^"]*)" that is (active|disabled)$`, func(id, state string) error {
				s.addKey(client.AccountKey{ID: id, Name: "ci", LimitUSD: 5, Reset: "weekly", Disabled: state == "disabled"})
				return nil
			})
			sc.Step(`^the account has an account key named "([^"]*)"$`, func(name string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: name, LimitUSD: 5, Reset: "weekly"})
				return nil
			})
			sc.Step(`^the broker refuses the change$`, func() error { s.refuseChange = true; return nil })
			sc.Step(`^the broker is slow to answer$`, func() error { s.slow = true; return nil })
			sc.Step(`^the browser refuses clipboard writes$`, func() error { s.clipFails = true; return nil })
			sc.Step(`^the account is logged out$`, func() error { s.loggedIn = false; return nil })
			sc.Step(`^the broker is unreachable$`, func() error { s.broker.Close(); return nil })
			sc.Step(`^the console is in LAN mode$`, func() error { s.lan = true; s.startConsole(); return nil })

			// --- givens: grants
			sc.Step(`^the account has the grants$`, func(tb *godog.Table) error {
				for _, r := range tb.Rows[1:] {
					s.grants = append(s.grants, keysGrant{ID: fmt.Sprintf("gr_%d", len(s.grants)), Name: r.Cells[0].Value, Free: r.Cells[1].Value == "free", Price: r.Cells[1].Value, Status: r.Cells[2].Value})
				}
				return nil
			})
			sc.Step(`^the account has a grant named "([^"]*)"$`, func(name string) error {
				s.grants = append(s.grants, keysGrant{ID: fmt.Sprintf("gr_%d", len(s.grants)), Name: name, Free: true, Price: "free", Status: "active"})
				return nil
			})
			sc.Step(`^the account has a grant named "([^"]*)" that is revoked$`, func(name string) error {
				s.grants = append(s.grants, keysGrant{ID: fmt.Sprintf("gr_%d", len(s.grants)), Name: name, Free: true, Price: "free", Status: "revoked"})
				return nil
			})
			sc.Step(`^the broker refuses grants for this account$`, func() error { s.grantsRefused = true; return nil })

			// --- what the operator does
			mint := func(name, limit string) {
				s.op("set", "key-name", name)
				s.op("set", "key-limit", limit)
				s.op("click", "Mint key")
			}
			sc.Step(`^the operator enters the name "([^"]*)", the limit "([^"]*)" and the window "([^"]*)" and presses "Mint key"$`, func(name, limit, window string) error {
				s.op("set", "key-name", name)
				s.op("set", "key-limit", limit)
				s.op("set", "key-reset", window)
				s.op("click", "Mint key")
				return nil
			})
			sc.Step(`^the operator enters the name "([^"]*)" and the limit "([^"]*)" and presses "Mint key"$`, func(name, limit string) error {
				mint(name, limit)
				return nil
			})
			sc.Step(`^the operator enters the name "([^"]*)", ticks "unlimited" and presses "Mint key"$`, func(name string) error {
				s.op("set", "key-name", name)
				s.op("tick", "key-nolimit")
				s.op("click", "Mint key")
				return nil
			})
			sc.Step(`^the operator enters the limit "([^"]*)" and ticks "unlimited"$`, func(limit string) error {
				s.op("set", "key-limit", limit)
				s.op("tick", "key-nolimit")
				return nil
			})
			sc.Step(`^the operator presses "Mint key" twice$`, func() error {
				s.op("set", "key-name", "ci")
				s.op("set", "key-limit", "5")
				s.op("click2", "Mint key")
				return nil
			})
			sc.Step(`^the operator presses "Mint key" and (opens another tab and comes back|logs out) before the answer$`, func(how string) error {
				s.op("set", "key-name", "ci")
				s.op("set", "key-limit", "5")
				s.op("race", "Mint key", map[string]string{"logs out": "logout"}[how])
				return nil
			})
			sc.Step(`^the operator presses "Create" and opens another tab and comes back before the answer$`, func() error {
				s.op("set", "grant-name", "friends")
				s.op("race", "Create", "")
				return nil
			})
			sc.Step(`^the operator presses "Create" twice$`, func() error {
				s.op("set", "grant-name", "friends")
				s.op("click2", "Create")
				return nil
			})
			sc.Step(`^the broker receives one grant create$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.grantCreates) != 1 {
					return fmt.Errorf("the broker received %d grant creates", len(s.grantCreates))
				}
				return nil
			})
			sc.Step(`^the broker receives a grant create with name "([^"]*)" and free true$`, func(name string) error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.grantCreates) != 1 {
					return fmt.Errorf("the broker received %d grant creates", len(s.grantCreates))
				}
				if g := s.grantCreates[0]; g["name"] != name || g["free"] != true {
					return fmt.Errorf("the broker received %v", g)
				}
				return nil
			})
			sc.Step(`^the "Grants" card has no "free" box$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if _, ok := s.page.Inputs["grant-free"]; ok {
					return fmt.Errorf("the grant form still has a free box")
				}
				return nil
			})
			sc.Step(`^the console receives a grant create request$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if n := s.pageSent("POST", "/api/grants"); n != 1 {
					return fmt.Errorf("the page sent %d grant create requests", n)
				}
				return nil
			})
			sc.Step(`^the "Grants" card does not say "([^"]*)"$`, func(w string) error {
				txt, err := s.card("Grants")
				if err != nil {
					return err
				}
				if txt == "" || strings.Contains(txt, w) {
					return fmt.Errorf("the Grants card reads:\n%s", txt)
				}
				return nil
			})
			sc.Step(`^the operator presses "revoke" on the live "([^"]*)" row and confirms$`, func(name string) error {
				s.op("rowclick", name, "revoke")
				s.op("rowclick", name, "keep") // the confirm is on the live row alone
				s.op("rowclick", name, "revoke")
				s.op("click", "revoke")
				return nil
			})
			sc.Step(`^the broker receives a revoke of the live grant "([^"]*)"$`, func(name string) error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.grantRevokes) != 1 || s.grantRevokes[0] != name {
					return fmt.Errorf("the broker revoked live grants %v", s.grantRevokes)
				}
				for _, g := range s.grants {
					if g.Status != "revoked" {
						return fmt.Errorf("grant %s (%s) is still %s", g.Name, g.ID, g.Status)
					}
				}
				return nil
			})
			sc.Step(`^the broker says the account is logged out only when keys are asked for$`, func() error { s.keysLoggedOut = true; return nil })
			sc.Step(`^another origin sends a preflight for "([^"]*)"$`, func(path string) error {
				req, _ := http.NewRequest("OPTIONS", s.console.URL+path, nil)
				req.Header.Set("Origin", "https://evil.example")
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "x-roger-token, content-type")
				r, err := send(req)
				s.last = r
				return err
			})
			sc.Step(`^the browser is given no permission to send the request$`, func() error {
				for _, h := range []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods"} {
					if s.last.hdr.Get(h) != "" {
						return fmt.Errorf("the preflight answers %s: %s", h, s.last.hdr.Get(h))
					}
				}
				if s.last.code == 200 || s.last.code == 204 {
					return fmt.Errorf("the preflight was answered %d", s.last.code)
				}
				return nil
			})
			sc.Step(`^the operator mints a key$`, func() error { mint("ci", "5"); return nil })
			sc.Step(`^the operator minted a key$`, func() error { mint("ci", "5"); return nil })
			sc.Step(`^the operator minted a key and the secret is showing$`, func() error {
				mint("ci", "5")
				sec, err := s.secret()
				if err != nil {
					return err
				}
				if !strings.Contains(s.page.Keys, sec) {
					return fmt.Errorf("the secret is not showing:\n%s", s.page.Keys)
				}
				return nil
			})
			sc.Step(`^the operator mints a second key$`, func() error { mint("two", "1"); return nil })
			sc.Step(`^the operator presses "([^"]*)"$`, func(label string) error { s.op("click", label); return nil })
			sc.Step(`^the operator opens another tab and comes back$`, func() error { s.op("leave"); s.op("load"); return nil })
			sc.Step(`^the operator reloads the page$`, func() error { s.op("remount"); return nil })
			sc.Step(`^the operator logs out$`, func() error {
				s.op("logout")
				return nil
			})
			sc.Step(`^the operator edits the row's limit to "([^"]*)" and saves$`, func(v string) error {
				s.op("rowset", "ci", "row-limit", v)
				s.op("rowclick", "ci", "save")
				return nil
			})
			sc.Step(`^the operator edits a row's limit and saves$`, func() error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci", LimitUSD: 5, Reset: "weekly"})
				s.op("rowset", "ci", "row-limit", "10")
				s.op("rowclick", "ci", "save")
				return nil
			})
			sc.Step(`^the operator changes the row's window to "([^"]*)" and saves$`, func(v string) error {
				s.op("rowset", "ci", "row-reset", v)
				s.op("rowclick", "ci", "save")
				return nil
			})
			sc.Step(`^the operator presses "([^"]*)" on its row$`, func(label string) error {
				name := "ci"
				s.mu.Lock()
				if len(s.keys) == 0 && len(s.grants) > 0 {
					name = s.grants[0].Name
				}
				s.mu.Unlock()
				s.op("rowclick", name, label)
				return nil
			})
			sc.Step(`^the operator was asked to confirm revoking "([^"]*)"$`, func(name string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: name, LimitUSD: 5, Reset: "weekly"})
				s.op("rowclick", name, "revoke")
				return nil
			})
			sc.Step(`^the operator was asked to confirm revoking grant "([^"]*)"$`, func(name string) error {
				s.grants = append(s.grants, keysGrant{ID: fmt.Sprintf("gr_%d", len(s.grants)), Name: name, Free: true, Price: "free", Status: "active"})
				s.op("rowclick", name, "revoke")
				return nil
			})
			createGrant := func() {
				s.op("set", "grant-name", "friends")
				s.op("click", "Create")
			}
			sc.Step(`^the operator creates a grant$`, func() error { createGrant(); return nil })
			sc.Step(`^the operator created a grant and the secret is showing$`, func() error {
				createGrant()
				sec, err := s.secret()
				if err != nil {
					return err
				}
				if !strings.Contains(s.page.Grants, sec) {
					return fmt.Errorf("the grant secret is not showing:\n%s", s.page.Grants)
				}
				return nil
			})

			// --- direct calls
			sc.Step(`^the page reads "([^"]*)"$`, func(path string) error { return s.call("GET", path, "", s.tok) })
			sc.Step(`^the page calls "([^"]*)"$`, func(path string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci"})
				switch path {
				case "/api/keys":
					return s.call("GET", path, "", s.tok)
				case "/api/keys/mint":
					return s.call("POST", path, `{"name":"n","limit_usd":1}`, s.tok)
				case "/api/keys/set":
					return s.call("POST", path, `{"id":"key_aaaa1","disabled":true}`, s.tok)
				case "/api/grants":
					return s.call("POST", path, `{"name":"n"}`, s.tok)
				}
				return s.call("POST", path, `{"id":"key_aaaa1"}`, s.tok)
			})
			sc.Step(`^"([A-Z]+) ([^"]*)" is called with (no token|a wrong token|the console token)$`, func(method, path, tok string) error {
				send := map[string]string{"no token": "", "a wrong token": "tok-wrong", "the console token": s.tok}[tok]
				body := ""
				if method == "POST" {
					body = `{"id":"key_aaaa1","name":"ci","limit_usd":5}`
				}
				if path == "/api/grants" && method != "GET" {
					body = `{"name":"friends"}`
				}
				return s.call(method, path, body, send)
			})
			sc.Step(`^"POST ([^"]*)" is called with the body (.*)$`, func(path, body string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci"})
				return s.call("POST", path, body, s.tok)
			})
			sc.Step(`^"POST ([^"]*)" is called with a body larger than 4 KiB$`, func(path string) error {
				return s.call("POST", path, `{"name":"`+strings.Repeat("a", 5000)+`","limit_usd":5}`, s.tok)
			})
			sc.Step(`^a page on another origin calls "([^"]*)" with the console token$`, func(path string) error {
				s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci"})
				req, _ := http.NewRequest("GET", s.console.URL+path, nil)
				req.Header.Set("Origin", "https://evil.example")
				req.Header.Set("X-Roger-Token", s.tok)
				r, err := send(req)
				s.last = r
				return err
			})
			sc.Step(`^"([^"]*)" is called from a non-loopback address$`, func(path string) error {
				req := httptest.NewRequest("GET", path, nil)
				req.RemoteAddr = "10.0.0.5:51000"
				req.Header.Set("X-Roger-Token", s.tok)
				rec := httptest.NewRecorder()
				s.srv.Handler().ServeHTTP(rec, req)
				s.last = plateReply{code: rec.Code, hdr: rec.Header(), raw: rec.Body.String()}
				return nil
			})

			// --- thens: the keys card
			sc.Step(`^the "Account keys" card has one row for "([^"]*)"$`, func(name string) error {
				_, err := s.keyRow(name)
				if err == nil && len(s.page.KeyRows) != 1 {
					return fmt.Errorf("%d rows, want 1", len(s.page.KeyRows))
				}
				return err
			})
			sc.Step(`^the row shows "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)"$`, func(a, b, c, d string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				for _, w := range []string{a, b, c, d} {
					if !strings.Contains(r.Text, w) {
						return fmt.Errorf("the row does not show %q:\n%s", w, r.Text)
					}
				}
				return nil
			})
			sc.Step(`^the "([^"]*)" card says "([^"]*)"$`, func(card, want string) error { return s.cardSays(card, want) })
			sc.Step(`^the card says "(.*)"$`, func(want string) error { return s.cardSays("Account keys", want) })
			sc.Step(`^that key's limit reads "([^"]*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains("\n"+r.Text+"\n", "\n"+want+"\n") {
					return fmt.Errorf("the limit does not read %q:\n%s", want, r.Text)
				}
				return nil
			})
			sc.Step(`^that key's used amount reads "([^"]*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains("\n"+r.Text+"\n", "\n"+want+" used\n") {
					return fmt.Errorf("the used amount does not read %q:\n%s", want, r.Text)
				}
				return nil
			})
			sc.Step(`^that key's state reads "([^"]*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				lines := strings.Split(r.Text, "\n")
				found := false
				for _, l := range lines {
					if l == "active" || l == "disabled" || l == "expired" {
						if l != want {
							return fmt.Errorf("the state reads %q, want %q", l, want)
						}
						found = true
					}
				}
				if !found {
					return fmt.Errorf("the row shows no state:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the row's state reads "([^"]*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains("\n"+r.Text+"\n", "\n"+want+"\n") {
					return fmt.Errorf("the state does not read %q:\n%s", want, r.Text)
				}
				return nil
			})
			sc.Step(`^the row shows that name as text$`, func() error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains(r.Text, "<img src=x onerror=alert(1)>") {
					return fmt.Errorf("the name is not shown as text:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the page has no new image element$`, func() error {
				for _, tag := range s.page.Tags {
					if tag != "div" && tag != "span" && tag != "ul" && tag != "li" && tag != "button" && tag != "input" && tag != "select" && tag != "option" && tag != "label" && tag != "code" {
						return fmt.Errorf("the page created a <%s>", tag)
					}
				}
				src, err := plateAsset("console-keys.js")
				if err != nil {
					return err
				}
				if strings.Contains(src, "innerHTML") {
					return fmt.Errorf("console-keys.js writes innerHTML")
				}
				return nil
			})
			sc.Step(`^the reply holds no field named "secret"$`, func() error {
				if s.last.code != 200 || strings.Contains(s.last.raw, `"secret"`) {
					return fmt.Errorf("the reply is %d %s", s.last.code, s.last.raw)
				}
				return nil
			})
			sc.Step(`^the reply holds no text starting with "rog-key_"$`, func() error {
				if strings.Contains(s.last.raw, "rog-key_") || !strings.Contains(s.last.raw, "key_aaaa1") {
					return fmt.Errorf("the reply is %s", s.last.raw)
				}
				return nil
			})

			// --- thens: minting
			sc.Step(`^the broker receives a mint with name "([^"]*)", limit_usd ([\d.]+) and reset "([^"]*)"$`, func(name string, limit float64, reset string) error {
				m, err := s.lastMint()
				if err != nil {
					return err
				}
				if m["name"] != name || m["limit_usd"] != limit || m["reset"] != reset {
					return fmt.Errorf("the mint carried %v", m)
				}
				return nil
			})
			sc.Step(`^the broker receives a mint with name "([^"]*)" and limit_usd ([\d.]+)$`, func(name string, limit float64) error {
				m, err := s.lastMint()
				if err != nil {
					return err
				}
				if m["name"] != name || m["limit_usd"] != limit {
					return fmt.Errorf("the mint carried %v", m)
				}
				return nil
			})
			sc.Step(`^the card shows the secret under "([^"]*)"$`, func(label string) error {
				sec, err := s.secret()
				if err != nil {
					return err
				}
				at, sat := strings.Index(s.page.Keys, label), strings.Index(s.page.Keys, sec)
				if at < 0 || sat < at {
					return fmt.Errorf("the card does not show the secret under %q:\n%s", label, s.page.Keys)
				}
				return nil
			})
			sc.Step(`^the new key is in the list with its hint$`, func() error {
				sec, err := s.secret()
				if err != nil {
					return err
				}
				r, err := s.keyRow("ci")
				if err != nil {
					return err
				}
				if !strings.Contains(r.Text, "..."+sec[len(sec)-4:]) || strings.Contains(r.Text, sec) {
					return fmt.Errorf("the row does not show the hint alone:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the clipboard holds exactly these lines, with the broker's address and the new secret$`, func(doc *godog.DocString) error {
				return s.exportBlock(doc.Content)
			})
			sc.Step(`^the page says "([^"]*)"$`, func(want string) error {
				if err := s.run(); err != nil {
					return err
				}
				if !strings.Contains(s.page.Keys+"\n"+s.page.Grants, want) {
					return fmt.Errorf("the page does not say %q:\n%s", want, s.page.Keys)
				}
				return nil
			})
			sc.Step(`^the secret is still showing$`, func() error {
				sec, err := s.secret()
				if err != nil {
					return err
				}
				if !strings.Contains(s.page.Keys, sec) {
					return fmt.Errorf("the secret is gone:\n%s", s.page.Keys)
				}
				return nil
			})
			sc.Step(`^the limit box is empty$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if in, ok := s.page.Inputs["key-limit"]; !ok || in.Value != "" {
					return fmt.Errorf("the limit box holds %q (present %v)", in.Value, ok)
				}
				return nil
			})
			sc.Step(`^the window choice is "no reset"$`, func() error {
				if in, ok := s.page.Inputs["key-reset"]; !ok || in.Value != "none" {
					return fmt.Errorf("the window is %q (present %v)", in.Value, ok)
				}
				if !strings.Contains(s.page.Keys, "no reset") {
					return fmt.Errorf("no choice reads %q", "no reset")
				}
				return nil
			})
			sc.Step(`^the "unlimited" box is not ticked$`, func() error {
				if in, ok := s.page.Inputs["key-nolimit"]; !ok || in.Checked {
					return fmt.Errorf("the box is ticked or missing (present %v)", ok)
				}
				return nil
			})
			sc.Step(`^the limit box is empty and cannot be typed in$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if in := s.page.Inputs["key-limit"]; in.Value != "" || !in.Disabled {
					return fmt.Errorf("the limit box holds %q, disabled %v", in.Value, in.Disabled)
				}
				return nil
			})
			sc.Step(`^the console receives no mint request$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if n := s.pageSent("POST", "/api/keys/mint"); n != 0 {
					return fmt.Errorf("the page sent %d mint requests", n)
				}
				return nil
			})
			sc.Step(`^the broker receives no (mint|change)$`, func(what string) error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if n := map[string]int{"mint": len(s.mints), "change": len(s.changes)}[what]; n != 0 {
					return fmt.Errorf("the broker received %d %s requests", n, what)
				}
				return nil
			})
			sc.Step(`^the broker receives one mint$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.mints) != 1 {
					return fmt.Errorf("the broker received %d mints", len(s.mints))
				}
				return nil
			})
			sc.Step(`^no secret is shown$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if strings.Contains(s.page.Keys+s.page.Grants, "rog-") {
					return fmt.Errorf("a secret is showing:\n%s", s.page.Keys)
				}
				return nil
			})

			// --- thens: where the secret is not
			routesHold := func(sec string) error {
				for _, p := range []string{"/api/keys", "/api/grants", "/api/state", "/api/account"} {
					r, err := s.do("GET", p, "", s.tok)
					if err != nil {
						return err
					}
					if strings.Contains(r.raw, sec) {
						return fmt.Errorf("GET %s returns the secret", p)
					}
				}
				return nil
			}
			sc.Step(`^the secret is not on the page$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				first := s.secrets[0]
				s.mu.Unlock()
				if strings.Contains(s.page.Keys+s.page.Grants, first) {
					return fmt.Errorf("the secret is still on the page:\n%s", s.page.Keys)
				}
				return nil
			})
			sc.Step(`^no console route returns the secret$`, func() error {
				s.mu.Lock()
				first := s.secrets[0]
				s.mu.Unlock()
				return routesHold(first)
			})
			sc.Step(`^the secret is not in (.+)$`, func(place string) error {
				sec, err := s.secret()
				if err != nil {
					return err
				}
				var hay string
				switch place {
				case "the reply to GET /api/state", "the reply to GET /api/keys":
					r, err := s.do("GET", strings.TrimPrefix(place, "the reply to GET "), "", s.tok)
					if err != nil {
						return err
					}
					if r.code != 200 {
						return fmt.Errorf("%s answered %d", place, r.code)
					}
					hay = r.raw
				case "the event stream":
					req, _ := http.NewRequest("GET", s.console.URL+"/api/events", nil)
					req.Header.Set("X-Roger-Token", s.tok)
					ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
					defer cancel()
					resp, err := http.DefaultClient.Do(req.WithContext(ctx))
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					b, _ := io.ReadAll(resp.Body)
					if len(b) == 0 {
						return fmt.Errorf("the event stream sent nothing to check")
					}
					hay = string(b)
				case "the page's address":
					for _, f := range s.page.Fetches {
						hay += f.Path + "\n"
					}
				case "localStorage or sessionStorage":
					src, err := plateAsset("console-keys.js")
					if err != nil {
						return err
					}
					for _, w := range []string{"localStorage", "sessionStorage", "indexedDB", "document.cookie"} {
						if strings.Contains(src, w) {
							return fmt.Errorf("console-keys.js uses %s", w)
						}
					}
					return nil
				case "the node log":
					hay = s.logBuf.String()
				case "the console's standard output":
					b, err := os.ReadFile(s.stdout.Name())
					if err != nil {
						return err
					}
					hay = string(b)
				default:
					return fmt.Errorf("unknown place %q", place)
				}
				if strings.Contains(hay, sec) {
					return fmt.Errorf("%s holds the secret", place)
				}
				return nil
			})
			sc.Step(`^the reply carries "Cache-Control: no-store"$`, func() error {
				if s.last.code != 200 || s.last.hdr.Get("Cache-Control") != "no-store" {
					return fmt.Errorf("the reply is %d with Cache-Control %q: %s", s.last.code, s.last.hdr.Get("Cache-Control"), s.last.raw)
				}
				return nil
			})

			// --- thens: changing and revoking
			sc.Step(`^the broker receives a change to "([^"]*)" with only limit_usd ([\d.]+)$`, func(id string, v float64) error {
				return s.changeOnly(id, map[string]any{"limit_usd": v})
			})
			sc.Step(`^the broker receives a change to "([^"]*)" with only reset "([^"]*)"$`, func(id, v string) error {
				return s.changeOnly(id, map[string]any{"reset": v})
			})
			sc.Step(`^the broker receives a change to "([^"]*)" with only disabled (true|false)$`, func(id, v string) error {
				return s.changeOnly(id, map[string]any{"disabled": v == "true"})
			})
			sc.Step(`^the row reads "([^"]*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains("\n"+r.Text+"\n", "\n"+want+"\n") {
					return fmt.Errorf("the row does not read %q:\n%s", want, r.Text)
				}
				return nil
			})
			sc.Step(`^the console receives no change request$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if n := s.pageSent("POST", "/api/keys/set"); n != 0 {
					return fmt.Errorf("the page sent %d change requests", n)
				}
				return nil
			})
			sc.Step(`^the row says "(.*)"$`, func(want string) error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains(r.Text, want) {
					return fmt.Errorf("the row does not say %q:\n%s", want, r.Text)
				}
				return nil
			})
			sc.Step(`^the row shows the broker's reason$`, func() error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains(r.Text, "the limit is above what this account may set") {
					return fmt.Errorf("the row does not show the broker's reason:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the row's limit reads what it read before$`, func() error {
				r, err := s.onlyKey()
				if err != nil {
					return err
				}
				if !strings.Contains("\n"+r.Text+"\n", "\n$5.00 weekly\n") {
					return fmt.Errorf("the limit changed:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the console receives no revoke request$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				if n := s.pageSent("POST", "/api/keys/revoke") + s.pageSent("POST", "/api/grants/revoke"); n != 0 {
					return fmt.Errorf("the page sent %d revoke requests", n)
				}
				return nil
			})
			sc.Step(`^the row asks "([^"]*)" with "([^"]*)" and "([^"]*)"$`, func(ask, yes, no string) error {
				if err := s.run(); err != nil {
					return err
				}
				all := append(append([]keysRow{}, s.page.KeyRows...), s.page.GrantRows...)
				for _, r := range all {
					if strings.Contains(r.Text, ask) {
						if strings.Join(r.Buttons, ",") != yes+","+no {
							return fmt.Errorf("the row offers %v, want %s and %s", r.Buttons, yes, no)
						}
						return nil
					}
				}
				return fmt.Errorf("no row asks %q:\n%s\n%s", ask, s.page.Keys, s.page.Grants)
			})
			sc.Step(`^the row is back to normal$`, func() error {
				r, err := s.keyRow("ci")
				if err != nil {
					return err
				}
				if got := strings.Join(r.Buttons, ","); got != "save,turn off,revoke" {
					return fmt.Errorf("the row offers %q", got)
				}
				if strings.Contains(r.Text, "Revoke ci?") {
					return fmt.Errorf("the row still asks:\n%s", r.Text)
				}
				return nil
			})
			sc.Step(`^the broker receives a revoke of "([^"]*)"$`, func(id string) error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.revokes) != 1 || s.revokes[0] != id {
					return fmt.Errorf("the broker received revokes %v", s.revokes)
				}
				return nil
			})
			sc.Step(`^the card no longer lists "([^"]*)"$`, func(name string) error {
				if err := s.run(); err != nil {
					return err
				}
				for _, r := range s.page.KeyRows {
					if r.Name == name {
						return fmt.Errorf("%q is still listed", name)
					}
				}
				return nil
			})

			// --- thens: grants
			sc.Step(`^the "Grants" card has a row "([^"]*)" showing "([^"]*)" and "([^"]*)"$`, func(name, a, b string) error {
				r, err := s.grantRow(name)
				if err != nil {
					return err
				}
				for _, w := range []string{name, a, b} {
					if !strings.Contains("\n"+r.Text+"\n", "\n"+w+"\n") {
						return fmt.Errorf("the row does not show %q:\n%s", w, r.Text)
					}
				}
				return nil
			})
			sc.Step(`^no row reads "grant"$`, func() error {
				for _, r := range s.page.GrantRows {
					if strings.Contains("\n"+r.Text+"\n", "\ngrant\n") {
						return fmt.Errorf("a row reads %q", "grant")
					}
				}
				return nil
			})
			sc.Step(`^each item has the fields "name", "price" and "status"$`, func() error {
				s.mu.Lock()
				s.grants = append(s.grants, keysGrant{ID: "gr_x", Name: "x", Free: true, Status: "active"})
				s.mu.Unlock()
				if err := s.call("GET", "/api/grants", "", s.tok); err != nil {
					return err
				}
				var items []map[string]any
				if err := json.Unmarshal([]byte(s.last.raw), &items); err != nil || len(items) != 1 {
					return fmt.Errorf("the reply is %d %s", s.last.code, s.last.raw)
				}
				if items[0]["name"] != "x" || items[0]["price"] != "free" || items[0]["status"] != "active" || len(items[0]) != 3 {
					return fmt.Errorf("the item is %v", items[0])
				}
				return nil
			})
			sc.Step(`^the broker receives a revoke of grant "([^"]*)"$`, func(name string) error {
				if err := s.run(); err != nil {
					return err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.grantRevokes) != 1 || s.grantRevokes[0] != name {
					return fmt.Errorf("the broker received grant revokes %v", s.grantRevokes)
				}
				return nil
			})
			sc.Step(`^the row's status reads "([^"]*)"$`, func(want string) error {
				if err := s.run(); err != nil {
					return err
				}
				if len(s.page.GrantRows) != 1 || !strings.Contains("\n"+s.page.GrantRows[0].Text+"\n", "\n"+want+"\n") {
					return fmt.Errorf("the grant rows read %v", s.page.GrantRows)
				}
				return nil
			})
			sc.Step(`^the row for "([^"]*)" has no "revoke" button$`, func(name string) error {
				r, err := s.grantRow(name)
				if err != nil {
					return err
				}
				if len(r.Buttons) != 0 {
					return fmt.Errorf("the row offers %v", r.Buttons)
				}
				return nil
			})
			sc.Step(`^the "Grants" card carries the broker's reason$`, func() error { return s.cardSays("Grants", keysGrantReason) })
			sc.Step(`^the "Account keys" card lists "([^"]*)"$`, func(name string) error { _, err := s.keyRow(name); return err })

			// --- thens: direct calls
			sc.Step(`^the reply is (\d+)$`, func(code int) error {
				if s.last.code != code {
					return fmt.Errorf("the reply is %d, want %d: %s", s.last.code, code, s.last.raw)
				}
				return nil
			})
			sc.Step(`^the reply is (\d+) with "(.*)"$`, func(code int, msg string) error {
				if s.last.code != code {
					return fmt.Errorf("the reply is %d, want %d: %s", s.last.code, code, s.last.raw)
				}
				got, _ := s.last.js["message"].(string)
				if got != msg {
					return fmt.Errorf("the message is %q, want %q", got, msg)
				}
				return nil
			})
			sc.Step(`^the reply does not contain "([^"]*)"$`, func(w string) error {
				if strings.Contains(s.last.raw, w) {
					return fmt.Errorf("the reply contains %q: %s", w, s.last.raw)
				}
				return nil
			})
			sc.Step(`^the broker is not asked$`, func() error {
				if s.usePage && len(s.ops) == 0 && s.last.code == 0 {
					// LAN mode on the page: nothing the page loaded reached the stand-in.
					if err := s.run(); err != nil {
						return err
					}
					s.mu.Lock()
					defer s.mu.Unlock()
					if s.hits != 0 {
						return fmt.Errorf("the broker stand-in saw %d requests", s.hits)
					}
					return nil
				}
				return s.brokerNotAsked()
			})
			sc.Step(`^the browser is given no permission to read the reply$`, func() error {
				// The reply went out, as it would to any local caller with the token. Without an
				// allow-origin header a browser will not hand it to another origin's page.
				if s.last.code != 200 || !strings.Contains(s.last.raw, "key_aaaa1") {
					return fmt.Errorf("the reply is %d %s", s.last.code, s.last.raw)
				}
				if got := s.last.hdr.Get("Access-Control-Allow-Origin"); got != "" {
					return fmt.Errorf("the reply allows origin %q", got)
				}
				return nil
			})
			sc.Step(`^the request is refused$`, func() error {
				if s.last.code != http.StatusForbidden || strings.Contains(s.last.raw, "keys") {
					return fmt.Errorf("the reply is %d %s", s.last.code, s.last.raw)
				}
				return nil
			})

			// --- thens: unavailable
			sc.Step(`^the mint form is hidden$`, func() error {
				if in, ok := s.page.Inputs["key-name"]; !ok || !in.Hidden {
					return fmt.Errorf("the mint form is showing (present %v)", ok)
				}
				for _, b := range []string{"Mint key"} {
					if strings.Contains(s.page.Keys, b) {
						return fmt.Errorf("the card still offers %q", b)
					}
				}
				return nil
			})
			sc.Step(`^the "Account keys" card says the broker could not be reached$`, func() error {
				return s.cardSays("Account keys", "The broker could not be reached.")
			})
			sc.Step(`^the rest of the ACCOUNT tab still loads$`, func() error {
				// The page asked for both lists and drew both cards; neither call threw.
				if s.pageSent("GET", "/api/keys") == 0 || s.pageSent("GET", "/api/grants") == 0 {
					return fmt.Errorf("the page did not ask for both lists: %v", s.page.Fetches)
				}
				return s.cardSays("Grants", "For someone else.")
			})

			// --- the endpoint plate
			sc.Step(`^the operator has a channel open from the CHAT tab on a market model$`, func() error {
				p := &plateBDD{t: s.t}
				p.reset()
				defer p.cleanup()
				p.setOffer("qwen3-32b", 0, 0, true)
				p.startConsole()
				if err := p.runPage("qwen3-32b", "open", "", false); err != nil {
					return err
				}
				s.plateTxt = p.page.Text
				return nil
			})
			sc.Step(`^the endpoint plate says "([^"]*)"$`, func(want string) error {
				if !strings.Contains(s.plateTxt, want) {
					return fmt.Errorf("the plate does not say %q:\n%s", want, s.plateTxt)
				}
				return nil
			})
			sc.Step(`^the endpoint plate does not contain "([^"]*)"$`, func(w string) error {
				if s.plateTxt == "" || strings.Contains(s.plateTxt, w) {
					return fmt.Errorf("the plate reads:\n%s", s.plateTxt)
				}
				return nil
			})
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"../../features/console/account_keys.feature"}, TestingT: t, Strict: true},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("features/console/account_keys.feature has failing scenarios")
	}
}
