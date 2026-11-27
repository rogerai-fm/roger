package client

// grant_adds_test.go: the caller side of a priced grant and the owner's reprice, against a
// broker stand-in that records what it was sent. The end-to-end CLI path against the REAL
// broker is pinned by cmd/rogerai-broker/grant_cli_bdd_test.go (grant_surfaces.feature).

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type addsStandIn struct {
	srv    *httptest.Server
	calls  []string
	bodies map[string]map[string]any
	added  string // the GET /account/grants grants array
	made   string // the GET /grants grants array
	status map[string]int
	reply  map[string]string
}

func newAddsStandIn(t *testing.T) *addsStandIn {
	t.Helper()
	prev := signedIn
	signedIn = func() bool { return true }
	t.Cleanup(func() { signedIn = prev })
	s := &addsStandIn{bodies: map[string]map[string]any{}, status: map[string]int{}, reply: map[string]string{},
		added: `[{"id":"g1","name":"friends","price_out":0.05,"state":"active","added_at":1}]`,
		made:  `[{"id":"m1","name":"club","price_in":0.01,"price_out":0.04,"added_by":2,"status":"active"}]`}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		s.calls = append(s.calls, key)
		require.NotEmpty(t, r.Header.Get("X-Roger-Sig"), "every call is signed")
		var b map[string]any
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &b)
		s.bodies[key] = b
		if st := s.status[key]; st != 0 {
			w.WriteHeader(st)
			_, _ = w.Write([]byte(s.reply[key]))
			return
		}
		switch {
		case key == "GET /account/grants":
			_, _ = w.Write([]byte(`{"grants":` + s.added + `,"max":10}`))
		case key == "GET /grants":
			_, _ = w.Write([]byte(`{"grants":` + s.made + `}`))
		case key == "POST /account/grants":
			_, _ = w.Write([]byte(`{"ok":true,"grant":{"id":"g2","name":"dear","price_out":0.2,"stations":[{"id":"olive-box","models":["qwen-small"],"public_out":0.1,"above_public":true}]}}`))
		case strings.HasPrefix(key, "DELETE /account/grants/"):
			_, _ = w.Write([]byte(`{"ok":true}`))
		case strings.HasSuffix(key, "/reprice"):
			_, _ = w.Write([]byte(`{"ok":true,"secret":"rog-grant_new"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func TestGrantAddShowsWhatTheBrokerShowed(t *testing.T) {
	s := newAddsStandIn(t)
	g, err := GrantAdd(s.srv.URL, "  rog-grant_abc \n")
	require.NoError(t, err)
	require.Equal(t, "dear", g.Name)
	require.True(t, g.Stations[0].AbovePublic)
	require.Equal(t, map[string]any{"secret": "rog-grant_abc"}, s.bodies["POST /account/grants"], "the secret is trimmed")
}

func TestGrantAddRefusals(t *testing.T) {
	s := newAddsStandIn(t)
	s.status["POST /account/grants"], s.reply["POST /account/grants"] = http.StatusConflict, `{"error":{"message":"an account can hold 10 grants - remove one first (`+"`roger grant remove <name>`"+`)"}}`
	_, err := GrantAdd(s.srv.URL, "x")
	require.ErrorIs(t, err, ErrGrantAddLimit)
	require.Contains(t, err.Error(), "an account can hold 10 grants")

	s.status["POST /account/grants"], s.reply["POST /account/grants"] = http.StatusConflict, `{"error":"something else"}`
	_, err = GrantAdd(s.srv.URL, "x")
	require.False(t, errors.Is(err, ErrGrantAddLimit), "only the limit refusal is the limit")

	s.status["POST /account/grants"], s.reply["POST /account/grants"] = http.StatusBadRequest, `{"error":"paste the grant secret"}`
	_, err = GrantAdd(s.srv.URL, "")
	require.EqualError(t, err, "paste the grant secret")
	require.False(t, errors.Is(err, ErrGrantAddLimit))

	s.status["POST /account/grants"], s.reply["POST /account/grants"] = http.StatusBadGateway, `not json`
	_, err = GrantAdd(s.srv.URL, "x")
	require.EqualError(t, err, "broker returned status 502")

	s.status["POST /account/grants"] = 0
	s.status["POST /account/grants"], s.reply["POST /account/grants"] = http.StatusOK, `{`
	_, err = GrantAdd(s.srv.URL, "x")
	require.ErrorContains(t, err, "could not read the reply")
}

func TestGrantCallsSendNothingWhenNotSignedIn(t *testing.T) {
	s := newAddsStandIn(t)
	signedIn = func() bool { return false }
	_, err := GrantAdd(s.srv.URL, "x")
	require.ErrorIs(t, err, ErrNotSignedIn)
	_, _, err = GrantAddedList(s.srv.URL)
	require.ErrorIs(t, err, ErrNotSignedIn)
	p := 0.08
	_, err = GrantReprice(s.srv.URL, "club", RepriceOpts{PriceOut: &p})
	require.ErrorIs(t, err, ErrNotSignedIn)
	require.Empty(t, s.calls)
	require.Equal(t, "log in first: roger login", MsgLoggedOut)
}

func TestGrantAddedListAndStateWords(t *testing.T) {
	s := newAddsStandIn(t)
	gs, max, err := GrantAddedList(s.srv.URL)
	require.NoError(t, err)
	require.Equal(t, 10, max)
	require.Equal(t, "active", gs[0].StateWords())
	require.Equal(t, "replaced - ask the owner for the new secret", AddedGrant{State: "replaced"}.StateWords())

	s.status["GET /account/grants"], s.reply["GET /account/grants"] = http.StatusUnauthorized, `{"error":"log in to add a grant"}`
	_, _, err = GrantAddedList(s.srv.URL)
	require.EqualError(t, err, "log in to add a grant")

	s.status["GET /account/grants"], s.reply["GET /account/grants"] = http.StatusOK, `[`
	_, _, err = GrantAddedList(s.srv.URL)
	require.ErrorContains(t, err, "could not read the reply")
}

func TestGrantRemove(t *testing.T) {
	s := newAddsStandIn(t)
	g, err := GrantRemove(s.srv.URL, "friends")
	require.NoError(t, err)
	require.Equal(t, "g1", g.ID)
	require.Contains(t, s.calls, "DELETE /account/grants/g1")

	_, err = GrantRemove(s.srv.URL, "g1")
	require.NoError(t, err, "by id too")

	_, err = GrantRemove(s.srv.URL, "club")
	require.EqualError(t, err, "\"club\" is a grant you made - `roger grant revoke club` ends it for everyone")

	_, err = GrantRemove(s.srv.URL, "nothing")
	require.EqualError(t, err, "no grant named \"nothing\" in your added grants")

	s.added = `[{"id":"g1","name":"friends"},{"id":"g9","name":"friends"}]`
	_, err = GrantRemove(s.srv.URL, "friends")
	require.EqualError(t, err, "2 added grants are named \"friends\" - remove one by id: g1, g9")

	s.added = `[{"id":"g1","name":"friends"}]`
	s.status["DELETE /account/grants/g1"], s.reply["DELETE /account/grants/g1"] = http.StatusNotFound, `{"error":"no such grant on this account"}`
	_, err = GrantRemove(s.srv.URL, "friends")
	require.EqualError(t, err, "no such grant on this account")

	s.status["GET /account/grants"], s.reply["GET /account/grants"] = http.StatusInternalServerError, `{"error":"store error"}`
	_, err = GrantRemove(s.srv.URL, "friends")
	require.EqualError(t, err, "store error")
}

func TestGrantAddedOnlyByMe(t *testing.T) {
	s := newAddsStandIn(t)
	require.True(t, GrantAddedOnlyByMe(s.srv.URL, "friends"))
	require.False(t, GrantAddedOnlyByMe(s.srv.URL, "club"), "a grant you made")
	require.False(t, GrantAddedOnlyByMe(s.srv.URL, "nothing"))
	s.status["GET /account/grants"] = http.StatusInternalServerError
	require.False(t, GrantAddedOnlyByMe(s.srv.URL, "friends"))
	s.status["GET /grants"] = http.StatusInternalServerError
	require.False(t, GrantAddedOnlyByMe(s.srv.URL, "friends"))
}

func TestGrantReprice(t *testing.T) {
	s := newAddsStandIn(t)
	_, err := GrantReprice(s.srv.URL, "club", RepriceOpts{})
	require.EqualError(t, err, MsgRepriceNoPrice)
	require.Empty(t, s.calls, "no price is refused before anything is sent")

	in, out := 0.02, 0.08
	nodes, models := []string{"n1", "n2"}, []string{}
	r, err := GrantReprice(s.srv.URL, "club", RepriceOpts{PriceIn: &in, PriceOut: &out, Nodes: &nodes, Models: &models})
	require.NoError(t, err)
	require.Equal(t, RepriceResult{ID: "m1", Name: "club", Secret: "rog-grant_new", Dropped: 2}, r)
	require.Equal(t, map[string]any{"price_in": 0.02, "price_out": 0.08, "nodes": []any{"n1", "n2"}, "models": []any{}}, s.bodies["POST /grants/m1/reprice"])

	// A price not stated keeps its current value; the broker would read it as 0.
	_, err = GrantReprice(s.srv.URL, "club", RepriceOpts{PriceOut: &out})
	require.NoError(t, err)
	require.Equal(t, map[string]any{"price_in": 0.01, "price_out": 0.08}, s.bodies["POST /grants/m1/reprice"])

	// A cut with no new scope is refused before anything is sent: it keeps the secret.
	s.calls = nil
	cut := 0.03
	_, err = GrantReprice(s.srv.URL, "club", RepriceOpts{PriceOut: &cut})
	require.ErrorIs(t, err, ErrGrantCut)
	require.NotContains(t, s.calls, "POST /grants/m1/reprice")
	same := 0.04
	_, err = GrantReprice(s.srv.URL, "club", RepriceOpts{PriceOut: &same, Nodes: &nodes})
	require.NoError(t, err, "a widening at the same price is a reprice")

	_, err = GrantReprice(s.srv.URL, "nothing", RepriceOpts{PriceOut: &out})
	require.EqualError(t, err, "no grant named \"nothing\" (run `roger grant list`)")

	s.status["POST /grants/m1/reprice"], s.reply["POST /grants/m1/reprice"] = http.StatusConflict, `{"error":"this grant is revoked - make a new grant instead"}`
	_, err = GrantReprice(s.srv.URL, "m1", RepriceOpts{PriceOut: &out})
	require.EqualError(t, err, "this grant is revoked - make a new grant instead")

	s.status["POST /grants/m1/reprice"], s.reply["POST /grants/m1/reprice"] = http.StatusOK, `{"ok":true}`
	_, err = GrantReprice(s.srv.URL, "m1", RepriceOpts{PriceOut: &out})
	require.ErrorContains(t, err, "could not read the new secret")

	s.status["GET /grants"] = http.StatusForbidden
	_, err = GrantReprice(s.srv.URL, "club", RepriceOpts{PriceOut: &out})
	require.ErrorContains(t, err, "linked operator account")
}

func TestGrantCurrentPricePrefersTheLiveGrant(t *testing.T) {
	s := newAddsStandIn(t)
	s.made = `[{"id":"old","name":"club","price_out":0.09,"status":"revoked"},{"id":"m1","name":"club","price_in":0.01,"price_out":0.04,"status":"active"}]`
	id, name, in, out, err := GrantCurrentPrice(s.srv.URL, "club")
	require.NoError(t, err)
	require.Equal(t, []any{"m1", "club", 0.01, 0.04}, []any{id, name, in, out})
	_, _, _, _, err = GrantCurrentPrice(s.srv.URL, "nope")
	require.Error(t, err)
}

func TestGrantAddsUnreachableBroker(t *testing.T) {
	prev := signedIn
	signedIn = func() bool { return true }
	t.Cleanup(func() { signedIn = prev })
	_, err := GrantAdd("http://127.0.0.1:1", "x")
	require.ErrorIs(t, err, ErrBrokerUnreachable)
}

func TestGrantSet(t *testing.T) {
	s := newAddsStandIn(t)
	_, err := GrantSet(s.srv.URL, "club", RepriceOpts{})
	require.EqualError(t, err, MsgNothingToChange)
	require.Empty(t, s.calls)

	s.status["PATCH /grants/m1"], s.reply["PATCH /grants/m1"] = http.StatusOK, `{"ok":true,"grant":{"name":"club","price_in":0.01,"price_out":0.03,"models":["a"]}}`
	cut, in := 0.03, 0.01
	nodes, models := []string{"n1"}, []string{"a"}
	g, err := GrantSet(s.srv.URL, "club", RepriceOpts{PriceIn: &in, PriceOut: &cut, Nodes: &nodes, Models: &models})
	require.NoError(t, err)
	require.Equal(t, GrantEdited{Name: "club", PriceIn: 0.01, PriceOut: 0.03, Models: []string{"a"}}, g)
	require.Equal(t, map[string]any{"price_in": 0.01, "price_out": 0.03, "nodes": []any{"n1"}, "models": []any{"a"}}, s.bodies["PATCH /grants/m1"])

	s.status["PATCH /grants/m1"], s.reply["PATCH /grants/m1"] = http.StatusConflict, `{"error":"a wider grant needs a new secret - reprice the grant"}`
	_, err = GrantSet(s.srv.URL, "club", RepriceOpts{Models: &models})
	require.EqualError(t, err, "a wider grant needs a new secret - reprice the grant")

	s.status["PATCH /grants/m1"], s.reply["PATCH /grants/m1"] = http.StatusOK, `{`
	_, err = GrantSet(s.srv.URL, "club", RepriceOpts{Models: &models})
	require.ErrorContains(t, err, "could not read the reply")

	_, err = GrantSet(s.srv.URL, "nothing", RepriceOpts{Models: &models})
	require.Error(t, err)
	signedIn = func() bool { return false }
	_, err = GrantSet(s.srv.URL, "club", RepriceOpts{Models: &models})
	require.ErrorIs(t, err, ErrNotSignedIn)
}

// Audit 2026-11-26: an added grant's name, stations and models are set by another account.
// Control and format characters are stripped as the reply is read, so no caller prints them.
func TestAddedGrantTextIsCleaned(t *testing.T) {
	s := newAddsStandIn(t)
	s.added = `[{"id":"g1","name":"fri\u001b[31mends\u0007‮","price_out":0.05,"state":"active","added_at":1,` +
		`"models":["qwen\u001b[2J"],"stations":[{"id":"olive\u0008-box","models":["qwen\u001b[2J"]}]}]`
	gs, _, err := GrantAddedList(s.srv.URL)
	require.NoError(t, err)
	require.Equal(t, "fri[31mends", gs[0].Name)
	require.Equal(t, []string{"qwen[2J"}, gs[0].Models)
	require.Equal(t, "olive-box", gs[0].Stations[0].ID)
	require.Equal(t, []string{"qwen[2J"}, gs[0].Stations[0].Models)
}
