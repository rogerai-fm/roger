package tui

// The @tui examples of features/cli/market_listing_kinds.feature (K9, approved 2026-11-25): a
// one-job (single_turn) model is never offered as a chat partner on the dial or in the agent
// picker. The offers are what /discover carries; groupBands builds the bands the real way.

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/protocol"
)

type specState struct {
	offers []offer
	list   []string
}

func (s *specState) add(o offer) {
	o.Online = true
	o.NodeID = fmt.Sprintf("n%d", len(s.offers)+1)
	s.offers = append(s.offers, o)
}

func (s *specState) model() model {
	m := seedFor(120, modeBrowse, false)
	m.offers = s.offers
	m.bands = groupBands(s.offers, m.limits)
	return m
}

func (s *specState) builds(surface string) error {
	m := s.model()
	s.list = nil
	switch surface {
	case "TUI dial chat section":
		for _, b := range m.visibleBands() {
			s.list = append(s.list, b.model)
		}
	case "TUI agent picker":
		for _, r := range m.agentPickerCandidates() {
			s.list = append(s.list, r.model)
		}
	default:
		return fmt.Errorf("surface %q is not a TUI surface", surface)
	}
	return nil
}

func (s *specState) notIn(name string) error {
	if slices.Contains(s.list, name) {
		return fmt.Errorf("%q is offered as a chat partner: %v", name, s.list)
	}
	return nil
}

func (s *specState) in(name string) error {
	if !slices.Contains(s.list, name) {
		return fmt.Errorf("%q is missing from the chat list: %v", name, s.list)
	}
	return nil
}

func TestSpecializedOffChatBDD(t *testing.T) {
	st := &specState{}
	single := func(name string) {
		o := offer{Model: name, Modality: protocol.ModalityChat}
		o.Interaction = protocol.InteractionSingleTurn
		st.add(o)
	}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*st = specState{}
				return ctx, nil
			})
			sc.Step(`^the broker lists a chat model "([^"]+)" at \$([\d.]+) in and \$([\d.]+) out per 1M tokens$`, func(n string, in, out float64) error {
				st.add(offer{Model: n, Modality: protocol.ModalityChat, PriceIn: in, PriceOut: out})
				return nil
			})
			sc.Step(`^a voice "([^"]+)" that (speaks|listens) at \$([\d.]+) per 1M (characters|audio bytes)$`, func(n, verb string, p float64, _ string) error {
				mod := protocol.ModalityTTS
				if verb == "listens" {
					mod = protocol.ModalitySTT
				}
				st.add(offer{Model: n, Modality: mod, PriceIn: p})
				return nil
			})
			sc.Step(`^a search station "([^"]+)" on engine "([^"]+)", free, scope "([^"]+)"$`, func(id, _, _ string) error {
				st.add(offer{Model: id, Modality: protocol.ModalitySearch})
				return nil
			})
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn" and example "(.+)"$`, func(n, _ string) error { single(n); return nil })
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn"$`, func(n string) error { single(n); return nil })
			sc.Step(`^the (.+) builds its chat list$`, st.builds)
			sc.Step(`^"([^"]+)" is not in it$`, st.notIn)
			sc.Step(`^"([^"]+)" is in it$`, st.in)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/cli/market_listing_kinds.feature"},
			Tags:     "@tui",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("specialized-off-chat @tui scenarios failed (see godog output above)")
	}
}

// The footnote names the one-job models the dial sets aside, so they are not simply gone.
func TestDialFootnoteNamesSpecialized(t *testing.T) {
	st := &specState{}
	o := offer{Model: "wave-pico-293m", Modality: protocol.ModalityChat}
	o.Interaction = protocol.InteractionSingleTurn
	st.add(o)
	st.add(offer{Model: "llama3.2", Modality: protocol.ModalityChat})
	m := st.model()
	if got := m.specializedFootnote(); got == "" {
		t.Fatal("no footnote for an on-air one-job model")
	}
	st2 := &specState{}
	st2.add(offer{Model: "llama3.2", Modality: protocol.ModalityChat})
	if got := st2.model().specializedFootnote(); got != "" {
		t.Fatalf("a footnote with no one-job model on air: %q", got)
	}
}

// A band's interaction is the broker's fold (protocol.FoldInteraction): only online stations
// vote, the majority wins and a tie is chat. One offline station cannot hide a chat band.
func TestBandSpecializedFollowsTheFold(t *testing.T) {
	mk := func(inter string, online bool) offer {
		o := offer{Model: "m", Modality: protocol.ModalityChat, Online: online}
		o.Interaction = inter
		return o
	}
	for _, c := range []struct {
		name   string
		offers []offer
		want   bool
	}{
		{"offline one-job vote does not count", []offer{mk("chat", true), mk("single_turn", false)}, false},
		{"majority one-job", []offer{mk("single_turn", true), mk("single_turn", true), mk("chat", true)}, true},
		{"tie is chat", []offer{mk("single_turn", true), mk("chat", true)}, false},
		{"all one-job", []offer{mk("single_turn", true)}, true},
		{"off the air, still one-job", []offer{mk("single_turn", false)}, true},
		{"off the air, the vote of every station", []offer{mk("single_turn", false), mk("chat", false)}, false},
	} {
		bands := groupBands(c.offers, nil)
		if len(bands) != 1 || bands[0].specialized != c.want {
			t.Errorf("%s: specialized=%v, want %v", c.name, len(bands) == 1 && bands[0].specialized, c.want)
		}
	}
}

// A one-job model tuned earlier in the session is never offered as the agent's brain.
func TestAgentSkipsATunedOneJobModel(t *testing.T) {
	st := &specState{}
	o := offer{Model: "fire", Modality: protocol.ModalityChat}
	o.Interaction = protocol.InteractionSingleTurn
	st.add(o)
	st.add(offer{Model: "llama3.2", Modality: protocol.ModalityChat})
	m := st.model()
	m.recentBands = map[string]bool{"fire": true}
	if slices.Contains(m.agentModelCandidates(), "fire") {
		t.Fatalf("a tuned one-job model is an agent candidate: %v", m.agentModelCandidates())
	}
}
