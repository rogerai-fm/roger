package towercli

// Contract: features/tower/tower_canary_offline_station.feature.
//
// Review fixes (2026-11-24) for `roger tower serve`'s refusals.

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/towercore/link"
	"rogerai.fm/roger/v7/internal/towerjoin"
)

// B1: a 403 that names no state - an old Core, an outage read as "unknown", a transient - is
// waited out, not taken as final on first sight. It becomes terminal only after it has lasted
// 24 hours of retries, and the operator is told that up front.
func TestAStatelessRefusalIsWaitedOutForADay(t *testing.T) {
	core := newCoreStub(t)
	st := servingTower(t)
	core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool {
		if n <= 3 {
			refuse(w, "")
			return true
		}
		return false
	}
	var mu sync.Mutex
	prev := serveWait
	serveWait = func(time.Duration) <-chan time.Time {
		mu.Lock()
		defer mu.Unlock()
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	t.Cleanup(func() { serveWait = prev })
	out := &syncBuffer{}
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- runLink(st, out, stop, manualTicker(nil), link.RelayPlane{}) }()
	deadline := time.Now().Add(10 * time.Second)
	for core.called("/tower/inventory") == 0 && time.Now().Before(deadline) {
		select {
		case err := <-done:
			t.Fatalf("serve stopped on a stateless 403 (%v):\n%s", err, out.String())
		default:
		}
		time.Sleep(5 * time.Millisecond)
	}
	require.NotZero(t, core.called("/tower/inventory"), "serve never linked after the refusals ended:\n%s", out.String())
	require.Contains(t, wrappedJoin(out.String()), "24 hours", "the operator was not told how long serve keeps trying")
	close(stop)
	require.NoError(t, <-done)
}

// B1: after 24 hours of the same stateless refusal, serve stops with the terminal status.
func TestAStatelessRefusalLastingADayIsTerminal(t *testing.T) {
	core := newCoreStub(t)
	st := servingTower(t)
	core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool { refuse(w, ""); return true }
	var elapsed atomic.Int64
	prevWait, prevNow := serveWait, serveNow
	start := time.Now()
	serveNow = func() time.Time { return start.Add(time.Duration(elapsed.Load())) }
	serveWait = func(d time.Duration) <-chan time.Time {
		elapsed.Add(int64(d))
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	t.Cleanup(func() { serveWait, serveNow = prevWait, prevNow })
	var b bytes.Buffer
	err := runLink(st, &b, make(chan struct{}), manualTicker(nil), link.RelayPlane{})
	require.Error(t, err)
	require.Equal(t, exitTerminal, exitCode(err), "a day of stateless refusals: %v", err)
	require.GreaterOrEqual(t, time.Duration(elapsed.Load()), 24*time.Hour)
}

// The classifier, row by row: only a NAMED terminal state, or a local fault, is terminal at once.
func TestOnlyANamedTerminalStateIsTerminalAtOnce(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{&towerjoin.Refusal{Status: 403, State: "state_revoked"}, refusalTerminal},
		{&towerjoin.Refusal{Status: 403, State: "cert_revoked"}, refusalTerminal},
		{&towerjoin.Refusal{Status: 403}, refusalStateless},
		{&towerjoin.Refusal{Status: 403, State: "state_suspended"}, refusalRecoverable},
		{&towerjoin.Refusal{Status: 429}, refusalRecoverable},
		{&towerjoin.Refusal{Status: 503}, refusalRecoverable},
		{fmt.Errorf("%w: boom", towerjoin.ErrUnreachable), refusalRecoverable},
		{errors.New("something else"), refusalOther},
	}
	for _, c := range cases {
		kind, words := classifyRefusal(c.err, "")
		require.Equal(t, c.want, kind, "%v (%s)", c.err, words)
	}
	_, w := classifyRefusal(&towerjoin.Refusal{Status: 403}, "")
	require.True(t, strings.Contains(wrappedJoin(w), "24 hours"), w)
}

