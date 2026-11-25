package station

// execute.go is the Station actually doing the work: checking that Roger Core really
// authorized it, running it, and signing for exactly what came back.
//
// # THE STATION IS THE LAST PLACE ANY OF THIS CAN BE CAUGHT
//
// By the time a request arrives here it has passed through the Tower, which is the one party
// in the exchange that is not trusted and the one holding every byte. So the Station does not
// take the Tower's word for anything:
//
//	the GRANT must be signed by CORE, using a key pinned into this Station out of band. A
//	relay cannot mint one, and cannot alter one it was given.
//	the GRANT must name THIS Station. A valid grant for a different machine is somebody
//	else's authorization, and pointing it here is exactly what a relay is positioned to do.
//	the REQUEST must be the bytes the grant commits to. Otherwise a relay could pair a real
//	grant with a request of its own, and the receipt would attest to work nobody asked for.
//
// Only then does it execute. What it signs afterwards is a digest of the exact bytes it is
// returning, so a relay that changes the answer on the way back invalidates the receipt and
// Core refuses the result.
//
// # THE PINNED KEY
//
// A Station never talks to Core. It only ever talks to its Tower - which is precisely the
// party a forged grant would come from - so it cannot fetch Core's key over that channel and
// learn anything. The key is pinned by the operator from Core's public endpoint, and that
// out-of-band step is what the whole verification rests on.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"rogerai.fm/roger/v7/internal/ctxsig"
	"rogerai.fm/roger/v7/internal/towercore/dispatch"
)

// ExecuteRequest is what a Tower hands a Station.
type ExecuteRequest struct {
	// Grant is Core's signed authorization, relayed verbatim.
	Grant json.RawMessage `json:"grant"`
	// Envelope is the request, SEALED to this Station's secure-session key. The Tower carries
	// it and cannot read it; the grant commits to a digest of what is inside.
	Envelope json.RawMessage `json:"envelope"`
}

// ExecuteResponse is what the Station hands back.
type ExecuteResponse struct {
	Receipt *dispatch.Receipt `json:"receipt,omitempty"`
	// Envelope is the result, SEALED to Roger Core's envelope key. The receipt commits to a
	// digest of the PLAINTEXT inside it, so Core checks after opening.
	Envelope json.RawMessage `json:"envelope,omitempty"`
	// Failure is set when the Station could not serve. It carries NO receipt: a failure is
	// not a result, and must never be capable of settling one.
	Failure string `json:"failure,omitempty"`
}

// Upstream is the local model this Station serves from. It is an interface so the executor
// can be tested against a real HTTP server without a real model, and so a Station can later
// serve something that is not an HTTP endpoint at all.
type Upstream interface {
	Serve(ctx context.Context, request []byte) ([]byte, error)
}

// Executor turns an authorized request into a signed result.
type Executor struct {
	Station *Station
	// CoreKey is Core's grant-signing public key, pinned by the operator. Without it nothing
	// can be verified and the Station refuses everything - which is the correct behaviour for
	// a Station that does not know who is allowed to give it work.
	CoreKey []byte
	// CoreEnvelopeKey is the X25519 key results are sealed to, pinned alongside CoreKey. A
	// Station cannot reach Core directly, so both come from the operator out of band.
	CoreEnvelopeKey []byte
	Network         string
	Upstream        Upstream
	Now             func() time.Time
}

func attemptOf(grant []byte) string {
	var obj struct {
		AttemptID string `json:"attempt_id"`
	}
	if err := json.Unmarshal(grant, &obj); err != nil {
		return ""
	}
	return obj.AttemptID
}

// HTTPUpstream serves from an OpenAI-compatible endpoint - the shape every local runner
// already speaks, so a Station is pointed at what the operator is running rather than
// requiring anything new of them.
//
// Contract: features/tower/tower_upstream_key.feature.
//
// Key is the bearer key a key-protected upstream requires (`roger share --upstream-key`). It
// is sent to the upstream and to nothing else, and it is REDACTED here, at the source, from
// every byte the upstream returns - body and error text alike - so the receipt digest, the
// token count, the audit transcript, the consumer's sealed answer and the operator's log line
// are all made from the same redacted bytes and none of them can carry it.
type HTTPUpstream struct {
	URL     string
	Key     string
	Client  *http.Client
	Timeout time.Duration
}

// RedactKey strips key from b in every form a JSON encoder, or a hand-rolled one, can write it:
// each CHARACTER of the key may appear verbatim or as \uXXXX (either case of hex digit; a
// character outside the Basic Multilingual Plane as its \uXXXX\uXXXX surrogate pair), "/" also
// as \/, and " and \ as \" and \\. That covers the verbatim key, json.Marshal's form with or
// without HTML escaping (< > & become \u003c \u003e \u0026), and any mix of escaped and plain
// characters. Per character, not per byte: a key need not be ASCII, and a pattern built from the
// bytes of a multi-byte character is not valid UTF-8 and does not compile (it panicked every
// serving path on the first reply - found by the pre-push audit 2026-11-24). A key that is not
// valid UTF-8 at all is matched verbatim only, since no JSON encoder writes invalid UTF-8 in an
// escaped form. A no-op for an empty key, which would match everywhere.
func RedactKey(b []byte, key string) []byte {
	if key == "" || len(b) == 0 {
		return b
	}
	if !utf8.ValidString(key) {
		return bytes.ReplaceAll(b, []byte(key), []byte("[redacted]"))
	}
	return keyPattern(key).ReplaceAll(b, []byte("[redacted]"))
}

