package station

// Unit coverage for the upstream key (features/tower/tower_upstream_key.feature): the key is sent
// to the upstream and redacted at the source, and the operator's hook hears upstream errors. The
// end-to-end behaviour through a joined Tower is the godog runner in internal/agent.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/towercore/dispatch"
)

const testUpKey = "sk-unit-0123456789KEY"

func TestRedactKeyCoversEveryForm(t *testing.T) {
	var lower, upper strings.Builder
	for _, c := range []byte(testUpKey) {
		fmt.Fprintf(&lower, `\u%04x`, c)
		fmt.Fprintf(&upper, `\u%04X`, c)
	}
	cases := map[string]string{
		"verbatim":        "key=" + testUpKey + ".",
		"escaped lower":   `{"k":"` + lower.String() + `"}`,
		"escaped upper":   `{"k":"` + upper.String() + `"}`,
		"twice, adjacent": testUpKey + testUpKey,
	}
	for name, in := range cases {
		out := string(RedactKey([]byte(in), testUpKey))
		require.NotContains(t, out, testUpKey, name)
		require.NotContains(t, strings.ToLower(out), strings.ToLower(lower.String()), name)
		require.Contains(t, out, "[redacted]", name)
	}
	require.Equal(t, "untouched", string(RedactKey([]byte("untouched"), "")), "no key, no change")
	require.Empty(t, RedactKey(nil, testUpKey))
}

