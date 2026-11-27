package client

// grant_adds.go: the caller side of a priced grant (features/grants/priced_grants.feature) and
// the owner's reprice, as data for the CLI, the TUI and the console
// (features/grants/grant_surfaces.feature). Every call is signed by the local user key; the
// broker routes are POST/GET /account/grants, DELETE /account/grants/{id} and
// POST /grants/{id}/reprice. Nothing here prints, and no secret is ever returned except the
// new one a reprice issues.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
)

// AddedGrantStation is one station a copy of a grant covers, with its public price beside it.
type AddedGrantStation struct {
	ID          string   `json:"id"`
	Models      []string `json:"models"`
	PublicIn    float64  `json:"public_in"`
	PublicOut   float64  `json:"public_out"`
	AbovePublic bool     `json:"above_public"`
}

// AddedGrant is a grant this account added, as the broker shows it: never the secret.
type AddedGrant struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Price    string              `json:"price"`
	PriceIn  float64             `json:"price_in"`
	PriceOut float64             `json:"price_out"`
	Models   []string            `json:"models"`
	Stations []AddedGrantStation `json:"stations"`
	State    string              `json:"state"`    // active, revoked, expired, replaced (list only)
	AddedAt  int64               `json:"added_at"` // unix seconds (list only)
}

// StateWords is a grant's state as a person reads it (ruling S9).
func (g AddedGrant) StateWords() string {
	if g.State == "replaced" {
		return "replaced - ask the owner for the new secret"
	}
	return g.State
}

// MsgLoggedOut is the CLI's refusal when this machine is not signed in (ruling S11). The TUI
// and the console say the same in their own words.
const MsgLoggedOut = "log in first: roger login"

// ErrNotSignedIn is returned, before anything is sent, when this machine has no sign-in.
var ErrNotSignedIn = errors.New(MsgLoggedOut)

// ErrGrantAddLimit marks the broker's "an account can hold N grants" refusal, so each surface
// can name its own remove action (ruling S10). Its text is the broker's.
var ErrGrantAddLimit = errors.New("grant add limit")

type grantAddLimitErr struct{ msg string }

func (e grantAddLimitErr) Error() string        { return e.msg }
func (e grantAddLimitErr) Is(target error) bool { return target == ErrGrantAddLimit }

// signedIn is SignedIn, replaceable in tests of this package.
var signedIn = func() bool { _, ok := SignedIn(); return ok }

// brokerError reads the broker's error text from a reply body (either error shape).
func brokerError(status int, raw []byte) error {
	var e struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &e) == nil && len(e.Error) > 0 {
		var s string
		if json.Unmarshal(e.Error, &s) == nil && s != "" {
			return errors.New(s)
		}
		var m struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(e.Error, &m) == nil && m.Message != "" {
			return errors.New(m.Message)
		}
	}
	return fmt.Errorf("broker returned status %d", status)
}

func grantAddsDo(method, broker, path string, payload any) (int, []byte, error) {
	if !signedIn() {
		return 0, nil, ErrNotSignedIn
	}
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}
	resp, err := signedDo(method, broker, path, body)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrBrokerUnreachable, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, raw, nil
}

// GrantAdd adds a priced grant to this account by its secret and returns what the broker
// showed: its price, stations and models (ruling S4). The secret is sent once and never kept.
func GrantAdd(broker, secret string) (AddedGrant, error) {
	st, raw, err := grantAddsDo(http.MethodPost, broker, "/account/grants", map[string]string{"secret": strings.TrimSpace(secret)})
	if err != nil {
		return AddedGrant{}, err
	}
	if st != http.StatusOK {
		e := brokerError(st, raw)
		if st == http.StatusConflict && strings.Contains(e.Error(), "grants - remove") {
			return AddedGrant{}, grantAddLimitErr{e.Error()}
		}
		return AddedGrant{}, e
	}
	var out struct {
		Grant AddedGrant `json:"grant"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return AddedGrant{}, fmt.Errorf("could not read the reply: %v", err)
	}
	out.Grant.clean()
	return out.Grant, nil
}

// cleanGrantText drops control and format characters from text another account chose.
func cleanGrantText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}

// clean makes an added grant safe to print: its name, stations and models come from the
// grant's owner, not from this account.
func (g *AddedGrant) clean() {
	g.Name = cleanGrantText(g.Name)
	for i := range g.Models {
		g.Models[i] = cleanGrantText(g.Models[i])
	}
	for i := range g.Stations {
		g.Stations[i].ID = cleanGrantText(g.Stations[i].ID)
		for j := range g.Stations[i].Models {
			g.Stations[i].Models[j] = cleanGrantText(g.Stations[i].Models[j])
		}
	}
}

// GrantAddedList is the grants this account added, oldest first, and how many it may hold.
func GrantAddedList(broker string) ([]AddedGrant, int, error) {
	st, raw, err := grantAddsDo(http.MethodGet, broker, "/account/grants", nil)
	if err != nil {
		return nil, 0, err
	}
	if st != http.StatusOK {
		return nil, 0, brokerError(st, raw)
	}
	var out struct {
		Grants []AddedGrant `json:"grants"`
		Max    int          `json:"max"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, 0, fmt.Errorf("could not read the reply: %v", err)
	}
	for i := range out.Grants {
		out.Grants[i].clean()
	}
	return out.Grants, out.Max, nil
}

