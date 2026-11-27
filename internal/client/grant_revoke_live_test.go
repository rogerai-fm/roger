package client

// A grant name can be used again after a revoke. Revoking by name means the live grant, and a
// name with only revoked grants is said to be revoked already. A logged-out key call matches
// ErrKeysLogin so a surface with its own login can say so in its own words.

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrantRevokeByNamePicksTheLiveGrant(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, tc := range []struct {
		name    string
		grants  []map[string]any
		wantDel string
		wantErr string
	}{
		{"revoked first, live second", []map[string]any{{"id": "g1", "name": "old", "status": "revoked"}, {"id": "g2", "name": "old", "status": "active"}}, "/grants/g2", ""},
		{"live first, revoked second", []map[string]any{{"id": "g1", "name": "old", "status": "active"}, {"id": "g2", "name": "old", "status": "revoked"}}, "/grants/g1", ""},
		{"an expired grant can still be revoked", []map[string]any{{"id": "g1", "name": "old", "status": "expired"}}, "/grants/g1", ""},
		{"only revoked", []map[string]any{{"id": "g1", "name": "old", "status": "revoked"}}, "", `grant "old" is already revoked`},
		{"another name", []map[string]any{{"id": "g1", "name": "other", "status": "active"}}, "", `no grant named "old"`},
		{"none", nil, "", `no grant named "old"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deleted := ""
			broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					deleted = r.URL.Path
					_, _ = w.Write([]byte(`{"ok":true,"revoked":true}`))
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"grants": tc.grants})
			}))
			defer broker.Close()
			err := GrantRevokeByName(broker.URL, "old")
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantDel, deleted)
		})
	}
}

func TestKeysCallLoggedOutIsErrKeysLogin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	broker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer broker.Close()
	_, err := KeysList(broker.URL)
	require.ErrorIs(t, err, ErrKeysLogin)
	require.EqualError(t, err, "log in first - run `roger login` (keys are per account)")
	require.False(t, errors.Is(err, ErrBrokerUnreachable))
}
