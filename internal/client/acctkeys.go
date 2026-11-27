package client

// acctkeys.go: account API keys (`rog-key_...`) over the broker's /account/keys endpoints,
// signed with the local user key - the same endpoints and fields the web account page uses
// (features/auth/key_guardrails.feature). A minted secret is returned to the caller exactly
// once and is never logged or stored here.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
)

// AccountKey is one key as GET /account/keys shows it (never the secret).
type AccountKey struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Hint           string   `json:"hint"`
	LimitUSD       float64  `json:"limit_usd"`
	LimitRemaining *float64 `json:"limit_remaining"`
	Reset          string   `json:"reset"`
	ExpiresAt      *string  `json:"expires_at"`
	Expired        bool     `json:"expired"`
	Disabled       bool     `json:"disabled"`
	Usage          float64  `json:"usage"`
	UsageDaily     float64  `json:"usage_daily"`
	UsageWeekly    float64  `json:"usage_weekly"`
	UsageMonthly   float64  `json:"usage_monthly"`
	AllowedModels  []string `json:"allowed_models"`
	AllowedNodes   []string `json:"allowed_nodes"`
}

// ErrKeysLogin is the broker refusing a key call because the account is not logged in. A
// surface with its own way to log in matches on it and says so in its own words.
var ErrKeysLogin = errors.New("log in first")

// The two ways a typed limit is refused. Each surface adds its own hint for how to ask for no
// limit (a flag value, a word, a tick box).
var (
	ErrKeyLimit         = errors.New("state a limit in dollars")
	ErrKeyLimitNegative = errors.New("a limit cannot be negative")
	// To the broker 0 means unlimited, the opposite of what someone typing 0 expects.
	ErrKeyLimitZero = errors.New("0 would mean unlimited, not no spending - use none for an unlimited key")
)

// At most six decimals: the broker rounds to six, and a smaller amount would round to 0,
// which it reads as unlimited.
var keyAmountRe = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)

var keyIDRe = regexp.MustCompile(`^key_[A-Za-z0-9]+$`)

// KeyIDOK reports whether id has the shape of a broker key id. Anything else never becomes
// part of a broker path.
func KeyIDOK(id string) bool { return keyIDRe.MatchString(id) }

// ParseKeyLimit reads a spend limit as a person writes it: "5", "$5", "5.50", or "none". It
// is the one reading of a limit, used by `roger keys --limit`, the TUI's /keys and the
// console. "none" (or "unlimited") is no limit, sent to the broker as 0. A typed 0 is refused
// (ErrKeyLimitZero). Digits only: "1e9" and "Inf" are not amounts.
func ParseKeyLimit(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "none" || s == "unlimited" {
		return 0, nil
	}
	amount, negative := strings.CutPrefix(s, "-")
	amount = strings.TrimPrefix(amount, "$")
	v, err := strconv.ParseFloat(amount, 64)
	switch {
	case !keyAmountRe.MatchString(amount) || err != nil:
		return 0, ErrKeyLimit
	case v == 0:
		return 0, ErrKeyLimitZero
	case negative:
		return 0, ErrKeyLimitNegative
	}
	return v, nil
}

// KeyResetOK reports whether s is a limit window the broker takes.
func KeyResetOK(s string) bool {
	switch s {
	case "daily", "weekly", "monthly", "none":
		return true
	}
	return false
}

// KeysBaseURL is the OpenAI base URL a key or a grant is used against: the broker's /v1.
func KeysBaseURL(broker string) string { return strings.TrimRight(broker, "/") + "/v1" }

// KeyUnlimitedNote is what a surface says when a key is made or left with no limit.
const KeyUnlimitedNote = "unlimited - this key can spend your whole wallet"

// ErrKeyNotFound is FindKey finding no key with the name asked for. Each surface adds how to
// see the list.
var ErrKeyNotFound = errors.New("no key named")

// FindKey matches what a person typed against the account's keys: an id first, then a name.
// A name on more than one key is refused with the ids to use instead.
func FindKey(keys []AccountKey, arg string) (AccountKey, error) {
	var ids []string
	var named AccountKey
	for _, k := range keys {
		if k.ID == arg {
			return k, nil
		}
		if k.Name == arg {
			named, ids = k, append(ids, k.ID)
		}
	}
	switch len(ids) {
	case 1:
		return named, nil
	case 0:
		return AccountKey{}, fmt.Errorf("%w %s", ErrKeyNotFound, arg)
	}
	return AccountKey{}, fmt.Errorf("%d keys are named %s - use the id: %s", len(ids), arg, strings.Join(ids, ", "))
}

