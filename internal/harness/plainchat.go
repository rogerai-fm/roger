package harness

import (
	"net/http"
	"strings"

	"rogerai.fm/roger/v7/internal/client"
)

// plainchat.go - THE AGENT ANSWERS AS PLAIN CHAT ON A BAND THAT CANNOT DRIVE TOOLS
// (features/agent/plain_chat_fallback.feature, approved 2026-11-24).
//
// Every agent turn sends the tools array. The broker reads a non-empty tools array as a
// requirement (features/routing/capability_gating.feature), so on a band with no tools-capable
// station every turn came back 503 "no node offers M with the tools capability", while the
// terminal app had just told the operator the agent "will answer as plain chat". A model server
// reached directly (llama.cpp without --jinja, vLLM without auto tool choice, an Ollama model
// without tool support) refuses the same field in its own words.
//
// A completer that sees one of those refusals returns a ToolsRefusedError, and the loop asks
// once more without tools and stays plain for the rest of the session on that model. Nothing
// else is classified here: every other error keeps its present path, untouched.

// ToolsRefusedError is a request refused only because it carried tools, by a route the operator
// did not require tools of. The loop retries it without tools; any other caller sees the
// refusal's own text.
type ToolsRefusedError struct {
	Model string // the band that refused, for the notice
	Err   error
}

func (e *ToolsRefusedError) Error() string { return e.Err.Error() }
func (e *ToolsRefusedError) Unwrap() error { return e.Err }

// brokerRefusedTools is the broker's capability no_match when tools is among the capabilities
// it could not find ("with the tools capability", "with the tools and vision capability"). One
// that names a routing filter (" under max_price") is the operator's own rule at work and is
// left as it is. The suffix also rides a refusal for a band that is simply off air, which is
// why the loop keeps plain chat only once the retry answers.
func brokerRefusedTools(status int, msg string) bool {
	low := strings.ToLower(msg)
	return status == http.StatusServiceUnavailable && strings.Contains(low, "no node") &&
		strings.Contains(low, "with the tools ") && !strings.Contains(low, " under ")
}

// localToolsRefusals are the words a model server uses when it will not take the tools field.
var localToolsRefusals = []string{
	"tools param requires --jinja", // llama.cpp
	"--enable-auto-tool-choice",    // vLLM
	"does not support tools",       // Ollama
}

func localRefusedTools(status int, msg string) bool {
	if status != http.StatusBadRequest && status < http.StatusInternalServerError {
		return false
	}
	low := strings.ToLower(msg)
	for _, w := range localToolsRefusals {
		if strings.Contains(low, w) {
			return true
		}
	}
	return false
}

// requiresTools is an explicit tools requirement in the operator's own routing (D3): it is
// never dropped.
func requiresTools(r client.Routing) bool {
	for _, c := range r.Require {
		if strings.EqualFold(c, "tools") {
			return true
		}
	}
	return false
}

// plainChatNotice is the one transcript line a fallback writes.
func plainChatNotice(model string) string {
	if model == "" {
		model = "this band"
	}
	return model + " cannot drive tools, so the agent is answering as plain chat for the rest of this session on it"
}

// ResetPlainChat sends tools again on the next turn. A front end that keeps one loop across a
// model change calls it (the terminal app's bindAgentEndpoint); Reset does it for a new session.
func (l *Loop) ResetPlainChat() { l.plainChat = false }
