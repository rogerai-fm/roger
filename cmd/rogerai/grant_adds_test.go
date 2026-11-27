package main

// grant_adds_test.go: `roger grant add / remove / reprice / list` in-process, against a broker
// stand-in, for the branches a subprocess run does not count toward coverage. The full path
// against the REAL broker is cmd/rogerai-broker/grant_cli_bdd_test.go (grant_surfaces.feature).

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type grantStandIn struct {
	srv   *httptest.Server
	calls []string
	added string
	made  string
	body  map[string]any
}

func newGrantStandIn(t *testing.T, signed bool) *grantStandIn {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	if signed {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "rogerai"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "rogerai", "auth.json"), []byte(`{"github_login":"frank"}`), 0o600))
	}
	s := &grantStandIn{
		added: `[{"id":"g1","name":"friends","price_out":0.05,"state":"replaced","added_at":86400}]`,
		made:  `[{"id":"m1","name":"club","price_out":0.05,"added_by":3,"status":"active","price":"$0/$0.05"}]`,
	}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		s.calls = append(s.calls, key)
		raw, _ := io.ReadAll(r.Body)
		s.body = nil
		_ = json.Unmarshal(raw, &s.body)
		switch {
		case key == "GET /account/grants":
			_, _ = w.Write([]byte(`{"grants":` + s.added + `,"max":10}`))
		case key == "GET /grants":
			_, _ = w.Write([]byte(`{"grants":` + s.made + `}`))
		case key == "POST /account/grants":
			if s.body["secret"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"paste the grant secret"}`))
				return
			}
			_, _ = w.Write([]byte(`{"grant":{"id":"g2","name":"dear","price_in":0.01,"price_out":0.2,"stations":[{"id":"olive-box","models":["qwen-small"],"public_out":0.1,"above_public":true}]}}`))
		case strings.HasPrefix(key, "DELETE /"):
			_, _ = w.Write([]byte(`{"ok":true}`))
		case r.Method == http.MethodPatch:
			if s.body["price_out"] == 0.09 {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":"a higher price needs a new secret - reprice the grant"}`))
				return
			}
			_, _ = w.Write([]byte(`{"ok":true,"grant":{"name":"club","price_out":0.03,"models":["m1"]}}`))
		case strings.HasSuffix(key, "/reprice"):
			_, _ = w.Write([]byte(`{"ok":true,"secret":"rog-grant_new"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func runGrant(t *testing.T, s *grantStandIn, stdin string, args ...string) (string, error) {
	t.Helper()
	prev := grantIn
	grantIn = strings.NewReader(stdin)
	t.Cleanup(func() { grantIn = prev })
	var err error
	out := captureStdout(t, func() { err = cmdGrant(config{Broker: s.srv.URL}, args) })
	return out, err
}

func TestGrantAddPrintsPriceStationsAndNeverTheSecret(t *testing.T) {
	s := newGrantStandIn(t, true)
	out, err := runGrant(t, s, "", "add", "rog-grant_abc")
	require.NoError(t, err)
	require.Contains(t, out, `added "dear": $0.01 per 1M input and $0.20 per 1M output tokens`)
	require.Contains(t, out, "olive-box")
	require.Contains(t, out, "note: on olive-box the grant price is above the public price.")
	require.NotContains(t, out, "rog-grant_abc")

	out, err = runGrant(t, s, "rog-grant_piped\n", "add")
	require.NoError(t, err)
	require.Equal(t, "rog-grant_piped", s.body["secret"], "a pipe is read for the secret")
	require.NotContains(t, out, "rog-grant_piped")

	_, err = runGrant(t, s, "", "add")
	require.EqualError(t, err, "paste the grant secret")
}

func TestGrantAddLoggedOutSendsNothing(t *testing.T) {
	s := newGrantStandIn(t, false)
	for _, args := range [][]string{{"add", "x"}, {"list"}, {"reprice", "club", "--price-out", "0.08", "--yes"}} {
		_, err := runGrant(t, s, "", args...)
		require.EqualError(t, err, "log in first: roger login", args)
	}
	require.Empty(t, s.calls)
}

func TestGrantListShowsBothSections(t *testing.T) {
	s := newGrantStandIn(t, true)
	out, err := runGrant(t, s, "", "list")
	require.NoError(t, err)
	i, j := strings.Index(out, "grants you made"), strings.Index(out, "grants you added")
	require.True(t, i >= 0 && j > i, out)
	require.Contains(t, out, "1970-01-02")
	require.Contains(t, out, "replaced - ask the owner for the new secret")

	s.added = `[]`
	out, err = runGrant(t, s, "", "list")
	require.NoError(t, err)
	require.Contains(t, out, "none - `roger grant add` adds a grant someone gave you")
}

func TestGrantRemoveAndRevokeNameEachOther(t *testing.T) {
	s := newGrantStandIn(t, true)
	out, err := runGrant(t, s, "", "remove", "friends")
	require.NoError(t, err)
	require.Contains(t, out, `removed "friends" from your account`)

	_, err = runGrant(t, s, "", "remove")
	require.EqualError(t, err, "usage: roger grant remove <name|id>")

	_, err = runGrant(t, s, "", "revoke", "friends")
	require.ErrorContains(t, err, "`roger grant remove friends`")

	_, err = runGrant(t, s, "", "revoke", "nothing")
	require.ErrorContains(t, err, `no grant named "nothing"`)
}

func TestGrantReprice(t *testing.T) {
	s := newGrantStandIn(t, true)
	out, err := runGrant(t, s, "n\n", "reprice", "club", "--price-out", "0.08")
	require.NoError(t, err)
	require.Contains(t, out, "nothing changed.")
	require.NotContains(t, s.calls, "POST /grants/m1/reprice")

	out, err = runGrant(t, s, "y\n", "reprice", "club", "--price-out", "$0.08", "--price-in", "0.01", "--nodes", "a,b", "--models", "")
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(out, "rog-grant_new"))
	require.Contains(t, out, "the old secret has stopped working.")
	require.Contains(t, out, "3 accounts that added it are on the public price until they add the new one.")
	require.Equal(t, map[string]any{"price_out": 0.08, "price_in": 0.01, "nodes": []any{"a", "b"}, "models": []any{}}, s.body)

	s.made = `[{"id":"m1","name":"club","price_out":0.05,"added_by":1,"status":"active"}]`
	out, err = runGrant(t, s, "", "reprice", "--yes", "club", "--price-out", "free")
	require.Error(t, err, "free is a cut")
	require.Empty(t, out)

	out, err = runGrant(t, s, "", "reprice", "club", "--price-out", "0.05", "--yes")
	require.NoError(t, err)
	require.Contains(t, out, "1 account that added it is on the public price")

	s.made = `[{"id":"m1","name":"club","price_out":0.05,"added_by":0,"status":"active"}]`
	out, err = runGrant(t, s, "", "reprice", "club", "--price-out", "0.06", "--yes")
	require.NoError(t, err)
	require.Contains(t, out, "no account had added it.")

	for _, bad := range [][]string{
		{"reprice", "club"},
		{"reprice", "club", "--price-out", "abc"},
		{"reprice", "club", "--price-in", "NaN"},
	} {
		_, err := runGrant(t, s, "", bad...)
		require.EqualError(t, err, `state a price in dollars per 1M tokens, or "free"`, bad)
	}
	_, err = runGrant(t, s, "", "reprice", "club", "--price-out", "-1")
	require.EqualError(t, err, "price cannot be negative")
	_, err = runGrant(t, s, "", "create", "--name", "x", "--price-out", "abc")
	require.EqualError(t, err, `state a price in dollars per 1M tokens, or "free"`, "create reads a price like reprice")
	_, err = runGrant(t, s, "", "reprice")
	require.ErrorContains(t, err, "usage: roger grant reprice")
	_, err = runGrant(t, s, "", "reprice", "club", "--bogus")
	require.Error(t, err)
	_, err = runGrant(t, s, "", "reprice", "nothing", "--price-out", "0.1", "--yes")
	require.ErrorContains(t, err, `no grant named "nothing"`)
}

func TestGrantShowSecretOnAPricedGrantNamesReprice(t *testing.T) {
	s := newGrantStandIn(t, true)
	s.made = `[{"id":"m1","name":"club","price_in":0.01,"price_out":0.05,"status":"active","price":"$0.01/$0.05"}]`
	_, err := runGrant(t, s, "", "show", "--secret", "club")
	require.ErrorContains(t, err, "`roger grant reprice club --price-out 0.05 --price-in 0.01`")
	require.NotContains(t, s.calls, "DELETE /grants/m1", "nothing is revoked")
}

func TestGrantPriceWords(t *testing.T) {
	require.Equal(t, "free", grantPriceWords(0, 0))
	require.Equal(t, "$0.30 per 1M input tokens", grantPriceWords(0.3, 0))
	require.Equal(t, "$1.00 per 1M output tokens", grantPriceWords(0, 1))
	require.Equal(t, "$0.125 per 1M output tokens", grantPriceWords(0, 0.125))
}

// A name reused after a revoke means the live grant: recovering it never revokes the live
// priced grant and mints a free one (2026-11-24 review).
func TestGrantShowSecretPicksTheLiveGrant(t *testing.T) {
	s := newGrantStandIn(t, true)
	s.made = `[{"id":"old","name":"friends","free":true,"price":"free","status":"revoked"},` +
		`{"id":"m1","name":"friends","price_out":0.05,"status":"active","price":"$0/$0.05"}]`
	_, err := runGrant(t, s, "", "show", "--secret", "friends")
	require.ErrorContains(t, err, "`roger grant reprice friends --price-out 0.05`")
	for _, c := range s.calls {
		require.False(t, strings.HasPrefix(c, "DELETE") || c == "POST /grants", "nothing is revoked or minted: %v", s.calls)
	}
}

func TestGrantSet(t *testing.T) {
	s := newGrantStandIn(t, true)
	out, err := runGrant(t, s, "", "set", "club", "--price-out", "0.03", "--models", "m1")
	require.NoError(t, err)
	require.Contains(t, out, `updated "club", same secret: $0.03 per 1M output tokens, models m1.`)
	require.Equal(t, map[string]any{"price_out": 0.03, "models": []any{"m1"}}, s.body, "only what is stated is sent")
	require.Contains(t, s.calls, "PATCH /grants/m1")

	_, err = runGrant(t, s, "", "set", "club", "--price-out", "0.09")
	require.EqualError(t, err, "a higher price needs a new secret - reprice the grant")

	s.calls = nil
	_, err = runGrant(t, s, "", "set", "club")
	require.EqualError(t, err, "nothing to change")
	require.Empty(t, s.calls)

	_, err = runGrant(t, s, "", "set", "friends", "--price-out", "0.01")
	require.ErrorContains(t, err, "`roger grant remove friends`")
	_, err = runGrant(t, s, "", "set", "nothing", "--price-out", "0.01")
	require.ErrorContains(t, err, `no grant named "nothing"`)
	_, err = runGrant(t, s, "", "set")
	require.ErrorContains(t, err, "usage: roger grant set")
	_, err = runGrant(t, s, "", "set", "club", "--price-out", "abc")
	require.Error(t, err)
	require.Equal(t, "any", scopeWords(nil))

	_, err = runGrant(t, s, "", "reprice", "club", "--price-out", "0.01")
	require.ErrorContains(t, err, "`roger grant set club --price-out 0.01` lowers it at once")
}
