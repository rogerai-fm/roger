package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A broker that accepts the call and never answers must not hang the caller: the console's
// key card waits on this read before it lets anyone mint.
func TestFetchBalanceGivesUpOnASilentBroker(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)

	was := balanceTimeout
	balanceTimeout = 100 * time.Millisecond
	defer func() { balanceTimeout = was }()

	start := time.Now()
	_, err := FetchBalance(srv.URL, "u")
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestFetchBalanceHasARealTimeout(t *testing.T) {
	require.Equal(t, 30*time.Second, balanceTimeout, "the same wait as every other signed account call")
}
