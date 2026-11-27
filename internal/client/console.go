package client

import (
	"encoding/json"
	"net/http"
	"time"
)

// This file exposes data-form readers used by the localhost web console (internal/webui),
// which renders these surfaces itself rather than printing them. They wrap the same broker
// calls the CLI/TUI already use, so there is one code path per surface.

// Discover fetches the broker's current open-market offer list. Exported data form of the
// internal failover discover.
func Discover(broker string) ([]Offer, error) { return discover(broker) }

// BandRangeOf is the live out-price range of model across offers already fetched: the data the
// connect screens quote from, for a caller that also reads the offers itself.
func BandRangeOf(offers []Offer, model string) (BandRange, bool) { return bandRange(offers, model) }

// BalanceInfo is the signed wallet read: the balance, whether the broker recognizes a real
// account (logged in), and the month-to-date spend cap + spend.
type BalanceInfo struct {
	Balance      float64 `json:"balance"`
	LoggedIn     bool    `json:"logged_in"`
	MonthlyCap   float64 `json:"monthly_cap"`
	MonthlySpend float64 `json:"monthly_spend"`
}

// balanceTimeout bounds the balance read, as signedDo bounds every other signed account call.
var balanceTimeout = 30 * time.Second

// FetchBalance reads the signed wallet balance for user from broker — the data form of the
// CLI's Balance print and the TUI's fetchBalance.
func FetchBalance(broker, user string) (BalanceInfo, error) {
	req, _ := http.NewRequest(http.MethodGet, broker+"/balance", nil)
	SignRequest(req, nil)
	req.Header.Set("X-Roger-User", user)
	resp, err := (&http.Client{Timeout: balanceTimeout}).Do(req)
	if err != nil {
		return BalanceInfo{}, err
	}
	defer resp.Body.Close()
	var b BalanceInfo
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return BalanceInfo{}, err
	}
	return b, nil
}