// KeyExportLines is the paste-ready block for a new key or grant: the broker is the endpoint.
func KeyExportLines(broker, secret string) string {
	return "export OPENAI_BASE_URL=" + KeysBaseURL(broker) + "\nexport OPENAI_API_KEY=" + secret
}

// LimitText is the key's limit as a list reads it: "$5.00 weekly", "$5.00 total" for a limit
// that never resets, "unlimited" for none.
func (k AccountKey) LimitText() string {
	if k.LimitUSD <= 0 {
		return "unlimited"
	}
	window := k.Reset
	if window == "" || window == "none" {
		window = "total"
	}
	return FormatUSD(k.LimitUSD) + " " + window
}

// UsedText is the spend in the window the limit counts, as a list reads it.
func (k AccountKey) UsedText() string { return FormatUSD(k.WindowUsage()) + " used" }

// keysCall runs one signed /account/keys request and decodes a 2xx body into out (nil = none).
func keysCall(method, broker, path string, body any, out any) error {
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	resp, err := signedDo(method, broker, path, raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBrokerUnreachable, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w - run `roger login` (keys are per account)", ErrKeysLogin)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error.Message != "" {
			return fmt.Errorf("%s", e.Error.Message)
		}
		return fmt.Errorf("broker returned status %d", resp.StatusCode)
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// KeysList reads the account's keys (newest first, as the broker orders them).
func KeysList(broker string) ([]AccountKey, error) {
	var out struct {
		Keys []AccountKey `json:"keys"`
	}
	err := keysCall(http.MethodGet, broker, "/account/keys", nil, &out)
	return out.Keys, err
}

// KeysMint mints a key with the given fields and returns it with its secret (shown once).
func KeysMint(broker string, fields map[string]any) (AccountKey, string, error) {
	var out struct {
		AccountKey
		Secret string `json:"secret"`
	}
	err := keysCall(http.MethodPost, broker, "/account/keys", fields, &out)
	return out.AccountKey, out.Secret, err
}

// KeysSet patches the named fields of a key and returns the updated key.
func KeysSet(broker, id string, fields map[string]any) (AccountKey, error) {
	var out AccountKey
	err := keysCall(http.MethodPatch, broker, "/account/keys/"+id, fields, &out)
	return out, err
}

// KeysRevoke deletes (revokes) a key.
func KeysRevoke(broker, id string) error {
	return keysCall(http.MethodDelete, broker, "/account/keys/"+id, nil, nil)
}

// WriteKeysTable prints keys as the mono table `roger keys list` shows: id, name, hint,
// limit, used (spend in the limit's current window), resets, expires and state.
func WriteKeysTable(w io.Writer, keys []AccountKey) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tHINT\tLIMIT\tUSED\tRESETS\tEXPIRES\tSTATE")
	for _, k := range keys {
		limit := "unlimited"
		if k.LimitUSD > 0 {
			limit = fmt.Sprintf("$%.2f", k.LimitUSD)
		}
		expires := "never"
		if k.ExpiresAt != nil && *k.ExpiresAt != "" {
			expires = *k.ExpiresAt
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t$%.2f\t%s\t%s\t%s\n", k.ID, k.Name, k.Hint, limit, k.WindowUsage(), k.Reset, expires, k.State())
	}
	_ = tw.Flush()
}

// WindowUsage is the spend in the window the limit counts: today, this week, this month, or
// the key's whole life for reset "none".
func (k AccountKey) WindowUsage() float64 {
	switch k.Reset {
	case "daily":
		return k.UsageDaily
	case "weekly":
		return k.UsageWeekly
	case "monthly":
		return k.UsageMonthly
	}
	return k.Usage
}

// State is "disabled", "expired" or "active".
func (k AccountKey) State() string {
	switch {
	case k.Disabled:
		return "disabled"
	case k.Expired:
		return "expired"
	}
	return "active"
}

// KeyTransportOK refuses to carry an account key in the clear: the broker must be https, or
// plain http only to this machine (a local Tower or a test broker).
func KeyTransportOK(broker string) error {
	u, err := url.Parse(broker)
	if err != nil {
		return fmt.Errorf("broker %q: %v", broker, err)
	}
	if u.Scheme == "https" {
		return nil
	}
	if ip := net.ParseIP(u.Hostname()); u.Scheme == "http" && (u.Hostname() == "localhost" || ip != nil && ip.IsLoopback()) {
		return nil
	}
	return fmt.Errorf("an account key is sent only over https (or to this machine), not to %s", broker)
}

// IsAccountKey reports whether s looks like an account key secret.
func IsAccountKey(s string) bool { return strings.HasPrefix(s, "rog-key_") && len(s) > len("rog-key_") }
