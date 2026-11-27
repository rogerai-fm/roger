package agent

// Contract: features/tower/station_mobility_idempotency.feature ("Every move reason has its own
// words in `roger share`"). The reasons are Core's move triggers as the nudge carries them.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

var moveTrigger = map[string]string{
	"gone": "tower_gone", "failing": "tower_failing", "better": "better_tower", "stage_cap": "stage_cap",
	"left_network": "left_network", "unreachable": "unreachable", "ineligible": "ineligible:suspended",
}

func TestMoveWordsBDD(t *testing.T) {
	var said string
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) { said = ""; return ctx, nil })
			sc.Step(`^a signed-in operator whose node is registered and co-signed by their account$`, func() error { return nil })
			sc.Step(`^two live joined Towers, A and B, that may take that node's Station$`, func() error { return nil })
			sc.Step(`^the node's Station S is attached to Tower A at epoch 4$`, func() error { return nil })
			sc.Step(`^S was moved for reason "([^"]*)"$`, func(reason string) error {
				trig, ok := moveTrigger[reason]
				if !ok {
					return fmt.Errorf("unknown reason %q", reason)
				}
				said = movedNotice("tw-a", "tw-b", trig)
				return nil
			})
			sc.Step("^`roger share` names that reason in plain words, not generic text$", func() error {
				if strings.Contains(said, "Roger Core asked it to attach again") {
					return fmt.Errorf("the share says the generic text: %s", said)
				}
				return nil
			})
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/station_mobility_idempotency.feature",
				bddrun.Titled(false, "Every move reason has its own words in `roger share`"))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("station_mobility_idempotency.feature (move words) failed")
	}
}
