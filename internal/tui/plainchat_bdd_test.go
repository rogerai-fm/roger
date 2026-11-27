package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/harness"
)

// plainchat_bdd_test.go - the @tui scenarios of features/agent/plain_chat_fallback.feature.
// The terminal app's REAL agent runtime (its per-call completer, routing and signing) runs a
// turn against an httptest broker that refuses any request carrying tools for a band with no
// tools-capable station. Events are fed through onAgentEvent, the same handler the live drain
// uses, so the transcript assertions read what the operator would see. The submit path (the
// key press and the goroutine that runs Send) is the existing one and is not re-driven here.

type tuiPlainChat struct {
	t       *testing.T
	srv     *httptest.Server
	mu      sync.Mutex
	bodies  []map[string]any
	noTools map[string]bool
	m       model
	answer  string
	err     error
	display *answersDisplayState
}

func (s *tuiPlainChat) reset() {
	if s.srv != nil {
		s.srv.Close()
	}
	*s = tuiPlainChat{t: s.t, noTools: map[string]bool{}, display: &answersDisplayState{}}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		s.mu.Lock()
		s.bodies = append(s.bodies, body)
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		tl, _ := body["tools"].([]any)
		if len(tl) > 0 && s.noTools[fmt.Sprint(body["model"])] {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":{"message":"no node offers ` + fmt.Sprint(body["model"]) + ` with the tools capability","code":"no_match"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"second reply"},"finish_reason":"stop"}]}`))
	}))
}

func (s *tuiPlainChat) onModel(mdl string) error {
	m := browseSeed(100)
	m.broker = s.srv.URL // before the runtime: its completer captures the model value
	m.agent = m.newAgentRuntime()
	m.pickAgentModel(mdl)
	if m.agent.model != mdl || m.agent.localChat != "" {
		return fmt.Errorf("the agent is on %q (local %q), want the broker band %q", m.agent.model, m.agent.localChat, mdl)
	}
	s.m = m
	return nil
}

func (s *tuiPlainChat) lacksTools(mdl string) error { s.noTools[mdl] = true; return nil }

func (s *tuiPlainChat) turn(text string) {
	var evs []harness.Event
	s.answer, s.err = s.m.agent.loop.Send(context.Background(), text, func(e harness.Event) { evs = append(evs, e) })
	for _, e := range evs {
		nm, _ := s.m.onAgentEvent(agentEventMsg(e))
		s.m = asModel(nm)
	}
}

func (s *tuiPlainChat) sendsRefused(text string) error {
	s.noTools[s.m.agent.model] = true
	s.turn(text)
	return nil
}

func (s *tuiPlainChat) showsSecondReply() error {
	if s.err != nil {
		return fmt.Errorf("the terminal app's turn failed: %v", s.err)
	}
	joined := stripANSI(strings.Join(s.m.agentLines, "\n"))
	if s.answer != "second reply" || !strings.Contains(joined, "second reply") {
		return fmt.Errorf("the transcript does not show the second reply (answer %q):\n%s", s.answer, joined)
	}
	return nil
}

func (s *tuiPlainChat) showsNotice() error {
	// One line must carry both: "solo-m" is already on screen from the switch note.
	for _, l := range s.m.agentLines {
		l = strings.ToLower(stripANSI(l))
		if strings.Contains(l, "plain chat") && strings.Contains(l, "solo-m") {
			return nil
		}
	}
	return fmt.Errorf("no transcript line names solo-m as answering in plain chat:\n%s", stripANSI(strings.Join(s.m.agentLines, "\n")))
}

func (s *tuiPlainChat) fellBack(mdl string) error {
	if err := s.onModel(mdl); err != nil {
		return err
	}
	_ = s.sendsRefused("earlier")
	if s.err != nil {
		return fmt.Errorf("the earlier turn did not fall back: %v", s.err)
	}
	return nil
}

func (s *tuiPlainChat) switchesTo(mdl string) error {
	s.m.pickAgentModel(mdl)
	if s.m.agent.model != mdl {
		return fmt.Errorf("/model did not move the agent to %q", mdl)
	}
	return nil
}

func (s *tuiPlainChat) clears() error {
	nm, _ := s.m.runAgentCommand("/clear")
	s.m = asModel(nm)
	return nil
}

func (s *tuiPlainChat) nextTurnHasTools() error {
	s.mu.Lock()
	first := len(s.bodies)
	s.mu.Unlock()
	s.turn("next")
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.bodies) <= first {
		return fmt.Errorf("the next turn sent nothing")
	}
	tl, _ := s.bodies[first]["tools"].([]any)
	if len(tl) == 0 {
		return fmt.Errorf("the next turn's first request carried no tools: %v", s.bodies[first])
	}
	return nil
}

func TestPlainChatFallbackTUIBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &tuiPlainChat{t: t}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				if s.srv != nil {
					s.srv.Close()
					s.srv = nil
				}
				return ctx, err
			})
			q := `"([^"]*)"`
			// Background: the band is named again by each scenario's own Given.
			sc.Step(`^the agent is on model `+q+`$`, func(string) error { return nil })
			sc.Step(`^no station of `+q+` has the tools capability$`, s.lacksTools)
			sc.Step(`^the terminal app's agent is on model `+q+`$`, s.onModel)
			sc.Step(`^the user sends `+q+` and the broker refuses it for tools$`, s.sendsRefused)
			sc.Step(`^the terminal app shows the second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the terminal app's transcript shows the plain chat notice$`, s.showsNotice)
			sc.Step(`^the terminal app's agent fell back to plain chat on `+q+`$`, s.fellBack)
			sc.Step(`^the operator switches the agent to `+q+` with /model$`, s.switchesTo)
			sc.Step(`^the operator clears the session with /clear$`, s.clears)
			sc.Step(`^the next turn sends a non-empty tools array$`, s.nextTurnHasTools)
			// The entry warning: the existing tui_display.feature steps, unchanged.
			sc.Step(`^the tuned band declares capabilities that do not include "tools"$`, func() error { return s.display.bandLacksTools() })
			sc.Step(`^the user enters the agent$`, func() error { return s.display.entersAgent() })
			sc.Step(`^the transcript says this band cannot drive tools$`, func() error { return s.display.saysCannotDriveTools() })
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/agent/plain_chat_fallback.feature"},
			Tags:     "@tui",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback.feature @tui failed (see godog output above)")
	}
}
