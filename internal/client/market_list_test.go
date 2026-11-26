package client

// Unit tests for the sectioned market listing (market_list.go). The behavior spec is
// features/cli/market_listing_kinds.feature, run end to end from cmd/rogerai.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func listOut(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan string, 1)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	err := fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done, err
}

// fakeMarket serves /discover and, when stations is non-nil, /search/stations.
func fakeMarket(t *testing.T, offers []map[string]any, stations []map[string]any) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/discover":
			_ = json.NewEncoder(w).Encode(map[string]any{"offers": offers})
		case r.URL.Path == "/search/stations" && stations != nil:
			_ = json.NewEncoder(w).Encode(map[string]any{"stations": stations})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func everyKind() []map[string]any {
	return []map[string]any{
		{"model": "llama3.2", "modality": "chat", "price_in": 0.1, "price_out": 0.3, "online": true, "tps": 40, "confidential": true, "free_now": true, "summary": "a chat model"},
		{"model": "old-model", "online": false},
		{"model": "fire", "modality": "chat", "interaction": "single_turn", "online": true,
			"summary": "one job", "input": map[string]any{"example": `{"tick": 1}`}},
		{"model": "fire-dark", "modality": "chat", "interaction": "Single-Turn", "online": false},
		{"model": "operator", "modality": "tts", "price_in": 1.0, "online": true},
		{"model": "whisper", "modality": "stt", "price_in": 0.5, "online": false},
		{"model": "free-ear", "modality": "stt", "online": true},
		{"model": "web", "modality": "search", "engine": "searxng", "price_in": 5000, "online": true},
		{"model": "mystery", "modality": "video", "online": true},
		{"model": "legacy", "modality": "text", "online": true},
	}
}

func TestListKind(t *testing.T) {
	for _, c := range []struct{ mod, inter, want string }{
		{"", "", kindModels}, {"chat", "", kindModels}, {"text", "", kindModels}, {"CHAT", "chat", kindModels},
		{"chat", "single_turn", kindSpecialized}, {"", "Single-Turn", kindSpecialized},
		{"tts", "", kindVoices}, {"stt", "single_turn", kindVoices},
		{"search", "", kindSearch}, {"video", "", kindOther},
	} {
		require.Equal(t, c.want, ListKind(c.mod, c.inter), "%q/%q", c.mod, c.inter)
	}
	require.True(t, ChatPartner("chat", ""))
	require.False(t, ChatPartner("chat", "single_turn"))
	require.False(t, ChatPartner("search", ""))
}

func TestValidListKind(t *testing.T) {
	for _, k := range []string{"", "models", "voices", "search"} {
		require.True(t, ValidListKind(k), k)
	}
	for _, k := range []string{"specialized", "other", "pictures"} {
		require.False(t, ValidListKind(k), k)
	}
}

func TestMarketPrintsEverySectionInOrder(t *testing.T) {
	t.Setenv("COLUMNS", "60")
	url := fakeMarket(t, everyKind(), []map[string]any{
		{"id": "@acme/web", "model": "web", "engine": "searxng", "scope": "the open web", "price_per_1k_queries": 0, "free": true, "operator_status": "verified"},
		{"model": "bare", "engine": "native", "price_per_1k_queries": 2.5},
	})
	out, err := listOut(t, func() error { return Market(url, ListOptions{}) })
	require.NoError(t, err)
	at := func(s string) int { i := strings.Index(out, s); require.GreaterOrEqual(t, i, 0, "missing %q:\n%s", s, out); return i }
	require.Less(t, at("MODELS"), at("SPECIALIZED"))
	require.Less(t, at("SPECIALIZED"), at("VOICES"))
	require.Less(t, at("VOICES"), at("SEARCH"))
	require.Less(t, at("SEARCH"), at("OTHER"))
	at("SPEAKS")
	at("LISTENS")
	at("0.001")
	at("0.48*")
	at("@acme/web")
	at("the open web")
	at("bare")
	at("2.5")
	at("send: ")
	at("* listeners bill")
	require.Regexp(t, `mystery +video`, out)
	// "legacy" is a model; "web" never appears under MODELS
	models := out[at("MODELS"):at("SPECIALIZED")]
	require.Contains(t, models, "legacy")
	require.NotContains(t, models, " web ")
}

