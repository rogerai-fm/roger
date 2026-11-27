package webui

// grant_adds.go: the ACCOUNT tab's "Grants you added" card and the owner's Reprice
// (features/grants/grant_surfaces.feature, the Console rule). The console signs the same
// /account/grants and /grants/{id}/reprice calls the CLI makes. A pasted secret is sent to the
// broker once and is in no reply; a reprice's new secret is in its one reply and nowhere else.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rogerai.fm/roger/v7/internal/client"
)

const grantsLogin = "log in to manage grants"

// addedRow is a grant this account added, as the card draws it. Never a secret.
type addedRow struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Price    string   `json:"price"`
	Stations []string `json:"stations"`
	Models   []string `json:"models"`
	State    string   `json:"state"`
	Added    string   `json:"added"`
	Notes    []string `json:"notes"`
}

func addedRowOf(g client.AddedGrant) addedRow {
	r := addedRow{ID: g.ID, Name: g.Name, Price: g.Price, Models: g.Models, State: g.StateWords(), Stations: []string{}, Notes: []string{}}
	if r.Models == nil {
		r.Models = []string{}
	}
	if g.AddedAt > 0 {
		r.Added = time.Unix(g.AddedAt, 0).UTC().Format("2006-01-02")
	}
	for _, st := range g.Stations {
		r.Stations = append(r.Stations, st.ID+" (public "+publicPrice(st.PublicIn, st.PublicOut)+")")
		if st.AbovePublic {
			r.Notes = append(r.Notes, st.ID+": the grant price is above its public price")
		}
	}
	if len(g.Stations) == 0 {
		r.Notes = append(r.Notes, "It covers no station yet. Add it again later.")
	}
	return r
}

// publicPrice is a station's public price in the words the CLI and the TUI use (S4).
func publicPrice(in, out float64) string {
	usd := func(v float64) string { return "$" + strconv.FormatFloat(v, 'f', -1, 64) }
	switch {
	case in == 0 && out == 0:
		return "free"
	case in == 0:
		return usd(out) + "/1M out"
	}
	return usd(in) + " in " + usd(out) + " out /1M"
}

// grantsFail reports a refusal in the broker's words, with the console's own words for being
// logged out and, at the add limit, the card's Remove button instead of the CLI command
// (rulings S10, S11). hide is a secret that must not be echoed back, or "".
func grantsFail(w http.ResponseWriter, err error, hide string) {
	msg := err.Error()
	if hide != "" {
		msg = strings.ReplaceAll(msg, hide, "[secret]")
	}
	switch {
	case errors.Is(err, client.ErrNotSignedIn), errors.Is(err, client.ErrKeysLogin):
		keysSay(w, http.StatusUnauthorized, grantsLogin)
	case errors.Is(err, client.ErrGrantAddLimit):
		head, _, _ := strings.Cut(msg, " - ")
		keysSay(w, http.StatusConflict, head+" - remove one with the Remove button first")
	case errors.Is(err, client.ErrBrokerUnreachable):
		keysSay(w, http.StatusServiceUnavailable, keysUnreachable)
	default:
		keysSay(w, http.StatusBadGateway, msg)
	}
}

// handleGrantsAdded lists the grants this account added (GET).
func (s *Server) handleGrantsAdded(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		keysSay(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.keysReady(w) {
		return
	}
	gs, max, err := client.GrantAddedList(s.opts.Broker)
	if err != nil {
		grantsFail(w, err, "")
		return
	}
	rows := make([]addedRow, 0, len(gs))
	for _, g := range gs {
		rows = append(rows, addedRowOf(g))
	}
	writeChannel(w, http.StatusOK, map[string]any{"grants": rows, "max": max})
}

// handleGrantAdd adds a grant by its secret. Body: {secret}. The reply shows what the broker
// returned and never the secret.
func (s *Server) handleGrantAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"secret"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	secret := strings.TrimSpace(req.Secret)
	if secret == "" {
		keysSay(w, http.StatusBadRequest, "paste the grant secret")
		return
	}
	if !s.keysReady(w) {
		return
	}
	g, err := client.GrantAdd(s.opts.Broker, secret)
	if err != nil {
		grantsFail(w, err, secret)
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true, "grant": addedRowOf(g)})
}

// handleGrantRemove drops one added grant from this account. Body: {id}.
func (s *Server) handleGrantRemove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	if req.ID == "" {
		keysSay(w, http.StatusBadRequest, "id required")
		return
	}
	if !s.keysReady(w) {
		return
	}
	g, err := client.GrantRemove(s.opts.Broker, req.ID)
	if err != nil {
		grantsFail(w, err, "")
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true, "grant": addedRowOf(g)})
}

// priceTyped reads a price as typed (text or a number) with the CLI's reader. ok is false
// when no price was given.
func priceTyped(raw json.RawMessage) (v float64, ok bool, err error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false, nil
	}
	text := string(raw)
	_ = json.Unmarshal(raw, &text) // typed text arrives quoted; a number is read as written
	if strings.TrimSpace(text) == "" {
		return 0, false, nil
	}
	v, err = client.ParseGrantPrice(text)
	return v, true, err
}

// repriceNote is what a reprice ended, in words (ruling S5).
func repriceNote(dropped int) string {
	switch dropped {
	case 0:
		return "The old secret has stopped working. No account had added it."
	case 1:
		return "The old secret has stopped working. 1 account that added it is on the public price until they add the new one."
	}
	return fmt.Sprintf("The old secret has stopped working. %d accounts that added it are on the public price until they add the new one.", dropped)
}

// handleGrantReprice gives a grant this account made a new output price and a new secret.
// Body: {name, price_out}. The page asks first; the new secret is in this one reply.
func (s *Server) handleGrantReprice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string          `json:"name"`
		Price json.RawMessage `json:"price_out"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	price, given, err := priceTyped(req.Price)
	switch {
	case req.Name == "":
		keysSay(w, http.StatusBadRequest, "name required")
		return
	case err != nil:
		keysSay(w, http.StatusBadRequest, err.Error())
		return
	case !given:
		keysSay(w, http.StatusBadRequest, client.MsgRepriceNoPrice)
		return
	}
	if !s.keysReady(w) {
		return
	}
	res, err := client.GrantReprice(s.opts.Broker, req.Name, client.RepriceOpts{PriceOut: &price})
	if err != nil {
		grantsFail(w, err, "")
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true, "name": res.Name, "secret": res.Secret,
		"export": client.KeyExportLines(s.opts.Broker, res.Secret), "dropped": res.Dropped, "message": repriceNote(res.Dropped)})
}
