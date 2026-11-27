package webui

// keys_test.go: the key routes' refusals that the feature file does not reach one by one - a
// console with no broker, a broker that fails each call, and ids that must never reach a
// broker path.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/node"
)

// keysConsole is a console over a stand-in broker whose /account/keys and /grants answers the
// test scripts. It returns the console's URL and how many key or grant calls the broker saw.
func keysConsole(t *testing.T, keysReply func(w http.ResponseWriter, r *http.Request)) (string, *atomic.Int32) {
	t.Helper()
	var asked atomic.Int32
	broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/balance" {
			_, _ = w.Write([]byte(`{"user":"op","balance":1,"logged_in":true}`))
			return
		}
		asked.Add(1)
		keysReply(w, r)
	}))
	t.Cleanup(broker.Close)
	srv := New(node.New(node.Config{Station: "amber-fox"}), Options{User: "op", Broker: broker.URL})
	srv.token = "tok"
	console := httptest.NewServer(srv.Handler())
	t.Cleanup(console.Close)
	return console.URL, &asked
}

func keysPost(t *testing.T, base, path, body string) (int, string) {
	t.Helper()
	method := http.MethodPost
	if body == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, base+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("X-Roger-Token", "tok")
	r, err := send(req)
	require.NoError(t, err)
	msg, _ := r.js["message"].(string)
	return r.code, msg
}

func TestKeyRoutesCarryABrokerFailureInItsWords(t *testing.T) {
	base, _ := keysConsole(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/grants" {
			_, _ = w.Write([]byte(`{"grants":[]}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "store error"}})
	})
	for _, tc := range []struct{ name, path, body, want string }{
		{"list", "/api/keys", "", "store error"},
		{"mint", "/api/keys/mint", `{"name":"ci","limit_usd":5}`, "store error"},
		{"set", "/api/keys/set", `{"id":"key_aaaa1","disabled":true}`, "store error"},
		{"revoke", "/api/keys/revoke", `{"id":"key_aaaa1"}`, "store error"},
		{"grant revoke of a grant that is not there", "/api/grants/revoke", `{"name":"nosuch"}`, `no grant named "nosuch"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, msg := keysPost(t, base, tc.path, tc.body)
			require.Equal(t, http.StatusBadGateway, code)
			require.Equal(t, tc.want, msg)
		})
	}
}

func TestKeyRoutesRefuseBeforeTheBrokerIsAsked(t *testing.T) {
	base, asked := keysConsole(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("the broker was asked: %s %s", r.Method, r.URL.Path)
	})
	for _, tc := range []struct{ name, path, body, want string }{
		{"set: an id with a path in it", "/api/keys/set", `{"id":"../../grants/x","disabled":true}`, "bad key id"},
		{"set: an id with a query in it", "/api/keys/set", `{"id":"key_a?x=1","disabled":true}`, "bad key id"},
		{"set: a negative limit", "/api/keys/set", `{"id":"key_aaaa1","limit_usd":-0.01}`, "a limit cannot be negative"},
		{"set: an unknown window", "/api/keys/set", `{"id":"key_aaaa1","reset":"hourly"}`, "reset is daily, weekly, monthly or none"},
		{"revoke: an empty id after trimming nothing", "/api/keys/revoke", `{"id":"key_"}`, "bad key id"},
		{"revoke: a grant id is not a key id", "/api/keys/revoke", `{"id":"gr_abc"}`, "bad key id"},
		{"mint: a name of spaces", "/api/keys/mint", `{"name":"   ","limit_usd":5}`, "name required"},
		{"grant revoke: not JSON", "/api/grants/revoke", `nope`, "bad request"},
		{"revoke: not JSON", "/api/keys/revoke", `nope`, "bad request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, msg := keysPost(t, base, tc.path, tc.body)
			require.Equal(t, http.StatusBadRequest, code)
			require.Equal(t, tc.want, msg)
		})
	}
	require.Zero(t, asked.Load())
}

func TestKeyListIsAnEmptyListNotNull(t *testing.T) {
	base, _ := keysConsole(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"keys":null}`)) })
	req, _ := http.NewRequest(http.MethodGet, base+"/api/keys", nil)
	req.Header.Set("X-Roger-Token", "tok")
	r, err := send(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.code)
	require.JSONEq(t, `{"keys":[]}`, r.raw)
}

func TestKeyRoutesWithNoBrokerConfigured(t *testing.T) {
	srv := New(node.New(node.Config{Station: "amber-fox"}), Options{User: "op"})
	srv.token = "tok"
	console := httptest.NewServer(srv.Handler())
	defer console.Close()
	for _, tc := range []struct{ path, body string }{
		{"/api/keys", ""},
		{"/api/keys/mint", `{"name":"ci","limit_usd":5}`},
		{"/api/keys/set", `{"id":"key_aaaa1","disabled":true}`},
		{"/api/keys/revoke", `{"id":"key_aaaa1"}`},
		{"/api/grants/revoke", `{"name":"friends"}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			code, _ := keysPost(t, console.URL, tc.path, tc.body)
			require.Equal(t, http.StatusServiceUnavailable, code)
		})
	}
}

// The page module drops a shown secret when told the operator left. These are the two places
// the console tells it: a tab change and a logout. The module itself is executed in the feature.
func TestConsolePageDropsAShownSecretOnTabChangeAndLogout(t *testing.T) {
	src, err := assetsFS.ReadFile("assets/console.js")
	require.NoError(t, err)
	js := string(src)
	require.Contains(t, js, `if (name === "account") loadAccount(); else if (keysCard) keysCard.leave();`)
	require.Contains(t, js, `toast("Signed out.", "ok"); if (keysCard) keysCard.leave(); loadAccount();`)
	html, err := assetsFS.ReadFile("assets/console.html")
	require.NoError(t, err)
	for _, want := range []string{`id="card-keys-body"`, `id="card-grants-body"`, `src="/assets/console-keys.js"`} {
		require.Contains(t, string(html), want)
	}
}
