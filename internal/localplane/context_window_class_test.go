package localplane

// Contract: features/tower/context_window_class.feature ("A Tower-local plane answers its own
// consumers the same way").
//
// A station on the plane reports a prompt its model refused as too long with the one fixed
// failure class, and the plane answers its consumer 400 context_exceeded. Anything else a
// station puts in the failure field changes nothing. The whole loop with the real station is
// in internal/agent; these drive the plane's own two delivery shapes.

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const cwClass = "the model did not answer (status 400, context-window)"

// cwJob posts one consumer request and has the station poll it. It returns the job id and the
// channel the consumer's response arrives on.
func cwJob(t *testing.T, srv *Server, client, station ed25519.PrivateKey, prompt string, streams bool) (string, chan *httptest.ResponseRecorder) {
	t.Helper()
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, signedBody(t, client, http.MethodPost, "/v1/chat/completions", []byte(prompt)))
		done <- rec
	}()
	var announce []byte
	if streams {
		announce = []byte(`{"streams":true}`)
	}
	pollRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(pollRec, signedBody(t, station, http.MethodPost, "/local/poll", announce))
	require.Equal(t, http.StatusOK, pollRec.Code, pollRec.Body.String())
	var job struct {
		JobID string `json:"job_id"`
	}
	require.NoError(t, json.Unmarshal(pollRec.Body.Bytes(), &job))
	require.NotEmpty(t, job.JobID)
	return job.JobID, done
}

func cwAwait(t *testing.T, done chan *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	t.Helper()
	select {
	case rec := <-done:
		return rec
	case <-time.After(3 * time.Second):
		t.Fatal("the consumer never received an answer")
		return nil
	}
}

// cwRefusal reads a context_exceeded refusal and holds it to its shape.
func cwRefusal(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var v struct {
		Error struct {
			Code     string         `json:"code"`
			Type     string         `json:"type"`
			Message  string         `json:"message"`
			Metadata map[string]any `json:"metadata"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &v), rec.Body.String())
	require.Equal(t, "context_exceeded", v.Error.Code)
	require.Equal(t, "invalid_request_error", v.Error.Type)
	require.Contains(t, v.Error.Message, "exceeds the context window")
	require.Equal(t, "station", v.Error.Metadata["source"])
	require.Equal(t, "0", rec.Header().Get("X-Roger-Cost"))
}

func TestCompletionWithTheContextWindowClassIsAnsweredContextExceeded(t *testing.T) {
	st := standaloneState(t)
	client := admitClient(t, st)
	station := attachKeyedStation(t, st, "st-1", []string{"m"})
	srv := New(st)

	id, done := cwJob(t, srv, client, station, `{"model":"m","messages":[{"role":"user","content":"hi"}]}`, false)
	body, _ := json.Marshal(map[string]any{
		"job_id": id, "failure": cwClass,
		"answer": json.RawMessage(`{"error":{"message":"SECRET-MODEL-WORDS context length exceeded"}}`),
	})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, signedBody(t, station, http.MethodPost, "/local/complete", body))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	got := cwAwait(t, done)
	cwRefusal(t, got)
	require.NotContains(t, got.Body.String(), "SECRET-MODEL-WORDS", "the refusal is the plane's own words")

	receipts, err := st.Receipts(0)
	require.NoError(t, err)
	require.Empty(t, receipts, "a refused prompt was not served, so it leaves no receipt")
}

// Only the exact class is read. Anything else in the failure field is ignored and the answer
// goes out as it always did.
func TestCompletionWithAnyOtherFailureKeepsTodaysAnswer(t *testing.T) {
	for _, failure := range []string{
		"",
		"the model did not answer (status 400)",
		"the model did not answer (status 429, retry-after 6)",
		"the model did not answer (status 500, context-window)",
		"the model did not answer (status 400, context-window, extra)",
		"the model did not answer (status 400, CONTEXT-WINDOW)",
		"context-window",
		"the model did not answer (status 400) context-window",
	} {
		st := standaloneState(t)
		client := admitClient(t, st)
		station := attachKeyedStation(t, st, "st-1", []string{"m"})
		srv := New(st)

		id, done := cwJob(t, srv, client, station, `{"model":"m","messages":[{"role":"user","content":"hi"}]}`, false)
		answer := `{"error":{"message":"unknown parameter foo"}}`
		body, _ := json.Marshal(map[string]any{"job_id": id, "failure": failure, "answer": json.RawMessage(answer)})
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, signedBody(t, station, http.MethodPost, "/local/complete", body))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		got := cwAwait(t, done)
		require.Equal(t, http.StatusOK, got.Code, "failure %q", failure)
		require.JSONEq(t, answer, got.Body.String(), "failure %q", failure)
	}
}

// A streamed job whose station reports the class before any chunk of the reply is the same
// JSON refusal, not an event stream.
func TestStreamedJobWithTheContextWindowClassIsTheSameJSONRefusal(t *testing.T) {
	st := standaloneState(t)
	client := admitClient(t, st)
	station := attachKeyedStation(t, st, "st-1", []string{"m"})
	srv := New(st)

	id, done := cwJob(t, srv, client, station, `{"model":"m","stream":true,"messages":[{"role":"user","content":"hi"}]}`, true)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, signedBody(t, station, http.MethodPost, "/local/chunk", chunkBody(id, 0, "", true, cwClass)))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	cwRefusal(t, cwAwait(t, done))
	receipts, err := st.Receipts(0)
	require.NoError(t, err)
	require.Empty(t, receipts)
}

// Any other station error on a streamed job is the plane's fixed sentence in an event stream,
// as before: a station's own error text is never echoed.
func TestStreamedJobWithAnyOtherErrorKeepsTodaysStream(t *testing.T) {
	st := standaloneState(t)
	client := admitClient(t, st)
	station := attachKeyedStation(t, st, "st-1", []string{"m"})
	srv := New(st)

	id, done := cwJob(t, srv, client, station, `{"model":"m","stream":true,"messages":[{"role":"user","content":"hi"}]}`, true)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, signedBody(t, station, http.MethodPost, "/local/chunk",
		chunkBody(id, 0, "", true, "the model did not answer (status 400, context-window) and a key sk-123")))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	got := cwAwait(t, done)
	require.Equal(t, http.StatusOK, got.Code)
	require.Contains(t, got.Header().Get("Content-Type"), "text/event-stream")
	require.Contains(t, got.Body.String(), "a local upstream failure ended the reply")
	require.NotContains(t, got.Body.String(), "sk-123")
	require.NotContains(t, got.Body.String(), "context_exceeded")
}
