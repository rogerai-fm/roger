package towercli

// Contract: features/tower/tower_canary_offline_station.feature.
//
// The `roger tower serve` half of tower_canary_offline_station.feature, Rule 3: a refusal Core
// may lift keeps the process up and asking again on the node re-attach backoff; one it cannot
// lift exits with exitTerminal, which the packaged unit does not restart. The Core half (which
// refusals name a state, and to whom) runs in cmd/rogerai-broker.
//
// Core is the package's HTTP stub, answering exactly as the broker does: a 403 envelope whose
// error.metadata.tower_state names the state for a request verified as the Tower's own, and the
// generic sentence otherwise.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/tower"
	"rogerai.fm/roger/v7/internal/towercore/link"
)

type refusalState struct {
	t       *testing.T
	core    *coreStub
	st      *tower.State
	out     *syncBuffer
	stop    chan struct{}
	beats   chan time.Time
	done    chan error
	err     error
	refusal string
	waitMu  sync.Mutex
	waits   []time.Duration
	held    atomic.Bool // the stub keeps refusing while this is set
	code    int
	stderr  string
	waited  time.Duration
}

func (s *refusalState) wait(d time.Duration) <-chan time.Time {
	s.waitMu.Lock()
	s.waits = append(s.waits, d)
	s.waitMu.Unlock()
	c := make(chan time.Time, 1)
	c <- time.Now()
	return c
}

// refuse answers like the broker's towerCallerOr: a 403 envelope, naming the state when the
// request was verified as the Tower's own (state != "").
func refuse(w http.ResponseWriter, state string) {
	w.WriteHeader(http.StatusForbidden)
	meta := map[string]any{}
	if state != "" {
		meta["tower_state"] = state
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
		"code": "tower_refused", "message": "this link requires a registered Tower's own signed request",
		"metadata": meta}})
}

// hangUp drops the connection: Core unreachable, from the Tower's side.
func hangUp(w http.ResponseWriter) {
	if hj, ok := w.(http.Hijacker); ok {
		if c, _, err := hj.Hijack(); err == nil {
			_ = c.Close()
		}
	}
}

func (s *refusalState) start() {
	s.out = &syncBuffer{}
	s.stop = make(chan struct{})
	s.beats = make(chan time.Time, 1)
	s.done = make(chan error, 1)
	go func() { s.done <- runLink(s.st, s.out, s.stop, manualTicker(s.beats), link.RelayPlane{}) }()
}

// until waits for cond while runLink keeps running; a runLink that returns first is a failure.
func (s *refusalState) until(cond func() bool, what string) error {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-s.done:
			s.done <- err
			return fmt.Errorf("serve returned (%v) before %s:\n%s", err, what, s.out.String())
		default:
		}
		if cond() {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s:\n%s", what, s.out.String())
}

func (s *refusalState) answer(state string, w http.ResponseWriter) {
	if state == "Core unreachable" {
		hangUp(w)
		return
	}
	refuse(w, state)
}

func (s *refusalState) meetsAt(refusal, step string) error {
	s.refusal = refusal
	s.core = newCoreStub(s.t)
	s.st = servingTower(s.t)
	const refusals = 3
	s.core.mu.Lock()
	switch step {
	case "the first session open":
		s.core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool {
			if n <= refusals {
				s.answer(refusal, w)
				return true
			}
			return false
		}
	case "the first inventory push":
		s.core.reply["/tower/inventory"] = func(w http.ResponseWriter, n int) bool {
			if n <= refusals {
				s.answer(refusal, w)
				return true
			}
			return false
		}
	case "the re-open after a refused heartbeat":
		s.core.reply["/tower/session/heartbeat"] = func(w http.ResponseWriter, n int) bool {
			refuse(w, "")
			return true
		}
		s.core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool {
			if n >= 2 && n <= 1+refusals {
				s.answer(refusal, w)
				return true
			}
			return false
		}
	case "the full push after a re-open":
		s.core.reply["/tower/session/heartbeat"] = func(w http.ResponseWriter, n int) bool {
			refuse(w, "")
			return true
		}
		s.core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool {
			if n == 1 {
				return false
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"session_id": fmt.Sprintf("sess-%d", n),
				"heartbeat_seconds": 30, "freshness_seconds": 180, "need_full_inventory": true})
			return true
		}
		s.core.reply["/tower/inventory"] = func(w http.ResponseWriter, n int) bool {
			if n >= 2 && n <= 1+refusals {
				s.answer(refusal, w)
				return true
			}
			return false
		}
	default:
		s.core.mu.Unlock()
		return fmt.Errorf("unknown step %q", step)
	}
	s.core.mu.Unlock()
	s.start()
	if strings.Contains(step, "heartbeat") || strings.Contains(step, "re-open") {
		if err := s.until(func() bool { return s.core.called("/tower/inventory") >= 1 }, "the first link"); err != nil {
			return err
		}
		s.beats <- time.Now()
	}
	var target func() bool
	switch step {
	case "the first session open":
		target = func() bool { return s.core.sessionCount() >= 1+refusals && s.core.called("/tower/inventory") >= 1 }
	case "the first inventory push":
		target = func() bool { return s.core.called("/tower/inventory") >= 1+refusals }
	case "the re-open after a refused heartbeat":
		target = func() bool { return s.core.sessionCount() >= 2+refusals }
	case "the full push after a re-open":
		target = func() bool { return s.core.called("/tower/inventory") >= 2+refusals }
	}
	return s.until(target, "the retries after "+refusal)
}

