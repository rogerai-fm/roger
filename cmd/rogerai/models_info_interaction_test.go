package main

// roger models <id> reads the interaction and the example as the broker does
// (features/details/interaction.feature, review of 2026-11-23): when the id has any chat
// offer, only chat offers vote.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func modelsInfoOut(t *testing.T, discover string, args ...string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(discover))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("COLUMNS", "")
	var err error
	out := captureStdout(t, func() { err = dispatch(config{Broker: srv.URL}, append([]string{"models"}, args...)) })
	require.NoError(t, err)
	return out
}

// A voice station on a chat model's id does not vote, so its label and example never show.
func TestModelsInfoVoiceOfferDoesNotVote(t *testing.T) {
	out := modelsInfoOut(t, `{"offers":[
		{"node_id":"s1","model":"m","modality":"chat","online":true},
		{"node_id":"s2","model":"m","modality":"stt","online":true,"interaction":"single_turn",
		 "input":{"accepts":["audio"],"example":"Voice words."}}]}`, "m")
	require.NotContains(t, out, "single turn")
	require.NotContains(t, out, "Voice words.")
}

// A chat station's declaration still shows when a voice station shares the id.
func TestModelsInfoChatOfferStillVotes(t *testing.T) {
	out := modelsInfoOut(t, `{"offers":[
		{"node_id":"s1","model":"m","modality":"chat","online":true,"interaction":"single_turn",
		 "input":{"accepts":["text"],"example":"Chat words."}},
		{"node_id":"s2","model":"m","modality":"stt","online":true,"interaction":"chat"}]}`, "m")
	require.Contains(t, out, "single turn")
	require.Contains(t, out, "Chat words.")
}

// A voice-only id keeps its own vote.
func TestModelsInfoVoiceOnlyIDKeepsItsVote(t *testing.T) {
	out := modelsInfoOut(t, `{"offers":[{"node_id":"s2","model":"v","modality":"stt","online":true,
		"interaction":"single_turn","input":{"accepts":["audio"],"example":"Say a word."}}]}`, "v")
	require.Contains(t, out, "single turn")
	require.Contains(t, out, "Say a word.")
}

// A voice station's description does not reach a chat id, in text or in --json, as on /market
// and /v1/models.
func TestModelsInfoVoiceOfferDescriptionStaysOff(t *testing.T) {
	const discover = `{"offers":[
		{"node_id":"s1","model":"m","modality":"chat","online":true},
		{"node_id":"s2","model":"m","modality":"stt","online":true,"summary":"Voice summary."}]}`
	require.NotContains(t, modelsInfoOut(t, discover, "m"), "Voice summary.")
	require.NotContains(t, modelsInfoOut(t, discover, "m", "--json"), "Voice summary.")
}
