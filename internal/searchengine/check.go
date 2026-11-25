package searchengine

// check.go is the search/1 contract check `roger search check` runs against a native service
// before it goes on air (features/search/native_provider.feature NP2). It talks to the service
// directly, not through Search, because it has to see what Search would quietly clean up.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Check names, in the order they run. The public provider guide lists exactly these
// (docs/search-providers.md); a test keeps the two from drifting.
const (
	CheckEndpoint   = "endpoint"
	CheckJSON       = "reply is JSON"
	CheckResults    = "results array"
	CheckCount      = "respects count"
	CheckURLs       = "http(s) urls only"
	CheckTitles     = "every result has a title"
	CheckPublished  = "published is RFC 3339"
	CheckTime       = "answers within 10 s"
	CheckSize       = "reply under 512 KiB"
	CheckFreshness  = "accepts freshness"
	CheckUnicode    = "accepts Unicode"
	checkCountAsked = 2
)

// CheckNames is every contract check, in order.
var CheckNames = []string{CheckEndpoint, CheckJSON, CheckResults, CheckCount, CheckURLs, CheckTitles,
	CheckPublished, CheckTime, CheckSize, CheckFreshness, CheckUnicode}

// CheckQueries are the only queries the check sends: fixed text, no personal data.
var CheckQueries = []string{"weather", "café 日本語 🌍"}

// CheckResult is one check's outcome. Status is "ok", "FAIL" or "warn".
type CheckResult struct {
	Name   string
	Status string
	Detail string
}

// probeReply is one raw exchange with the service.
type probeReply struct {
	status  int
	body    []byte
	results []map[string]any
	err     error
}

// CheckContract runs the search/1 checks against a native service and returns every outcome.
// A failure early in the chain (no endpoint, not JSON) skips the checks that depend on it.
func CheckContract(ctx context.Context, e Engine) []CheckResult {
	var out []CheckResult
	add := func(name, status, detail string) { out = append(out, CheckResult{name, status, detail}) }

	base := e.probe(ctx, Request{Query: CheckQueries[0], Count: checkCountAsked})
	switch {
	case errors.Is(base.err, ErrTimeout):
		add(CheckTime, "FAIL", "no answer within "+EngineTimeout.String())
		return out
	case base.err != nil:
		add(CheckEndpoint, "FAIL", "could not reach POST /v1/search")
		return out
	case base.status != http.StatusOK:
		add(CheckEndpoint, "FAIL", "POST /v1/search answered HTTP "+http.StatusText(base.status))
		return out
	}
	add(CheckEndpoint, "ok", "")
	add(CheckTime, "ok", "")
	if len(base.body) > ReplyMax {
		add(CheckSize, "FAIL", "the reply is over 512 KiB")
		return out
	}
	add(CheckSize, "ok", "")
	var top map[string]json.RawMessage
	if json.Unmarshal(base.body, &top) != nil {
		add(CheckJSON, "FAIL", "the reply is not a JSON object")
		return out
	}
	add(CheckJSON, "ok", "")
	if base.results == nil {
		add(CheckResults, "FAIL", `the reply has no "results" array`)
		return out
	}
	add(CheckResults, "ok", "")

	pass := func(name string, ok bool, detail string) {
		if ok {
			add(name, "ok", "")
		} else {
			add(name, "FAIL", detail)
		}
	}
	pass(CheckCount, len(base.results) <= checkCountAsked, "asked for 2 results, got more")
	urlsOK, titlesOK, publishedOK := true, true, true
	for _, r := range base.results {
		u, _ := r["url"].(string)
		if p, err := url.Parse(u); err != nil || (p.Scheme != "http" && p.Scheme != "https") || p.Host == "" {
			urlsOK = false
		}
		if t, _ := r["title"].(string); strings.TrimSpace(t) == "" {
			titlesOK = false
		}
		if p, has := r["published"]; has && p != nil {
			s, _ := p.(string)
			if !rfc3339Date(s) {
				publishedOK = false
			}
		}
	}
	pass(CheckURLs, urlsOK, "a result's url is not an absolute http or https URL")
	pass(CheckTitles, titlesOK, "a result has no title")
	pass(CheckPublished, publishedOK, "a published value is not an RFC 3339 date or date-time")

	fresh := e.probe(ctx, Request{Query: CheckQueries[0], Count: checkCountAsked, Freshness: "week"})
	pass(CheckFreshness, fresh.err == nil && fresh.status == http.StatusOK, `a query with freshness "week" failed`)
	uni := e.probe(ctx, Request{Query: CheckQueries[1], Count: checkCountAsked})
	pass(CheckUnicode, uni.err == nil && uni.status == http.StatusOK, "a query with non-ASCII text failed")

	// Soft findings: filters the service accepts but appears to ignore.
	country := e.probe(ctx, Request{Query: CheckQueries[0], Count: checkCountAsked, Country: "US"})
	var ignored []string
	if fresh.err == nil && reflect.DeepEqual(fresh.results, base.results) {
		ignored = append(ignored, "freshness")
	}
	if country.err == nil && reflect.DeepEqual(country.results, base.results) {
		ignored = append(ignored, "country")
	}
	if len(ignored) > 0 {
		add(strings.Join(ignored, " and ")+" ignored", "warn", "same results with and without them, so a caller who filters by them gets unfiltered results")
	}
	return out
}

// Failed reports whether any check failed.
func Failed(rs []CheckResult) bool {
	for _, r := range rs {
		if r.Status == "FAIL" {
			return true
		}
	}
	return false
}

// probe sends one raw search/1 request and reads the reply without cleaning it.
func (e Engine) probe(ctx context.Context, req Request) probeReply {
	ctx, cancel := context.WithTimeout(ctx, EngineTimeout)
	defer cancel()
	hr, err := e.nativeRequest(ctx, req)
	if err != nil {
		return probeReply{err: ErrUnreachable}
	}
	resp, err := e.client().Do(hr)
	if err != nil {
		return probeReply{err: transportErr(ctx, err)}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, ReplyMax+1))
	if err != nil {
		return probeReply{err: transportErr(ctx, err)}
	}
	p := probeReply{status: resp.StatusCode, body: body}
	var top struct {
		Results []map[string]any `json:"results"`
	}
	if json.Unmarshal(bytes.TrimSpace(body), &top) == nil {
		p.results = top.Results
	}
	return p
}

func rfc3339Date(s string) bool {
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return true
	}
	_, err := time.Parse(time.RFC3339, s)
	return err == nil
}