func (s *refusalState) phrase() string {
	switch s.refusal {
	case "state_suspended":
		return "RogerAI has suspended this Tower"
	case "lease_lapsed":
		return "this Tower's lease has lapsed"
	case "Core unreachable":
		return "could not reach RogerAI"
	}
	return s.refusal
}

func (s *refusalState) printsOnce() error {
	out := s.out.String()
	if n := strings.Count(out, s.phrase()); n != 1 {
		return fmt.Errorf("%q printed %d times:\n%s", s.phrase(), n, out)
	}
	if !strings.Contains(out, "asks again") {
		return fmt.Errorf("serve did not say what happens next:\n%s", out)
	}
	return nil
}

func (s *refusalState) backsOff() error {
	s.waitMu.Lock()
	waits := append([]time.Duration(nil), s.waits...)
	s.waitMu.Unlock()
	if len(waits) != 3 {
		return fmt.Errorf("serve waited %d times, want 3: %v", len(waits), waits)
	}
	for i, d := range waits {
		full := 30 * time.Second << i
		if d <= full/2 || d > full {
			return fmt.Errorf("wait %d was %v, want in (%v, %v]", i, d, full/2, full)
		}
	}
	for i := 0; i < 64; i++ {
		if d := linkBackoff(20); d <= 7*time.Minute+30*time.Second || d > 15*time.Minute {
			return fmt.Errorf("the capped backoff was %v, want in (7m30s, 15m]", d)
		}
	}
	return nil
}

func (s *refusalState) keepsRunning() error {
	select {
	case err := <-s.done:
		return fmt.Errorf("serve exited while it should be waiting: %v", err)
	default:
	}
	close(s.stop)
	select {
	case err := <-s.done:
		if err != nil {
			return fmt.Errorf("serve ended with %v after a clean stop", err)
		}
	case <-time.After(5 * time.Second):
		return errors.New("serve did not stop")
	}
	return nil
}

// ---- terminal refusals ----

