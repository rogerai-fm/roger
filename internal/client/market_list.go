package client

// The CLI market listing (features/cli/market_listing_kinds.feature, approved 2026-11-25):
// every kind of station in its own section, priced in its own unit. Chat models, one-job
// (single_turn) models, voices, search stations and any kind this copy does not know yet are
// never mixed into one table, so a per-query or per-character price never sits under a token
// column.

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"rogerai.fm/roger/v7/internal/pricetier"
	"rogerai.fm/roger/v7/internal/protocol"
)

// ListOptions narrows and shapes the listing.
type ListOptions struct {
	Kind string // "", "models", "voices" or "search"
	JSON bool   // print the offers grouped by kind
}

// Listing kinds, in print order.
const (
	kindModels      = "models"
	kindSpecialized = "specialized"
	kindVoices      = "voices"
	kindSearch      = "search"
	kindOther       = "other"
)

// ErrBadKind is the refusal for an unknown --kind.
var ErrBadKind = errors.New("--kind must be models, voices or search")

// ValidListKind reports whether k is a kind --kind accepts ("" means every kind).
func ValidListKind(k string) bool {
	switch k {
	case "", kindModels, kindVoices, kindSearch:
		return true
	}
	return false
}

// listOffer is one /discover offer as the listing reads it. Raw keeps the whole object for
// --json, so a script sees every field the broker sent.
type listOffer struct {
	NodeID       string  `json:"node_id"`
	Region       string  `json:"region"`
	Model        string  `json:"model"`
	Modality     string  `json:"modality"`
	Engine       string  `json:"engine"`
	PriceIn      float64 `json:"price_in"`
	PriceOut     float64 `json:"price_out"`
	PriceTier    int     `json:"price_tier"`
	Ctx          int     `json:"ctx"`
	Online       bool    `json:"online"`
	Confidential bool    `json:"confidential"`
	FreeNow      bool    `json:"free_now"`
	TPS          float64 `json:"tps"`
	Signal       int     `json:"signal"`
	Summary      string  `json:"summary"`
	Interaction  string  `json:"interaction"`
	Input        struct {
		Example string `json:"example"`
	} `json:"input"`
	Raw json.RawMessage `json:"-"`
}

// searchStation is one GET /search/stations row.
type searchStation struct {
	ID             string  `json:"id"`
	Model          string  `json:"model"`
	Engine         string  `json:"engine"`
	Scope          string  `json:"scope"`
	PricePer1K     float64 `json:"price_per_1k_queries"`
	Free           bool    `json:"free"`
	OperatorStatus string  `json:"operator_status"`
}

// ListKind files an offer under its listing section. A chat offer (or one from an older broker
// with no modality, or the older "text" label) is a model, unless its published interaction is
// single_turn: then it is a one-job model (K9). A category alone never makes one specialized.
func ListKind(modality, interaction string) string {
	switch strings.ToLower(strings.TrimSpace(modality)) {
	case "", protocol.ModalityChat, "text":
		if protocol.CanonicalInteraction(interaction) == protocol.InteractionSingleTurn {
			return kindSpecialized
		}
		return kindModels
	case protocol.ModalityTTS, protocol.ModalitySTT:
		return kindVoices
	case protocol.ModalitySearch:
		return kindSearch
	}
	return kindOther
}

// FoldInteraction is a model's interaction for the listing: the broker's fold
// (protocol.FoldInteraction: online stations vote, the majority wins, a tie is chat), and,
// when no online station voted, the same fold over every station, so a one-job model that is
// off the air is still listed as one.
func FoldInteraction(ballots []protocol.Ballot) string {
	if inter, _ := protocol.FoldInteraction(ballots); inter != "" {
		return inter
	}
	all := make([]protocol.Ballot, len(ballots))
	for i, b := range ballots {
		b.Online = true
		all[i] = b
	}
	inter, _ := protocol.FoldInteraction(all)
	return inter
}

// ChatPartner reports whether a model may be offered as a general chat partner: a chat-kind
// offer whose interaction is not single_turn. Every chat picker shares this one rule (K9).
func ChatPartner(modality, interaction string) bool {
	return ListKind(modality, interaction) == kindModels
}

