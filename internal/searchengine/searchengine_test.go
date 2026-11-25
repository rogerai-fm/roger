package searchengine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/protocol"
)

// engineServer answers every request with status and body.
func engineServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSearchMapsEngineFailures(t *testing.T) {
	cases := []struct {
		name   string
		kind   string
		status int
		body   string
		want   error
		reason string
	}{
		{"rate limited", protocol.EngineNative, 429, "", ErrRateLimited, "engine rate limited"},
		{"searxng json off", protocol.EngineSearXNG, 403, "", ErrJSONOff, "the engine's JSON output is turned off: add json to search.formats in SearXNG's settings.yml"},
		{"brave key refused", protocol.EngineBrave, 401, "", ErrKeyRefused, "engine refused the key"},
		{"native forbidden is a key refusal", protocol.EngineNative, 403, "", ErrKeyRefused, "engine refused the key"},
		{"other status", protocol.EngineNative, 500, "stack trace", ErrStatus, "engine error (status 500)"},
		{"redirect is a status, never followed", protocol.EngineNative, 302, "", ErrStatus, "engine error (status 302)"},
		{"not json", protocol.EngineNative, 200, "<html>", ErrUnreadable, "engine reply unreadable"},
		{"native without results", protocol.EngineNative, 200, `{"hits":[]}`, ErrUnreadable, "engine reply unreadable"},
		{"searxng not json", protocol.EngineSearXNG, 200, "nope", ErrUnreadable, "engine reply unreadable"},
		{"brave not json", protocol.EngineBrave, 200, "nope", ErrUnreadable, "engine reply unreadable"},
		{"too large", protocol.EngineNative, 200, `{"results":[{"title":"` + strings.Repeat("x", ReplyMax) + `"}]}`, ErrTooLarge, "engine reply too large"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := engineServer(t, c.status, c.body)
			_, err := Engine{Kind: c.kind, Upstream: srv.URL, Key: "k"}.Search(context.Background(), Request{Query: "q", Count: 3})
			require.ErrorIs(t, err, c.want)
			require.Equal(t, c.reason, err.Error())
		})
	}
}

