package towerhub

// Contract: features/tower/station_mobility_idempotency.feature (review fixes 2026-11-24).

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func unknownPollRig(t *testing.T, hook func(string)) (*Server, *httptest.Server) {
	t.Helper()
	s := NewServer(New(), stubCheck, ServerOptions{TowerID: testTowerID, EpochKey: testHubKey,
		SubmitTTL: time.Second, PollTTL: 20 * time.Millisecond})
	s.OnUnknownStation = hook
	mux := http.NewServeMux()
	mux.HandleFunc(PathPoll, s.Poll)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return s, srv
}

// A failure on a Station the hub DOES know (a wrong key, a stale epoch after a hub restart) is
// never a reason to ask Core.
func TestAFailedPollFromAKnownStationNeverAsksCore(t *testing.T) {
	var asked atomic.Int64
	s, srv := unknownPollRig(t, func(string) { asked.Add(1) })
	owner, stranger := newTestNode(t), newTestNode(t)
	s.RegisterNode("st-known", owner.auth())
	for i := 0; i < 5; i++ {
		_, _, err := stranger.client(srv.URL, 0).PollJob(context.Background(), "st-known")
		require.Error(t, err)
		time.Sleep(250 * time.Millisecond)
	}
	require.Zero(t, asked.Load(), "a refused poll from a known Station asked Core")
}

// Unknown-id polls never queue behind an ask in progress: they get today's 401 at once.
func TestUnknownPollsNeverWaitOnAnotherAsk(t *testing.T) {
	var asked atomic.Int64
	_, srv := unknownPollRig(t, func(string) { asked.Add(1); time.Sleep(400 * time.Millisecond) })
	id := newTestNode(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var slow []time.Duration
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			_, _, _ = id.client(srv.URL, 0).PollJob(context.Background(), fmt.Sprintf("st-x-%d", i))
			if d := time.Since(start); d > 300*time.Millisecond {
				mu.Lock()
				slow = append(slow, d)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	require.LessOrEqual(t, asked.Load(), int64(1), "200 unknown polls in one second asked Core %d times", asked.Load())
	require.LessOrEqual(t, len(slow), 1, "polls waited on another poll's ask: %v", slow)
}
