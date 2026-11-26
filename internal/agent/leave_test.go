package agent

// Contract: features/tower/tower_canary_offline_station.feature.
//
// Unit coverage for leave.go, the share's half of the signed leave. The end-to-end behaviour
// (a real broker applying the leave, Ctrl-C through the real signal handler) is covered by the
// broker and cmd/rogerai runners; these pin the node side in this package: what is held, what
// is sent and signed, and that a stopping share never attaches again.

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/station"
)

type leaveCall struct {
	signer string
	body   map[string]any
}

// leaveBroker answers POST /tower/edge/leave with status and records each request whose
// signature verifies over exactly what was sent.
func leaveBroker(t *testing.T, status int) (*httptest.Server, func() []leaveCall) {
	t.Helper()
	var mu sync.Mutex
	var calls []leaveCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		ts, _ := strconv.ParseInt(r.Header.Get(protocol.HeaderTS), 10, 64)
		pub := r.Header.Get(protocol.HeaderPubkey)
		if _, ok := protocol.VerifyRequest(pub, r.Header.Get(protocol.HeaderSig), ts, r.Method, r.URL.Path, raw); !ok ||
			r.URL.Path != "/tower/edge/leave" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		mu.Lock()
		calls = append(calls, leaveCall{signer: pub, body: body})
		mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv, func() []leaveCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]leaveCall(nil), calls...)
	}
}

func resetHeld(t *testing.T) {
	t.Helper()
	held.Lock()
	held.m = map[string]leaveTarget{}
	held.Unlock()
	t.Cleanup(func() {
		held.Lock()
		held.m = map[string]leaveTarget{}
		held.Unlock()
		leaving.Store(false)
	})
}

func nodeKeyForTest(t *testing.T) (ed25519.PrivateKey, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return priv, hex.EncodeToString(pub)
}

func TestLeaveAllSendsOneSignedLeavePerHeldAttachment(t *testing.T) {
	resetHeld(t)
	srv, calls := leaveBroker(t, http.StatusOK)
	priv, pubHex := nodeKeyForTest(t)
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-1", TowerID: "tw-a", Epoch: 4})
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-2", TowerID: "tw-b", Epoch: 7})
	// A re-attach overwrites its Station's entry: the latest Tower and epoch are the ones left.
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-1", TowerID: "tw-c", Epoch: 5})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.Equal(t, 2, LeaveAll(ctx))

	got := map[string]map[string]any{}
	for _, c := range calls() {
		require.Equal(t, pubHex, c.signer, "a leave is signed by the node key it attached with")
		got[c.body["station_id"].(string)] = c.body
	}
	require.Len(t, got, 2)
	require.Equal(t, "tw-c", got["st-1"]["tower_id"])
	require.EqualValues(t, 5, got["st-1"]["epoch"])
	require.Equal(t, "tw-b", got["st-2"]["tower_id"])
	require.EqualValues(t, 7, got["st-2"]["epoch"])

	// Each attachment is left once: a second LeaveAll has nothing to send.
	require.Equal(t, 0, LeaveAll(ctx))
	require.Len(t, calls(), 2)
}

func TestHoldAttachmentIgnoresAnAttachmentThatCannotBeLeft(t *testing.T) {
	resetHeld(t)
	srv, calls := leaveBroker(t, http.StatusOK)
	priv, _ := nodeKeyForTest(t)
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "", TowerID: "tw-a", Epoch: 4})
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-1", TowerID: "tw-a", Epoch: 0})
	require.Equal(t, 0, LeaveAll(context.Background()))
	require.Empty(t, calls())
}

func TestARefusedOrUnreachableLeaveIsNotCountedAndIsNotRetried(t *testing.T) {
	resetHeld(t)
	srv, calls := leaveBroker(t, http.StatusConflict)
	priv, _ := nodeKeyForTest(t)
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-1", TowerID: "tw-a", Epoch: 4})

	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()
	holdAttachment(deadURL, priv, TowerAttachment{StationID: "st-2", TowerID: "tw-a", Epoch: 4})
	holdAttachment("http://[::1]:namedport", priv, TowerAttachment{StationID: "st-3", TowerID: "tw-a", Epoch: 4})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.Equal(t, 0, LeaveAll(ctx), "a 409, a dead broker and a bad URL are none of them accepted")
	require.Len(t, calls(), 1, "only the reachable broker saw a leave")
	require.Equal(t, 0, LeaveAll(ctx), "a failed leave is not retried; the liveness predicate covers it")
}