func TestSearchConfigAndTransportFailures(t *testing.T) {
	_, err := Engine{Kind: "google", Upstream: "http://x"}.Search(context.Background(), Request{Query: "q"})
	require.ErrorIs(t, err, ErrConfig)

	for _, kind := range []string{protocol.EngineBrave, protocol.EngineSearXNG, protocol.EngineNative} {
		_, err = Engine{Kind: kind, Upstream: "http://bad host\x7f"}.Search(context.Background(), Request{Query: "q"})
		require.ErrorIs(t, err, ErrConfig, kind)
	}

	srv := engineServer(t, 200, "{}")
	srv.Close()
	_, err = Engine{Kind: protocol.EngineNative, Upstream: srv.URL}.Search(nil, Request{Query: "q"}) //nolint:staticcheck // nil ctx is accepted
	require.ErrorIs(t, err, ErrUnreachable)

	prev := EngineTimeout
	EngineTimeout = 100 * time.Millisecond
	t.Cleanup(func() { EngineTimeout = prev })
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	t.Cleanup(slow.Close)
	_, err = Engine{Kind: protocol.EngineNative, Upstream: slow.URL}.Search(context.Background(), Request{Query: "q"})
	require.ErrorIs(t, err, ErrTimeout)

	// headers sent, then the body stalls past the deadline
	stall := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[`))
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	t.Cleanup(stall.Close)
	_, err = Engine{Kind: protocol.EngineNative, Upstream: stall.URL}.Search(context.Background(), Request{Query: "q"})
	require.ErrorIs(t, err, ErrTimeout)
}

func TestSearchDecodesEachEngineAndNormalizes(t *testing.T) {
	brave := engineServer(t, 200, `{"web":{"results":[
		{"title":"<b>A</b>","url":"https://a.example/","description":"one\ntwo","page_age":"2026-11-01"},
		{"title":"B","url":"ftp://b.example/","description":"dropped"},
		{"title":"C","url":"https://c.example/","description":"c"}]}}`)
	rs, err := Engine{Kind: protocol.EngineBrave, Upstream: brave.URL, Key: "k"}.Search(context.Background(), Request{Query: "q", Count: 5})
	require.NoError(t, err)
	require.Equal(t, []Result{
		{Title: "A", URL: "https://a.example/", Snippet: "one two", Published: "2026-11-01"},
		{Title: "C", URL: "https://c.example/", Snippet: "c"},
	}, rs)

	searx := engineServer(t, 200, `{"results":[
		{"title":"A","url":"https://a.example/","content":"x","publishedDate":"2026-01-02"},
		{"title":"B","url":"https://b.example/","content":"y"}]}`)
	rs, err = Engine{Kind: protocol.EngineSearXNG, Upstream: searx.URL + "/"}.Search(context.Background(), Request{Query: "q", Count: 1, Freshness: "day"})
	require.NoError(t, err)
	require.Equal(t, []Result{{Title: "A", URL: "https://a.example/", Snippet: "x", Published: "2026-01-02"}}, rs)

	native := engineServer(t, 200, `{"results":[{"title":"N","url":" https://n.example/ ","snippet":"s"},{"title":"x","url":"https://"}]}`)
	rs, err = Engine{Kind: protocol.EngineNative, Upstream: native.URL}.Search(context.Background(), Request{Query: "q"})
	require.NoError(t, err)
	require.Equal(t, []Result{{Title: "N", URL: "https://n.example/", Snippet: "s"}}, rs)

	noWeb := engineServer(t, 200, `{"query":{}}`)
	rs, err = Engine{Kind: protocol.EngineBrave, Upstream: noWeb.URL}.Search(context.Background(), Request{Query: "q", Count: 3})
	require.NoError(t, err)
	require.Empty(t, rs)
}

func TestBraveRequestShapeAndDefaultEndpoint(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		_, _ = w.Write([]byte(`{"web":{"results":[]}}`))
	}))
	t.Cleanup(srv.Close)
	_, err := Engine{Kind: protocol.EngineBrave, Upstream: srv.URL, Key: "k1"}.Search(context.Background(),
		Request{Query: "tides", Count: 7, Freshness: "year", Country: "DE"})
	require.NoError(t, err)
	require.Equal(t, "tides", got.URL.Query().Get("q"))
	require.Equal(t, "7", got.URL.Query().Get("count"))
	require.Equal(t, "py", got.URL.Query().Get("freshness"))
	require.Equal(t, "DE", got.URL.Query().Get("country"))
	require.Equal(t, "k1", got.Header.Get("X-Subscription-Token"))

	hr, err := Engine{Kind: protocol.EngineBrave}.braveRequest(context.Background(), Request{Query: "q"})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(hr.URL.String(), BraveEndpoint))
	require.Empty(t, hr.URL.Query().Get("freshness"))
	require.Empty(t, hr.URL.Query().Get("country"))
}

func TestServeJob(t *testing.T) {
	srv := engineServer(t, 200, `{"results":[{"title":"T","url":"https://t.example/","snippet":"s"}]}`)
	engines := map[string]Engine{"go-docs": {Kind: protocol.EngineNative, Upstream: srv.URL}}

	st, body := ServeJob(context.Background(), engines, protocol.Job{Body: []byte("not json")})
	require.Equal(t, http.StatusBadRequest, st)
	require.JSONEq(t, `{"error":"unreadable search job"}`, string(body))

	st, body = ServeJob(context.Background(), engines, protocol.Job{Body: []byte(`{"model":"other","query":"q"}`)})
	require.Equal(t, http.StatusNotFound, st)
	require.Contains(t, string(body), `no search offer \"other\" on this node`)

	st, body = ServeJob(context.Background(), engines, protocol.Job{Body: []byte(`{"model":"go-docs","query":"q","count":3}`)})
	require.Equal(t, http.StatusOK, st)
	var out struct {
		Results []Result `json:"results"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	require.Len(t, out.Results, 1)

	bad := map[string]Engine{"go-docs": {Kind: protocol.EngineNative, Upstream: engineServer(t, 429, "").URL}}
	st, body = ServeJob(context.Background(), bad, protocol.Job{Body: []byte(`{"model":"go-docs","query":"q"}`)})
	require.Equal(t, http.StatusBadGateway, st)
	require.JSONEq(t, `{"error":"engine rate limited"}`, string(body))
}

// contractServer answers search/1 requests through reply, which sees the decoded request.
func contractServer(t *testing.T, reply func(w http.ResponseWriter, req Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		_ = json.NewDecoder(r.Body).Decode(&req)
		reply(w, req)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func statusOf(rs []CheckResult, name string) string {
	for _, r := range rs {
		if r.Name == name {
			return r.Status
		}
	}
	return ""
}

func TestCheckContract(t *testing.T) {
	good := func(n int) string {
		items := []string{}
		for i := 0; i < n; i++ {
			items = append(items, `{"title":"T","url":"https://e.example/`+string(rune('a'+i))+`","snippet":"s","published":"2026-11-01"}`)
		}
		return `{"results":[` + strings.Join(items, ",") + `]}`
	}
	t.Run("a conforming service that honours its filters passes with no warning", func(t *testing.T) {
		srv := contractServer(t, func(w http.ResponseWriter, req Request) {
			n := req.Count
			if req.Freshness != "" || req.Country != "" {
				n = 1 // the filters change the answer
			}
			_, _ = w.Write([]byte(good(n)))
		})
		rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: srv.URL})
		require.False(t, Failed(rs))
		require.Len(t, rs, len(CheckNames))
		for _, r := range rs {
			require.Equal(t, "ok", r.Status, r.Name)
		}
	})
	t.Run("ignored filters are a warning, not a failure", func(t *testing.T) {
		srv := contractServer(t, func(w http.ResponseWriter, req Request) { _, _ = w.Write([]byte(good(2))) })
		rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: srv.URL})
		require.False(t, Failed(rs))
		require.Equal(t, "warn", statusOf(rs, "freshness and country ignored"))
	})
	cases := []struct {
		name  string
		reply func(w http.ResponseWriter, req Request)
		fails string
	}{
		{"endpoint", func(w http.ResponseWriter, _ Request) { w.WriteHeader(404) }, CheckEndpoint},
		{"size", func(w http.ResponseWriter, _ Request) { _, _ = w.Write([]byte(strings.Repeat(" ", ReplyMax+10))) }, CheckSize},
		{"json", func(w http.ResponseWriter, _ Request) { _, _ = w.Write([]byte("<html>")) }, CheckJSON},
		{"results", func(w http.ResponseWriter, _ Request) { _, _ = w.Write([]byte(`{"hits":[]}`)) }, CheckResults},
		{"count", func(w http.ResponseWriter, _ Request) { _, _ = w.Write([]byte(good(5))) }, CheckCount},
		{"urls", func(w http.ResponseWriter, _ Request) {
			_, _ = w.Write([]byte(`{"results":[{"title":"T","url":"javascript:x"}]}`))
		}, CheckURLs},
		{"titles", func(w http.ResponseWriter, _ Request) {
			_, _ = w.Write([]byte(`{"results":[{"title":" ","url":"https://e.example/"}]}`))
		}, CheckTitles},
		{"published", func(w http.ResponseWriter, _ Request) {
			_, _ = w.Write([]byte(`{"results":[{"title":"T","url":"https://e.example/","published":"last tuesday"}]}`))
		}, CheckPublished},
		{"freshness", func(w http.ResponseWriter, req Request) {
			if req.Freshness != "" {
				w.WriteHeader(500)
				return
			}
			_, _ = w.Write([]byte(good(1)))
		}, CheckFreshness},
		{"unicode", func(w http.ResponseWriter, req Request) {
			if req.Query != CheckQueries[0] {
				w.WriteHeader(500)
				return
			}
			_, _ = w.Write([]byte(good(1)))
		}, CheckUnicode},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := contractServer(t, c.reply)
			rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: srv.URL})
			require.True(t, Failed(rs))
			require.Equal(t, "FAIL", statusOf(rs, c.fails), "%+v", rs)
		})
	}
	t.Run("unreachable", func(t *testing.T) {
		srv := contractServer(t, nil)
		srv.Close()
		rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: srv.URL})
		require.Equal(t, "FAIL", statusOf(rs, CheckEndpoint))
	})
	t.Run("bad address", func(t *testing.T) {
		rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: "http://bad host\x7f"})
		require.Equal(t, "FAIL", statusOf(rs, CheckEndpoint))
	})
	t.Run("too slow", func(t *testing.T) {
		prev := EngineTimeout
		EngineTimeout = 100 * time.Millisecond
		t.Cleanup(func() { EngineTimeout = prev })
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
		}))
		t.Cleanup(srv.Close)
		rs := CheckContract(context.Background(), Engine{Kind: protocol.EngineNative, Upstream: srv.URL})
		require.Equal(t, "FAIL", statusOf(rs, CheckTime))
	})
}

func TestRFC3339Date(t *testing.T) {
	for in, want := range map[string]bool{
		"2026-11-01": true, "2026-11-01T10:00:00Z": true, "2026-11-01T10:00:00+02:00": true,
		"2026-13-45": false, "yesterday": false, "": false,
	} {
		require.Equal(t, want, rfc3339Date(in), in)
	}
}