// Market prints the live market (GET /discover, plus GET /search/stations for the search rows)
// in sections: MODELS, SPECIALIZED, VOICES, SEARCH, OTHER. An empty section is not printed.
func Market(broker string, opt ListOptions) error {
	if !ValidListKind(opt.Kind) {
		return ErrBadKind
	}
	var d struct {
		Offers []json.RawMessage `json:"offers"`
	}
	if err := getJSON(broker, "/discover", "", &d); err != nil {
		return err
	}
	var all []listOffer
	ballots := map[string][]protocol.Ballot{} // per model: the broker's interaction fold
	for _, raw := range d.Offers {
		var o listOffer
		if json.Unmarshal(raw, &o) != nil {
			continue
		}
		o.Raw = raw
		all = append(all, o)
		ballots[o.Model] = append(ballots[o.Model], protocol.Ballot{Online: o.Online,
			T: protocol.Taxonomy{Interaction: o.Interaction, Example: o.Input.Example}})
	}
	folded := map[string]string{}
	for model, bs := range ballots {
		folded[model] = FoldInteraction(bs)
	}
	groups := map[string][]listOffer{}
	var feedSearch []listOffer
	for _, o := range all {
		// One model, one section: its kind is its stations' fold, never one station's word.
		k := ListKind(o.Modality, folded[o.Model])
		if k == kindSearch {
			feedSearch = append(feedSearch, o)
			continue
		}
		groups[k] = append(groups[k], o)
	}
	want := func(k string) bool {
		switch opt.Kind {
		case "":
			return true
		case kindModels:
			return k == kindModels
		}
		return k == opt.Kind
	}
	var stations []searchStation
	if want(kindSearch) { // a listing without search never asks for it
		stations = searchStations(broker, feedSearch)
	}
	if opt.JSON {
		return printListJSON(groups, stations, want)
	}

	width := listWidth()
	printed := false
	gap := func() {
		if printed {
			fmt.Println()
		}
		printed = true
	}
	if want(kindModels) && len(groups[kindModels]) > 0 {
		gap()
		printModels(groups[kindModels], width)
	}
	if want(kindSpecialized) && len(groups[kindSpecialized]) > 0 {
		gap()
		printSpecialized(groups[kindSpecialized], width)
	}
	if want(kindVoices) && len(groups[kindVoices]) > 0 {
		gap()
		printVoices(groups[kindVoices])
	}
	if want(kindSearch) && len(stations) > 0 {
		gap()
		printSearch(stations)
	}
	if want(kindOther) && len(groups[kindOther]) > 0 {
		gap()
		printOther(groups[kindOther])
	}
	if !printed {
		if opt.Kind == "" {
			fmt.Println("no offers yet - run `roger share` on a box with a local model")
		} else {
			fmt.Printf("no %s on the air right now\n", opt.Kind)
		}
	}
	return nil
}

// searchStations reads GET /search/stations. A broker without it (older than search) gets
// rows built from the /discover offers, keyed by model, priced from price_in per 1M queries.
func searchStations(broker string, feed []listOffer) []searchStation {
	var r struct {
		Stations []searchStation `json:"stations"`
	}
	if err := getJSON(broker, "/search/stations", "", &r); err == nil && r.Stations != nil {
		return r.Stations
	}
	var out []searchStation
	for _, o := range feed {
		out = append(out, searchStation{ID: o.Model, Model: o.Model, Engine: o.Engine,
			PricePer1K: o.PriceIn / 1000, Free: o.PriceIn == 0})
	}
	return out
}

