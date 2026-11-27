package client

// Units for ParseGrantPrice and GrantCreatePriced (features/grants/grant_surfaces.feature S8,
// features/grants/priced_grants.feature's price examples). A loopback broker stand-in; no mocks.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGrantPrice(t *testing.T) {
	cases := []struct {
		typed string
		want  float64
		err   string
	}{
		{"0.05", 0.05, ""},
		{" $0.05 ", 0.05, ""},
		{"0", 0, ""},
		{"$0", 0, ""},
		{"0.00", 0, ""},
		{"free", 0, ""},
		{"FREE", 0, ""},
		{"12", 12, ""},
		{"-1", 0, MsgGrantPriceNegative},
		{"-$0.5", 0, MsgGrantPriceNegative},
		{"cheap", 0, MsgRepriceNoPrice},
		{"abc", 0, MsgRepriceNoPrice},
		{"1e9", 0, MsgRepriceNoPrice},
		{"NaN", 0, MsgRepriceNoPrice},
		{"Inf", 0, MsgRepriceNoPrice},
		{"", 0, MsgRepriceNoPrice},
		{"$-1", 0, MsgRepriceNoPrice},
		{"0.", 0, MsgRepriceNoPrice},
		{".5", 0, MsgRepriceNoPrice},
		{"1,5", 0, MsgRepriceNoPrice},
		{"0x10", 0, MsgRepriceNoPrice},
	}
	for _, tc := range cases {
		t.Run(tc.typed, func(t *testing.T) {
			v, err := ParseGrantPrice(tc.typed)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, v)
		})
	}
}

func TestGrantCreatePriced(t *testing.T) {
	// The call is signed: keep any key it makes out of the real config dir.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var got map[string]any
	refuse := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		if refuse {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "price above the ceiling"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"grant": map[string]any{"name": "club", "price": "$0.05/1M out"}, "secret": "rog-grant_TESTONLYc"})
	}))
	defer srv.Close()

	price, secret, err := GrantCreatePriced(srv.URL, "club", 0.05)
	require.NoError(t, err)
	require.Equal(t, "$0.05/1M out", price)
	require.Equal(t, "rog-grant_TESTONLYc", secret)
	require.Equal(t, map[string]any{"name": "club", "free": false, "price_out": 0.05}, got)

	refuse = true
	_, _, err = GrantCreatePriced(srv.URL, "club", 99)
	require.EqualError(t, err, "price above the ceiling")
}
