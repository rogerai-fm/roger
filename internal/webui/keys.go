package webui

// keys.go: account keys (rog-key_...) on the ACCOUNT tab, plus grant revoke
// (features/console/account_keys.feature). The console signs the same /account/keys and
// /grants calls the CLI makes. A minted secret passes through one reply and is kept nowhere.

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"rogerai.fm/roger/v7/internal/client"
)

const (
	keysUnreachable = "The broker could not be reached."
	keysLogin       = "log in to manage keys"
)

func keysSay(w http.ResponseWriter, code int, msg string) {
	writeChannel(w, code, map[string]any{"message": msg})
}

// keysBody reads a key route's body into v.
func keysBody(w http.ResponseWriter, r *http.Request, v any) bool {
	code, msg := strictBody(w, r, v, "bad request")
	if code != 0 {
		keysSay(w, code, msg)
	}
	return code == 0
}

// strictBody reads a small JSON body into v. It returns 0, or the status and message that
// refuse a body that is oversized, malformed (in the caller's word for it) or carries a field
// v does not name.
func strictBody(w http.ResponseWriter, r *http.Request, v any, malformed string) (int, string) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, channelBodyCap))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err == nil {
		return 0, ""
	}
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		return http.StatusRequestEntityTooLarge, "request too large"
	}
	// The decoder names an unknown field; nothing else in its words is worth showing.
	if f, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		return http.StatusBadRequest, "unknown field " + f
	}
	return http.StatusBadRequest, malformed
}

// keysReady refuses when the console has no account to ask about: LAN mode, no broker, or
// logged out. The broker's own logged-out wording names a terminal command, so it is not used.
func (s *Server) keysReady(w http.ResponseWriter) bool {
	switch {
	case s.opts.LANDial != nil:
		keysSay(w, http.StatusConflict, "Keys need the public network. Not available in LAN mode.")
	case s.opts.Broker == "":
		keysSay(w, http.StatusServiceUnavailable, "no broker configured")
	default:
		bal, err := client.FetchBalance(s.opts.Broker, s.opts.User)
		if err != nil {
			keysSay(w, http.StatusServiceUnavailable, keysUnreachable)
		} else if !bal.LoggedIn {
			keysSay(w, http.StatusUnauthorized, keysLogin)
		}
		return err == nil && bal.LoggedIn
	}
	return false
}

// keysFail reports a broker-side refusal in the broker's own words, except its logged-out
// wording, which names a terminal command.
func keysFail(w http.ResponseWriter, err error) {
	if errors.Is(err, client.ErrKeysLogin) {
		keysSay(w, http.StatusUnauthorized, keysLogin)
		return
	}
	keysSay(w, http.StatusBadGateway, err.Error())
}

// keyLimits reads a limit and a window, either of which may be absent. limit_usd is a number
// or the text the operator typed; both are read as `roger keys --limit` reads them, so the
// page has no rules of its own. none is how this route says to ask for unlimited. msg is the
// refusal, or "".
func keyLimits(raw json.RawMessage, reset *string, none string) (limit *float64, msg string) {
	if len(raw) != 0 && string(raw) != "null" {
		text := string(raw)
		_ = json.Unmarshal(raw, &text) // typed text arrives quoted; a number is read as written
		v, err := client.ParseKeyLimit(text)
		switch {
		case errors.Is(err, client.ErrKeyLimit):
			return nil, err.Error() + none
		case err != nil:
			return nil, err.Error()
		}
		limit = &v
	}
	if reset != nil && !client.KeyResetOK(*reset) {
		return nil, "reset is daily, weekly, monthly or none"
	}
	return limit, ""
}

// keyRow is a key as the page draws it: the broker's fields plus the words for its limit, its
// spend and its state, written once in the client so the page, the TUI and the CLI agree.
type keyRow struct {
	client.AccountKey
	LimitText string `json:"limit_text"`
	UsedText  string `json:"used_text"`
	StateText string `json:"state"`
}

func rowOf(k client.AccountKey) keyRow {
	return keyRow{k, k.LimitText(), k.UsedText(), k.State()}
}

// handleKeys lists the account's keys (GET). Never a secret: the broker keeps none to give.
func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.keysReady(w) {
		return
	}
	keys, err := client.KeysList(s.opts.Broker)
	if err != nil {
		keysFail(w, err)
		return
	}
	rows := make([]keyRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, rowOf(k))
	}
	writeChannel(w, http.StatusOK, map[string]any{"keys": rows})
}

