package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Grant keys client (GRANT-KEYS-DESIGN section 6.1). These call the broker's
// owner-auth /grants endpoints (signed + GitHub-bound, gated like priced share).
// The lean surface a typical provider sees is `grant create --name <label>`;
// everything else is defaulted or behind --advanced (see cmd/rogerai).

// GrantCreateOpts is the full create payload. The CLI fills Name + (Free or a
// price) by default and tucks the rest behind --advanced.
type GrantCreateOpts struct {
	Name       string
	Free       bool
	FreeSet    bool // whether --free was explicitly passed (vs price-derived default)
	PriceIn    float64
	PriceOut   float64
	Models     []string
	Nodes      []string
	RPM        float64
	Burst      float64
	DailyCap   int64
	MonthlyCap int64
	ExpiresAt  int64
	Self       bool
}

// grantJSON is the broker's secret-free grant view.
type grantJSON struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Nodes      []string `json:"nodes"`
	Models     []string `json:"models"`
	Free       bool     `json:"free"`
	Self       bool     `json:"self"`
	Price      string   `json:"price"`
	PriceIn    float64  `json:"price_in"`
	PriceOut   float64  `json:"price_out"`
	AddedBy    int      `json:"added_by"` // how many accounts added it, never who
	RPM        float64  `json:"rpm"`
	DailyCap   int64    `json:"daily_cap"`
	MonthlyCap int64    `json:"monthly_cap"`
	ExpiresAt  int64    `json:"expires_at"`
	Status     string   `json:"status"`
	Usage      struct {
		DayTokens   int64 `json:"day_tokens"`
		MonthTokens int64 `json:"month_tokens"`
	} `json:"usage"`
}