func TestMarketKindFilterAndEmpty(t *testing.T) {
	url := fakeMarket(t, everyKind(), nil) // an older broker: no /search/stations
	out, err := listOut(t, func() error { return Market(url, ListOptions{Kind: "search"}) })
	require.NoError(t, err)
	require.Contains(t, out, "SEARCH")
	require.Contains(t, out, "web ") // built from the /discover offer
	require.Regexp(t, `web +searxng +5\b`, out) // 5000 per 1M = $5 per 1K
	require.NotContains(t, out, "MODELS")

	out, err = listOut(t, func() error { return Market(url, ListOptions{Kind: "models"}) })
	require.NoError(t, err)
	require.Contains(t, out, "MODELS")
	require.NotContains(t, out, "SPECIALIZED")

	empty := fakeMarket(t, nil, nil)
	out, _ = listOut(t, func() error { return Market(empty, ListOptions{}) })
	require.Contains(t, out, "no offers yet")
	out, _ = listOut(t, func() error { return Market(empty, ListOptions{Kind: "voices"}) })
	require.Contains(t, out, "no voices on the air")

	require.ErrorIs(t, Market(url, ListOptions{Kind: "pictures"}), ErrBadKind)
}

func TestMarketJSON(t *testing.T) {
	url := fakeMarket(t, everyKind(), []map[string]any{{"id": "@acme/web", "engine": "searxng", "free": true}})
	out, err := listOut(t, func() error { return Market(url, ListOptions{JSON: true}) })
	require.NoError(t, err)
	var m map[string][]map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &m))
	for _, k := range []string{"models", "specialized", "voices", "search", "other"} {
		require.Contains(t, m, k)
	}
	require.Equal(t, "@acme/web", m["search"][0]["id"])
	require.Len(t, m["specialized"], 2)

	out, err = listOut(t, func() error { return Market(url, ListOptions{JSON: true, Kind: "voices"}) })
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(out), &m))
	require.Empty(t, m["models"])
	require.Empty(t, m["search"])
	require.Len(t, m["voices"], 3)
}

func TestMarketBrokerDown(t *testing.T) {
	require.Error(t, Market("http://127.0.0.1:1", ListOptions{}))
}

func TestPriceAndCut(t *testing.T) {
	require.Equal(t, "FREE", price(0))
	require.Equal(t, "FREE", price(-1))
	require.Equal(t, "0.001", price(0.001))
	require.Equal(t, "0.48", price(0.5*960000/1e6))
	require.Equal(t, "5", price(5))
	require.Equal(t, "abc", cutTo("abc", 0))
	require.Equal(t, "abcdefghij", cutTo("abcdefghij", 20))
	require.Equal(t, "abcdefg…", cutTo("abcdefghijkl", 8))
	require.Equal(t, "abcdefghijkl", cutTo("abcdefghijkl", 4)) // too narrow to cut usefully
}

// A model's kind is its stations' fold (protocol.FoldInteraction), so one model never lands
// in two sections: online stations vote, the majority wins, a tie is chat.
func TestMarketFoldsAModelIntoOneSection(t *testing.T) {
	offers := []map[string]any{
		{"model": "mixed", "modality": "chat", "interaction": "chat", "online": true, "node_id": "a"},
		{"model": "mixed", "modality": "chat", "interaction": "single_turn", "online": false, "node_id": "b"},
		{"model": "job", "modality": "chat", "interaction": "single_turn", "online": true, "node_id": "c"},
		{"model": "job", "modality": "chat", "online": false, "node_id": "d"},
	}
	out, err := listOut(t, func() error { return Market(fakeMarket(t, offers, nil), ListOptions{}) })
	require.NoError(t, err)
	spec := strings.Index(out, "SPECIALIZED")
	require.Greater(t, spec, 0, out)
	require.Equal(t, 2, strings.Count(out[:spec], "mixed"), "both mixed rows under MODELS:\n%s", out)
	require.NotContains(t, out[spec:], "mixed")
	row := regexp.MustCompile(`(?m)^[◉○]\s+\S+\s+job\s`) // a station row (status glyph) whose model is "job"
	require.Len(t, row.FindAllString(out[spec:], -1), 2, "both job rows under SPECIALIZED:\n%s", out)
	require.Empty(t, row.FindAllString(out[:spec], -1))
}