func printListJSON(groups map[string][]listOffer, stations []searchStation, want func(string) bool) error {
	out := map[string]any{}
	for _, k := range []string{kindModels, kindSpecialized, kindVoices, kindOther} {
		rows := []json.RawMessage{}
		if want(k) {
			for _, o := range groups[k] {
				rows = append(rows, o.Raw)
			}
		}
		out[k] = rows
	}
	if !want(kindSearch) || stations == nil {
		stations = []searchStation{}
	}
	out[kindSearch] = stations
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func modelRow(o listOffer) string {
	status := glyphOnAir
	if !o.Online {
		status = glyphOffAir
	}
	tps := "-"
	if o.TPS > 0 {
		tps = fmt.Sprintf("%.0f", o.TPS)
	}
	flags := ""
	if o.Confidential {
		flags += glyphVerify + " verified "
	}
	if o.FreeNow {
		flags += "FREE-now"
	}
	return fmt.Sprintf("%-8s %-12s %-22s %-9.2f %-9.2f %-13s %-7s %-7d %-7s %-7s %s",
		status, signalTower(o.Signal, o.TPS, o.Online), o.Model, o.PriceIn, o.PriceOut,
		pricetier.Label(o.PriceTier, o.PriceOut), tps, o.Ctx, o.Region, o.NodeID, flags)
}

// Station rows mirror the TUI band table's instrument language so the piped CLI reads as a
// terminal twin of the on-screen one. Plain text (no color), so it degrades cleanly.
const modelHeader = "%-8s %-12s %-22s %-9s %-9s %-13s %-7s %-7s %-7s %-7s %s\n"

func printModelHeader() {
	fmt.Printf(modelHeader, "STATUS", "SIGNAL", "MODEL", "$/1M in", "$/1M out", "TIER", "TOK/S", "CTX", "REGION", "NODE", "FLAGS")
}

func printModels(rows []listOffer, width int) {
	fmt.Println("MODELS  chat, priced per 1M tokens")
	printModelHeader()
	for _, o := range rows {
		fmt.Println(withSummary(modelRow(o), o.Summary, width))
	}
}

// printSpecialized lists one-job models: same columns as models, then the summary and the
// example input, so a reader sees what to send. They are not chat partners.
func printSpecialized(rows []listOffer, width int) {
	fmt.Println("SPECIALIZED  one job each: send one input like the example, not a conversation")
	printModelHeader()
	for _, o := range rows {
		fmt.Println(strings.TrimRight(modelRow(o), " "))
		if s := protocol.DetailsText(o.Summary, protocol.SummaryMax); s != "" {
			fmt.Println("         " + s)
		}
		if ex := strings.TrimSpace(o.Input.Example); ex != "" {
			fmt.Println("         send: " + cutTo(ex, width-len("         send: ")))
		}
	}
}

// cutTo shortens s to n runes with "…" when the terminal is narrower than s; piped (n <= 0)
// output is never cut.
func cutTo(s string, n int) string {
	r := []rune(s)
	if n <= 0 || len(r) <= n || n < 8 {
		return s
	}
	return string(r[:n-1]) + "…"
}

// bytesPerMinute is the TUI's listening-post estimate: ~128 kbps of uploaded audio.
const bytesPerMinute = 960000

func printVoices(rows []listOffer) {
	fmt.Println("VOICES  speak (text to speech) and listen (speech to text)")
	fmt.Printf("%-8s %-8s %-22s %-11s %-9s %-7s %s\n", "STATUS", "KIND", "MODEL", "$/1K chars", "~$/min*", "REGION", "NODE")
	listens := false
	for _, o := range rows {
		status := glyphOnAir
		if !o.Online {
			status = glyphOffAir
		}
		kind, chars, perMin := "SPEAKS", price(o.PriceIn/1000), "-"
		if strings.EqualFold(o.Modality, protocol.ModalitySTT) {
			kind, chars, perMin = "LISTENS", "-", price(o.PriceIn*bytesPerMinute/1e6)
			if perMin != "FREE" {
				perMin += "*"
			}
			listens = true
		}
		fmt.Printf("%-8s %-8s %-22s %-11s %-9s %-7s %s\n", status, kind, o.Model, chars, perMin, o.Region, o.NodeID)
	}
	if listens {
		fmt.Println("* listeners bill by uploaded audio bytes; per minute is an estimate at ~128 kbps")
	}
}

func printSearch(rows []searchStation) {
	fmt.Println("SEARCH  web search stations, answered on POST /v1/search")
	fmt.Printf("%-26s %-8s %-13s %-11s %s\n", "STATION", "ENGINE", "$/1K queries", "OPERATOR", "SCOPE")
	for _, s := range rows {
		p := price(s.PricePer1K)
		if s.Free {
			p = "FREE"
		}
		id := s.ID
		if id == "" {
			id = s.Model
		}
		fmt.Println(strings.TrimRight(fmt.Sprintf("%-26s %-8s %-13s %-11s %s", id, s.Engine, p, s.OperatorStatus, s.Scope), " "))
	}
}

func printOther(rows []listOffer) {
	fmt.Println("OTHER  kinds this copy of roger does not know yet (upgrade to use them)")
	fmt.Printf("%-22s %-10s %s\n", "MODEL", "KIND", "NODE")
	for _, o := range rows {
		fmt.Printf("%-22s %-10s %s\n", o.Model, o.Modality, o.NodeID)
	}
}

// price prints dollars with no trailing zeros, rounded to a millionth: 0.001, 0.48, 5. Zero is
// FREE.
func price(d float64) string {
	if d <= 0 || math.IsNaN(d) || math.IsInf(d, 0) {
		return "FREE"
	}
	return strconv.FormatFloat(math.Round(d*1e6)/1e6, 'f', -1, 64)
}