// S6: a suspension that comes back after the link was restored is announced again.
func TestASecondSuspensionIsAnnouncedAgain(t *testing.T) {
	core := newCoreStub(t)
	st := servingTower(t)
	var suspended atomic.Bool
	var opened, refused atomic.Int32
	core.reply["/tower/session/heartbeat"] = func(w http.ResponseWriter, _ int) bool { refuse(w, ""); return true }
	core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool {
		if suspended.Load() {
			refused.Add(1)
			refuse(w, "state_suspended")
			return true
		}
		opened.Add(1)
		return false
	}
	prev := serveWait
	serveWait = func(time.Duration) <-chan time.Time {
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	t.Cleanup(func() { serveWait = prev })
	out := &syncBuffer{}
	stop := make(chan struct{})
	beats := make(chan time.Time, 1)
	done := make(chan error, 1)
	go func() { done <- runLink(st, out, stop, manualTicker(beats), link.RelayPlane{}) }()

	until := func(cond func() bool, what string) {
		deadline := time.Now().Add(10 * time.Second)
		for !cond() && time.Now().Before(deadline) {
			time.Sleep(2 * time.Millisecond)
		}
		require.True(t, cond(), "%s:\n%s", what, out.String())
	}
	until(func() bool { return opened.Load() >= 1 }, "the first link")
	for round := int32(1); round <= 2; round++ {
		suspended.Store(true)
		refusedBefore := refused.Load()
		beats <- time.Now()
		until(func() bool { return refused.Load() >= refusedBefore+2 }, "two refused re-opens")
		suspended.Store(false)
		until(func() bool { return opened.Load() >= round+1 }, "the re-open that lands")
	}
	close(stop)
	require.NoError(t, <-done)
	require.Equal(t, 2, strings.Count(wrappedJoin(out.String()), "RogerAI has suspended this Tower"),
		"suspend -> restore -> suspend should be announced twice:\n%s", out.String())
}

// Pre-push audit MAJOR: a Tower that boots while Core is briefly unreachable waits for Core and
// then starts its hub - it does not exit 78. Here Core's grant key is down for the first three
// asks and answers after; the link is then refused as revoked only so that serve returns.
func TestAHubWhoseGrantKeyCannotBeFetchedWaitsForCore(t *testing.T) {
	core := newCoreStub(t)
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	core.reply["/tower/dispatch/key"] = func(w http.ResponseWriter, n int) bool {
		if n <= 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"dispatch_key": hex.EncodeToString(pub)})
		return true
	}
	core.answerHubNodes(`{"nodes":[]}`)
	core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool { refuse(w, "state_revoked"); return true }
	runADay(t)
	st := servingTower(t)
	out := &syncBuffer{}
	err = serveJoined(st, out, "203.0.113.9:8444", hubOptions{Addr: "127.0.0.1:0"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "revoked", "serve ended on the hub, not on the link:\n%s", out.String())
	require.Equal(t, 4, core.called("/tower/dispatch/key"), "the grant key was not asked again until Core answered")
	require.Equal(t, 1, core.sessionCount(), "the link opened before the hub was up, or never")
	require.Equal(t, 1, strings.Count(out.String(), "grant key"), "the wait was not said exactly once:\n%s", out.String())
	for _, advice := range []string{"fix --hub", "different --hub", "cannot bind"} {
		require.NotContains(t, out.String(), advice, "a Core outage was described as a bind fault")
	}
}

// Pre-push audit MAJOR: a hub address that cannot be bound is a local fault - exit 78, and the
// line names the bind.
func TestAHubThatCannotBindExitsTerminalNamingTheBind(t *testing.T) {
	core := newCoreStub(t)
	core.answerDispatchKey(t)
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = taken.Close() })
	dir := joinedRegisteredTower(t)
	var out, errb bytes.Buffer
	code := Run([]string{"serve", "--dir", dir, "--hub", taken.Addr().String()}, &out, &errb)
	require.Equal(t, exitTerminal, code, errb.String())
	require.Contains(t, errb.String(), "the hub cannot bind "+taken.Addr().String(),
		"the operator line must name the bind in its own words, not only Go's")
}

