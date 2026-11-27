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
	"sync/atomic"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

// plainchat_lan_bdd_test.go - the @webui scenario of
// features/agent/plain_chat_fallback_lan.feature. A real /api/agent turn, through the console's
// own loop and completer, with the REAL LAN relay set over this machine's own serve: an httptest
// model server that refuses the tools field the way llama.cpp does without --jinja.

type webPlainChatLAN struct {
	webPlainChat
	self   *httptest.Server
	market atomic.Int32
}

func (s *webPlainChatLAN) close() {
	client.SetLANRelay(nil)
	for _, srv := range []*httptest.Server{s.self, s.up} {
		if srv != nil {
			srv.Close()
		}
	}
	s.self, s.up = nil, nil
}

func (s *webPlainChatLAN) start() {
	s.close()
	s.webPlainChat = webPlainChat{t: s.t}
	s.market.Store(0)
	s.up = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.market.Add(1)
		http.Error(w, "the market was asked in LAN mode", http.StatusTeapot)
	}))
}

func (s *webPlainChatLAN) selfRefusesTools(_ string, status int, msg string) error {
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

func (s *webPlainChatLAN) sends(text string) error {
	srv := New(testCtrl(), Options{Broker: s.up.URL, User: "u"})
	body, _ := json.Marshal(map[string]string{"model": s.model, "message": text})
	res, evs := agentPost(s.t, srv, string(body))
	s.status, s.events = res.StatusCode, evs
	if n := s.market.Load(); n != 0 {
		return fmt.Errorf("the market broker received %d requests in LAN mode", n)
	}
	return nil
}

func TestPlainChatFallbackLANWebuiBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	feature, err := filepath.Abs("../../features/agent/plain_chat_fallback_lan.feature")
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
			s := &webPlainChatLAN{}
			s.t = t
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.start()
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
			sc.Step(`^the web console runs with LAN mode on and its agent is on model `+q+`$`, func(m string) error { s.model = m; return nil })
			sc.Step(`^this machine serves `+q+` and refuses tools with (\d+) `+q+`$`, s.selfRefusesTools)
			sc.Step(`^the user sends `+q+` in the web console$`, s.sends)
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
		t.Fatal("plain_chat_fallback_lan.feature @webui failed (see godog output above)")
	}
}
