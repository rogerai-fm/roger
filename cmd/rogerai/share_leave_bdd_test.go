package main

// Contract: features/tower/tower_canary_offline_station.feature.
//
// The `roger share` stop sequence: on the first Ctrl-C the share sends its leaves, waiting at
// most leaveLimit, and a second Ctrl-C exits at once. What a leave does at Core runs in
// cmd/rogerai-broker against the real share code (agent.LeaveAll).

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

// TestShareLeaveHelper is the child process: the REAL Ctrl-C handler (acquireOnAirLock), with a
// leave that never lands, so the parent can signal it and time the exit.
func TestShareLeaveHelper(t *testing.T) {
	if os.Getenv("ROGER_LEAVE_HELPER") == "" {
		t.Skip("helper-process entry point (spawned by the share-stop BDD)")
	}
	shareLeave = func(ctx context.Context) int { <-ctx.Done(); return 0 }
	if _, err := acquireOnAirLock(os.Getenv("ROGER_LEAVE_NODE"), "leave-test", "m"); err != nil {
		fmt.Println("REFUSED:", err)
		os.Exit(3)
	}
	fmt.Println("READY")
	time.Sleep(30 * time.Second) // the parent signals us; bounded so nothing leaks
	os.Exit(5)
}

type shareLeaveState struct {
	t      *testing.T
	cmd    *exec.Cmd
	out    *bufio.Reader
	start  time.Time
	waited chan error
	took   time.Duration
	code   int
}

func (s *shareLeaveState) inFlight() error {
	cfg := s.t.TempDir()
	s.t.Setenv("XDG_CONFIG_HOME", cfg)
	s.t.Setenv("ROGER_CONFIG_DIR", filepath.Join(cfg, "roger"))
	s.cmd = exec.Command(os.Args[0], "-test.run", "^TestShareLeaveHelper$")
	s.cmd.Env = append(os.Environ(), "ROGER_LEAVE_HELPER=1",
		fmt.Sprintf("ROGER_LEAVE_NODE=leave-test-%d", time.Now().UnixNano()))
	pipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := s.cmd.Start(); err != nil {
		return err
	}
	s.t.Cleanup(func() { _ = s.cmd.Process.Kill() })
	s.out = bufio.NewReader(pipe)
	for {
		line, err := s.out.ReadString('\n')
		if err != nil {
			return fmt.Errorf("the helper never became ready: %v", err)
		}
		if strings.HasPrefix(line, "READY") {
			break
		}
	}
	s.waited = make(chan error, 1)
	go func() { _, _ = io.Copy(io.Discard, s.out); s.waited <- s.cmd.Wait() }()
	// The first Ctrl-C: the real handler starts the leave, which will not land.
	s.start = time.Now()
	if err := s.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (s *shareLeaveState) secondCtrlC() error {
	return s.cmd.Process.Signal(syscall.SIGINT)
}

func (s *shareLeaveState) exitsAtOnce() error {
	select {
	case err := <-s.waited:
		s.took = time.Since(s.start)
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			s.code = ee.ExitCode()
		}
	case <-time.After(5 * time.Second):
		return errors.New("the share did not exit after the second Ctrl-C")
	}
	if s.code != 130 {
		return fmt.Errorf("the share exited %d, want 130", s.code)
	}
	if s.took >= leaveLimit {
		return fmt.Errorf("the share waited %v, the full leave limit, despite the second Ctrl-C", s.took)
	}
	return nil
}

func TestShareLeaveOnStopBDD(t *testing.T) {
	s := &shareLeaveState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = shareLeaveState{t: t}
				return ctx, nil
			})
			// The feature's Background is about Core; nothing here reads it.
			sc.Step(`^a live joined Tower T that may take work$`, func() error { return nil })
			sc.Step(`^nodeTTL is 45 seconds$`, func() error { return nil })
			sc.Step(`^the leave is still in flight$`, s.inFlight)
			sc.Step(`^the operator presses Ctrl-C again$`, s.secondCtrlC)
			sc.Step(`^the share exits at once, and the liveness predicate covers S$`, s.exitsAtOnce)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/tower_canary_offline_station.feature",
				bddrun.Titled(false, "A second Ctrl-C exits at once"))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("tower_canary_offline_station.feature (share stop) failed")
	}
}

// The other two exits of leaveBeforeExit: the leaves land, or the limit passes.
func TestLeaveBeforeExitReturnsWhenTheLeavesLandOrTheLimitPasses(t *testing.T) {
	start := time.Now()
	leaveBeforeExit(func(context.Context) int { return 1 }, make(chan os.Signal), time.Minute)
	if d := time.Since(start); d > time.Second {
		t.Fatalf("waited %v after the leaves landed", d)
	}
	start = time.Now()
	leaveBeforeExit(func(ctx context.Context) int { <-ctx.Done(); return 0 }, make(chan os.Signal), 100*time.Millisecond)
	if d := time.Since(start); d > time.Second || d < 100*time.Millisecond {
		t.Fatalf("a leave that never lands held the share %v, want the 100ms limit", d)
	}
}