// GrantRemove drops one grant this account added, named by its name or id, from this account
// only (ruling S2). A name matching a grant this account MADE, and none it added, is refused
// naming `revoke`. A name two added grants share is refused listing their ids.
func GrantRemove(broker, ref string) (AddedGrant, error) {
	added, _, err := GrantAddedList(broker)
	if err != nil {
		return AddedGrant{}, err
	}
	var hits []AddedGrant
	for _, g := range added {
		if g.ID == ref {
			hits = []AddedGrant{g}
			break
		}
		if g.Name == ref {
			hits = append(hits, g)
		}
	}
	switch len(hits) {
	case 0:
		if made, _ := fetchGrants(broker); hasGrantNamed(made, ref) {
			return AddedGrant{}, fmt.Errorf("%q is a grant you made - `roger grant revoke %s` ends it for everyone", ref, ref)
		}
		return AddedGrant{}, fmt.Errorf("no grant named %q in your added grants", ref)
	case 1:
	default:
		ids := make([]string, len(hits))
		for i, h := range hits {
			ids[i] = h.ID
		}
		return AddedGrant{}, fmt.Errorf("%d added grants are named %q - remove one by id: %s", len(hits), ref, strings.Join(ids, ", "))
	}
	st, raw, err := grantAddsDo(http.MethodDelete, broker, "/account/grants/"+hits[0].ID, nil)
	if err != nil {
		return AddedGrant{}, err
	}
	if st != http.StatusOK {
		return AddedGrant{}, brokerError(st, raw)
	}
	return hits[0], nil
}

func hasGrantNamed(gs []grantJSON, ref string) bool {
	for _, g := range gs {
		if (g.Name == ref || g.ID == ref) && g.Status != "revoked" {
			return true
		}
	}
	return false
}

// GrantAddedOnlyByMe reports that ref names a grant this account added and none it made, so
// `roger grant revoke` can name `remove` instead (ruling S2).
func GrantAddedOnlyByMe(broker, ref string) bool {
	if made, err := fetchGrants(broker); err != nil || hasGrantNamed(made, ref) {
		return false
	}
	added, _, err := GrantAddedList(broker)
	if err != nil {
		return false
	}
	for _, g := range added {
		if g.Name == ref || g.ID == ref {
			return true
		}
	}
	return false
}

// RepriceOpts is a reprice: at least one price (nil is "not stated"), and optional new station
// and model lists (nil keeps the grant's own).
type RepriceOpts struct {
	PriceIn, PriceOut *float64
	Nodes, Models     *[]string
}

// RepriceResult is what a reprice did: the new secret (shown once), and how many accounts had
// added the old one and are now on the public price until they add the new one (ruling S5).
type RepriceResult struct {
	ID, Name string
	Secret   string
	Dropped  int
}

// ErrGrantCut refuses a reprice that only lowers the price: a cut keeps the same secret.
var ErrGrantCut = errors.New("a price cut keeps the same secret, so it is not a reprice")

// MsgRepriceNoPrice is the refusal for a reprice with no price, before anything is sent.
const MsgRepriceNoPrice = `state a price in dollars per 1M tokens, or "free"`

// GrantCurrentPrice is a grant this account made, by name or id: its id, name and prices.
func GrantCurrentPrice(broker, ref string) (id, name string, in, out float64, err error) {
	g, err := ownGrantRef(broker, ref)
	if err != nil {
		return "", "", 0, 0, err
	}
	return g.ID, g.Name, g.PriceIn, g.PriceOut, nil
}