func (s *refusalState) meets(refusal string) error {
	s.refusal = refusal
	s.core = newCoreStub(s.t)
	switch refusal {
	case "a joined Tower started in standalone mode, or the reverse":
		dir := filepath.Join(s.t.TempDir(), "tw")
		var b bytes.Buffer
		if err := run([]string{"init", "--dir", dir, "--mode", "standalone"}, &b); err != nil {
			return err
		}
		st, release, err := openDir(dir)
		if err != nil {
			return err
		}
		s.t.Cleanup(func() { _ = release() })
		s.st = st
	default:
		s.st = servingTower(s.t)
		state := ""
		if refusal == "state_revoked" || refusal == "cert_revoked" {
			state = refusal
		}
		s.core.mu.Lock()
		s.core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool { refuse(w, state); return true }
		s.core.mu.Unlock()
	}
	// A refusal that names no state is waited out for 24 hours before it is terminal: run serve's
	// clock forward by every wait it takes, so the day passes in the test.
	var elapsed atomic.Int64
	start := time.Now()
	prevNow, prevWait := serveNow, serveWait
	serveNow = func() time.Time { return start.Add(time.Duration(elapsed.Load())) }
	serveWait = func(d time.Duration) <-chan time.Time {
		elapsed.Add(int64(d))
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	defer func() { serveNow, serveWait = prevNow, prevWait }()
	var b bytes.Buffer
	s.err = runLink(s.st, &b, make(chan struct{}), manualTicker(nil), link.RelayPlane{})
	s.waited = time.Duration(elapsed.Load())
	return nil
}

func (s *refusalState) printsReason() error {
	if s.err == nil {
		return errors.New("serve did not stop")
	}
	want := map[string]string{
		"state_revoked":                     "revoked this Tower permanently",
		"cert_revoked":                      "revoked this Tower's certificate",
		"unknown_tower":                     "without saying why for 24 hours",
		"a request signed by the wrong key": "without saying why for 24 hours",
		"a joined Tower started in standalone mode, or the reverse": "this Tower is standalone",
	}[s.refusal]
	if !strings.Contains(s.err.Error(), want) {
		return fmt.Errorf("serve said %q, want it to say %q", s.err, want)
	}
	stateless := s.refusal == "unknown_tower" || s.refusal == "a request signed by the wrong key"
	if stateless && s.waited < 24*time.Hour {
		return fmt.Errorf("a refusal that names no state was final after %v, want 24 hours of retries", s.waited)
	}
	if !stateless && s.waited != 0 {
		return fmt.Errorf("a named terminal state waited %v before stopping", s.waited)
	}
	return nil
}

func (s *refusalState) exitsTerminal() error {
	if c := exitCode(s.err); c != exitTerminal {
		return fmt.Errorf("serve would exit %d, want %d", c, exitTerminal)
	}
	return nil
}

func (s *refusalState) unitPreventsRestart() error {
	raw, err := os.ReadFile("../../packaging/tower/roger-tower.service")
	if err != nil {
		return err
	}
	if !strings.Contains(string(raw), fmt.Sprintf("RestartPreventExitStatus=%d", exitTerminal)) {
		return errors.New("the packaged unit does not list the terminal status in RestartPreventExitStatus")
	}
	return nil
}

// ---- the hub cannot bind ----

func (s *refusalState) hubCannotBind() error {
	s.core = newCoreStub(s.t)
	s.core.answerDispatchKey(s.t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.t.Cleanup(func() { _ = ln.Close() })
	dir := joinedRegisteredTower(s.t)
	var out, errb bytes.Buffer
	s.code = Run([]string{"serve", "--dir", dir, "--hub", ln.Addr().String()}, &out, &errb)
	s.stderr = errb.String()
	return nil
}

func (s *refusalState) hubTerminal() error {
	if s.code != exitTerminal || !strings.Contains(s.stderr, "the hub cannot bind") {
		return fmt.Errorf("serve exited %d saying %q", s.code, s.stderr)
	}
	return nil
}

// ---- restored while waiting ----

func (s *refusalState) backingOffSuspended() error {
	s.refusal = "state_suspended"
	s.core = newCoreStub(s.t)
	s.st = servingTower(s.t)
	s.held.Store(true)
	s.core.mu.Lock()
	s.core.reply["/tower/session"] = func(w http.ResponseWriter, n int) bool {
		if s.held.Load() {
			refuse(w, "state_suspended")
			return true
		}
		return false
	}
	s.core.mu.Unlock()
	s.start()
	return s.until(func() bool { return s.core.sessionCount() >= 2 }, "two refused asks")
}

func (s *refusalState) adminRestores() error { s.held.Store(false); return nil }

func (s *refusalState) carriesAgain() error {
	if err := s.until(func() bool {
		return s.core.called("/tower/inventory") >= 1 && strings.Contains(s.out.String(), "linked to RogerAI")
	}, "the restored Tower to link"); err != nil {
		return err
	}
	return s.keepsRunning()
}

var refusalServeTitles = []string{
	"A recoverable refusal keeps `roger tower serve` running",
	"A terminal refusal stops `roger tower serve` with a status that is not restarted",
	"A hub that cannot start is a terminal local error",
	"A refused Tower that is restored carries traffic again without a restart",
}

func TestTowerRefusalServeBDD(t *testing.T) {
	s := &refusalState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = refusalState{t: t}
				prev := serveWait
				serveWait = s.wait
				t.Cleanup(func() { serveWait = prev })
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				serveWait = func(d time.Duration) <-chan time.Time { return time.After(d) }
				if s.stop != nil {
					select {
					case <-s.stop:
					default:
						close(s.stop)
					}
				}
				return ctx, nil
			})
			// The feature's Background is about Core; nothing here reads it.
			sc.Step(`^a live joined Tower T that may take work$`, func() error { return nil })
			sc.Step(`^nodeTTL is 45 seconds$`, func() error { return nil })

			sc.Step("^`roger tower serve` meets \"([^\"]*)\" at \"([^\"]*)\"$", s.meetsAt)
			sc.Step(`^it prints the state in plain words, with what the operator can do, once per state$`, s.printsOnce)
			sc.Step(`^it stays running and asks Core again on the node re-attach backoff \(30 s, capped at 15 min, jittered\)$`, s.backsOff)
			sc.Step(`^it does not exit and is not restarted by its service manager$`, s.keepsRunning)

			sc.Step("^`roger tower serve` meets \"([^\"]*)\"$", s.meets)
			sc.Step(`^it prints the reason in plain words and what the operator must do$`, s.printsReason)
			sc.Step(`^it exits with the dedicated terminal status$`, s.exitsTerminal)
			sc.Step(`^the packaged service unit lists that status in RestartPreventExitStatus$`, s.unitPreventsRestart)

			sc.Step(`^the hub cannot bind its address$`, s.hubCannotBind)
			sc.Step("^`roger tower serve` prints why and exits with the terminal status$", s.hubTerminal)

			sc.Step("^`roger tower serve` is backing off after \"state_suspended\"$", s.backingOffSuspended)
			sc.Step(`^an administrator restores T and it holds a fresh lease$`, s.adminRestores)
			sc.Step(`^the running process takes work again on its next retry$`, s.carriesAgain)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/tower_canary_offline_station.feature",
				bddrun.Titled(false, refusalServeTitles...))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("tower_canary_offline_station.feature (serve) failed")
	}
}