// Pre-push audit MINOR: a refused first push does not leave sessions open behind it, and the
// operator is told "linked" once.
func TestARefusedFirstPushDoesNotLeakSessions(t *testing.T) {
	core := newCoreStub(t)
	st := servingTower(t)
	core.reply["/tower/inventory"] = func(w http.ResponseWriter, n int) bool {
		if n <= 2 {
			refuse(w, "lease_lapsed")
			return true
		}
		return false
	}
	prev := serveWait
	serveWait = func(time.Duration) <-chan time.Time {
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	t.Cleanup(func() { serveWait = prev })
	out := &syncBuffer{}
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- runLink(st, out, stop, manualTicker(nil), link.RelayPlane{}) }()
	deadline := time.Now().Add(10 * time.Second)
	for core.called("/tower/inventory") < 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	close(stop)
	require.NoError(t, <-done)
	opened := core.sessionCount()
	require.GreaterOrEqual(t, core.called("/tower/session/close"), opened, "%d sessions opened, %d closed:\n%s",
		opened, core.called("/tower/session/close"), out.String())
	require.Equal(t, 1, strings.Count(out.String(), "linked to RogerAI"), out.String())
}

// Pre-push audit (2026-11-24): after a refused HEARTBEAT the Tower just re-opens. It sends no
// close: Core may already hold the replacement session elsewhere, and a close is what drops a
// Tower's routable fleet.
func TestAHeartbeatRefusalReopensWithoutAClose(t *testing.T) {
	core := newCoreStub(t)
	st := servingTower(t)
	core.reply["/tower/session/heartbeat"] = func(w http.ResponseWriter, n int) bool {
		if n == 1 {
			refuse(w, "")
			return true
		}
		return false
	}
	out := &syncBuffer{}
	stop := make(chan struct{})
	beats := make(chan time.Time, 1)
	done := make(chan error, 1)
	go func() { done <- runLink(st, out, stop, manualTicker(beats), link.RelayPlane{}) }()
	deadline := time.Now().Add(10 * time.Second)
	for core.called("/tower/inventory") < 1 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	beats <- time.Now()
	for core.sessionCount() < 2 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	require.Equal(t, 2, core.sessionCount(), out.String())
	require.Zero(t, core.called("/tower/session/close"), "a heartbeat refusal sent a close before re-opening")
	close(stop)
	require.NoError(t, <-done)
	require.Equal(t, 1, core.called("/tower/session/close"), "the drain on the way out")
}

// Pre-push audit (2026-11-24): the grant-key endpoint is public, so NO answer from it is a fault
// on this machine. A 429, a 404 from a proxy mid-deploy, or a body that is not a key each keep
// serve running, and the hub starts once the key answers.
func TestEveryGrantKeyFailureIsWaitedOut(t *testing.T) {
	cases := map[string]func(w http.ResponseWriter){
		"429":     func(w http.ResponseWriter) { w.WriteHeader(http.StatusTooManyRequests) },
		"404":     func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) },
		"garbage": func(w http.ResponseWriter) { _, _ = w.Write([]byte(`<html>not a key</html>`)) },
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			core := newCoreStub(t)
			pub, _, err := ed25519.GenerateKey(rand.Reader)
			require.NoError(t, err)
			core.reply["/tower/dispatch/key"] = func(w http.ResponseWriter, n int) bool {
				if n <= 2 {
					bad(w)
					return true
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"dispatch_key": hex.EncodeToString(pub)})
				return true
			}
			core.answerHubNodes(`{"nodes":[]}`)
			core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool { refuse(w, "state_revoked"); return true }
			runADay(t)
			out := &syncBuffer{}
			err = serveJoined(servingTower(t), out, "203.0.113.9:8444", hubOptions{Addr: "127.0.0.1:0"})
			require.Error(t, err)
			require.Contains(t, err.Error(), "revoked", "serve ended on the grant key, not the link:\n%s", out.String())
			require.Equal(t, 3, core.called("/tower/dispatch/key"))
			require.Equal(t, 1, core.sessionCount(), "the hub never started")
			require.Contains(t, out.String(), "WAITING - could not reach RogerAI for its grant key", out.String())
			require.Contains(t, out.String(), "checks work with (", "the operator line lost the reason:\n%s", out.String())
			require.NotContains(t, out.String(), "<nil>", "the operator line lost the reason")
		})
	}
}
