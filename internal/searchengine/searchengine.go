// Package searchengine is the share node's side of the search modality: it answers a
// broker search job by asking the provider's engine (Brave, SearXNG, or a native service
// speaking the search/1 contract) and returns the normalized result list.
// Specs: features/search/node_bridge.feature, native_provider.feature.
//
// The node normalizes, but it is not the caller's trust boundary: the broker re-validates
// every reply (features/search/result_validation.feature).
package searchengine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rogerai.fm/roger/v7/internal/protocol"
)

// ContractVersion is the native contract this node speaks (native_provider.feature NP1).
const ContractVersion = "search/1"

// BraveEndpoint is the Brave Search API web endpoint, used when a Brave share sets none.
const BraveEndpoint = "https://api.search.brave.com/res/v1/web/search"

// DefaultEngineTimeout bounds one engine call (N3): well under the broker's relay wait, so the
// broker answers before the edge does.
const DefaultEngineTimeout = 10 * time.Second

// EngineTimeout is the bound in force. A var (the shellTimeout precedent) so a test can
// shorten it; production never assigns it.
var EngineTimeout = DefaultEngineTimeout

// ReplyMax caps an engine's reply (N3).
const ReplyMax = 512 << 10

// Request is a search job as the broker sends it (already validated and bounded there).
type Request struct {
	Model     string `json:"model,omitempty"`
	Query     string `json:"query"`
	Count     int    `json:"count"`
	Freshness string `json:"freshness,omitempty"`
	Country   string `json:"country,omitempty"`
}

// Result is one normalized result.
type Result struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet"`
	Published string `json:"published,omitempty"`
}

// Engine is one configured search backend.
type Engine struct {
	Kind     string // protocol.EngineBrave | EngineSearXNG | EngineNative
	Upstream string
	Key      string
	Client   *http.Client
}

// The kinds of engine failure. Each error Search returns wraps one (errors.Is), and its text
// is a short fixed reason: never the engine's body, URL or key.
var (
	ErrUnreachable = errors.New("engine unreachable")
	ErrTimeout     = errors.New("engine timed out")
	ErrRateLimited = errors.New("engine rate limited")
	ErrKeyRefused  = errors.New("engine refused the key")
	ErrJSONOff     = errors.New("the engine's JSON output is turned off")
	ErrStatus      = errors.New("engine error")
	ErrTooLarge    = errors.New("engine reply too large")
	ErrUnreadable  = errors.New("engine reply unreadable")
	ErrConfig      = errors.New("engine misconfigured")
)

// engineErr is one failure: its kind and the reason a caller is shown.
type engineErr struct {
	kind   error
	reason string
}

func (e engineErr) Error() string { return e.reason }
func (e engineErr) Unwrap() error { return e.kind }

func fail(kind error) error { return engineErr{kind, kind.Error()} }

// ServeJob answers one broker search job using the engine configured for its model. It never
// passes the job's pseudonym or id to the engine (privacy.feature P3).
func ServeJob(ctx context.Context, engines map[string]Engine, job protocol.Job) (status int, body []byte) {
	var req Request
	if json.Unmarshal(job.Body, &req) != nil {
		return http.StatusBadRequest, errBody("unreadable search job")
	}
	e, ok := engines[req.Model]
	if !ok {
		return http.StatusNotFound, errBody("no search offer " + strconv.Quote(req.Model) + " on this node")
	}
	results, err := e.Search(ctx, req)
	if err != nil {
		return http.StatusBadGateway, errBody(err.Error())
	}
	b, _ := json.Marshal(map[string]any{"results": results})
	return http.StatusOK, b
}

func errBody(reason string) []byte {
	b, _ := json.Marshal(map[string]string{"error": reason})
	return b
}

// Search asks the engine once (no retry, ever: a 429 must not be answered by hammering a
// rate-limited engine) and returns at most req.Count normalized results.
func (e Engine) Search(ctx context.Context, req Request) ([]Result, error) {
	rs, _, err := e.search(ctx, req)
	return rs, err
}

// Check is Search for the go-live check: it also returns how many results the engine sent
// before the unusable ones (no http(s) URL) were dropped, so a well-formed empty reply can be
// told apart from a reply whose every result was unusable.
func (e Engine) Check(ctx context.Context, req Request) (usable []Result, sent int, err error) {
	return e.search(ctx, req)
}

