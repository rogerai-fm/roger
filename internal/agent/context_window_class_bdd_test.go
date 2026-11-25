package agent

// Executable spec: the self-hosted plane scenario of
// features/tower/context_window_class.feature ("A Tower-local plane answers its own consumers
// the same way").
//
// No mocks of our own code: a tower.State on disk, the real localplane handler over HTTP, the
// real ServeLocalTower polling it, and an HTTP model stand-in that refuses the prompt. The
// Station and broker scenarios of the same feature run in internal/station and in the
// broker's own package.

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/localplane"
	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/tower"
)

const cwFeature = "../../features/tower/context_window_class.feature"

// cwPlane is a standalone Tower with one admitted client and one attached station whose model
// answers every request with the given status and text.
type cwPlane struct {
	url      string
	clientPv ed25519.PrivateKey
	hits     chan struct{}
}

func newCWPlane(t *testing.T, status int, text string) *cwPlane {
	t.Helper()
	dir := t.TempDir()
	_, err := tower.Init(dir, tower.ModeStandalone)
	require.NoError(t, err)
	st, err := tower.Open(dir)
	require.NoError(t, err)
	p := &cwPlane{hits: make(chan struct{}, 16)}
	_, p.clientPv, _ = ed25519.GenerateKey(nil)
	inv, code, err := st.CreateInvitation(idFromPriv(p.clientPv), time.Hour, 5)
	require.NoError(t, err)
	_, err = st.ConsumeInvitation(inv.ID, code, idFromPriv(p.clientPv))
	require.NoError(t, err)
	_, nodePv, _ := ed25519.GenerateKey(nil)
	_, err = st.AttachStation("shed", idFromPriv(nodePv), []string{"m"})
	require.NoError(t, err)
	plane := httptest.NewServer(localplane.New(st).Handler())
	t.Cleanup(plane.Close)
	p.url = plane.URL
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		p.hits <- struct{}{}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, text)
	}))
	t.Cleanup(model.Close)
	serveStation(t, Config{Broker: plane.URL, Upstream: model.URL, Model: "m"}, nodePv)
	return p
}

// ask posts one signed chat request and returns the status, the content type and the body.
func (p *cwPlane) ask(stream bool) (int, string, []byte, error) {
	body := []byte(`{"model":"m","messages":[{"role":"user","content":"hi"}]}`)
	if stream {
		body = []byte(`{"model":"m","messages":[{"role":"user","content":"hi"}],"stream":true}`)
	}
	req, _ := http.NewRequest(http.MethodPost, p.url+"/v1/chat/completions", bytes.NewReader(body))
	nonce := protocol.NewNonce()
	pub, ts, sig := protocol.SignRequestNonce(p.clientPv, http.MethodPost, "/v1/chat/completions", body, nonce)
	req.Header.Set(protocol.HeaderPubkey, pub)
	req.Header.Set(protocol.HeaderTS, strconv.FormatInt(ts, 10))
	req.Header.Set(protocol.HeaderSig, sig)
	req.Header.Set(protocol.HeaderNonce, nonce)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	select {
	case <-p.hits:
	default:
		return 0, "", nil, fmt.Errorf("the Station's model was never asked (status %d: %.200s)", resp.StatusCode, out)
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), out, nil
}

func cwErrorCode(body []byte) string {
	var v struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &v)
	return v.Error.Code
}

// cwMetadata is the error's metadata object, or an error when the body carries none.
func cwMetadata(body []byte) (map[string]any, error) {
	var v struct {
		Error struct {
			Metadata map[string]any `json:"metadata"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &v); err != nil || v.Error.Metadata == nil {
		return nil, fmt.Errorf("the response carries no error.metadata: %.300s", body)
	}
	return v.Error.Metadata, nil
}

func TestTowerContextClassStandaloneBDD(t *testing.T) {
	var (
		plane  *cwPlane
		status int
		body   []byte
	)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
				plane, status, body = nil, 0, nil
				return c, nil
			})
			sc.Step(`^a standalone Tower whose Station's model replies (\d+) with "([^"]*)"$`, func(code, text string) error {
				n, err := strconv.Atoi(code)
				plane = newCWPlane(t, n, text)
				return err
			})
			sc.Step(`^a local consumer relays through the Tower's own plane$`, func() error {
				var err error
				status, _, body, err = plane.ask(false)
				return err
			})
			sc.Step(`^the response is (\d+) with error code "([^"]+)"$`, func(code, want string) error {
				if n, _ := strconv.Atoi(code); status != n || cwErrorCode(body) != want {
					return fmt.Errorf("the response is %d with error code %q, want %s %q (%.300s)", status, cwErrorCode(body), code, want, body)
				}
				return nil
			})
			sc.Step(`^error\.metadata\.source is "([^"]+)"$`, func(want string) error {
				md, err := cwMetadata(body)
				if err != nil {
					return err
				}
				if md["source"] != want {
					return fmt.Errorf("error.metadata.source is %v, want %q", md["source"], want)
				}
				return nil
			})
			sc.Step(`^error\.metadata has no "([^"]+)" key$`, func(key string) error {
				md, err := cwMetadata(body)
				if err != nil {
					return err
				}
				if v, ok := md[key]; ok {
					return fmt.Errorf("error.metadata.%s is present (%v)", key, v)
				}
				return nil
			})
		},
		Options: &godog.Options{
			Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios(cwFeature, bddrun.Titled(false,
				"A Tower-local plane answers its own consumers the same way",
				"The self-hosted plane sends the code without numbers")),
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("context_window_class.feature (self-hosted plane) failed")
	}
}

// A streamed request gets the same JSON refusal, before any frame.
func TestLocalStreamedOverflowIsTheSameJSONRefusal(t *testing.T) {
	p := newCWPlane(t, http.StatusBadRequest, "the request exceeds the available context size")
	status, ctype, body, err := p.ask(true)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, status, string(body))
	require.Equal(t, "application/json", ctype)
	require.Equal(t, "context_exceeded", cwErrorCode(body))
	require.NotContains(t, string(body), "available context size", "the refusal is the plane's own words")
}

// Only a 400 in the overflow vocabulary is the class. Every other reply keeps today's answer
// on both delivery shapes.
func TestLocalOtherModelFailuresKeepTodaysAnswer(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		text   string
	}{
		{"a 400 that is not an overflow", http.StatusBadRequest, "unknown parameter foo"},
		{"a 413 in the overflow vocabulary", http.StatusRequestEntityTooLarge, "context length exceeded"},
		{"a 500 in the overflow vocabulary", http.StatusInternalServerError, "context length exceeded"},
		{"a 429 in the overflow vocabulary", http.StatusTooManyRequests, "too many tokens per minute"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newCWPlane(t, tc.status, tc.text)
			status, _, body, err := p.ask(false)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status, string(body))
			require.Contains(t, string(body), "local_station_error")
			require.NotEqual(t, "context_exceeded", cwErrorCode(body))

			status, ctype, body, err := p.ask(true)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status, string(body))
			require.Contains(t, ctype, "text/event-stream")
			require.Contains(t, string(body), "a local upstream failure ended the reply")
			require.NotContains(t, string(body), "context_exceeded")
		})
	}
}
