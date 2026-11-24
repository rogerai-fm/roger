package towercli

// Contract: features/tower/station_attach_signer.feature.
//
// The operator half of features/tower/station_attach_signer.feature: what `roger tower status`
// says a house Tower can hold. The broker half runs in cmd/rogerai-broker.
//
// Core reports a house Tower's trust as the stage "house" and nothing else (operatorStageView
// in the broker), because no trust cap applies to it. Status used to read the absent caps as
// zero and print "no RogerAI Stations" on the one kind of Tower that carries them.

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

type houseStatusState struct {
	t   *testing.T
	dir string
	out string
}

func (s *houseStatusState) houseTower() error {
	core := newCoreStub(s.t)
	s.dir = joinedRegisteredTower(s.t)
	core.reply["/tower/status"] = func(w http.ResponseWriter, _ int) bool {
		_, _ = w.Write([]byte(`{"towers":[{"tower_id":"tw-1","state":"active","link_live":true,
			"carries_traffic":true,"trust":{"stage":"house"}}]}`))
		return true
	}
	return nil
}

func (s *houseStatusState) runsStatus() error {
	out, err := runCLI(s.t, "status", "--dir", s.dir)
	s.out = out
	return err
}

func (s *houseStatusState) trustLine() string {
	for _, line := range strings.Split(s.out, "\n") {
		if strings.Contains(line, "trust stage:") {
			return line
		}
	}
	return ""
}

func (s *houseStatusState) saysNoCap() error {
	line := s.trustLine()
	if !strings.Contains(line, "trust stage:   house") || !strings.Contains(line, "RogerAI Stations without a trust cap") {
		return fmt.Errorf("the trust line is %q:\n%s", line, s.out)
	}
	return nil
}

func (s *houseStatusState) doesNotSayNone() error {
	if strings.Contains(s.out, "no RogerAI Stations") {
		return fmt.Errorf("status says no RogerAI Stations on a house Tower:\n%s", s.out)
	}
	return nil
}

func TestHouseTowerStatusBDD(t *testing.T) {
	s := &houseStatusState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = houseStatusState{t: t}
				return ctx, nil
			})
			// The feature's Background is about the broker's attach; nothing here reads it.
			sc.Step(`^a live joined Tower that may take Stations$`, func() error { return nil })
			sc.Step(`^the attaching Station sends a body that is valid in every other respect, co-signed by its assertion key$`, func() error { return nil })
			sc.Step(`^a joined Tower whose owner is a house account$`, s.houseTower)
			sc.Step(`^the operator runs roger tower status$`, s.runsStatus)
			sc.Step(`^the trust line says house Towers take RogerAI Stations without a trust cap$`, s.saysNoCap)
			sc.Step(`^it does not say "no RogerAI Stations"$`, s.doesNotSayNone)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/station_attach_signer.feature",
				bddrun.Titled(false, "roger tower status on a house Tower does not say it holds no RogerAI Stations"))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("station_attach_signer.feature (status) failed")
	}
}