// grantMint creates a grant (signed POST /grants) and returns it with its secret, shown once.
// It is the one create call: `roger grant create`, the TUI and the console all land here.
func grantMint(broker string, payload map[string]any) (grantJSON, string, error) {
	resp, err := postSigned(broker+"/grants", payload)
	if err != nil {
		return grantJSON{}, "", err
	}
	defer resp.Body.Close()
	var out struct {
		Grant  grantJSON `json:"grant"`
		Secret string    `json:"secret"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusOK || out.Secret == "" {
		if out.Error.Message != "" {
			return grantJSON{}, "", fmt.Errorf("%s", out.Error.Message)
		}
		return grantJSON{}, "", fmt.Errorf("grant create failed (status %d)", resp.StatusCode)
	}
	return out.Grant, out.Secret, nil
}

// GrantCreate mints a grant and prints the secret once + ready-to-paste env lines.
func GrantCreate(broker string, o GrantCreateOpts) error {
	free := o.Free
	if !o.FreeSet && (o.PriceIn > 0 || o.PriceOut > 0) {
		free = false
	}
	g, secret, err := grantMint(broker, map[string]any{
		"name": o.Name, "free": free,
		"price_in": o.PriceIn, "price_out": o.PriceOut,
		"models": o.Models, "nodes": o.Nodes,
		"rpm": o.RPM, "burst": o.Burst,
		"daily_cap": o.DailyCap, "monthly_cap": o.MonthlyCap,
		"expires_at": o.ExpiresAt, "self": o.Self,
	})
	if err != nil {
		return err
	}
	kind := "free"
	if g.Self {
		kind = "self ($0 on your own boxes)"
	} else if !g.Free {
		kind = "priced " + g.Price
	}
	fmt.Printf("\ncreated grant %q  (%s)\n\n", g.Name, kind)
	fmt.Printf("  %s\n", secret)
	fmt.Printf("  save it now - it is shown only once.\n\n")
	fmt.Printf("  point any OpenAI app at your models with no login:\n")
	fmt.Printf("    OPENAI_API_BASE=%s\n", KeysBaseURL(broker))
	fmt.Printf("    OPENAI_API_KEY=%s\n", secret)
	if g.DailyCap > 0 {
		fmt.Printf("\n  daily cap: %d tokens/day  (roger grant show %s)\n", g.DailyCap, g.Name)
	}
	return nil
}

// GrantInfo is a compact grant summary for programmatic callers (the in-TUI list).
type GrantInfo struct {
	Name   string `json:"name"`
	Price  string `json:"price"`
	Status string `json:"status"`
}

// GrantCreateSecret mints a grant and returns ONLY the secret (the data form of
// GrantCreate, for the TUI and the console). free=true makes it a free key.
func GrantCreateSecret(broker, name string, free bool) (string, error) {
	_, secret, err := grantMint(broker, map[string]any{"name": name, "free": free})
	return secret, err
}

// grantPriceText is a price as a person types it: dollars per 1M tokens, an optional "$" and
// an optional minus (so a negative price is refused for what it is). The broker's reader
// (cmd/rogerai-broker/grants_http.go) takes the same shapes.
var grantPriceText = regexp.MustCompile(`^-?\$?[0-9]+(\.[0-9]+)?$`)

// MsgGrantPriceNegative is the refusal for a negative grant price, in the broker's words.
const MsgGrantPriceNegative = "price cannot be negative"

// ParseGrantPrice reads a grant price as typed ("0.05", "$0.05", "0", "free") into dollars per
// 1M tokens. It is the one reader for `roger grant create --price-out`, `roger grant reprice`
// and the console, so each refuses a bad price in the same words: MsgRepriceNoPrice for text
// that is not a price (including "1e9" and "NaN"), MsgGrantPriceNegative for a negative one.
// The ceiling is the broker's to judge.
func ParseGrantPrice(s string) (float64, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "free" {
		return 0, nil
	}
	if !grantPriceText.MatchString(t) {
		return 0, errors.New(MsgRepriceNoPrice)
	}
	v, err := strconv.ParseFloat(strings.Replace(t, "$", "", 1), 64)
	switch {
	case err != nil:
		return 0, errors.New(MsgRepriceNoPrice)
	case v < 0:
		return 0, errors.New(MsgGrantPriceNegative)
	}
	return v, nil
}

// GrantCreatePriced mints a grant at priceOut dollars per 1M output tokens (0 makes it free)
// and returns its price as the broker labels it and its secret, shown once. The data form of
// `roger grant create --price-out`, for the console.
func GrantCreatePriced(broker, name string, priceOut float64) (price, secret string, err error) {
	g, secret, err := grantMint(broker, map[string]any{"name": name, "free": priceOut == 0, "price_out": priceOut})
	if err != nil {
		return "", "", err
	}
	return priceLabel(g), secret, nil
}

// GrantListRows returns the owner's grants as compact rows (the in-TUI /grant list).
func GrantListRows(broker string) ([]GrantInfo, error) {
	gs, err := fetchGrants(broker)
	if err != nil {
		return nil, err
	}
	rows := make([]GrantInfo, 0, len(gs))
	for _, g := range gs {
		rows = append(rows, GrantInfo{Name: g.Name, Price: priceLabel(g), Status: g.Status})
	}
	return rows, nil
}

// GrantList prints the caller-owner's grants as a table.
func GrantList(broker string) error {
	gs, err := fetchGrants(broker)
	if err != nil {
		return err
	}
	if len(gs) == 0 {
		fmt.Println("no grants yet - `roger grant create --name my-bots` mints a free key for your bots/family.")
		return nil
	}
	sort.Slice(gs, func(i, j int) bool { return gs[i].Name < gs[j].Name })
	fmt.Printf("%-16s %-8s %-14s %-18s %-10s %s\n", "NAME", "PRICE", "CAPS(rpm/day)", "USED(day/month)", "EXPIRES", "STATUS")
	for _, g := range gs {
		caps := fmt.Sprintf("%s/%s", numOrDash(g.RPM), capOrDash(g.DailyCap))
		used := fmt.Sprintf("%d/%d", g.Usage.DayTokens, g.Usage.MonthTokens)
		fmt.Printf("%-16s %-8s %-14s %-18s %-10s %s\n",
			trunc(g.Name, 16), priceLabel(g), caps, used, expiresLabel(g.ExpiresAt), g.Status)
	}
	return nil
}

// GrantShow prints one grant's full scope + caps + usage (never the secret).
func GrantShow(broker, name string) error {
	g, ok, err := findGrant(broker, name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no grant named %q", name)
	}
	fmt.Printf("grant %q  [%s]\n", g.Name, g.Status)
	fmt.Printf("  id        %s\n", g.ID)
	fmt.Printf("  price     %s\n", priceLabel(g))
	fmt.Printf("  nodes     %s\n", scopeLabel(g.Nodes))
	fmt.Printf("  models    %s\n", scopeLabel(g.Models))
	fmt.Printf("  rpm       %s\n", numOrDash(g.RPM))
	fmt.Printf("  daily     %s tokens/day\n", capOrDash(g.DailyCap))
	fmt.Printf("  monthly   %s tokens/month\n", capOrDash(g.MonthlyCap))
	fmt.Printf("  expires   %s\n", expiresLabel(g.ExpiresAt))
	fmt.Printf("  used      %d tokens today, %d this month\n", g.Usage.DayTokens, g.Usage.MonthTokens)
	return nil
}

// GrantRevoke revokes a grant by name and says so on stdout (the CLI's `roger grant revoke`).
func GrantRevoke(broker, name string) error {
	if err := GrantRevokeByName(broker, name); err != nil {
		return err
	}
	fmt.Printf("revoked %q - the next request with its key is rejected.\n", name)
	return nil
}

// GrantRevokeByName revokes a grant by name (DELETE /grants/{id}) and prints nothing: the data
// form of GrantRevoke, for the TUI and the console.
func GrantRevokeByName(broker, name string) error {
	g, found, err := findGrant(broker, name)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no grant named %q", name)
	}
	if g.Status == "revoked" {
		return fmt.Errorf("grant %q is already revoked", name)
	}
	req, _ := http.NewRequest(http.MethodDelete, broker+"/grants/"+g.ID, nil)
	signRequest(req, nil)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke failed (status %d)", resp.StatusCode)
	}
	return nil
}

// fetchGrants lists the owner's grants (signed GET /grants).
func fetchGrants(broker string) ([]grantJSON, error) {
	req, _ := http.NewRequest(http.MethodGet, broker+"/grants", nil)
	signRequest(req, nil)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		// NEVER suggest "link GitHub" here: an Apple-funded wallet would be stranded by the
		// GitHub-wins precedence. Grants need a linked account of EITHER kind.
		return nil, fmt.Errorf("grants require a linked operator account - sign in (GitHub or Apple) first")
	}
	// Anything but a list is a failed read. Calling it an empty list would tell an owner
	// with grants that they have none.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not read grants (status %d)", resp.StatusCode)
	}
	var out struct {
		Grants []grantJSON `json:"grants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("could not read grants: %v", err)
	}
	return out.Grants, nil
}

// findGrant resolves a grant by its human name (label). A name can be reused after a revoke,
// so the live grant is the one meant; a revoked one is returned only when no other has the name.
func findGrant(broker, name string) (grantJSON, bool, error) {
	gs, err := fetchGrants(broker)
	if err != nil {
		return grantJSON{}, false, err
	}
	var g grantJSON
	found := false
	for _, c := range gs {
		if c.Name == name && (!found || g.Status == "revoked") {
			g, found = c, true
		}
	}
	return g, found, nil
}

func priceLabel(g grantJSON) string {
	if g.Self {
		return "self"
	}
	if g.Free {
		return "free"
	}
	return g.Price
}

func scopeLabel(s []string) string {
	if len(s) == 0 {
		return "any"
	}
	return strings.Join(s, ",")
}

func numOrDash(v float64) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%g", v)
}

func capOrDash(v int64) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", v)
}

func expiresLabel(unix int64) string {
	if unix == 0 {
		return "never"
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}

func trunc(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}
