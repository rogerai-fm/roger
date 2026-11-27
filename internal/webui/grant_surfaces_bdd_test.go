package webui

// Executable spec: the Console Rule of features/grants/grant_surfaces.feature - the ACCOUNT
// tab's "Grants you added" card, a priced create and Reprice in the grants-you-made card. The
// REAL console server over HTTP, a broker stand-in on loopback that serves /balance, /grants,
// /grants/{id}/reprice and /account/grants, and the shipped console-keys.js EXECUTED with node
// against that server. No mocks. Every secret here is made up.

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
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/node"
)

const gsFriendsSecret = "rog-grant_TESTONLYfriends0000000000000000"

// gsMade is a grant the console's account made, as GET /grants shows it.
type gsMade struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Free     bool     `json:"free"`
	Price    string   `json:"price"`
	PriceOut float64  `json:"price_out"`
	Nodes    []string `json:"nodes"`
	Models   []string `json:"models"`
	AddedBy  int      `json:"added_by"`
	Status   string   `json:"status"`
}

type gsFetch struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Status int    `json:"status"`
	Reply  string `json:"reply"`
}

type gsRow struct {
	Name    string   `json:"name"`
	Text    string   `json:"text"`
	Buttons []string `json:"buttons"`
}

type gsSnap struct {
	Grants  string `json:"grants"`
	Added   string `json:"added"`
	Fetches int    `json:"fetches"`
}

type gsPage struct {
	Grants    string               `json:"grants"`
	Added     string               `json:"added"`
	GrantRows []gsRow              `json:"grantRows"`
	AddedRows []gsRow              `json:"addedRows"`
	Buttons   map[string][]string  `json:"buttons"`
	Inputs    map[string]keysInput `json:"inputs"`
	Clipboard string               `json:"clipboard"`
	Fetches   []gsFetch            `json:"fetches"`
	Snaps     []gsSnap             `json:"snaps"`
	Login     bool                 `json:"login"`
}

type gsBDD struct {
	t *testing.T

	mu       sync.Mutex
	loggedIn bool
	made     []gsMade
	added    []client.AddedGrant
	creates  []map[string]any
	reprices []map[string]any
	removes  []string
	hits     int
	hitsWas  int
	secrets  []string

	broker  *httptest.Server
	console *httptest.Server
	tok     string
	logBuf  *bytes.Buffer

	ops  [][]string
	ran  bool
	page gsPage
	last plateReply
}

func (s *gsBDD) authFile() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "rogerai", "auth.json")
}

// signIn records (or forgets) this machine's roger sign-in, as `roger login` would.
func (s *gsBDD) signIn(on bool) {
	s.mu.Lock()
	s.loggedIn = on
	s.mu.Unlock()
	if !on {
		_ = os.Remove(s.authFile())
		return
	}
	_ = os.MkdirAll(filepath.Dir(s.authFile()), 0o700)
	_ = os.WriteFile(s.authFile(), []byte(`{"github_login":"frank","github_id":7}`), 0o600)
}

func gsFriendsCopy(state string) client.AddedGrant {
	return client.AddedGrant{ID: "gr_friends", Name: "friends", Price: "$0.05/1M out", PriceOut: 0.05, Models: []string{"qwen-small"},
		Stations: []client.AddedGrantStation{{ID: "olive-box", Models: []string{"qwen-small"}, PublicOut: 0.10}}, State: state, AddedAt: 1795996800}
}

