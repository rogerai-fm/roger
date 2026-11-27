package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

// plainchat_bdd_test.go - the @webui scenario of features/agent/plain_chat_fallback.feature.
// A real /api/agent turn, through the console's own loop and completer, against an httptest
// broker that refuses any request carrying tools for a band with no tools-capable station.

type webPlainChat struct {
	t       *testing.T
	up      *httptest.Server
	noTools map[string]bool
	model   string
	events  []agentEvent
	status  int
}

func (s *webPlainChat) reset() {
	if s.up != nil {
		s.up.Close()
	}
	*s = webPlainChat{t: s.t, noTools: map[string]bool{}}
	s.up = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
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

func (s *webPlainChat) sendsRefused(text string) error {
	s.noTools[s.model] = true
	srv := New(testCtrl(), Options{Broker: s.up.URL, User: "u"})
	body, _ := json.Marshal(map[string]string{"model": s.model, "message": text})
	res, evs := agentPost(s.t, srv, string(body))
	s.status, s.events = res.StatusCode, evs
	return nil
}

func (s *webPlainChat) showsSecondReply() error {
	if s.status != http.StatusOK {
		return fmt.Errorf("the console's turn answered %d", s.status)
	}
	for _, e := range s.events {
		if e.Kind == "final" {
			if e.Text != "second reply" {
				return fmt.Errorf("the console's answer is %q, want the second reply", e.Text)
			}
			return nil
		}
	}
	return fmt.Errorf("the console streamed no answer: %+v", s.events)
}

func (s *webPlainChat) streamsNotice() error {
	for _, e := range s.events {
		low := strings.ToLower(e.Text)
		if e.Kind == "notice" && strings.Contains(low, "plain chat") && strings.Contains(e.Text, s.model) {
			return nil
		}
	}
	return fmt.Errorf("the console streamed no plain-chat notice naming %q: %+v", s.model, s.events)
}

func TestPlainChatFallbackWebuiBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	feature, err := filepath.Abs("../../features/agent/plain_chat_fallback.feature")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir() // the console's agent is rooted at the working directory
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &webPlainChat{t: t}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				if s.up != nil {
					s.up.Close()
					s.up = nil
				}
				return ctx, err
			})
			q := `"([^"]*)"`
			sc.Step(`^the agent is on model `+q+`$`, func(string) error { return nil })
			sc.Step(`^no station of `+q+` has the tools capability$`, func(m string) error { s.noTools[m] = true; return nil })
			sc.Step(`^the web console's agent is on model `+q+`$`, func(m string) error { s.model = m; return nil })
			sc.Step(`^the user sends `+q+` and the broker refuses it for tools$`, s.sendsRefused)
			sc.Step(`^the web console shows the second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the web console streams the plain chat notice$`, s.streamsNotice)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{feature},
			Tags:     "@webui",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback.feature @webui failed (see godog output above)")
	}
}
