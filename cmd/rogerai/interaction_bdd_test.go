package main

// interaction_bdd_test.go makes the @cli scenarios of features/details/interaction.feature
// EXECUTABLE on the details CLI runner's seams (details_bdd_test.go): the REAL cmdShare with
// --interaction and --example, and the REAL dispatch of `roger models <id>` against an
// httptest broker that answers /discover in the shape the broker publishes (interaction at
// the top of the offer, the example under input).

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

const ixDiscover = `{"offers":[{"node_id":"s1","model":"smoke-check","modality":"chat","online":true,
	"capabilities":["vision"],"interaction":"single_turn",
	"input":{"accepts":["image","text"],"example":"Is there smoke in this picture? Answer yes or no."}}]}`

func (s *detailsCLIState) shareInteraction(interaction, example string) error {
	return s.runShare(`--interaction ` + interaction + ` --example "` + example + `"`)
}
func (s *detailsCLIState) shareInteractionOnly(interaction string) error {
	return s.runShare(`--interaction ` + interaction)
}
func (s *detailsCLIState) registrationCarries(interaction, example string) error {
	if s.share.err != nil {
		return fmt.Errorf("share failed: %v", s.share.err)
	}
	b, _ := json.Marshal(s.share.got.Taxonomy)
	var got map[string]any
	_ = json.Unmarshal(b, &got)
	if got["interaction"] != interaction || got["example"] != example {
		return fmt.Errorf("the registration carries %s, want interaction %q and example %q", b, interaction, example)
	}
	return nil
}
func (s *detailsCLIState) failsNamingBoth() error {
	if s.share.err == nil {
		return fmt.Errorf("share did not fail")
	}
	for _, v := range []string{"chat", "single_turn"} {
		if !strings.Contains(s.share.err.Error(), v) {
			return fmt.Errorf("the error does not name %q: %v", v, s.share.err)
		}
	}
	return s.share.nothingOnAir()
}
func (s *detailsCLIState) modelsFor(id string) error {
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ixDiscover))
	}))
	s.share.t.Setenv("COLUMNS", "")
	s.out = captureStdout(s.share.t, func() { s.err = dispatch(config{Broker: s.srv.URL}, []string{"models", id}) })
	return s.err
}
func (s *detailsCLIState) interactionLines() error {
	want := map[string]string{
		"  turns     single turn, one job per message":                  "interaction",
		"  example   Is there smoke in this picture? Answer yes or no.": "example",
	}
	for _, l := range strings.Split(s.out, "\n") {
		delete(want, l)
	}
	for l, what := range want {
		return fmt.Errorf("no %s line %q in:\n%s", what, l, s.out)
	}
	return nil
}

func TestInteractionCLIBDD(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			st := &detailsCLIState{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*st = detailsCLIState{share: shareMetaState{t: t}}
				st.share.install()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				st.share.teardown()
				if st.srv != nil {
					st.srv.Close()
				}
				return ctx, nil
			})
			q := `"([^"]*)"`
			sc.Step(`^a broker with an empty in-memory node registry$`, func() error { return nil })
			sc.Step(`^the operator runs roger share with --interaction (\S+) and --example `+q+`$`, st.shareInteraction)
			sc.Step(`^the operator runs roger share with --interaction (\S+)$`, st.shareInteractionOnly)
			sc.Step(`^the registration carries interaction `+q+` and example `+q+`$`, st.registrationCarries)
			sc.Step(`^the command fails and names the two accepted values$`, st.failsNamingBoth)
			sc.Step(`^a user runs roger models (\S+)$`, st.modelsFor)
			sc.Step(`^the output has a line for the interaction and a line for the example$`, st.interactionLines)
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"../../features/details/interaction.feature"},
			Tags: "@cli", TestingT: t, Strict: true},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("interaction @cli behavior scenarios failed (see godog output above)")
	}
}
