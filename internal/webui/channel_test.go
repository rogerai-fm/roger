package webui

// channel_test.go: the endpoint plate's edges the feature's scenarios do not walk - a wrong
// method on a read, and a broker that cannot be reached.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/node"
)

func plateCall(t *testing.T, s *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:50000"
	req.Header.Set("X-Roger-Token", s.Token())
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func TestChannelReadsAndAnUnreachableBroker(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close() // a broker address nothing answers on
	down := New(node.New(node.Config{}), Options{Broker: dead.URL, User: "op"})
	bare := New(node.New(node.Config{}), Options{})
	lan := New(node.New(node.Config{}), Options{Broker: dead.URL, LANDial: func() client.LANDial { return client.LANDial{} }})

	for _, tc := range []struct {
		name         string
		srv          *Server
		method, path string
		body         string
		code         int
		says         string
	}{
		{"a POST to the status is refused", down, "POST", "/api/channel", "", 405, "method not allowed"},
		{"a POST to the quote is refused", down, "POST", "/api/channel/quote?model=m", "", 405, "method not allowed"},
		{"a quote with no broker", bare, "GET", "/api/channel/quote?model=m", "", 503, "needs one"},
		{"a quote in LAN mode", lan, "GET", "/api/channel/quote?model=m", "", 200, "LAN mode"},
		{"a quote the broker cannot answer", down, "GET", "/api/channel/quote?model=m", "", 502, "connect"},
		{"an open the broker cannot answer", down, "POST", "/api/channel/open", `{"model":"m"}`, 502, "connect"},
		{"the status with nothing open", down, "GET", "/api/channel", "", 200, `"open":false`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := plateCall(t, tc.srv, tc.method, tc.path, tc.body)
			require.Equal(t, tc.code, rec.Code, rec.Body.String())
			require.Contains(t, rec.Body.String(), tc.says)
		})
	}
	// Nothing above bound an endpoint or minted a key.
	require.Equal(t, 409, plateCall(t, down, "POST", "/api/channel/key", "").Code)
}

func TestChannelLoginFallsBackToWhatThisRunKnows(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	ctrl := node.New(node.Config{})
	s := New(ctrl, Options{Broker: dead.URL})
	require.False(t, s.loggedIn(), "the broker cannot say and this run never logged in")
	ctrl.SetLoggedIn(true)
	require.True(t, s.loggedIn(), "the broker cannot say, but this run logged in")
}

func TestChannelCapIsTheOperatorsOrTheConsumerDefault(t *testing.T) {
	limits := map[string]SpendLimit{"capped": {MaxOut: 1.5, MinTPS: 40}, "floor-only": {MinTPS: 20}}
	with := New(node.New(node.Config{}), Options{ReadLimits: func() map[string]SpendLimit { return limits }})
	without := New(node.New(node.Config{}), Options{})
	for _, tc := range []struct {
		srv        *Server
		model      string
		capOut, tp float64
	}{
		{with, "capped", 1.5, 40},
		{with, "floor-only", client.ConsumerDefaultMaxOut, 20},
		{with, "unset", client.ConsumerDefaultMaxOut, 0},
		{without, "capped", client.ConsumerDefaultMaxOut, 0},
	} {
		capOut, minTPS := tc.srv.capFor(tc.model)
		require.Equal(t, tc.capOut, capOut, tc.model)
		require.Equal(t, tc.tp, minTPS, tc.model)
	}
}

// plateBroker is a broker stand-in for the two reads an open makes: the offers and the login.
func plateBroker(t *testing.T, offer client.Offer, loggedIn *bool, onBalance func()) *httptest.Server {
	t.Helper()
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/discover":
			_ = json.NewEncoder(w).Encode(map[string]any{"offers": []client.Offer{offer}})
		case "/balance":
			was := *loggedIn // answered as it stood when asked; onBalance is what happens next
			if onBalance != nil {
				onBalance()
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"logged_in": was})
		default:
			_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
		}
	}))
	t.Cleanup(b.Close)
	return b
}

func TestChannelTypedPriceMatchesAtTheShownPrecision(t *testing.T) {
	in := true
	broker := plateBroker(t, client.Offer{NodeID: "n", Model: "big", PriceOut: 25.125, Online: true}, &in, nil)
	limits := map[string]SpendLimit{"big": {MaxOut: 30}}
	s := New(node.New(node.Config{}), Options{Broker: broker.URL, ReadLimits: func() map[string]SpendLimit { return limits }})
	t.Cleanup(func() { s.closeChannel(false) })
	for _, tc := range []struct {
		typed string
		ok    bool
	}{
		{"25.12", false}, // the page shows 25.13
		{"25.14", false},
		{"25.13", true},
	} {
		rec := plateCall(t, s, "POST", "/api/channel/open", `{"model":"big","confirm_price_out":`+tc.typed+`}`)
		var out channelReply
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		require.Equal(t, tc.ok, out.OK, tc.typed)
		require.Equal(t, !tc.ok, out.ConfirmNeeded, tc.typed)
	}
}

func TestChannelOnAFreeWindowOfAPaidBandClosesAtLogout(t *testing.T) {
	out := false // not logged in: only the free-now window lets it open
	broker := plateBroker(t, client.Offer{NodeID: "n", Model: "m", PriceOut: 0.5, Online: true, FreeNow: true}, &out, nil)
	restore := logoutReturn
	logoutReturn = func() error { return nil }
	t.Cleanup(func() { logoutReturn = restore })
	s := New(node.New(node.Config{}), Options{Broker: broker.URL})
	t.Cleanup(func() { s.closeChannel(false) })

	require.Contains(t, plateCall(t, s, "POST", "/api/channel/open", `{"model":"m"}`).Body.String(), `"ok":true`)
	require.Equal(t, 200, plateCall(t, s, "POST", "/api/account/logout", "").Code)
	// The window ends and the channel would outlive it on a paid band, so the logout closes it.
	require.Contains(t, plateCall(t, s, "GET", "/api/channel", "").Body.String(), `"open":false`)
}

func TestChannelOpenChecksAgainWhenALogoutLandsMidOpen(t *testing.T) {
	in := true
	var s *Server
	asked := 0
	broker := plateBroker(t, client.Offer{NodeID: "n", Model: "m", PriceOut: 0.5, Online: true}, &in, func() {
		asked++
		if asked == 1 { // the logout lands after this open's login check has been answered "yes"
			in = false
			s.closeChannel(true)
		}
	})
	s = New(node.New(node.Config{}), Options{Broker: broker.URL})
	t.Cleanup(func() { s.closeChannel(false) })

	rec := plateCall(t, s, "POST", "/api/channel/open", `{"model":"m"}`)
	require.Equal(t, 2, asked, "the open asked about the login again")
	require.Contains(t, rec.Body.String(), `"login_needed":true`)
	require.Contains(t, plateCall(t, s, "GET", "/api/channel", "").Body.String(), `"open":false`)
}
