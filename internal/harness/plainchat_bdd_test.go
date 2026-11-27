package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

// plainchat_bdd_test.go - features/agent/plain_chat_fallback.feature, the untagged scenarios.
// The REAL Loop runs on the REAL BrokerCompleterRoute (the completer the terminal app builds per
// call) or LocalCompleter, against an httptest server that answers the way the broker, llama.cpp,
// vLLM and Ollama answer. Every assertion is made on what reached the wire.

const plainReply = "second reply"

type plainReq struct {
	body   map[string]any
	status int
}

type plainOverride struct {
	status int
	msg    string
}

type plainChatState struct {
	t   *testing.T
	srv *httptest.Server
	mu  sync.Mutex

	reqs      []plainReq
	overrides []plainOverride
	noTools   map[string]bool // broker: models with no tools-capable station
	model     string
	routing   client.Routing

	// local model server (D5)
	local         bool
	localRefusal  *plainOverride // refusal for any request that carries tools
	localAlwaysEr *plainOverride // every request fails this way

	replyToolCalls bool
	sse            bool           // answer successes as text/event-stream
	toolCallOnce   bool           // the first tools request of the turn gets a tool call back
	noToolsAfter   string         // ...after which this model has no tools station
	lastFail       *plainOverride // the last refusal the server wrote

	loop      *Loop
	pending   string
	turnFirst int // index into reqs where the current turn began
	events    []Event
	turnEvts  []Event
	answer    string
	err       error
	cost      float64
}

func (s *plainChatState) reset() {
	if s.srv != nil {
		s.srv.Close()
	}
	*s = plainChatState{t: s.t, noTools: map[string]bool{}}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
}

func hasTools(body map[string]any) bool {
	tl, ok := body["tools"].([]any)
	return ok && len(tl) > 0
}

