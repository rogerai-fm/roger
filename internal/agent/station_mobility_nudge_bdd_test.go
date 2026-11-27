package agent

// Contract: features/tower/station_mobility_idempotency.feature.
//
// The node half of R3: a nudge is heard while the node backs off, and a nudge from before the
// current attach is dropped. The broker half (inline moves, the skip) runs in cmd/rogerai-broker.
// The nudge travels the real path a node uses: deliverNudge, which the cancel long-poll calls, onto
// the node's nudge channel, read by reattachWait (ServeTower's backoff waits) and by the tenancy.

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

var smnSeq atomic.Int64

type smnState struct {
	node    string
	waited  chan time.Duration
	cancel  context.CancelFunc
	started time.Time
	n       nudge
	read    bool
}

func (s *smnState) backoff(minutes int) func() error {
	return func() error {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		s.waited = make(chan time.Duration, 1)
		begin := time.Now()
		go func() {
			if reattachWait(ctx, time.Duration(minutes)*time.Minute, s.node) {
				s.waited <- time.Since(begin)
			}
		}()
		time.Sleep(20 * time.Millisecond) // the node is now inside its wait
		return nil
	}
}

func (s *smnState) nudgeArrives() error {
	deliverNudge(s.node, "better_tower")
	return nil
}

func (s *smnState) reattachedWithin(limit time.Duration) error {
	select {
	case d := <-s.waited:
		if d > limit {
			return fmt.Errorf("the node re-attached after %s, want within %s", d, limit)
		}
		return nil
	case <-time.After(limit):
		return fmt.Errorf("the node was still waiting out its backoff %s after the nudge", limit)
	}
}

// The broker's undo after 3 unanswered nudges is the approved station_mobility.feature behaviour
// and is asserted by its runner. What the node owes it is that silence stays silence: a node that
// hears no nudge does not cut its wait, so a move whose node never answers is left to that undo.
func (s *smnState) silentNodeStaysInBackoff() error {
	defer s.cancel()
	quiet := &smnState{node: s.node + "-quiet"}
	if err := quiet.backoff(15)(); err != nil {
		return err
	}
	defer quiet.cancel()
	select {
	case <-quiet.waited:
		return fmt.Errorf("a node that heard no nudge left its backoff")
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}

func (s *smnState) bufferedBeforeAttach() error {
	deliverNudge(s.node, "better_tower")
	time.Sleep(2 * time.Millisecond)
	s.started = time.Now() // the attach that followed the nudge
	return nil
}

func (s *smnState) tenancyStarts() error {
	select {
	case s.n = <-nudgeChan(s.node):
		s.read = true
	default:
		return fmt.Errorf("the buffered nudge is not on the node's channel")
	}
	return nil
}

func (s *smnState) nudgeIgnored() error {
	if s.read && nudgeActionable(s.n, s.started, time.Now()) {
		return fmt.Errorf("a nudge from before the current attach would cause another re-attach")
	}
	return nil
}

func TestStationMobilityNudgeBDD(t *testing.T) {
	s := &smnState{}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = smnState{node: "smn-node-" + strconv.FormatInt(smnSeq.Add(1), 10)}
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				if s.cancel != nil {
					s.cancel()
				}
				return ctx, nil
			})
			// The feature's Background is the broker's state; the node half reads none of it.
			sc.Step(`^a signed-in operator whose node is registered and co-signed by their account$`, func() error { return nil })
			sc.Step(`^two live joined Towers, A and B, that may take that node's Station$`, func() error { return nil })
			sc.Step(`^the node's Station S is attached to Tower A at epoch 4$`, func() error { return nil })

			sc.Step(`^S's node is waiting out a re-attach backoff of 4 minutes$`, s.backoff(4))
			sc.Step(`^a nudge for S arrives$`, s.nudgeArrives)
			sc.Step(`^the node re-attaches within 5 seconds instead of finishing the wait$`, func() error { return s.reattachedWithin(5 * time.Second) })

			sc.Step(`^a move for S is pending and S's node is in a 15-minute backoff$`, s.backoff(15))
			sc.Step(`^the first nudge reaches it$`, s.nudgeArrives)
			// The broker re-sends a nudge every 30 seconds (moveNudgeEvery).
			sc.Step(`^it re-attaches before the second nudge is due, so the approved 3-nudge undo is not reached$`, func() error { return s.reattachedWithin(30 * time.Second) })
			sc.Step(`^a node that has truly gone silent is still undone after 3 nudges, as approved$`, s.silentNodeStaysInBackoff)

			sc.Step(`^a nudge for S was buffered before S's last successful attach$`, s.bufferedBeforeAttach)
			sc.Step(`^S's tenancy starts$`, s.tenancyStarts)
			sc.Step(`^that nudge does not cause another re-attach$`, s.nudgeIgnored)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/station_mobility_idempotency.feature", bddrun.Titled(false,
				"A nudge cuts a backoff wait short",
				"A node that hears its nudge re-attaches inside the undo window",
				"A nudge older than the current attach is dropped",
			))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("station_mobility_idempotency.feature (node nudges) failed")
	}
}