func (s *gsBDD) startBroker() {
	s.broker = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fields map[string]any
		_ = json.Unmarshal(body, &fields)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.hits++
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch {
		case r.URL.Path == "/balance":
			_ = enc.Encode(map[string]any{"user": "op", "balance": 5.0, "logged_in": s.loggedIn})
		case !s.loggedIn:
			keysErr(w, http.StatusUnauthorized, "", "log in first - run `roger login`")
		case r.URL.Path == "/grants" && r.Method == http.MethodGet:
			_ = enc.Encode(map[string]any{"grants": s.made})
		case r.URL.Path == "/grants" && r.Method == http.MethodPost:
			s.creates = append(s.creates, fields)
			name, _ := fields["name"].(string)
			out, _ := fields["price_out"].(float64)
			free, _ := fields["free"].(bool)
			g := gsMade{ID: fmt.Sprintf("gr_%d", len(s.made)), Name: name, Free: free, Price: "free", PriceOut: out, Status: "active"}
			if !free {
				g.Price = fmt.Sprintf("$%.2f/1M out", out)
			}
			s.made = append(s.made, g)
			secret := fmt.Sprintf("rog-grant_TESTONLYmade%04dqq", len(s.secrets)+1)
			s.secrets = append(s.secrets, secret)
			_ = enc.Encode(map[string]any{"grant": g, "secret": secret})
		case strings.HasPrefix(r.URL.Path, "/grants/") && strings.HasSuffix(r.URL.Path, "/reprice"):
			s.reprices = append(s.reprices, fields)
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/grants/"), "/reprice")
			for i := range s.made {
				if s.made[i].ID == id {
					s.made[i].PriceOut, _ = fields["price_out"].(float64)
					s.made[i].Price = fmt.Sprintf("$%.2f/1M out", s.made[i].PriceOut)
					s.made[i].AddedBy = 0
				}
			}
			for i := range s.added {
				if s.added[i].ID == id {
					s.added[i].State = "replaced"
				}
			}
			secret := fmt.Sprintf("rog-grant_TESTONLYrepriced%04dzz", len(s.secrets)+1)
			s.secrets = append(s.secrets, secret)
			_ = enc.Encode(map[string]any{"grant": map[string]any{"id": id}, "secret": secret})
		case r.URL.Path == "/account/grants" && r.Method == http.MethodGet:
			_ = enc.Encode(map[string]any{"grants": s.added, "max": 10})
		case r.URL.Path == "/account/grants" && r.Method == http.MethodPost:
			secret, _ := fields["secret"].(string)
			switch {
			case len(s.added) >= 10:
				keysErr(w, http.StatusConflict, "", "an account can hold 10 grants - remove one first (`roger grant remove <name>`)")
			case secret != gsFriendsSecret:
				keysErr(w, http.StatusBadRequest, "", "grant key invalid or revoked")
			default:
				g := gsFriendsCopy("active")
				s.added = append(s.added, g)
				g.State, g.AddedAt = "", 0
				_ = enc.Encode(map[string]any{"grant": g})
			}
		case strings.HasPrefix(r.URL.Path, "/account/grants/") && r.Method == http.MethodDelete:
			id := strings.TrimPrefix(r.URL.Path, "/account/grants/")
			s.removes = append(s.removes, id)
			for i := range s.added {
				if s.added[i].ID == id {
					s.added = append(s.added[:i], s.added[i+1:]...)
					break
				}
			}
			_ = enc.Encode(map[string]any{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
}

func (s *gsBDD) startConsole() {
	srv := New(node.New(node.Config{Station: "amber-fox"}), Options{User: "op", Broker: s.broker.URL})
	srv.token = s.tok
	s.console = httptest.NewServer(srv.Handler())
}

func (s *gsBDD) reset() {
	s.cleanup()
	*s = gsBDD{t: s.t, tok: "tok-run"}
	s.logBuf = &bytes.Buffer{}
	log.SetOutput(s.logBuf)
	s.startBroker()
	s.startConsole()
}

func (s *gsBDD) cleanup() {
	if s.logBuf != nil {
		log.SetOutput(os.Stderr)
	}
	if s.console != nil {
		s.console.Close()
	}
	if s.broker != nil {
		s.broker.Close()
	}
	_ = os.Remove(s.authFile())
}

func (s *gsBDD) op(parts ...string) { s.ops = append(s.ops, parts); s.ran = false }

// ---- the page, executed --------------------------------------------------------------

const gsPageScript = `
const api = require(process.argv[1]);
const [base, tok, opsJSON] = process.argv.slice(2);
const ops = JSON.parse(opsJSON);
const out = { grants: "", added: "", grantRows: [], addedRows: [], buttons: {}, inputs: {}, clipboard: "", fetches: [], snaps: [], login: false };
let inflight = 0;
function el(tag) {
  return { tag, children: [], attrs: {}, hidden: false, textContent: "", className: "", value: "", checked: false, disabled: false, listeners: {},
    appendChild(c) { this.children.push(c); return c; },
    setAttribute(k, v) { this.attrs[k] = String(v); }, getAttribute(k) { return this.attrs[k]; },
    addEventListener(t, f) { (this.listeners[t] = this.listeners[t] || []).push(f); },
    replaceChildren(...cs) { this.children = cs; } };
}
function fire(e, t) { (e.listeners[t] || []).forEach((f) => f({ target: e })); }
function walk(e, f) { if (e.hidden) return; f(e); (e.children || []).forEach((c) => walk(c, f)); }
function has(e, cls) { return String(e.className || "").split(" ").includes(cls); }
function find(root, pred) { let hit = null; walk(root, (e) => { if (!hit && pred(e)) hit = e; }); return hit; }
const hosts = { keys: el("div"), grants: el("div"), added: el("div") };
const doc = { getElementById: (id) => ({ "card-keys-body": hosts.keys, "card-grants-body": hosts.grants, "card-added-body": hosts.added })[id] || null, createElement: el };
function text(root) { const t = []; walk(root, (e) => { if (e.textContent) t.push(e.textContent); if (e.tag === "input" && e.value) t.push(e.value); }); return t.join("\n"); }
function buttons(root) { const b = []; walk(root, (c) => { if (c.tag === "button") b.push(c.textContent); }); return b; }
function rows(root, cls) {
  const r = [];
  walk(root, (e) => { if (has(e, cls)) r.push({ name: e.attrs["data-name"] || "", text: text(e), buttons: buttons(e) }); });
  return r;
}
function row(host, name) {
  const r = find(hosts[host], (e) => e.attrs["data-name"] === name && (has(e, "grant-row") || has(e, "added-row")));
  if (!r) throw new Error("no row " + name + " in: " + text(hosts[host]));
  return r;
}
function button(root, label) {
  const b = find(root, (e) => e.tag === "button" && e.textContent === label);
  if (!b) throw new Error("no button " + label + " in: " + text(root));
  return b;
}
function field(root, cls) {
  const f = find(root, (e) => has(e, cls));
  if (!f) throw new Error("no field " + cls + " in: " + text(root));
  return f;
}
function press(b) { if (!b.disabled) fire(b, "click"); }
function type(f, v) { f.value = v; fire(f, "input"); fire(f, "change"); }
const actions = {
  fetch: (path, opts) => {
    opts = opts || {};
    const rec = { method: opts.method || "GET", path, status: 0, reply: "" };
    out.fetches.push(rec);
    const h = Object.assign({ "X-Roger-Token": tok }, opts.headers || {});
    inflight++;
    return fetch(base + path, Object.assign({}, opts, { headers: h })).then((r) =>
      r.clone().text().then((t) => { rec.status = r.status; rec.reply = t; return r; })).finally(() => { inflight--; });
  },
  clipboard: { writeText: (t) => { out.clipboard = t; return Promise.resolve(); } },
  login: () => { out.login = true; },
};
async function settle() {
  for (let i = 0; i < 400; i++) {
    await new Promise((r) => setTimeout(r, 5));
    if (inflight === 0) { await new Promise((r) => setTimeout(r, 5)); if (inflight === 0) return; }
  }
  throw new Error("the page never settled");
}
(async () => {
  const card = await api.mount(doc, actions);
  await card.load();
  await settle();
  for (const [what, a, b, c] of ops) {
    if (what === "set") type(field(hosts[a], b), c);
    else if (what === "click") press(button(hosts[a], b));
    else if (what === "rowtype") { const [host, name, cls] = a.split("|"); type(field(row(host, name), cls), b); }
    else if (what === "rowclick") press(button(row(a, b), c));
    else if (what === "copyif") { const b = find(hosts.grants, (e) => e.tag === "button" && e.textContent === "copy"); if (b) press(b); }
    else if (what === "snap") out.snaps.push({ grants: text(hosts.grants), added: text(hosts.added), fetches: out.fetches.length });
    else if (what === "load") await card.load();
    else throw new Error("unknown op " + what);
    await settle();
  }
  out.grants = text(hosts.grants); out.added = text(hosts.added);
  out.grantRows = rows(hosts.grants, "grant-row"); out.addedRows = rows(hosts.added, "added-row");
  for (const [k, h] of Object.entries(hosts)) {
    out.buttons[k] = buttons(h);
    (function all(e, hid) {
      hid = hid || e.hidden;
      String(e.className || "").split(" ").filter(Boolean).forEach((cls) => {
        if (e.tag === "input" && !(cls in out.inputs)) out.inputs[cls] = { Value: e.value, Checked: e.checked, Disabled: e.disabled, Hidden: hid };
      });
      (e.children || []).forEach((ch) => all(ch, hid));
    })(h, false);
  }
  process.stdout.write(JSON.stringify(out), () => process.exit(0));
})().catch((e) => { console.error(e); process.exit(2); });
`

func (s *gsBDD) run() error {
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
	s.mu.Lock()
	s.hitsWas = s.hits
	s.mu.Unlock()
	b, err := exec.Command(nodeBin, "-e", gsPageScript, "./assets/console-keys.js", s.console.URL, s.tok, string(raw)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("console-keys.js failed: %v: %s", err, b)
	}
	s.ran = true
	s.page = gsPage{}
	return json.Unmarshal(b, &s.page)
}

func (s *gsBDD) addedRow(name string) (gsRow, bool, error) {
	if err := s.run(); err != nil {
		return gsRow{}, false, err
	}
	for _, r := range s.page.AddedRows {
		if r.Name == name {
			return r, true, nil
		}
	}
	return gsRow{}, false, nil
}

func (s *gsBDD) grantRow(name string) (gsRow, error) {
	if err := s.run(); err != nil {
		return gsRow{}, err
	}
	for _, r := range s.page.GrantRows {
		if r.Name == name {
			return r, nil
		}
	}
	return gsRow{}, fmt.Errorf("no grant row %q; the card reads:\n%s", name, s.page.Grants)
}

func (s *gsBDD) lastSecret() (string, error) {
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

// shownOnceWithCopy checks the newest secret is drawn exactly once in the grants card, that
// its copy button (pressed as the last step, "copyif") put it on the clipboard, and that no
// list reply carried it.
func (s *gsBDD) shownOnceWithCopy() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	if n := strings.Count(s.page.Grants, sec); n != 1 {
		return fmt.Errorf("the secret is drawn %d times; the card reads:\n%s", n, s.page.Grants)
	}
	if !strings.Contains(s.page.Clipboard, sec) {
		return fmt.Errorf("copy put %q on the clipboard", s.page.Clipboard)
	}
	for _, f := range s.page.Fetches {
		if f.Method == "GET" && strings.Contains(f.Reply, sec) {
			return fmt.Errorf("the list reply %s carried the secret", f.Path)
		}
	}
	return nil
}

// cliPriceRefusal is what `roger grant create --price-out <typed>` says, per
// features/grants/priced_grants.feature, read with the CLI's own reader.
func cliPriceRefusal(typed string) (string, error) {
	_, err := client.ParseGrantPrice(typed)
	if err == nil {
		return "", fmt.Errorf("the CLI reader takes %q", typed)
	}
	want := map[string]string{"-1": "price cannot be negative", "abc": client.MsgRepriceNoPrice, "1e9": client.MsgRepriceNoPrice}[typed]
	if want != "" && err.Error() != want {
		return "", fmt.Errorf("the CLI reader says %q for %q, the broker spec says %q", err, typed, want)
	}
	return err.Error(), nil
}

func TestGrantSurfacesConsoleFeature(t *testing.T) {
	s := &gsBDD{t: t}
	const file = "../../features/grants/grant_surfaces.feature"
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.cleanup()
				s.console, s.broker, s.logBuf = nil, nil, nil
				return ctx, nil
			})

			// --- background
			sc.Step(`^olive made a grant "friends" at \$0\.05 per 1M output tokens on "olive-box"$`, func() error {
				s.made = append(s.made, gsMade{ID: "gr_friends", Name: "friends", Price: "$0.05/1M out", PriceOut: 0.05,
					Nodes: []string{"olive-box"}, Models: []string{"qwen-small"}, AddedBy: 1, Status: "active"})
				return nil
			})
			sc.Step(`^frank is logged in$`, func() error { s.signIn(true); return nil })

			// --- givens
			sc.Step(`^frank added "friends"$`, func() error { s.added = append(s.added, gsFriendsCopy("active")); return nil })
			sc.Step(`^frank added "friends" and olive repriced it$`, func() error {
				s.added = append(s.added, gsFriendsCopy("replaced"))
				return nil
			})
			sc.Step(`^frank has added 10 grants$`, func() error {
				for i := 0; i < 10; i++ {
					g := gsFriendsCopy("active")
					g.ID, g.Name = fmt.Sprintf("gr_other%d", i), fmt.Sprintf("other-%d", i)
					s.added = append(s.added, g)
				}
				return nil
			})
			sc.Step(`^frank is logged out$`, func() error { s.signIn(false); return nil })

			// --- whens
			add := func() error {
				s.op("set", "added", "added-secret", gsFriendsSecret)
				s.op("snap")
				s.op("click", "added", "Add")
				return s.run()
			}
			sc.Step(`^frank pastes the "friends" secret into the add box and presses Add$`, add)
			sc.Step(`^frank adds the "friends" secret in the console$`, add)
			sc.Step(`^a request without the console token calls the added-grants routes$`, func() error {
				s.mu.Lock()
				s.hitsWas = s.hits
				s.mu.Unlock()
				calls := [][2]string{{"GET", "/api/grants/added"}, {"POST", "/api/grants/add"}, {"POST", "/api/grants/remove"}, {"POST", "/api/grants/reprice"}}
				for _, c := range calls {
					for _, tok := range []string{"", "wrong-token"} {
						req, _ := http.NewRequest(c[0], s.console.URL+c[1], strings.NewReader(`{"secret":"`+gsFriendsSecret+`","id":"gr_friends","name":"friends","price_out":"0.08"}`))
						if tok != "" {
							req.Header.Set("X-Roger-Token", tok)
						}
						r, err := send(req)
						if err != nil {
							return err
						}
						if r.code != http.StatusForbidden {
							return fmt.Errorf("%s %s with token %q answered %d %s", c[0], c[1], tok, r.code, r.raw)
						}
					}
				}
				return nil
			})
			sc.Step(`^frank presses Remove on "friends" and confirms$`, func() error {
				s.op("rowclick", "added", "friends", "Remove")
				s.op("snap")
				s.op("rowclick", "added", "friends", "yes, remove")
				return s.run()
			})
			sc.Step(`^olive makes a grant "club" with price "([^"]*)" in the console$`, func(typed string) error {
				s.op("set", "grants", "grant-name", "club")
				s.op("set", "grants", "grant-price", typed)
				s.op("click", "grants", "Create")
				s.op("copyif")
				return s.run()
			})
			sc.Step(`^olive presses Reprice on "friends", enters 0\.08 and confirms$`, func() error {
				s.op("rowclick", "grants", "friends", "Reprice")
				s.op("rowtype", "grants|friends|reprice-price", "0.08")
				s.op("snap")
				s.op("rowclick", "grants", "friends", "yes, reprice")
				s.op("copyif")
				return s.run()
			})
			sc.Step(`^frank opens the ACCOUNT tab$`, func() error { return s.run() })
			sc.Step(`^frank opens the "Grants you added" card$`, func() error { return s.run() })

			// --- thens
			sc.Step(`^the card lists "friends" at \$0\.05 with "olive-box" and "qwen-small"$`, func() error {
				r, ok, err := s.addedRow("friends")
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("no added row \"friends\"; the card reads:\n%s", s.page.Added)
				}
				for _, w := range []string{"$0.05", "olive-box", "qwen-small", "active"} {
					if !strings.Contains(r.Text, w) {
						return fmt.Errorf("the row does not say %q:\n%s", w, r.Text)
					}
				}
				return nil
			})
			sc.Step(`^the add box is cleared$`, func() error {
				in, ok := s.page.Inputs["added-secret"]
				if !ok || in.Value != "" || in.Hidden {
					return fmt.Errorf("the add box is %+v (present %v)", in, ok)
				}
				if len(s.page.Snaps) == 0 || !strings.Contains(s.page.Snaps[0].Added, gsFriendsSecret) {
					return fmt.Errorf("the secret was never typed into the add box")
				}
				return nil
			})
			sc.Step(`^the secret is in no response the page received after the add$`, func() error {
				after := s.page.Fetches[s.page.Snaps[0].Fetches:]
				if len(after) == 0 || after[0].Path != "/api/grants/add" || after[0].Status != 200 {
					return fmt.Errorf("the add was not sent: %+v", after)
				}
				for _, f := range s.page.Fetches {
					if strings.Contains(f.Reply, gsFriendsSecret) {
						return fmt.Errorf("the reply to %s %s carried the secret", f.Method, f.Path)
					}
				}
				if strings.Contains(s.page.Added, gsFriendsSecret) || strings.Contains(s.logBuf.String(), gsFriendsSecret) {
					return fmt.Errorf("the secret is still on the page or in the log")
				}
				return nil
			})
			sc.Step(`^it is refused and nothing reaches the broker$`, func() error {
				s.mu.Lock()
				defer s.mu.Unlock()
				if s.hits != s.hitsWas {
					return fmt.Errorf("the broker stand-in saw %d requests", s.hits-s.hitsWas)
				}
				return nil
			})
			sc.Step(`^the card no longer lists "friends"$`, func() error {
				if _, ok, err := s.addedRow("friends"); err != nil || ok {
					return fmt.Errorf("still listed (err %v):\n%s", err, s.page.Added)
				}
				snap := s.page.Snaps[0]
				if !strings.Contains(snap.Added, "Remove friends from your account?") {
					return fmt.Errorf("Remove did not ask first; the card read:\n%s", snap.Added)
				}
				for _, f := range s.page.Fetches[:snap.Fetches] {
					if f.Path == "/api/grants/remove" {
						return fmt.Errorf("Remove was sent before it was confirmed")
					}
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.removes) != 1 || s.removes[0] != "gr_friends" {
					return fmt.Errorf("the broker removed %v", s.removes)
				}
				return nil
			})
			sc.Step(`^the refusal says to remove one with the Remove button$`, func() error {
				if !strings.Contains(s.page.Added, "remove one with the Remove button") || strings.Contains(s.page.Added, "roger grant remove") {
					return fmt.Errorf("the card reads:\n%s", s.page.Added)
				}
				if !strings.Contains(s.page.Added, "an account can hold 10 grants") {
					return fmt.Errorf("the broker's words were lost:\n%s", s.page.Added)
				}
				return nil
			})
			sc.Step(`^"club" is priced at \$0\.05 per 1M output tokens$`, func() error {
				r, err := s.grantRow("club")
				if err != nil {
					return err
				}
				if !strings.Contains(r.Text, "$0.05") {
					return fmt.Errorf("the row reads %q", r.Text)
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.creates) != 1 || s.creates[0]["price_out"] != 0.05 || s.creates[0]["free"] != false {
					return fmt.Errorf("the broker received %v", s.creates)
				}
				return nil
			})
			sc.Step(`^the new secret is shown once with a copy button$`, s.shownOnceWithCopy)
			sc.Step(`^the console refuses it with the same message as \x60roger grant create --price-out ([^\x60]*)\x60$`, func(typed string) error {
				want, err := cliPriceRefusal(typed)
				if err != nil {
					return err
				}
				if err := s.run(); err != nil {
					return err
				}
				if !strings.Contains(s.page.Grants, want) {
					return fmt.Errorf("the card does not say %q:\n%s", want, s.page.Grants)
				}
				return nil
			})
			sc.Step(`^no grant is made$`, func() error {
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.creates) != 0 {
					return fmt.Errorf("the broker received %v", s.creates)
				}
				return nil
			})
			sc.Step(`^the page says the old secret has stopped working and how many accounts it dropped$`, func() error {
				snap := s.page.Snaps[0]
				if !strings.Contains(snap.Grants, "Reprice friends?") {
					return fmt.Errorf("Reprice did not ask first; the card read:\n%s", snap.Grants)
				}
				for _, f := range s.page.Fetches[:snap.Fetches] {
					if f.Path == "/api/grants/reprice" {
						return fmt.Errorf("the reprice was sent before it was confirmed")
					}
				}
				for _, w := range []string{"The old secret has stopped working.", "1 account that added it is on the public price until they add the new one."} {
					if !strings.Contains(s.page.Grants, w) {
						return fmt.Errorf("the card does not say %q:\n%s", w, s.page.Grants)
					}
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				if len(s.reprices) != 1 || s.reprices[0]["price_out"] != 0.08 {
					return fmt.Errorf("the broker received %v", s.reprices)
				}
				return nil
			})
			sc.Step(`^"friends" reads "([^"]*)"$`, func(want string) error {
				r, ok, err := s.addedRow("friends")
				if err != nil || !ok {
					return fmt.Errorf("no added row \"friends\" (err %v):\n%s", err, s.page.Added)
				}
				if !strings.Contains(r.Text, want) {
					return fmt.Errorf("the row reads %q", r.Text)
				}
				return nil
			})
			sc.Step(`^it shows the login button and no add box$`, func() error {
				if err := s.run(); err != nil {
					return err
				}
				hasLogin := false
				for _, b := range s.page.Buttons["added"] {
					hasLogin = hasLogin || b == "Log in"
					if b == "Add" {
						return fmt.Errorf("the card offers Add:\n%s", s.page.Added)
					}
				}
				if !hasLogin {
					return fmt.Errorf("no login button:\n%s", s.page.Added)
				}
				if in, ok := s.page.Inputs["added-secret"]; ok && !in.Hidden {
					return fmt.Errorf("the add box is showing")
				}
				for _, f := range s.page.Fetches {
					if f.Path == "/api/grants/add" {
						return fmt.Errorf("the page sent an add")
					}
				}
				s.op("click", "added", "Log in")
				if err := s.run(); err != nil {
					return err
				}
				if !s.page.Login {
					return fmt.Errorf("the login button did not start the console's login")
				}
				return nil
			})
		},
		Options: &godog.Options{Format: "pretty", Paths: bddrun.ScenariosInRule(file, "The console"), TestingT: t, Strict: true},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("the Console rule of features/grants/grant_surfaces.feature has failing scenarios")
	}
}
