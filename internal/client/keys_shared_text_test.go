package client

// The words and lines every surface shows for a key or a grant come from here, so the CLI, the
// TUI and the console cannot drift apart. Every secret here is made up.

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountKeyLimitAndUsedText(t *testing.T) {
	for _, tc := range []struct {
		name        string
		key         AccountKey
		limit, used string
	}{
		{"weekly", AccountKey{LimitUSD: 5, Reset: "weekly", UsageWeekly: 1.5, Usage: 9}, "$5.00 weekly", "$1.50 used"},
		{"daily", AccountKey{LimitUSD: 0.5, Reset: "daily", UsageDaily: 0.25}, "$0.50 daily", "$0.25 used"},
		{"monthly", AccountKey{LimitUSD: 20, Reset: "monthly", UsageMonthly: 3}, "$20.00 monthly", "$3.00 used"},
		{"never resets", AccountKey{LimitUSD: 5, Reset: "none", Usage: 2}, "$5.00 total", "$2.00 used"},
		{"no window named", AccountKey{LimitUSD: 5, Usage: 2}, "$5.00 total", "$2.00 used"},
		{"unlimited", AccountKey{Reset: "weekly"}, "unlimited", "$0.00 used"},
		{"unlimited, never resets", AccountKey{Reset: "none", Usage: 4}, "unlimited", "$4.00 used"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.limit, tc.key.LimitText())
			require.Equal(t, tc.used, tc.key.UsedText())
		})
	}
}

func TestKeyExportLines(t *testing.T) {
	want := "export OPENAI_BASE_URL=https://broker.example/v1\nexport OPENAI_API_KEY=rog-key_madeup"
	require.Equal(t, want, KeyExportLines("https://broker.example", "rog-key_madeup"))
	require.Equal(t, want, KeyExportLines("https://broker.example/", "rog-key_madeup"))
}

func TestFindGrantPrefersTheLiveGrant(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, tc := range []struct {
		name   string
		grants []map[string]any
		wantID string
	}{
		{"revoked first, live second", []map[string]any{{"id": "g1", "name": "old", "status": "revoked"}, {"id": "g2", "name": "old", "status": "active"}}, "g2"},
		{"live first, revoked second", []map[string]any{{"id": "g1", "name": "old", "status": "active"}, {"id": "g2", "name": "old", "status": "revoked"}}, "g1"},
		{"only revoked", []map[string]any{{"id": "g1", "name": "old", "status": "revoked"}}, "g1"},
		{"another name", []map[string]any{{"id": "g1", "name": "other", "status": "active"}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"grants": tc.grants})
			}))
			defer broker.Close()
			g, ok, err := findGrant(broker.URL, "old")
			require.NoError(t, err)
			require.Equal(t, tc.wantID != "", ok)
			require.Equal(t, tc.wantID, g.ID)
		})
	}
}

// One create call backs `roger grant create`, the TUI and the console, so a refusal reads the
// same wherever it lands.
func TestGrantMintRefusalsReadTheSame(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, tc := range []struct {
		name, body string
		code       int
		want       string
	}{
		{"the broker says why", `{"error":{"message":"need owner"}}`, http.StatusForbidden, "need owner"},
		{"no reason given", `{}`, http.StatusBadRequest, "grant create failed (status 400)"},
		{"ok but no secret", `{"ok":true}`, http.StatusOK, "grant create failed (status 200)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer broker.Close()
			_, err := GrantCreateSecret(broker.URL, "x", true)
			require.EqualError(t, err, tc.want)
			require.EqualError(t, GrantCreate(broker.URL, GrantCreateOpts{Name: "x"}), tc.want)
		})
	}
}

// A key is named by its id or its name on every surface that takes one typed.
func TestFindKey(t *testing.T) {
	keys := []AccountKey{{ID: "key_a", Name: "ci"}, {ID: "key_b", Name: "twin"}, {ID: "key_c", Name: "twin"}, {ID: "key_d", Name: "key_a"}}
	for _, tc := range []struct {
		arg, wantID, wantErr string
	}{
		{"ci", "key_a", ""},
		{"key_b", "key_b", ""},
		{"key_a", "key_a", ""}, // an id wins over another key's name
		{"nope", "", "no key named nope"},
		{"", "", "no key named "},
		{"twin", "", "2 keys are named twin - use the id: key_b, key_c"},
	} {
		k, err := FindKey(keys, tc.arg)
		if tc.wantErr != "" {
			require.EqualError(t, err, tc.wantErr, tc.arg)
			continue
		}
		require.NoError(t, err, tc.arg)
		require.Equal(t, tc.wantID, k.ID, tc.arg)
	}
	_, err := FindKey(keys, "nope")
	require.True(t, errors.Is(err, ErrKeyNotFound))
	_, err = FindKey(keys, "twin")
	require.False(t, errors.Is(err, ErrKeyNotFound))
	require.Equal(t, "unlimited - this key can spend your whole wallet", KeyUnlimitedNote)
}
