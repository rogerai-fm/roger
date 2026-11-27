package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

// plainchat_lan_bdd_test.go - the @tui scenario of features/agent/plain_chat_fallback_lan.feature.
// The terminal app's REAL agent runtime runs a turn with the REAL LAN relay set, as `roger --lan`
// sets it, over this machine's own serve: an httptest model server that refuses the tools field
// the way llama.cpp does without --jinja. The market broker (the embedded fake) must hear nothing.

type tuiPlainChatLAN struct {
	tuiPlainChat
	self *httptest.Server
}

func (s *tuiPlainChatLAN) close() {
	client.SetLANRelay(nil)
	if s.self != nil {
		s.self.Close()
		s.self = nil
	}
	if s.srv != nil {
		s.srv.Close()
		s.srv = nil
	}
}

func (s *tuiPlainChatLAN) selfRefusesTools(_ string, status int, msg string) error {
	s.self = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		if tl, _ := body["tools"].([]any); len(tl) > 0 {
			w.WriteHeader(status)
			out, _ := json.Marshal(map[string]any{"error": map[string]any{"message": msg}})
			_, _ = w.Write(out)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"second reply"},"finish_reason":"stop"}]}`))
	}))
	client.SetLANRelay(&client.ProxyOptions{LANMode: true,
		LANSelf: func(string) (string, string) { return s.self.URL + "/v1/chat/completions", "" }})
	return nil
}

func (s *tuiPlainChatLAN) sends(text string) error {
	s.turn(text)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.bodies) != 0 {
		return fmt.Errorf("the market broker received %d requests in LAN mode", len(s.bodies))
	}
	return nil
}

func TestPlainChatFallbackLANTUIBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &tuiPlainChatLAN{tuiPlainChat: tuiPlainChat{t: t}}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.close()
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				s.close()
				return ctx, err
			})
			q := `"([^"]*)"`
			// Background: the relay is set by the step that declares this machine's serve.
			sc.Step(`^LAN mode is on$`, func() error { return nil })
			sc.Step(`^the agent is on model `+q+`$`, func(string) error { return nil })
			sc.Step(`^the terminal app runs with LAN mode on and its agent is on model `+q+`$`, s.onModel)
			sc.Step(`^this machine serves `+q+` and refuses tools with (\d+) `+q+`$`, s.selfRefusesTools)
			sc.Step(`^the user sends `+q+` in the terminal app$`, s.sends)
			sc.Step(`^the terminal app shows the second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the terminal app's transcript shows the plain chat notice$`, s.showsNotice)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/agent/plain_chat_fallback_lan.feature"},
			Tags:     "@tui",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback_lan.feature @tui failed (see godog output above)")
	}
}