func (s *plainChatState) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	s.mu.Lock()
	defer s.mu.Unlock()

	fail := func(status int, msg string) {
		s.reqs = append(s.reqs, plainReq{body: body, status: status})
		s.lastFail = &plainOverride{status, msg}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(failBody(msg))
	}
	switch {
	case !s.local && s.toolCallOnce && hasTools(body):
		s.toolCallOnce = false
		s.noTools[s.noToolsAfter] = true
		s.reqs = append(s.reqs, plainReq{body: body, status: 200})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"c1","type":"function","function":{"name":"list_dir","arguments":"{\"path\":\".\"}"}}]},"finish_reason":"tool_calls"}]}`))
		return
	case s.local && s.localAlwaysEr != nil:
		fail(s.localAlwaysEr.status, s.localAlwaysEr.msg)
		return
	case s.local && s.localRefusal != nil && hasTools(body):
		fail(s.localRefusal.status, s.localRefusal.msg)
		return
	case !s.local && len(s.overrides) > 0:
		o := s.overrides[0]
		s.overrides = s.overrides[1:]
		fail(o.status, o.msg)
		return
	case !s.local && hasTools(body) && s.noTools[fmt.Sprint(body["model"])]:
		fail(http.StatusServiceUnavailable, "no node offers "+fmt.Sprint(body["model"])+" with the tools capability")
		return
	}
	s.reqs = append(s.reqs, plainReq{body: body, status: 200})
	msg := map[string]any{"role": "assistant", "content": plainReply}
	if s.replyToolCalls {
		msg["tool_calls"] = []any{map[string]any{"id": "c1", "type": "function",
			"function": map[string]any{"name": "read_file", "arguments": `{"path":"notes.md"}`}}}
	}
	if s.sse && !s.replyToolCalls {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"" + plainReply + "\"}}]}\n\n" +
			"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if !s.local {
		w.Header().Set("X-RogerAI-Cost", "0.25")
		w.Header().Set("X-RogerAI-Tokens-In", "10")
		w.Header().Set("X-RogerAI-Tokens-Out", "2")
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": msg, "finish_reason": "stop"}}})
}

// failBody is what the fake writes for a refusal: the broker's and llama.cpp's error object, or
// Ollama's bare string.
func failBody(msg string) []byte {
	var b []byte
	if strings.Contains(msg, "ollama") {
		b, _ = json.Marshal(map[string]any{"error": msg})
	} else {
		b, _ = json.Marshal(map[string]any{"error": map[string]any{"message": msg, "code": "no_match"}})
	}
	return b
}

// todaysError is the error a completer gives for this refusal without the fallback: the same
// parse every relayed turn already goes through.
func todaysError(o plainOverride) string {
	_, err := parseCompletion(failBody(o.msg), o.status)
	return err.Error()
}

func (s *plainChatState) completer() Completer {
	return func(ctx context.Context, messages []Message, tools []map[string]any) (Message, error) {
		if s.local {
			return LocalCompleter(s.srv.URL+"/v1/chat/completions", "", s.model)(ctx, messages, tools)
		}
		// Built per call with the live model, exactly as the terminal app does.
		return BrokerCompleterRoute(BrokerRoute{
			Broker: s.srv.URL, User: "u", Model: s.model, Routing: s.routing,
			OnCost:          func(c float64, _, _ int, _ float64) { s.cost += c },
			FallbackTimeout: brokerTimeout,
		})(ctx, messages, tools)
	}
}

func (s *plainChatState) ensureLoop() {
	if s.loop == nil {
		s.loop = NewLoop(s.t.TempDir(), "You are a test persona.", s.completer(), func(string, map[string]any) bool { return false })
	}
}

func (s *plainChatState) send(text string) {
	s.ensureLoop()
	s.mu.Lock()
	s.turnFirst = len(s.reqs)
	s.mu.Unlock()
	s.turnEvts = nil
	s.cost = 0
	s.answer, s.err = s.loop.Send(context.Background(), text, func(e Event) {
		s.events = append(s.events, e)
		s.turnEvts = append(s.turnEvts, e)
	})
}

// flush runs a turn that a When step queued, so the steps that configure the broker's answer
// can come after it in the scenario's words.
func (s *plainChatState) flush() {
	if s.pending != "" {
		t := s.pending
		s.pending = ""
		s.send(t)
	}
}

func (s *plainChatState) turnReqs() []plainReq {
	s.flush()
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]plainReq(nil), s.reqs[s.turnFirst:]...)
}

func (s *plainChatState) nthReq(n int) (plainReq, error) {
	rs := s.turnReqs()
	if len(rs) < n {
		return plainReq{}, fmt.Errorf("the server received %d request(s) this turn, want at least %d", len(rs), n)
	}
	return rs[n-1], nil
}

// --- steps ---------------------------------------------------------------------------------

func (s *plainChatState) onModel(m string) error           { s.model = m; return nil }
func (s *plainChatState) stationLacksTools(m string) error { s.noTools[m] = true; return nil }
func (s *plainChatState) stationHasTools(m string) error   { delete(s.noTools, m); return nil }

func (s *plainChatState) userSends(text string) error {
	s.flush()
	s.pending = text
	return nil
}

func (s *plainChatState) userSendsRefused(text string) error {
	s.flush()
	s.noTools[s.model] = true
	s.send(text)
	return nil
}

func (s *plainChatState) firstHasTools(m string) error {
	r, err := s.nthReq(1)
	if err != nil {
		return err
	}
	if r.body["model"] != m || !hasTools(r.body) {
		return fmt.Errorf("first request: model %v, tools present %v; want %q with a non-empty tools array", r.body["model"], hasTools(r.body), m)
	}
	return nil
}

// In scenario 1 this is an observation of what came back; after a When that has not run yet it
// configures the broker's answer.
func (s *plainChatState) brokerAnswersNoMatch(msg string) error {
	if s.pending != "" {
		s.overrides = append(s.overrides, plainOverride{http.StatusServiceUnavailable, msg})
		return nil
	}
	r, err := s.nthReq(1)
	if err != nil {
		return err
	}
	if r.status != http.StatusServiceUnavailable {
		return fmt.Errorf("the first request got %d, want 503", r.status)
	}
	return nil
}

func (s *plainChatState) brokerAnswers(status int, msg string) error {
	s.overrides = append(s.overrides, plainOverride{status, msg})
	return nil
}

func noToolKeys(r plainReq) error {
	if _, ok := r.body["tools"]; ok {
		return fmt.Errorf("the request still carries a tools key: %v", r.body["tools"])
	}
	if _, ok := r.body["tool_choice"]; ok {
		return fmt.Errorf("the request still carries a tool_choice key")
	}
	return nil
}

func (s *plainChatState) secondPlain(m string) error {
	r, err := s.nthReq(2)
	if err != nil {
		return err
	}
	if r.body["model"] != m {
		return fmt.Errorf("second request is for %v, want %q", r.body["model"], m)
	}
	return noToolKeys(r)
}

func (s *plainChatState) secondPlainStream() error {
	r, err := s.nthReq(2)
	if err != nil {
		return err
	}
	if err := noToolKeys(r); err != nil {
		return err
	}
	if r.body["stream"] != true {
		return fmt.Errorf("the second request is not streamed: stream=%v", r.body["stream"])
	}
	return nil
}

func (s *plainChatState) showsSecondReply() error {
	s.flush()
	if s.err != nil {
		return fmt.Errorf("the turn failed: %v", s.err)
	}
	if s.answer != plainReply {
		return fmt.Errorf("answer %q, want %q", s.answer, plainReply)
	}
	return nil
}

func (s *plainChatState) personaAndTwoTurns() error {
	s.ensureLoop()
	return s.loop.RestoreConversation([]Message{
		{Role: "user", Content: "one"}, {Role: "assistant", Content: "first answer"},
		{Role: "user", Content: "two"}, {Role: "assistant", Content: "second answer"},
	})
}

func (s *plainChatState) sameMessages() error {
	a, err := s.nthReq(1)
	if err != nil {
		return err
	}
	b, err := s.nthReq(2)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(a.body["messages"], b.body["messages"]) {
		return fmt.Errorf("the messages changed between the requests:\n first  %v\n second %v", a.body["messages"], b.body["messages"])
	}
	if ms, _ := a.body["messages"].([]any); len(ms) != 6 {
		return fmt.Errorf("the first request carried %d messages, want persona + 4 earlier + 1 new", len(ms))
	}
	return nil
}

func (s *plainChatState) plainNotices(model string) int {
	n := 0
	for _, e := range s.events {
		low := strings.ToLower(e.Text)
		if e.Kind == EventNotice && strings.Contains(low, "plain chat") && strings.Contains(low, "tools") &&
			(model == "" || strings.Contains(e.Text, model)) {
			n++
		}
	}
	return n
}

func (s *plainChatState) noticeFor(model string) error {
	s.flush()
	if s.plainNotices(model) != 1 {
		return fmt.Errorf("want one plain-chat notice naming %q, got %d in %+v", model, s.plainNotices(model), s.events)
	}
	return nil
}

func (s *plainChatState) noticeAny() error { return s.noticeFor("") }

func (s *plainChatState) noSecondNotice() error {
	s.flush()
	if n := s.plainNotices(""); n != 1 {
		return fmt.Errorf("the plain-chat notice was shown %d times, want once", n)
	}
	return nil
}

func (s *plainChatState) earlierTurnRefused() error {
	s.noTools[s.model] = true
	s.send("earlier")
	if s.err != nil {
		return fmt.Errorf("the earlier turn did not fall back: %v", s.err)
	}
	return nil
}

func (s *plainChatState) exactlyN(n int) func() error {
	return func() error {
		if rs := s.turnReqs(); len(rs) != n {
			return fmt.Errorf("the server received %d request(s) this turn, want %d", len(rs), n)
		}
		return nil
	}
}

func (s *plainChatState) onlyReqPlain() error {
	r, err := s.nthReq(1)
	if err != nil {
		return err
	}
	return noToolKeys(r)
}

// What every front end does when the agent moves to another model: the web console builds a
// new loop, and the terminal app keeps its loop and calls this hook (bindAgentEndpoint).
func (s *plainChatState) movesTo(m string) error {
	s.model = m
	r, ok := any(s.loop).(interface{ ResetPlainChat() })
	if !ok {
		return fmt.Errorf("the loop has no ResetPlainChat hook for a model move")
	}
	r.ResetPlainChat()
	return nil
}

func (s *plainChatState) newSession(m string) error {
	s.model = m
	s.loop.Reset() // the terminal app's /clear; the web console builds a new loop
	return nil
}

func (s *plainChatState) refuseFirstForTools() error {
	s.overrides = append(s.overrides, plainOverride{http.StatusServiceUnavailable, "no node offers " + s.model + " with the tools capability"})
	return nil
}

func (s *plainChatState) refuseSecond(msg string) error {
	if len(s.overrides) == 0 {
		_ = s.refuseFirstForTools()
	}
	s.overrides = append(s.overrides, plainOverride{http.StatusServiceUnavailable, msg})
	return nil
}

func (s *plainChatState) endsWithSecondError() error {
	rs := s.turnReqs()
	if s.err == nil {
		return fmt.Errorf("the turn succeeded, want the second refusal's error")
	}
	if len(rs) != 2 || s.lastFail == nil {
		return fmt.Errorf("want the turn to end on its second request, got %d", len(rs))
	}
	if want := todaysError(*s.lastFail); s.err.Error() != want {
		return fmt.Errorf("the turn ended with %q, want the second refusal %q", s.err, want)
	}
	return nil
}

func (s *plainChatState) noNotice() error {
	s.flush()
	if n := s.plainNotices(""); n != 0 {
		return fmt.Errorf("a plain-chat notice was shown %d time(s) for a retry that failed", n)
	}
	return nil
}

func (s *plainChatState) firstReplyCallsTool() error {
	s.toolCallOnce = true
	return nil
}

func (s *plainChatState) noToolsAfterReply(m string) error {
	delete(s.noTools, m) // the station that drove the first step had tools
	s.noToolsAfter = m
	return nil
}

func (s *plainChatState) costOnlySecond() error {
	s.flush()
	if s.err != nil {
		return fmt.Errorf("the turn failed: %v", s.err)
	}
	if s.cost != 0.25 {
		return fmt.Errorf("the turn's cost is %g, want only the answered request's 0.25", s.cost)
	}
	return nil
}

func (s *plainChatState) routingRequires(c string) error {
	s.routing.Require = append(s.routing.Require, c)
	return nil
}

func (s *plainChatState) endsWithToolsRefusal() error {
	s.flush()
	if s.err == nil || !strings.Contains(s.err.Error(), "with the tools capability") {
		return fmt.Errorf("the turn ended with %v, want the tools refusal", s.err)
	}
	return nil
}

func (s *plainChatState) endsWithTodaysError() error {
	s.flush()
	if s.err == nil {
		return fmt.Errorf("the turn succeeded; want it to end with the error it gives today")
	}
	if s.lastFail == nil {
		return fmt.Errorf("the server wrote no refusal")
	}
	if want := todaysError(*s.lastFail); s.err.Error() != want {
		return fmt.Errorf("the turn ended with %q, want today's %q", s.err, want)
	}
	if s.plainNotices("") != 0 {
		return fmt.Errorf("a plain-chat notice was shown for an error that is not the tools refusal")
	}
	return nil
}

func (s *plainChatState) oneWithTools() error {
	rs := s.turnReqs()
	if len(rs) != 1 || !hasTools(rs[0].body) {
		return fmt.Errorf("want exactly one request with tools, got %d (tools on the first: %v)", len(rs), len(rs) > 0 && hasTools(rs[0].body))
	}
	if s.err != nil {
		return fmt.Errorf("the turn failed: %v", s.err)
	}
	return nil
}

func (s *plainChatState) secondReplyHasToolCalls() error {
	s.replyToolCalls = true
	s.send("and again")
	return nil
}

func (s *plainChatState) noToolRun() error {
	for _, e := range s.turnEvts {
		if e.Kind == EventToolCall || e.Kind == EventToolResult {
			return fmt.Errorf("a tool step ran on a plain-chat turn: %+v", e)
		}
	}
	return nil
}

func (s *plainChatState) showsReplyText() error { return s.showsSecondReply() }

func (s *plainChatState) noToolCallsKept() error {
	for _, m := range s.loop.messages {
		if len(m.ToolCalls) > 0 {
			return fmt.Errorf("the conversation kept tool_calls from a plain-chat reply: %+v", m.ToolCalls)
		}
	}
	return nil
}

func (s *plainChatState) localRefuses(status int, msg string) error {
	s.local = true
	s.localRefusal = &plainOverride{status, msg}
	return nil
}

func (s *plainChatState) localAlways(status int, msg string) error {
	s.local = true
	s.localAlwaysEr = &plainOverride{status, msg}
	return nil
}

func (s *plainChatState) localSecondPlain() error {
	r, err := s.nthReq(2)
	if err != nil {
		return err
	}
	if !hasTools(s.turnReqs()[0].body) {
		return fmt.Errorf("the first local request carried no tools")
	}
	return noToolKeys(r)
}

func (s *plainChatState) bareTick() error {
	s.flush()
	s.noTools[s.model] = true
	s.send(`{"time_local":"14:05","wind_kmh":30,"wind_from":"NE","wind_toward_structures":true,"rh_pct":34,"temp_c":27,"camera_margin_prev":0.82,"camera_margin_now":0.9,"pm25":95,"co_ppm":11,"voc_idx":40,"thermal_delta_c":3,"sensor_fault":false,"escalation_tried_all":false,"acknowledged":false,"minutes_unacked":0}`)
	return nil
}

func (s *plainChatState) tickSecondRequest() error {
	r, err := s.nthReq(2)
	if err != nil {
		return err
	}
	if err := noToolKeys(r); err != nil {
		return err
	}
	ms, _ := r.body["messages"].([]any)
	if len(ms) != 2 {
		return fmt.Errorf("want the persona and the tick, got %d messages", len(ms))
	}
	first, _ := ms[0].(map[string]any)
	last, _ := ms[1].(map[string]any)
	if first["role"] != "system" || last["role"] != "user" || !strings.HasPrefix(fmt.Sprint(last["content"]), `{"time_local"`) {
		return fmt.Errorf("the second request is not persona + tick: %v", ms)
	}
	return nil
}

func TestPlainChatFallbackBDD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			s := &plainChatState{t: t}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				s.reset()
				return ctx, nil
			})
			sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				if s.srv != nil {
					s.srv.Close()
					s.srv = nil
				}
				if s.loop != nil {
					s.loop.Close()
				}
				return ctx, err
			})
			q := `"([^"]*)"`
			sc.Step(`^the agent is on model `+q+`$`, s.onModel)
			sc.Step(`^no station of `+q+` has the tools capability$`, s.stationLacksTools)
			sc.Step(`^a station of `+q+` has the tools capability$`, s.stationHasTools)
			sc.Step(`^the user sends `+q+` to the agent$`, s.userSends)
			sc.Step(`^the user sends `+q+` to the agent and the broker refuses it for tools$`, s.userSendsRefused)
			sc.Step(`^the broker receives a request for `+q+` with a non-empty tools array$`, s.firstHasTools)
			sc.Step(`^the broker answers 503 no_match `+q+`$`, s.brokerAnswersNoMatch)
			sc.Step(`^the broker answers (\d+) with `+q+`$`, s.brokerAnswers)
			sc.Step(`^the broker then receives a second request for `+q+` with no "tools" key and no "tool_choice" key$`, s.secondPlain)
			sc.Step(`^the broker then receives a second request with no "tools" key and "stream" true$`, s.secondPlainStream)
			sc.Step(`^the agent shows the second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the agent's session holds a system persona and two earlier turns$`, s.personaAndTwoTurns)
			sc.Step(`^the second request's messages equal the first request's messages$`, s.sameMessages)
			sc.Step(`^the transcript shows a notice that `+q+` cannot drive tools and is answering as plain chat$`, s.noticeFor)
			sc.Step(`^the transcript shows a notice that the model cannot drive tools and is answering as plain chat$`, s.noticeAny)
			sc.Step(`^no second notice is shown$`, s.noSecondNotice)
			sc.Step(`^the broker refused an earlier turn of this session for tools$`, s.earlierTurnRefused)
			sc.Step(`^the broker refused an earlier turn of an earlier session for tools$`, s.earlierTurnRefused)
			sc.Step(`^the broker receives exactly one request for that turn$`, s.exactlyN(1))
			sc.Step(`^the broker received exactly one request for that turn$`, s.exactlyN(1))
			sc.Step(`^the broker received exactly two requests for that turn$`, s.exactlyN(2))
			sc.Step(`^the local server received exactly one request for that turn$`, s.exactlyN(1))
			sc.Step(`^it has no "tools" key and no "tool_choice" key$`, s.onlyReqPlain)
			sc.Step(`^the agent moves to model `+q+`$`, s.movesTo)
			sc.Step(`^a new agent session starts on `+q+`$`, s.newSession)
			sc.Step(`^the broker refuses the first request for tools$`, s.refuseFirstForTools)
			sc.Step(`^the broker refuses the second request with 503 no_match `+q+`$`, s.refuseSecond)
			sc.Step(`^the turn ends with the error the second refusal gives$`, s.endsWithSecondError)
			sc.Step(`^no plain chat notice is shown$`, s.noNotice)
			sc.Step(`^the first reply of the turn calls a tool$`, s.firstReplyCallsTool)
			sc.Step(`^no station of `+q+` has the tools capability after that reply$`, s.noToolsAfterReply)
			sc.Step(`^the agent shows the streamed second reply as its answer$`, s.showsSecondReply)
			sc.Step(`^the turn's cost counts only the second request$`, s.costOnlySecond)
			sc.Step(`^the agent streams its turns$`, func() error { s.sse = true; return nil })
			sc.Step(`^the user's routing requires `+q+`$`, s.routingRequires)
			sc.Step(`^the turn ends with that refusal, as today$`, s.endsWithToolsRefusal)
			sc.Step(`^the turn ends with the error it gives today$`, s.endsWithTodaysError)
			sc.Step(`^the broker receives exactly one request for that turn, with a non-empty tools array$`, s.oneWithTools)
			sc.Step(`^the second reply carries a tool_calls array$`, s.secondReplyHasToolCalls)
			sc.Step(`^no tool is run$`, s.noToolRun)
			sc.Step(`^the agent shows the reply's text as its answer$`, s.showsReplyText)
			sc.Step(`^the conversation keeps no tool_calls from that reply$`, s.noToolCallsKept)
			sc.Step(`^the agent is on a local model whose server refuses tools with (\d+) "(.*)"$`, s.localRefuses)
			sc.Step(`^the agent is on a local model whose server answers (\d+) `+q+`$`, s.localAlways)
			sc.Step(`^the local server then receives a second request with no "tools" key and no "tool_choice" key$`, s.localSecondPlain)
			sc.Step(`^the user sends a bare tick to the agent$`, s.bareTick)
			sc.Step(`^the broker's second request carries the persona, the tick and no tools$`, s.tickSecondRequest)
			sc.Step(`^the agent shows the station's reply as its answer$`, s.showsSecondReply)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/agent/plain_chat_fallback.feature"},
			Tags:     "~@tui && ~@webui",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("plain_chat_fallback.feature failed (see godog output above)")
	}
}