// handleKeyMint mints a key. Body: {name, limit_usd, reset?}. The limit must be stated;
// "none" is unlimited and 0 is refused, as on the CLI. The secret is in this one reply and nowhere else.
func (s *Server) handleKeyMint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string          `json:"name"`
		Limit json.RawMessage `json:"limit_usd"`
		Reset *string         `json:"reset"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Reset == nil || *req.Reset == "" {
		none := "none"
		req.Reset = &none
	}
	limit, msg := keyLimits(req.Limit, req.Reset, `, or tick "unlimited"`)
	switch {
	case req.Name == "":
		msg = "name required"
	case limit == nil && msg == "":
		msg = "limit_usd required"
	}
	if msg != "" {
		keysSay(w, http.StatusBadRequest, msg)
		return
	}
	if !s.keysReady(w) {
		return
	}
	k, secret, err := client.KeysMint(s.opts.Broker, map[string]any{"name": req.Name, "limit_usd": *limit, "reset": *req.Reset})
	if err != nil {
		keysFail(w, err)
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true, "key": rowOf(k), "secret": secret, "export": client.KeyExportLines(s.opts.Broker, secret)})
}

// handleKeySet changes a key. Body: {id, limit_usd?, reset?, disabled?}; only the fields sent
// are changed.
func (s *Server) handleKeySet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string          `json:"id"`
		Limit    json.RawMessage `json:"limit_usd"`
		Reset    *string         `json:"reset"`
		Disabled *bool           `json:"disabled"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	limit, msg := keyLimits(req.Limit, req.Reset, `, or "none" for unlimited`)
	fields := map[string]any{}
	if limit != nil {
		fields["limit_usd"] = *limit
	}
	if req.Reset != nil {
		fields["reset"] = *req.Reset
	}
	if req.Disabled != nil {
		fields["disabled"] = *req.Disabled
	}
	switch {
	case req.ID == "":
		msg = "id required"
	case !client.KeyIDOK(req.ID):
		msg = "bad key id"
	case len(fields) == 0 && msg == "":
		msg = "nothing to change"
	}
	if msg != "" {
		keysSay(w, http.StatusBadRequest, msg)
		return
	}
	if !s.keysReady(w) {
		return
	}
	k, err := client.KeysSet(s.opts.Broker, req.ID, fields)
	if err != nil {
		keysFail(w, err)
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true, "key": rowOf(k)})
}

// handleKeyRevoke revokes a key. Body: {id}.
func (s *Server) handleKeyRevoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	switch {
	case req.ID == "":
		keysSay(w, http.StatusBadRequest, "id required")
		return
	case !client.KeyIDOK(req.ID):
		keysSay(w, http.StatusBadRequest, "bad key id")
		return
	}
	if !s.keysReady(w) {
		return
	}
	if err := client.KeysRevoke(s.opts.Broker, req.ID); err != nil {
		keysFail(w, err)
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true})
}

// handleGrants lists the account's grants (GET) or makes a free one (POST {name}), whose
// secret is in that one reply and nowhere else. Guarded like the key routes (ruling C10). A
// price_out, as typed, makes it priced (features/grants/grant_surfaces.feature, ruling S8).
func (s *Server) handleGrants(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if !s.keysReady(w) {
			return
		}
		rows, err := client.GrantListRows(s.opts.Broker)
		if err != nil {
			keysFail(w, err)
			return
		}
		writeJSON(w, rows)
	case http.MethodPost:
		var req struct {
			Name  string          `json:"name"`
			Price json.RawMessage `json:"price_out"`
		}
		if !keysBody(w, r, &req) {
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			keysSay(w, http.StatusBadRequest, "name required")
			return
		}
		// A price is read with the CLI's reader, so a price `roger grant create --price-out`
		// refuses is refused here in the same words (ruling S8). None, or 0, is free.
		price, _, err := priceTyped(req.Price)
		if err != nil {
			keysSay(w, http.StatusBadRequest, err.Error())
			return
		}
		if !s.keysReady(w) {
			return
		}
		var secret string
		if price > 0 {
			_, secret, err = client.GrantCreatePriced(s.opts.Broker, name, price)
		} else {
			secret, err = client.GrantCreateSecret(s.opts.Broker, name, true)
		}
		if err != nil {
			keysFail(w, err)
			return
		}
		writeChannel(w, http.StatusOK, map[string]any{"ok": true, "secret": secret, "export": client.KeyExportLines(s.opts.Broker, secret)})
	default:
		keysSay(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleGrantRevoke revokes a grant by name. Body: {name}.
func (s *Server) handleGrantRevoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if !keysBody(w, r, &req) {
		return
	}
	if req.Name == "" {
		keysSay(w, http.StatusBadRequest, "name required")
		return
	}
	if !s.keysReady(w) {
		return
	}
	if err := client.GrantRevokeByName(s.opts.Broker, req.Name); err != nil {
		keysFail(w, err)
		return
	}
	writeChannel(w, http.StatusOK, map[string]any{"ok": true})
}
