package webui

// Units for grant_adds.go and the priced create in keys.go: every refusal path of the added
// grants routes and Reprice, against the same loopback broker stand-in the Console rule of
// features/grants/grant_surfaces.feature runs on. No mocks. Every secret here is made up.

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/client"
)

func newGrantAddsRig(t *testing.T) *gsBDD {
	t.Helper()
	s := &gsBDD{t: t}
	s.reset()
	t.Cleanup(func() { s.cleanup() })
	s.made = append(s.made, gsMade{ID: "gr_friends", Name: "friends", Price: "$0.05/1M out", PriceOut: 0.05, AddedBy: 3, Status: "active"})
	s.signIn(true)
	return s
}

func TestGrantAddsRoutes(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		path     string
		body     string
		setup    func(*gsBDD)
		code     int
		says     string
		notSays  string
		brokerOK bool // the broker may be asked
	}{
		{name: "added list refuses POST", method: "POST", path: "/api/grants/added", body: `{}`, code: 405, says: "method not allowed"},
		{name: "added list logged out", method: "GET", path: "/api/grants/added", setup: func(s *gsBDD) { s.signIn(false) }, code: 401, brokerOK: true},
		{name: "added list without a local sign-in", method: "GET", path: "/api/grants/added",
			setup: func(s *gsBDD) { s.signIn(false); s.loggedIn = true }, code: 401, says: grantsLogin, brokerOK: true},
		{name: "added list broker refusal", method: "GET", path: "/api/grants/added",
			setup: func(s *gsBDD) { s.broker.Config.Handler = brokenList(s.broker.Config.Handler) }, code: 502, says: "list is down", brokerOK: true},
		{name: "added list", method: "GET", path: "/api/grants/added", setup: func(s *gsBDD) { s.added = append(s.added, gsFriendsCopy("expired")) },
			code: 200, says: `"state":"expired"`, brokerOK: true},
		{name: "add empty secret", method: "POST", path: "/api/grants/add", body: `{"secret":"  "}`, code: 400, says: "paste the grant secret"},
		{name: "add malformed", method: "POST", path: "/api/grants/add", body: `{"secret":1}`, code: 400, says: "bad request"},
		{name: "add unknown field", method: "POST", path: "/api/grants/add", body: `{"secret":"x","name":"y"}`, code: 400, says: `unknown field`},
		{name: "add refused in the broker's words, secret not echoed", method: "POST", path: "/api/grants/add", body: `{"secret":"rog-grant_TESTONLYnope"}`,
			code: 502, says: "grant key invalid or revoked", notSays: "TESTONLYnope", brokerOK: true},
		{name: "add logged out", method: "POST", path: "/api/grants/add", body: `{"secret":"rog-grant_TESTONLYnope"}`,
			setup: func(s *gsBDD) { s.signIn(false) }, code: 401, notSays: "TESTONLYnope", brokerOK: true},
		{name: "remove id required", method: "POST", path: "/api/grants/remove", body: `{}`, code: 400, says: "id required"},
		{name: "remove malformed", method: "POST", path: "/api/grants/remove", body: `nope`, code: 400, says: "bad request"},
		{name: "remove one never added", method: "POST", path: "/api/grants/remove", body: `{"id":"gr_none"}`, code: 502, says: "no grant named", brokerOK: true},
		{name: "remove logged out", method: "POST", path: "/api/grants/remove", body: `{"id":"gr_friends"}`, setup: func(s *gsBDD) { s.signIn(false) }, code: 401, brokerOK: true},
		{name: "remove", method: "POST", path: "/api/grants/remove", body: `{"id":"gr_friends"}`,
			setup: func(s *gsBDD) { s.added = append(s.added, gsFriendsCopy("active")) }, code: 200, says: `"name":"friends"`, brokerOK: true},
		{name: "reprice name required", method: "POST", path: "/api/grants/reprice", body: `{"price_out":"0.08"}`, code: 400, says: "name required"},
		{name: "reprice malformed", method: "POST", path: "/api/grants/reprice", body: `[`, code: 400, says: "bad request"},
		{name: "reprice no price", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends"}`, code: 400, says: "state a price in dollars per 1M tokens"},
		{name: "reprice blank price", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends","price_out":" "}`, code: 400, says: "state a price"},
		{name: "reprice negative", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends","price_out":"-1"}`, code: 400, says: "price cannot be negative"},
		{name: "reprice exponent", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends","price_out":1e9}`, code: 400, says: "state a price"},
		{name: "reprice logged out", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends","price_out":"0.08"}`,
			setup: func(s *gsBDD) { s.signIn(false) }, code: 401, brokerOK: true},
		{name: "reprice unknown grant", method: "POST", path: "/api/grants/reprice", body: `{"name":"club","price_out":"0.08"}`, code: 502, says: "no grant named", brokerOK: true},
		{name: "reprice as a number, many dropped", method: "POST", path: "/api/grants/reprice", body: `{"name":"friends","price_out":0.08}`,
			code: 200, says: "3 accounts that added it are on the public price", brokerOK: true},
		{name: "create priced", method: "POST", path: "/api/grants", body: `{"name":"club","price_out":"$0.05"}`, code: 200, says: `"secret":"rog-grant_TESTONLYmade`, brokerOK: true},
		{name: "create free price", method: "POST", path: "/api/grants", body: `{"name":"club","price_out":"free"}`, code: 200, says: `"ok":true`, brokerOK: true},
		{name: "create refused price", method: "POST", path: "/api/grants", body: `{"name":"club","price_out":"abc"}`, code: 400, says: "state a price in dollars per 1M tokens"},
		{name: "create priced, broker down", method: "POST", path: "/api/grants", body: `{"name":"club","price_out":"0.05"}`,
			setup: func(s *gsBDD) { s.broker.Config.Handler = brokenCreate(s.broker.Config.Handler) }, code: 502, says: "no more grants today", brokerOK: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newGrantAddsRig(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			s.mu.Lock()
			was := s.hits
			s.mu.Unlock()
			r, err := s.do(tc.method, tc.path, tc.body)
			require.NoError(t, err)
			require.Equal(t, tc.code, r.code, r.raw)
			require.Contains(t, r.raw, tc.says)
			if tc.notSays != "" {
				require.NotContains(t, r.raw, tc.notSays)
			}
			if !tc.brokerOK {
				s.mu.Lock()
				require.Equal(t, was, s.hits, "the broker was asked")
				s.mu.Unlock()
			}
		})
	}
}

func (s *gsBDD) do(method, path, body string) (plateReply, error) {
	req, err := http.NewRequest(method, s.console.URL+path, strings.NewReader(body))
	if err != nil {
		return plateReply{}, err
	}
	req.Header.Set("X-Roger-Token", s.tok)
	return send(req)
}

// brokenList and brokenCreate wrap the stand-in so one route answers with a refusal.
func brokenList(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/account/grants" && r.Method == http.MethodGet {
			keysErr(w, http.StatusInternalServerError, "", "the added list is down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func brokenCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/grants" && r.Method == http.MethodPost {
			keysErr(w, http.StatusTooManyRequests, "", "no more grants today")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TestGrantsFailWords(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
		says string
	}{
		{"unreachable", fmt.Errorf("%w: dial", client.ErrBrokerUnreachable), 503, keysUnreachable},
		{"not signed in", client.ErrNotSignedIn, 401, grantsLogin},
		{"keys login", client.ErrKeysLogin, 401, grantsLogin},
		{"other", errors.New("rog-grant_TESTONLYx is odd"), 502, "[secret] is odd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			grantsFail(w, tc.err, "rog-grant_TESTONLYx")
			require.Equal(t, tc.code, w.Code)
			require.Contains(t, w.Body.String(), tc.says)
			require.NotContains(t, w.Body.String(), "TESTONLYx")
		})
	}
}

func TestRepriceNote(t *testing.T) {
	require.Equal(t, "The old secret has stopped working. No account had added it.", repriceNote(0))
	require.Contains(t, repriceNote(1), "1 account that added it is on the public price")
	require.Contains(t, repriceNote(4), "4 accounts that added it are on the public price")
}

func TestAddedRowOf(t *testing.T) {
	g := client.AddedGrant{ID: "g", Name: "n", Price: "$0.20/1M out", State: "replaced",
		Stations: []client.AddedGrantStation{{ID: "olive-box", AbovePublic: true, PublicOut: 0.1},
			{ID: "olive-2", PublicIn: 0.02, PublicOut: 0.3}, {ID: "olive-3"}}}
	r := addedRowOf(g)
	// S4: each station with its public price beside it, as the CLI and the TUI show it.
	require.Equal(t, []string{"olive-box (public $0.1/1M out)", "olive-2 (public $0.02 in $0.3 out /1M)", "olive-3 (public free)"}, r.Stations)
	require.Equal(t, []string{}, r.Models)
	require.Equal(t, "", r.Added)
	require.Equal(t, "replaced - ask the owner for the new secret", r.State)
	require.Equal(t, []string{"olive-box: the grant price is above its public price"}, r.Notes)

	none := addedRowOf(client.AddedGrant{Name: "n", AddedAt: 1795996800})
	require.Equal(t, []string{"It covers no station yet. Add it again later."}, none.Notes)
	require.Equal(t, "2026-11-30", none.Added)
}

// Audit 2026-11-26: a merge left a conflict marker in console.css and dropped a rule. No
// embedded asset may carry one.
func TestAssetsHaveNoConflictMarkers(t *testing.T) {
	err := fs.WalkDir(assetsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(assetsFS, path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(raw), "\n") {
			if line == "=======" || strings.HasPrefix(line, "<<<<<<< ") || strings.HasPrefix(line, ">>>>>>> ") {
				t.Errorf("%s:%d has a merge conflict marker", path, i+1)
			}
		}
		return nil
	})
	require.NoError(t, err)
}