// keyPattern is the RE2 pattern RedactKey matches: one alternation per key character. RE2 runs in time
// linear in the body, so a large body cannot make redaction slow. Compiled once per key: the
// classic streaming path redacts every line.
func keyPattern(key string) *regexp.Regexp {
	if re, ok := keyPatterns.Load(key); ok {
		return re.(*regexp.Regexp)
	}
	re := compileKeyPattern(key)
	keyPatterns.Store(key, re)
	return re
}

// keyPatterns caches compiled patterns by key. A process serves one or a few keys.
var keyPatterns sync.Map

func compileKeyPattern(key string) *regexp.Regexp {
	var p strings.Builder
	for _, c := range key {
		alts := []string{regexp.QuoteMeta(string(c)), jsonEscapePattern(c)}
		switch c {
		case '/':
			alts = append(alts, `\\/`)
		case '"':
			alts = append(alts, `\\"`)
		case '\\':
			alts = append(alts, `\\\\`)
		}
		p.WriteString("(?:" + strings.Join(alts, "|") + ")")
	}
	return regexp.MustCompile(p.String())
}

// jsonEscapePattern matches c written as a JSON \u escape, hex digits in either case: \uXXXX for a
// character in the Basic Multilingual Plane, the \uXXXX\uXXXX surrogate pair for one beyond it.
func jsonEscapePattern(c rune) string {
	units := []rune{c}
	if c > 0xFFFF {
		hi, lo := utf16.EncodeRune(c)
		units = []rune{hi, lo}
	}
	var p strings.Builder
	for _, u := range units {
		p.WriteString(`\\u`)
		for _, d := range []byte(fmt.Sprintf("%04x", u)) {
			p.WriteString(hexClass(d))
		}
	}
	return p.String()
}

// hexClass matches one hex digit in either case.
func hexClass(d byte) string {
	if d >= 'a' && d <= 'f' {
		return "[" + string([]byte{d}) + string([]byte{d - 'a' + 'A'}) + "]"
	}
	return string([]byte{d})
}

// upstreamReadLimit bounds what one upstream reply may be, so a reply that streams forever cannot
// exhaust the Station's memory.
const upstreamReadLimit = 8 << 20

// ErrUpstreamTooLarge refuses a reply past upstreamReadLimit. It is refused, not cut: a cut could
// leave a prefix of the upstream key that redaction no longer recognises.
var ErrUpstreamTooLarge = errors.New("the model returned more than this Station reads")

func (u HTTPUpstream) Serve(ctx context.Context, request []byte) ([]byte, error) {
	timeout := u.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.URL, bytes.NewReader(request))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if u.Key != "" {
		req.Header.Set("Authorization", "Bearer "+u.Key)
	}
	client := u.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Bounded: an upstream that streams forever must not be able to exhaust this Station's
	// memory on one request.
	body, err := io.ReadAll(io.LimitReader(resp.Body, upstreamReadLimit+1))
	if err != nil {
		return nil, err
	}
	if len(body) > upstreamReadLimit {
		return nil, ErrUpstreamTooLarge
	}
	// Redacted BEFORE anything is measured, signed, kept or shortened: a cut made first could
	// leave part of the key that a whole-key replace no longer matches.
	body = RedactKey(body, u.Key)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &UpstreamStatusError{Status: resp.StatusCode, RetryAfter: retryAfterSeconds(resp.Header.Get("Retry-After")),
			Detail: truncate(string(body), 200), ContextWindow: ctxsig.IsRefusal(resp.StatusCode, string(body))}
	}
	if len(body) == 0 {
		return nil, errors.New("the model returned an empty body")
	}
	return body, nil
}

// UpstreamStatusError is a non-2xx reply from the model behind a Station. The status and the
// Retry-After (seconds, 0 when absent) are what a relay needs to answer a consumer honestly
// (a 429 with its wait, not a generic failure); Detail is the model's own words, for the
// operator's log only - it never crosses a blind Tower (see EdgeExecutor.ServeSealed).
// ContextWindow is whether the model's whole reply was a context-window refusal
// (ctxsig.IsRefusal): one bit, read here because only the Station sees the text.
type UpstreamStatusError struct {
	Status        int
	RetryAfter    int
	Detail        string
	ContextWindow bool
}

func (e *UpstreamStatusError) Error() string {
	return fmt.Sprintf("the model replied %d: %s", e.Status, e.Detail)
}

// Class is the failure as it may cross a blind Tower: the class, the status and the wait,
// never the model's words. A 400 that was a context-overflow complaint is its own class, so
// the relay can tell the consumer the prompt is too long; no other status carries the word.
func (e *UpstreamStatusError) Class() string {
	if e.ContextWindow && e.Status == http.StatusBadRequest {
		return ctxsig.TowerClass
	}
	if e.RetryAfter > 0 {
		return fmt.Sprintf("the model did not answer (status %d, retry-after %d)", e.Status, e.RetryAfter)
	}
	return fmt.Sprintf("the model did not answer (status %d)", e.Status)
}

func retryAfterSeconds(v string) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// tokenUsageOf parses the model's own token counts from an OpenAI-compatible response body
// (its "usage" object), for the Option C per-token receipt. A missing usage object, a non-JSON
// body, or negative counts yield zero - the per-token settle path then bills nothing for this
// request and the byte cap + audit govern, exactly as an un-tokened receipt. The node signs
// whatever this returns; Core clamps it to the grant's token ceiling and the Tower's
// byte-attestation, so an inflated figure here cannot exceed what was authorized.
func tokenUsageOf(body []byte) dispatch.Usage {
	var parsed struct {
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	_ = json.Unmarshal(body, &parsed)
	in, out := parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens
	if in < 0 {
		in = 0
	}
	if out < 0 {
		out = 0
	}
	return dispatch.Usage{In: in, Out: out}
}
