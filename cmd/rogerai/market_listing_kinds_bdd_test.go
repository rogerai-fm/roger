package main

// Step definitions for features/cli/market_listing_kinds.feature (approved 2026-11-25): the CLI
// lists models, one-job models, voices and search stations apart, each priced in its own unit.
// The broker is a real httptest server serving /discover from the scenario's offers. The @tui,
// @console and @site examples run in the packages that own those surfaces.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

type mlkState struct {
	mu     sync.Mutex
	offers []map[string]any
	// stations is GET /search/stations: the search rows' source (callsign id, engine, scope).
	stations []map[string]any
	srv      *httptest.Server
	out      string
	err      error
	chats    []string // models of the chat completions the broker received
}

func (s *mlkState) reset() {
	if s.srv != nil {
		s.srv.Close()
	}
	*s = mlkState{}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.URL.Path {
		case "/discover":
			_ = json.NewEncoder(w).Encode(map[string]any{"offers": s.offers})
		case "/search/stations":
			_ = json.NewEncoder(w).Encode(map[string]any{"stations": s.stations})
		case "/v1/chat/completions":
			var body struct {
				Model string `json:"model"`
			}
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &body)
			s.chats = append(s.chats, body.Model)
			_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

func (s *mlkState) add(o map[string]any) {
	o["online"] = true
	if _, ok := o["node_id"]; !ok {
		o["node_id"] = "n" + fmt.Sprint(len(s.offers)+1)
	}
	s.offers = append(s.offers, o)
}

func (s *mlkState) remove(keep func(map[string]any) bool) {
	var out []map[string]any
	for _, o := range s.offers {
		if keep(o) {
			out = append(out, o)
		}
	}
	s.offers = out
}

func (s *mlkState) chatModel(name string, in, out float64) error {
	s.add(map[string]any{"model": name, "modality": "chat", "price_in": in, "price_out": out, "ctx": 8192})
	return nil
}

func (s *mlkState) voice(name, verb string, perM float64, _ string) error {
	mod := "tts"
	if verb == "listens" {
		mod = "stt"
	}
	s.add(map[string]any{"model": name, "modality": mod, "price_in": perM})
	return nil
}

func (s *mlkState) searchStation(id, engine, scope string) error {
	model := id[strings.Index(id, "/")+1:]
	s.add(map[string]any{"model": model, "modality": "search", "engine": engine, "unit": "query",
		"price_in": 0.0, "operator_status": "verified"})
	s.stations = append(s.stations, map[string]any{"id": id, "model": model, "station": id[:strings.Index(id, "/")],
		"engine": engine, "scope": scope, "price_per_1k_queries": 0, "free": true, "operator_status": "verified"})
	return nil
}

func (s *mlkState) singleTurn(name, example string) error {
	s.add(map[string]any{"model": name, "modality": "chat", "interaction": "single_turn",
		"input": map[string]any{"accepts": []string{"text"}, "example": strings.ReplaceAll(example, `\"`, `"`)}})
	return nil
}

func (s *mlkState) run(cmd string) error {
	args := strings.Fields(strings.TrimPrefix(cmd, "roger "))
	cfg := config{Broker: s.srv.URL}
	s.out, s.err = captureEdgeStdout(func() error { return dispatch(cfg, args) })
	return nil
}

// sections splits the output into section name -> body lines. A section header is a line that
// is exactly an upper-case word, optionally followed by two spaces and a note.
var mlkHeader = regexp.MustCompile(`^(MODELS|SPECIALIZED|VOICES|SEARCH|OTHER)\b`)

func (s *mlkState) sections() (order []string, body map[string][]string) {
	body = map[string][]string{}
	cur := ""
	for _, l := range strings.Split(s.out, "\n") {
		if m := mlkHeader.FindStringSubmatch(l); m != nil {
			cur = m[1]
			order = append(order, cur)
			body[cur] = append(body[cur], l)
			continue
		}
		if cur != "" {
			body[cur] = append(body[cur], l)
		}
	}
	return order, body
}

func (s *mlkState) rowIn(section, name string) (string, bool) {
	_, body := s.sections()
	for _, l := range body[section] {
		for _, f := range strings.Fields(l) {
			if f == name {
				return l, true
			}
		}
	}
	return "", false
}

func (s *mlkState) onlyUnder(name, section string) error {
	order, _ := s.sections()
	found := ""
	for _, sec := range order {
		if _, ok := s.rowIn(sec, name); ok {
			if found != "" {
				return fmt.Errorf("%q is under both %s and %s:\n%s", name, found, sec, s.out)
			}
			found = sec
		}
	}
	if found != section {
		return fmt.Errorf("%q is under %q, want %q:\n%s", name, found, section, s.out)
	}
	return nil
}

func (s *mlkState) onlyUnderLabeled(name, section, label string) error {
	if err := s.onlyUnder(name, section); err != nil {
		return err
	}
	if row, _ := s.rowIn(section, name); !strings.Contains(row, label) {
		return fmt.Errorf("%q row does not say %q: %q", name, label, row)
	}
	return nil
}

func (s *mlkState) searchRow(name, engine, price string) error {
	if err := s.onlyUnder(name, "SEARCH"); err != nil {
		return err
	}
	row, _ := s.rowIn("SEARCH", name)
	for _, w := range []string{engine, price} {
		if !strings.Contains(row, w) {
			return fmt.Errorf("search row lacks %q: %q", w, row)
		}
	}
	return nil
}

func (s *mlkState) sectionsInOrder(a, b, c string) error {
	order, _ := s.sections()
	want := []string{a, b, c}
	j := 0
	for _, sec := range order {
		if j < len(want) && sec == want[j] {
			j++
		}
	}
	if j != len(want) {
		return fmt.Errorf("sections %v, want %v in that order:\n%s", order, want, s.out)
	}
	return nil
}

func (s *mlkState) noSection(name string) error {
	order, _ := s.sections()
	for _, sec := range order {
		if sec == name {
			return fmt.Errorf("unexpected %s section:\n%s", name, s.out)
		}
	}
	return nil
}

func (s *mlkState) onlySection(name string) error {
	if s.err != nil {
		return fmt.Errorf("command failed: %v", s.err)
	}
	order, _ := s.sections()
	if len(order) != 1 || order[0] != name {
		return fmt.Errorf("sections %v, want only %s:\n%s", order, name, s.out)
	}
	return nil
}

func (s *mlkState) notSays(text string) error {
	if strings.Contains(s.out, text) {
		return fmt.Errorf("output says %q:\n%s", text, s.out)
	}
	return nil
}

func (s *mlkState) underWithKind(name, section, kind string) error {
	return s.onlyUnderLabeled(name, section, kind)
}

func (s *mlkState) notUnder(name, section string) error {
	if _, ok := s.rowIn(section, name); ok {
		return fmt.Errorf("%q is under %s:\n%s", name, section, s.out)
	}
	return nil
}

func (s *mlkState) under(name, section string) error {
	if _, ok := s.rowIn(section, name); !ok {
		return fmt.Errorf("%q is not under %s:\n%s", name, section, s.out)
	}
	return nil
}

func (s *mlkState) headerShows(section, unit string) error {
	_, body := s.sections()
	if len(body[section]) < 2 || !strings.Contains(body[section][0]+"\n"+body[section][1], unit) {
		return fmt.Errorf("%s header does not show %q:\n%s", section, unit, s.out)
	}
	return nil
}

func (s *mlkState) rowShowsPrice(name, price string) error {
	order, _ := s.sections()
	for _, sec := range order {
		if row, ok := s.rowIn(sec, name); ok {
			for _, f := range strings.Fields(row) {
				if strings.TrimPrefix(strings.TrimSuffix(f, "*"), "$") == price || f == price {
					return nil
				}
			}
			return fmt.Errorf("%q row lacks price %q: %q", name, price, row)
		}
	}
	return fmt.Errorf("no row for %q:\n%s", name, s.out)
}

func (s *mlkState) modelsHoldNoVoiceOrSearch() error {
	for _, o := range s.offers {
		if m := o["modality"]; m == "tts" || m == "stt" || m == "search" {
			if err := s.notUnder(fmt.Sprint(o["model"]), "MODELS"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *mlkState) fails(msg string) error {
	if s.err == nil || !strings.Contains(s.err.Error(), msg) {
		return fmt.Errorf("want failure %q, got %v\n%s", msg, s.err, s.out)
	}
	return nil
}

func (s *mlkState) jsonKeys(a, b, c, d string) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s.out), &m); err != nil {
		return fmt.Errorf("not JSON: %v\n%s", err, s.out)
	}
	for _, k := range []string{a, b, c, d} {
		if _, ok := m[k]; !ok {
			return fmt.Errorf("JSON lacks key %q: %s", k, s.out)
		}
	}
	return nil
}

func (s *mlkState) jsonHas(name, group, key, val string) error {
	var m map[string][]map[string]any
	if err := json.Unmarshal([]byte(s.out), &m); err != nil {
		return err
	}
	for _, o := range m[group] {
		if o["id"] == name || o["model"] == name {
			if fmt.Sprint(o[key]) != val {
				return fmt.Errorf("%s %s = %v, want %s", name, key, o[key], val)
			}
			return nil
		}
	}
	return fmt.Errorf("%q not in %q: %s", name, group, s.out)
}

func (s *mlkState) specializedWithExample(name string) error {
	if err := s.onlyUnder(name, "SPECIALIZED"); err != nil {
		return err
	}
	if !strings.Contains(s.out, `{"tick": 1}`) {
		return fmt.Errorf("the example is not shown:\n%s", s.out)
	}
	return nil
}

func (s *mlkState) noInteraction(name string) error {
	s.add(map[string]any{"model": name, "modality": "chat"})
	return nil
}

func (s *mlkState) withCategory(name, cat string) error {
	parts := strings.SplitN(cat, "/", 2)
	s.add(map[string]any{"model": name, "modality": "chat", "category": parts[0], "subcategory": parts[len(parts)-1]})
	return nil
}

func (s *mlkState) sendChat(name string) error {
	client.SetLANRelay(nil) // an internet chat: no LAN relay left over from another test
	_, err := client.ChatDetailed(s.srv.URL, "", name, "hello", false, 0)
	return err
}

func (s *mlkState) routedAsBefore() error {
	if len(s.chats) != 1 || s.chats[0] != "wave-nano-fire-controller" {
		return fmt.Errorf("broker received %v", s.chats)
	}
	return nil
}

func (s *mlkState) hintFirst() error {
	first := strings.SplitN(strings.TrimSpace(s.out), "\n", 2)[0]
	if !strings.Contains(strings.ToLower(first), "search the web") {
		return fmt.Errorf("first line is not a web-search hint: %q", first)
	}
	return nil
}

func (s *mlkState) marketAfter() error {
	order, _ := s.sections()
	if len(order) == 0 {
		return fmt.Errorf("no market listed after the hint:\n%s", s.out)
	}
	return nil
}

func TestMarketListingKindsBDD(t *testing.T) {
	t.Setenv("COLUMNS", "")
	// A clean home: no saved LAN pins, Towers or local models from another test steer the chat.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st := &mlkState{}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) { st.reset(); return ctx, nil })
			sc.Step(`^the broker lists a chat model "([^"]+)" at \$([\d.]+) in and \$([\d.]+) out per 1M tokens$`, st.chatModel)
			sc.Step(`^a voice "([^"]+)" that (speaks|listens) at \$([\d.]+) per 1M (characters|audio bytes)$`, st.voice)
			sc.Step(`^a search station "([^"]+)" on engine "([^"]+)", free, scope "([^"]+)"$`, st.searchStation)
			sc.Step(`^I run "([^"]+)"$`, st.run)
			sc.Step(`^the output has sections "([^"]+)", "([^"]+)" and "([^"]+)" in that order$`, st.sectionsInOrder)
			sc.Step(`^"([^"]+)" is only under "([^"]+)"$`, st.onlyUnder)
			sc.Step(`^"([^"]+)" is only under "([^"]+)", labeled "([^"]+)"$`, st.onlyUnderLabeled)
			sc.Step(`^"([^"]+)" is only under "SEARCH", with engine "([^"]+)" and "([^"]+)"$`, st.searchRow)
			sc.Step(`^no voices are on the broker$`, func() error {
				st.remove(func(o map[string]any) bool { return o["modality"] != "tts" && o["modality"] != "stt" })
				return nil
			})
			sc.Step(`^the output has no "([^"]+)" section$`, st.noSection)
			sc.Step(`^only the search station is on the broker$`, func() error {
				st.remove(func(o map[string]any) bool { return o["modality"] == "search" })
				return nil
			})
			sc.Step(`^the output has only the "([^"]+)" section$`, st.onlySection)
			sc.Step(`^it does not say "([^"]+)"$`, st.notSays)
			sc.Step(`^an offer "([^"]+)" with modality "([^"]+)"$`, func(name, mod string) error {
				st.add(map[string]any{"model": name, "modality": mod})
				return nil
			})
			sc.Step(`^"([^"]+)" is under "([^"]+)" with kind "([^"]+)"$`, st.underWithKind)
			sc.Step(`^"([^"]+)" is not under "([^"]+)"$`, st.notUnder)
			sc.Step(`^an offer "([^"]+)" with no modality field$`, func(name string) error {
				st.add(map[string]any{"model": name})
				return nil
			})
			sc.Step(`^"([^"]+)" is under "([^"]+)"$`, st.under)
			sc.Step(`^the "([^"]+)" header shows "([^"]+)"$`, st.headerShows)
			sc.Step(`^"([^"]+)" shows the price "([^"]+)"$`, st.rowShowsPrice)
			sc.Step(`^no row under "MODELS" belongs to a voice or a search station$`, st.modelsHoldNoVoiceOrSearch)
			sc.Step(`^the command fails with "([^"]+)"$`, st.fails)
			sc.Step(`^the JSON has keys "([^"]+)", "([^"]+)", "([^"]+)" and "([^"]+)"$`, st.jsonKeys)
			sc.Step(`^"([^"]+)" is in "([^"]+)" with "([^"]+)" "([^"]+)"$`, st.jsonHas)
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn" and example "(.+)"$`, st.singleTurn)
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn"$`, func(n string) error { return st.singleTurn(n, "") })
			sc.Step(`^"([^"]+)" is only under "SPECIALIZED", showing its example$`, st.specializedWithExample)
			sc.Step(`^"([^"]+)" is still under "([^"]+)"$`, st.under)
			sc.Step(`^a signed caller sends a chat completion to "([^"]+)"$`, st.sendChat)
			sc.Step(`^it is routed to the station exactly as before$`, st.routedAsBefore)
			sc.Step(`^a chat-kind model "([^"]+)" with no interaction$`, st.noInteraction)
			sc.Step(`^a chat-kind model "([^"]+)" with category "([^"]+)" and no interaction$`, st.withCategory)
			sc.Step(`^the output starts with a one-line hint about searching the web$`, st.hintFirst)
			sc.Step(`^the market is listed after it$`, st.marketAfter)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/cli/market_listing_kinds.feature"},
			Tags:     "~@tui && ~@console && ~@site",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("market listing kinds scenarios failed (see godog output above)")
	}
	if st.srv != nil {
		st.srv.Close()
	}
}
