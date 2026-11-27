package agent

// Contract: features/tower/station_mobility_idempotency.feature.
//
// The node half of "A node is never stranded by a 401 or an endless empty poll": relayWatch
// decides, on the node's clock, what a "no node is registered" 401, a poll, and a nudge it could
// not act on mean; hubFailureStreak is the real standing-window streak that turns counted hub
// failures into a re-attach. The hub half (asking Core on an unknown poll) runs in towerhub and
// towercli.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/towerhub"
)

type smsState struct {
	clock   time.Time
	w       *relayWatch
	said    []string
	counted bool
	tripped bool
	again   bool
}

var noNodeErr = &towerhub.HTTPError{Status: 401, Body: `{"error":"no node is registered for this Station on this tower"}`}

func (s *smsState) attachedAgo(d time.Duration) {
	s.clock = time.Unix(1_700_000_000, 0)
	s.w = newRelayWatch("tw-b", s.clock.Add(-d), func(err error) { s.said = append(s.said, err.Error()) })
	s.w.now = func() time.Time { return s.clock }
}

func (s *smsState) refusedPastGrace() error {
	s.attachedAgo(2 * noNodeGrace)
	return nil
}

func (s *smsState) standingWindowPasses() error {
	streak := &hubFailureStreak{}
	for t := time.Duration(0); t <= hubStandingWindow+5*time.Second; t += 2 * time.Second {
		now := s.clock.Add(t)
		s.clock = now
		if s.w.onNoNode(noNodeErr) {
			s.counted = true
			if streak.observe(noNodeErr, now) {
				s.tripped = true
			}
		}
	}
	return nil
}

func (s *smsState) reattachesLikeTransport() error {
	if !s.counted || !s.tripped {
		return fmt.Errorf("a hub that keeps saying no node is registered did not trip the standing window (counted=%v tripped=%v)", s.counted, s.tripped)
	}
	return nil
}

func (s *smsState) attachedRecently() error {
	s.attachedAgo(10 * time.Second)
	return nil
}

func (s *smsState) hubSaysNoNode() error {
	s.counted = s.w.onNoNode(noNodeErr)
	return nil
}

func (s *smsState) quietRetry() error {
	if s.counted || len(s.said) != 0 {
		return fmt.Errorf("inside the grace the node counted the 401 (%v) or said %v", s.counted, s.said)
	}
	return nil
}

func (s *smsState) refusedPastGraceSaid() error {
	s.attachedAgo(2 * noNodeGrace)
	for i := 0; i < 5; i++ {
		s.w.onNoNode(noNodeErr)
	}
	return nil
}

func (s *smsState) oneTrueLine() error {
	if len(s.said) != 1 || !strings.Contains(s.said[0], "does not know this node yet") {
		return fmt.Errorf("the share said %q, want one line that the relay does not know this node yet", s.said)
	}
	return nil
}

func (s *smsState) noFalseBlame() error {
	for _, l := range s.said {
		for _, blame := range []string{"older than signed hub polls", "revoked attachment", "system clock"} {
			if strings.Contains(l, blame) {
				return fmt.Errorf("the warning blames %q: %s", blame, l)
			}
		}
	}
	return nil
}

func (s *smsState) recoveryLine() error {
	s.w.onPoll(false)
	s.w.onPoll(false)
	if len(s.said) != 2 || !strings.Contains(s.said[1], "serving through") {
		return fmt.Errorf("after polls succeeded the share said %q, want one recovery line", s.said)
	}
	return nil
}

func (s *smsState) nudgeThenEmpty() error {
	s.attachedAgo(time.Hour)
	s.w.onNudgeHeard()
	return nil
}

func (s *smsState) fiveMinutesPass() error {
	for t := time.Duration(0); t <= emptyAfterNudge; t += 10 * time.Second {
		s.clock = s.clock.Add(10 * time.Second)
		if s.w.onPoll(false) {
			s.again = true
		}
	}
	return nil
}

func (s *smsState) reattachesThroughCore() error {
	if !s.again {
		return errors.New("five minutes of empty polls after a nudge did not lead to a re-attach")
	}
	return nil
}

func (s *smsState) emptyNoNudge() error {
	s.attachedAgo(time.Hour)
	return s.fiveMinutesPass()
}

func (s *smsState) noReattachForThat() error {
	if s.again {
		return errors.New("a quiet Station with no nudge re-attached")
	}
	return nil
}

func TestStationMobilityStrandedBDD(t *testing.T) {
	s := &smsState{}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = smsState{}
				return ctx, nil
			})
			sc.Step(`^a signed-in operator whose node is registered and co-signed by their account$`, func() error { return nil })
			sc.Step(`^two live joined Towers, A and B, that may take that node's Station$`, func() error { return nil })
			sc.Step(`^the node's Station S is attached to Tower A at epoch 4$`, func() error { return nil })

			sc.Step(`^S's node polls Tower A's hub and gets 401 "no node is registered" past the grace below$`, s.refusedPastGrace)
			sc.Step(`^the standing window passes$`, s.standingWindowPasses)
			sc.Step(`^the node re-attaches through Core with backoff, as on a transport failure$`, s.reattachesLikeTransport)

			sc.Step(`^S's node attached to Tower B less than 60 seconds ago$`, s.attachedRecently)
			sc.Step(`^Tower B's hub answers 401 "no node is registered"$`, s.hubSaysNoNode)
			sc.Step(`^the node retries quietly and prints nothing$`, s.quietRetry)

			sc.Step(`^Tower B's hub has refused S's node past the grace$`, s.refusedPastGraceSaid)
			sc.Step("^`roger share` prints one line saying the relay does not know this node yet$", s.oneTrueLine)
			sc.Step(`^it does not blame an old Tower, a revoked attachment or the clock unless that is the cause$`, s.noFalseBlame)
			sc.Step(`^when polls succeed again it prints one line saying the node is serving through the relay$`, s.recoveryLine)

			sc.Step(`^S's node has heard a nudge and then gets only empty polls for 5 minutes$`, s.nudgeThenEmpty)
			sc.Step(`^the 5 minutes pass$`, s.fiveMinutesPass)
			sc.Step(`^the node re-attaches through Core$`, s.reattachesThroughCore)
			sc.Step(`^S's node gets only empty polls and has heard no nudge$`, s.emptyNoNudge)
			sc.Step(`^the node does not re-attach for that reason$`, s.noReattachForThat)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/station_mobility_idempotency.feature", bddrun.Titled(false,
				`A hub that says "no node is registered" counts as a hub failure`,
				"A new attachment gets a short grace for the hub to learn it",
				"The share's warning says what is true, and says when it recovers",
				"Endless empty polls after a nudge lead to a re-attach",
				"Endless empty polls with no nudge are just a quiet Station",
			))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("station_mobility_idempotency.feature (never stranded, node) failed")
	}
}
