package webui

// The @console example of features/cli/market_listing_kinds.feature (K9, approved 2026-11-25):
// the node web console's chat picker never offers a one-job (single_turn) model. The test runs
// the console's own picker predicate (chatIsVoice + chatCannotTalk, cut from assets/console.js)
// under node against the scenario's /discover offers, filtered the way chatMarketOffers does.

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
)

type consoleSpec struct {
	offers []map[string]any
	list   []string
}

// pickerFns cuts the two predicate functions out of console.js, verbatim.
func pickerFns(t *testing.T) string {
	js := chatJS(t)
	var out []string
	for _, name := range []string{"chatIsVoice", "chatCannotTalk", "chatFoldInteraction", "chatOneJobModels"} {
		re := regexp.MustCompile(`(?s)  function ` + name + `\(.*?\n  \}\n`)
		m := re.FindString(js)
		if m == "" {
			t.Fatalf("console.js has no function %s", name)
		}
		out = append(out, m)
	}
	return strings.Join(out, "\n")
}

func (c *consoleSpec) build(t *testing.T) error {
	node, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("node is needed to run the console's picker code: %v", err)
	}
	offers, _ := json.Marshal(c.offers)
	script := pickerFns(t) + `
const offers = ` + string(offers) + `;
const oneJob = chatOneJobModels(offers);
console.log(JSON.stringify(offers.filter(function (o) {
  return o.model && o.online === true && !oneJob[o.model] && !chatCannotTalk(o.modality, "");
}).map(function (o) { return o.model; })));`
	out, err := exec.Command(node, "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("node: %v: %s", err, out)
	}
	c.list = nil
	return json.Unmarshal(out, &c.list)
}

func TestConsoleSpecializedOffChatBDD(t *testing.T) {
	c := &consoleSpec{}
	add := func(o map[string]any) { o["online"] = true; c.offers = append(c.offers, o) }
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*c = consoleSpec{}
				return ctx, nil
			})
			sc.Step(`^the broker lists a chat model "([^"]+)" at \$([\d.]+) in and \$([\d.]+) out per 1M tokens$`, func(n string, _, _ float64) error {
				add(map[string]any{"model": n, "modality": "chat"})
				return nil
			})
			sc.Step(`^a voice "([^"]+)" that (speaks|listens) at \$([\d.]+) per 1M (characters|audio bytes)$`, func(n, verb string, _ float64, _ string) error {
				mod := "tts"
				if verb == "listens" {
					mod = "stt"
				}
				add(map[string]any{"model": n, "modality": mod})
				return nil
			})
			sc.Step(`^a search station "([^"]+)" on engine "([^"]+)", free, scope "([^"]+)"$`, func(id, _, _ string) error {
				add(map[string]any{"model": id, "modality": "search"})
				return nil
			})
			single := func(n string) error {
				add(map[string]any{"model": n, "modality": "chat", "interaction": "single_turn"})
				return nil
			}
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn" and example "(.+)"$`, func(n, _ string) error { return single(n) })
			sc.Step(`^a chat-kind model "([^"]+)" with interaction "single_turn"$`, single)
			sc.Step(`^the node web console picker builds its chat list$`, func() error { return c.build(t) })
			sc.Step(`^"([^"]+)" is not in it$`, func(n string) error {
				if slices.Contains(c.list, n) {
					return fmt.Errorf("%q is offered as a chat partner: %v", n, c.list)
				}
				return nil
			})
			sc.Step(`^"([^"]+)" is in it$`, func(n string) error {
				if !slices.Contains(c.list, n) {
					return fmt.Errorf("%q is missing: %v", n, c.list)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/cli/market_listing_kinds.feature"},
			Tags:     "@console",
			TestingT: t,
			Strict:   true,
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("specialized-off-chat @console scenarios failed (see godog output above)")
	}
}

// A local row (node.RowView) carries its interaction under declared or inferred, never at the
// top level. A locally declared one-job model is not offered as a chat partner either.
func TestConsoleLocalOneJobRowNotOffered(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatalf("node is needed to run the console's picker code: %v", err)
	}
	js := chatJS(t)
	fn := regexp.MustCompile(`(?s)function chatRowInteraction\(.*?\n\s*\}\n`).FindString(js)
	if fn == "" {
		t.Fatal("console.js has no chatRowInteraction")
	}
	if !strings.Contains(js, "!chatCannotTalk(r.modality, chatRowInteraction(r))") {
		t.Fatal("chatLocalRows must read a local row's interaction through chatRowInteraction")
	}
	rows := `[{"model":"chat-8b","upstream":"u","declared":{"interaction":"chat"}},
	          {"model":"fire","upstream":"u","declared":{"interaction":"single_turn"}},
	          {"model":"guess","upstream":"u","inferred":{"interaction":"single_turn"}},
	          {"model":"plain","upstream":"u"}]`
	script := pickerFns(t) + fn + `
console.log(JSON.stringify(` + rows + `.filter(function (r) {
  return r.model && r.upstream && !chatCannotTalk(r.modality, chatRowInteraction(r));
}).map(function (r) { return r.model; })));`
	out, err := exec.Command(node, "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v: %s", err, out)
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if want := []string{"chat-8b", "plain"}; !slices.Equal(got, want) {
		t.Fatalf("local chat rows %v, want %v", got, want)
	}
}

// K9 is per MODEL: the console folds every station's vote the way the broker and the CLI do
// (client.FoldInteraction): online stations vote, the majority wins, a tie is chat, and only
// when no online station says anything do the offline ones decide. One station's word never
// decides a model's kind.
func TestConsoleOneJobFoldsPerModel(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is needed to run the console's picker code")
	}
	js := chatJS(t)
	var fns []string
	for _, name := range []string{"chatIsVoice", "chatCannotTalk", "chatFoldInteraction", "chatOneJobModels"} {
		m := regexp.MustCompile(`(?s)  function ` + name + `\(.*?\n  \}\n`).FindString(js)
		if m == "" {
			t.Fatalf("console.js has no function %s", name)
		}
		fns = append(fns, m)
	}
	on := func(m, inter string) map[string]any {
		return map[string]any{"model": m, "online": true, "interaction": inter}
	}
	off := func(m, inter string) map[string]any {
		return map[string]any{"model": m, "online": false, "interaction": inter}
	}
	for _, c := range []struct {
		name   string
		offers []map[string]any
		oneJob bool
	}{
		{"one station of three says one-job", []map[string]any{on("m", "single_turn"), on("m", "chat"), on("m", "chat")}, false},
		{"majority one-job", []map[string]any{on("m", "single_turn"), on("m", "single_turn"), on("m", "chat")}, true},
		{"tie is chat", []map[string]any{on("m", "single_turn"), on("m", "chat")}, false},
		{"offline one-job vote does not count", []map[string]any{on("m", "chat"), off("m", "single_turn")}, false},
		{"online silent, offline decide", []map[string]any{on("m", ""), off("m", "single_turn")}, true},
		{"one-job, one station", []map[string]any{on("m", "single_turn")}, true},
	} {
		offers, _ := json.Marshal(c.offers)
		script := strings.Join(fns, "\n") + "\nconsole.log(JSON.stringify(!!chatOneJobModels(" + string(offers) + ")[\"m\"]));"
		out, err := exec.Command(node, "-e", script).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: node: %v: %s", c.name, err, out)
		}
		if got := strings.TrimSpace(string(out)) == "true"; got != c.oneJob {
			t.Errorf("%s: one-job=%v, want %v", c.name, got, c.oneJob)
		}
	}
}