func TestHTTPUpstreamSendsTheKeyAndRedactsWhatComesBack(t *testing.T) {
	var auth string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"echo":"` + testUpKey + `"}`))
	}))
	defer up.Close()
	body, err := HTTPUpstream{URL: up.URL, Key: testUpKey}.Serve(context.Background(), []byte(`{}`))
	require.NoError(t, err)
	require.Equal(t, "Bearer "+testUpKey, auth)
	require.Equal(t, `{"echo":"[redacted]"}`, string(body))

	_, err = HTTPUpstream{URL: up.URL}.Serve(context.Background(), []byte(`{}`))
	require.NoError(t, err)
	require.Empty(t, auth, "no key configured, no Authorization header")
}

// The error detail is redacted BEFORE it is shortened: with the key across byte 200, a cut first
// would leave a prefix of it that a whole-key replace no longer matches.
func TestHTTPUpstreamRedactsTheErrorBeforeShorteningIt(t *testing.T) {
	pad := strings.Repeat("x", 200-len(testUpKey)/2)
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, pad+testUpKey+strings.Repeat("y", 50), http.StatusUnauthorized)
	}))
	defer up.Close()
	_, err := HTTPUpstream{URL: up.URL, Key: testUpKey}.Serve(context.Background(), []byte(`{}`))
	var se *UpstreamStatusError
	require.True(t, errors.As(err, &se))
	require.Equal(t, http.StatusUnauthorized, se.Status)
	for i := 8; i <= len(testUpKey); i++ {
		require.NotContains(t, se.Detail, testUpKey[:i])
	}
}

// The operator's hook hears an upstream status error on both serve paths, and only those.
func TestOnUpstreamErrorHearsStatusErrorsOnBothPaths(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	var heard []int
	hook := func(se *UpstreamStatusError) { heard = append(heard, se.Status) }

	e, s, g, _, up := sealedRig(t, now, nil)
	e.OnUpstreamError = hook
	up.err = &UpstreamStatusError{Status: 500, Detail: "boom"}
	_, _, failure := e.ServeSealed(context.Background(), g.Signed, sealFor(t, s, g.AttemptID, []byte(`{"p":1}`)))
	require.Equal(t, "the model did not answer (status 500)", failure)
	require.Equal(t, []int{500}, heard)

	e2, s2, g2, _, up2 := sealedRig(t, now, nil)
	e2.OnUpstreamError = hook
	up2.err = errors.New("connection refused")
	_, _, failure = e2.ServeSealed(context.Background(), g2.Signed, sealFor(t, s2, g2.AttemptID, []byte(`{"p":1}`)))
	require.Equal(t, "the model did not answer", failure)
	require.Equal(t, []int{500}, heard, "a transport error is not an upstream reply")

	// No hook set: nothing to call, nothing panics.
	e3, s3, g3, _, up3 := sealedRig(t, now, nil)
	up3.err = &UpstreamStatusError{Status: 502}
	_, _, failure = e3.ServeSealed(context.Background(), g3.Signed, sealFor(t, s3, g3.AttemptID, []byte(`{"p":1}`)))
	require.NotEmpty(t, failure)
}

func TestSignAttachProofNamesThisStationsKeys(t *testing.T) {
	s := initStation(t)
	proof := s.SignAttachProof("roger-public", "ab", 1_700_000_000, []byte(`{"x":1}`))
	require.NotEmpty(t, proof)
	require.NotEqual(t, proof, s.SignAttachProof("roger-public", "ab", 1_700_000_001, []byte(`{"x":1}`)),
		"a different timestamp is a different proof")
}

// A key straddling the read limit must not survive as a prefix: an oversize body is refused
// rather than cut and served.
func TestABodyPastTheReadLimitIsRefusedNotCut(t *testing.T) {
	const limit = 8 << 20
	body := strings.Repeat("x", limit-len(testUpKey)/2) + testUpKey + "yyyy"
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer up.Close()

	now := time.Unix(1_700_000_000, 0)
	e, s, g, _, _ := sealedRig(t, now, func(tg *dispatch.EdgeTarget) { tg.MaxOut = 16 << 20 })
	e.Upstream = HTTPUpstream{URL: up.URL, Key: testUpKey}
	env, rec, failure := e.ServeSealed(context.Background(), g.Signed, sealFor(t, s, g.AttemptID, []byte(`{"p":1}`)))
	require.Equal(t, "the model returned more than this grant allows", failure)
	require.Empty(t, env)
	require.Empty(t, rec)

	got, err := HTTPUpstream{URL: up.URL, Key: testUpKey}.Serve(context.Background(), []byte(`{}`))
	require.Error(t, err)
	for i := 8; i <= len(testUpKey); i++ {
		require.NotContains(t, string(got), testUpKey[:i])
	}
}

// Every way a JSON encoder (or a hand-rolled one) can write the key is redacted.
func TestRedactKeyCoversPartialAndEncoderEscapes(t *testing.T) {
	plain := "sk-part/0123456789KEY"
	special := `sk-<a>&"b\c/0123456789`
	marshal := func(k string, html bool) string {
		var b strings.Builder
		enc := json.NewEncoder(&b)
		enc.SetEscapeHTML(html)
		_ = enc.Encode(k)
		return strings.Trim(strings.TrimSpace(b.String()), `"`)
	}
	partial := func(k string) string { // every other byte escaped, the rest verbatim
		var b strings.Builder
		for i, c := range []byte(k) {
			if i%2 == 0 {
				fmt.Fprintf(&b, `\u%04x`, c)
			} else {
				b.WriteByte(c)
			}
		}
		return b.String()
	}
	cases := []struct{ name, key, form string }{
		{"partial \\u00XX escaping", plain, partial(plain)},
		{"slash written as \\/", plain, strings.ReplaceAll(plain, "/", `\/`)},
		{"json.Marshal, HTML escaped", special, marshal(special, true)},
		{"json.Marshal, no HTML escaping", special, marshal(special, false)},
		{"quote and backslash escaped", special, strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(special)},
		{"marshal form with \\/ too", special, strings.ReplaceAll(marshal(special, true), "/", `\/`)},
	}
	for _, c := range cases {
		out := string(RedactKey([]byte(`{"k":"`+c.form+`"}`), c.key))
		require.Equal(t, `{"k":"[redacted]"}`, out, "%s: %q", c.name, c.form)
	}
}