func ownGrantRef(broker, ref string) (grantJSON, error) {
	gs, err := fetchGrants(broker)
	if err != nil {
		return grantJSON{}, err
	}
	var g grantJSON
	found := false
	for _, c := range gs {
		if c.ID == ref {
			return c, nil
		}
		if c.Name == ref && (!found || g.Status == "revoked") {
			g, found = c, true
		}
	}
	if !found {
		return grantJSON{}, fmt.Errorf("no grant named %q (run `roger grant list`)", ref)
	}
	return g, nil
}

// GrantReprice gives a grant this account made a new price, and a new secret that ends the old
// one. It does not ask: the surface asks first (ruling S5).
func GrantReprice(broker, ref string, o RepriceOpts) (RepriceResult, error) {
	if o.PriceIn == nil && o.PriceOut == nil {
		return RepriceResult{}, errors.New(MsgRepriceNoPrice)
	}
	if !signedIn() {
		return RepriceResult{}, ErrNotSignedIn
	}
	g, err := ownGrantRef(broker, ref)
	if err != nil {
		return RepriceResult{}, err
	}
	// A price not stated keeps its current value: the broker reads a missing price as 0.
	in, out := g.PriceIn, g.PriceOut
	if o.PriceIn != nil {
		in = *o.PriceIn
	}
	if o.PriceOut != nil {
		out = *o.PriceOut
	}
	// A cut with no new scope keeps the secret, so it is an edit, never a reprice that would
	// drop everyone who added the grant (priced_grants.feature P5).
	if o.Nodes == nil && o.Models == nil && in <= g.PriceIn && out <= g.PriceOut && (in < g.PriceIn || out < g.PriceOut) {
		return RepriceResult{}, ErrGrantCut
	}
	body := map[string]any{"price_in": in, "price_out": out}
	if o.Nodes != nil {
		body["nodes"] = *o.Nodes
	}
	if o.Models != nil {
		body["models"] = *o.Models
	}
	st, raw, err := grantAddsDo(http.MethodPost, broker, "/grants/"+g.ID+"/reprice", body)
	if err != nil {
		return RepriceResult{}, err
	}
	if st != http.StatusOK {
		return RepriceResult{}, brokerError(st, raw)
	}
	var rep struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(raw, &rep); err != nil || rep.Secret == "" {
		return RepriceResult{}, fmt.Errorf("could not read the new secret from the reply")
	}
	return RepriceResult{ID: g.ID, Name: g.Name, Secret: rep.Secret, Dropped: g.AddedBy}, nil
}

// MsgNothingToChange refuses a `grant set` that names no change, before anything is sent.
const MsgNothingToChange = "nothing to change"

// GrantEdited is a grant after an edit that kept its secret.
type GrantEdited struct {
	Name              string
	PriceIn, PriceOut float64
	Models            []string
}

// GrantSet edits a grant this account made and keeps its secret (PATCH /grants/{id}): a price
// cut and a narrowing apply at once; the broker refuses a raise or a widening of a priced
// grant, naming reprice (features/grants/grant_surfaces.feature S12). Only what is stated is
// sent; the broker keeps the rest.
func GrantSet(broker, ref string, o RepriceOpts) (GrantEdited, error) {
	body := map[string]any{}
	if o.PriceIn != nil {
		body["price_in"] = *o.PriceIn
	}
	if o.PriceOut != nil {
		body["price_out"] = *o.PriceOut
	}
	if o.Nodes != nil {
		body["nodes"] = *o.Nodes
	}
	if o.Models != nil {
		body["models"] = *o.Models
	}
	if len(body) == 0 {
		return GrantEdited{}, errors.New(MsgNothingToChange)
	}
	if !signedIn() {
		return GrantEdited{}, ErrNotSignedIn
	}
	g, err := ownGrantRef(broker, ref)
	if err != nil {
		return GrantEdited{}, err
	}
	st, raw, err := grantAddsDo(http.MethodPatch, broker, "/grants/"+g.ID, body)
	if err != nil {
		return GrantEdited{}, err
	}
	if st != http.StatusOK {
		return GrantEdited{}, brokerError(st, raw)
	}
	var rep struct {
		Grant grantJSON `json:"grant"`
	}
	if err := json.Unmarshal(raw, &rep); err != nil {
		return GrantEdited{}, fmt.Errorf("could not read the reply: %v", err)
	}
	return GrantEdited{Name: rep.Grant.Name, PriceIn: rep.Grant.PriceIn, PriceOut: rep.Grant.PriceOut, Models: rep.Grant.Models}, nil
}