func (e Engine) search(ctx context.Context, req Request) ([]Result, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, EngineTimeout)
	defer cancel()
	var hreq *http.Request
	var err error
	switch e.Kind {
	case protocol.EngineBrave:
		hreq, err = e.braveRequest(ctx, req)
	case protocol.EngineSearXNG:
		hreq, err = e.searxRequest(ctx, req)
	case protocol.EngineNative:
		hreq, err = e.nativeRequest(ctx, req)
	default:
		return nil, 0, engineErr{ErrConfig, "unknown search engine"}
	}
	if err != nil {
		return nil, 0, engineErr{ErrConfig, "bad engine address"}
	}
	resp, err := e.client().Do(hreq)
	if err != nil {
		return nil, 0, transportErr(ctx, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, 0, fail(ErrRateLimited)
	case resp.StatusCode == http.StatusForbidden && e.Kind == protocol.EngineSearXNG:
		return nil, 0, engineErr{ErrJSONOff, "the engine's JSON output is turned off: add json to search.formats in SearXNG's settings.yml"}
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, 0, fail(ErrKeyRefused)
	case resp.StatusCode >= 300:
		return nil, 0, engineErr{ErrStatus, fmt.Sprintf("engine error (status %d)", resp.StatusCode)}
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, ReplyMax+1))
	if err != nil {
		return nil, 0, transportErr(ctx, err)
	}
	if len(raw) > ReplyMax {
		return nil, 0, fail(ErrTooLarge)
	}
	results, err := e.decode(raw)
	if err != nil {
		return nil, 0, fail(ErrUnreadable)
	}
	return normalize(results, req.Count), len(results), nil
}

// transportErr names a failure to reach or hear from the engine.
func transportErr(ctx context.Context, err error) error {
	var ne net.Error
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
		return fail(ErrTimeout)
	}
	return fail(ErrUnreachable)
}

// client never follows a redirect: an engine that answers 3xx gets that status reported, and
// no other host is contacted.
func (e Engine) client() *http.Client {
	c := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if e.Client != nil {
		c.Transport = e.Client.Transport
	}
	return &c
}

var braveFreshness = map[string]string{"day": "pd", "week": "pw", "month": "pm", "year": "py"}

func (e Engine) braveRequest(ctx context.Context, req Request) (*http.Request, error) {
	endpoint := e.Upstream
	if endpoint == "" {
		endpoint = BraveEndpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", req.Query)
	q.Set("count", strconv.Itoa(req.Count))
	if f := braveFreshness[req.Freshness]; f != "" {
		q.Set("freshness", f)
	}
	if req.Country != "" {
		q.Set("country", req.Country)
	}
	u.RawQuery = q.Encode()
	hr, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	hr.Header.Set("Accept", "application/json")
	hr.Header.Set("X-Subscription-Token", e.Key)
	return hr, nil
}

func (e Engine) searxRequest(ctx context.Context, req Request) (*http.Request, error) {
	u, err := url.Parse(strings.TrimRight(e.Upstream, "/") + "/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", req.Query)
	q.Set("format", "json")
	if req.Freshness != "" {
		q.Set("time_range", req.Freshness) // SearXNG takes day | week | month | year verbatim
	}
	u.RawQuery = q.Encode() // SearXNG has no country parameter: it is not sent
	hr, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	hr.Header.Set("Accept", "application/json")
	return hr, nil
}

func (e Engine) nativeRequest(ctx context.Context, req Request) (*http.Request, error) {
	body, _ := json.Marshal(Request{Query: req.Query, Count: req.Count, Freshness: req.Freshness, Country: req.Country})
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(e.Upstream, "/")+"/v1/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("X-Search-Contract", ContractVersion)
	if e.Key != "" {
		hr.Header.Set("Authorization", "Bearer "+e.Key)
	}
	return hr, nil
}

// decode maps the engine's own reply shape onto Result.
func (e Engine) decode(raw []byte) ([]Result, error) {
	switch e.Kind {
	case protocol.EngineBrave:
		var p struct {
			Web *struct {
				Results []struct {
					Title       string `json:"title"`
					URL         string `json:"url"`
					Description string `json:"description"`
					PageAge     string `json:"page_age"`
				} `json:"results"`
			} `json:"web"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		out := []Result{}
		if p.Web != nil { // no "web" section is zero results, not an error
			for _, r := range p.Web.Results {
				out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Description, Published: r.PageAge})
			}
		}
		return out, nil
	case protocol.EngineSearXNG:
		var p struct {
			Results []struct {
				Title         string `json:"title"`
				URL           string `json:"url"`
				Content       string `json:"content"`
				PublishedDate string `json:"publishedDate"`
			} `json:"results"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		out := make([]Result, 0, len(p.Results))
		for _, r := range p.Results {
			out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Content, Published: r.PublishedDate})
		}
		return out, nil
	default:
		var p struct {
			Results *[]Result `json:"results"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		if p.Results == nil {
			return nil, errors.New("no results array")
		}
		return *p.Results, nil
	}
}

// normalize keeps http(s) results only, flattens their text, and cuts to count in the
// engine's order.
func normalize(in []Result, count int) []Result {
	out := make([]Result, 0, len(in))
	for _, r := range in {
		if count > 0 && len(out) == count {
			break
		}
		u, err := url.Parse(strings.TrimSpace(r.URL))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			continue
		}
		out = append(out, Result{
			Title: protocol.FlattenText(r.Title), URL: strings.TrimSpace(r.URL),
			Snippet: protocol.FlattenText(r.Snippet), Published: strings.TrimSpace(r.Published),
		})
	}
	return out
}