func TestLeaveAllGivesUpWhenItsContextEnds(t *testing.T) {
	resetHeld(t)
	block := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-block:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(block); slow.Close() })
	priv, _ := nodeKeyForTest(t)
	holdAttachment(slow.URL, priv, TowerAttachment{StationID: "st-1", TowerID: "tw-a", Epoch: 4})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	require.Equal(t, 0, LeaveAll(ctx))
	require.Less(t, time.Since(start), 2*time.Second, "LeaveAll must not outlive its context")
}

func TestAStoppingShareLeavesALateAttachmentAtOnceAndNeverAttachesAgain(t *testing.T) {
	resetHeld(t)
	srv, calls := leaveBroker(t, http.StatusOK)
	priv, _ := nodeKeyForTest(t)

	resume := StopAttaching()
	// An attach that was in flight lands after leaving began: it is left at once, not held.
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-late", TowerID: "tw-a", Epoch: 2})
	require.Len(t, calls(), 1)
	require.Equal(t, "st-late", calls()[0].body["station_id"])
	require.Equal(t, 0, LeaveAll(context.Background()), "the late attachment was never held")

	// And no new attach is sent while leaving.
	_, _, err := attachTowerAvoiding(Config{Broker: srv.URL}, priv, t.TempDir(), TowerAttachment{})
	require.ErrorIs(t, err, ErrLeaving)

	resume()
	holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-again", TowerID: "tw-a", Epoch: 3})
	require.Len(t, calls(), 1, "after resume an attachment is held again, not left at once")
}

func TestTranscriptSourceForwardsTheStationsRecords(t *testing.T) {
	// No transcript store: nothing evicted and nothing found.
	none := transcriptSource{e: station.EdgeExecutor{}}
	require.Equal(t, 0, none.EvictedYoung())
	_, _, _, ok, _ := none.SignedTranscript("att-missing")
	require.False(t, ok)

	// A real transcript store: an unknown attempt is not found, and the eviction count is read.
	src := transcriptSource{e: station.EdgeExecutor{Transcripts: station.NewTranscripts(0, 0)}}
	require.Equal(t, 0, src.EvictedYoung())
	_, _, _, ok, err := src.SignedTranscript("att-missing")
	require.NoError(t, err)
	require.False(t, ok)
}

// An attach that lands while the share starts leaving is left exactly once: holdAttachment
// decides "hold" or "leave now" under the same lock LeaveAll snapshots with, so no interleaving
// leaves a Station held and never left (pre-push audit, 2026-11-25). Deterministic: the test
// holds the lock so the attach is parked at it, and the share begins leaving meanwhile.
func TestAnAttachParkedAtTheLockWhenLeavingBeginsIsLeftAtOnce(t *testing.T) {
	resetHeld(t)
	srv, calls := leaveBroker(t, http.StatusOK)
	priv, _ := nodeKeyForTest(t)

	held.Lock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		holdAttachment(srv.URL, priv, TowerAttachment{StationID: "st-parked", TowerID: "tw-a", Epoch: 1})
	}()
	time.Sleep(50 * time.Millisecond) // let the attach reach the lock
	leaving.Store(true)               // the share begins leaving while the attach waits
	held.Unlock()
	<-done

	require.Equal(t, 0, LeaveAll(context.Background()), "the parked attach must not be left held")
	require.Len(t, calls(), 1, "it is left at once instead")
	require.Equal(t, "st-parked", calls()[0].body["station_id"])
}

// StopAttaching waits for an attach that is being recorded, so LeaveAll, which runs after it,
// always sees that attach.
func TestStopAttachingWaitsForAnAttachBeingRecorded(t *testing.T) {
	resetHeld(t)
	held.Lock()
	stopped := make(chan struct{})
	go func() { StopAttaching(); close(stopped) }()
	select {
	case <-stopped:
		held.Unlock()
		t.Fatal("StopAttaching must take the same lock holdAttachment records under")
	case <-time.After(50 * time.Millisecond):
	}
	held.Unlock()
	<-stopped
	require.True(t, leaving.Load())
}
