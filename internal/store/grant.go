package store

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Grant is an owner-issued private access key: a labeled bearer credential
// (rog-grant_<secret>) a grantee sets as their API key and uses with no login,
// no account, no wallet. The broker authenticates the grant by its secret hash,
// resolves it to the issuing owner, routes only to that owner's nodes at the
// grant's price (free = 0/0), and enforces the grant's caps. See
// docs-internal/GRANT-KEYS-DESIGN.md. The secret itself is shown ONCE at create;
// only its sha256 is ever stored.
type Grant struct {
	ID         string   `json:"id"`          // "grant_<rand>" - the DB id (NOT the secret)
	SecretHash string   `json:"-"`           // sha256(secret); the secret is shown once at create
	Owner      string   `json:"owner"`       // issuing owner pubkey (store.Owner.Pubkey)
	Label      string   `json:"label"`       // "petlings", "friend-jane", "self:hermes-box"
	Nodes      []string `json:"nodes"`       // allowed node ids; empty = ALL of this owner's nodes
	Models     []string `json:"models"`      // allowed models; empty = any model the nodes offer
	Free       bool     `json:"free"`        // true => price 0/0 (skips wallet debit entirely)
	PriceIn    float64  `json:"price_in"`    // custom/discounted $/1M in (ignored when Free)
	PriceOut   float64  `json:"price_out"`   // custom/discounted $/1M out (ignored when Free)
	RPM        float64  `json:"rpm"`         // rate-limit sustained req/min (0 = broker default)
	Burst      float64  `json:"burst"`       // rate-limit bucket depth (0 = broker default)
	DailyCap   int64    `json:"daily_cap"`   // max tokens/UTC-day (0 = unlimited)
	MonthlyCap int64    `json:"monthly_cap"` // max tokens/UTC-month (0 = unlimited)
	Self       bool     `json:"self"`        // a --self grant (owner's own boxes/agents; always $0)
	ExpiresAt  int64    `json:"expires_at"`  // unix; 0 = never
	Revoked    bool     `json:"revoked"`
	CreatedAt  int64    `json:"created_at"`
	// Gen is the secret's generation. A reprice gives the grant a new secret and the next
	// generation, which ends every copy an account added before it (GrantAdd.Current).
	Gen int64 `json:"gen,omitempty"`

	// --- Edge grant fields (features/edge/control.feature) --------------------------------------
	// An Edge grant is the same record, SIGNED BY THE EDGE'S AUTHORITY so any member verifies it with
	// the public root - no broker, which is what an airgapped Edge requires. These are empty/zero on a
	// broker grant, which never signs or verifies them.
	//
	// Caller is the node id (or node/instance) allowed to INVOKE under this grant; empty = any of this
	// owner's own instances. Caps are the capabilities it may exercise on the callee. MaxHops bounds
	// onward delegation (0 = no onward delegation). Sig is the authority's signature over the grant's
	// canonical bytes; it is NOT part of the canonical form.
	Caller  string   `json:"caller,omitempty"`
	Caps    []string `json:"caps,omitempty"`
	MaxHops int      `json:"max_hops,omitempty"`
	Sig     string   `json:"sig,omitempty"`
}

// GrantUsage is a per-grant rollup row used by the daily/monthly cap check and the
// dashboard. Tokens are prompt+completion summed for the UTC window.
type GrantUsage struct {
	DayTokens   int64 `json:"day_tokens"`   // tokens served in the current UTC day
	MonthTokens int64 `json:"month_tokens"` // tokens served in the current UTC month
}

// GrantPatch is the set of editable grant fields (PATCH /grants/{id}). A nil
// pointer field means "leave unchanged"; this lets an owner toggle revoked or
// adjust caps/price/scope without resending the whole grant.
type GrantPatch struct {
	Label      *string   `json:"label,omitempty"`
	Nodes      *[]string `json:"nodes,omitempty"`
	Models     *[]string `json:"models,omitempty"`
	Free       *bool     `json:"free,omitempty"`
	PriceIn    *float64  `json:"price_in,omitempty"`
	PriceOut   *float64  `json:"price_out,omitempty"`
	RPM        *float64  `json:"rpm,omitempty"`
	Burst      *float64  `json:"burst,omitempty"`
	DailyCap   *int64    `json:"daily_cap,omitempty"`
	MonthlyCap *int64    `json:"monthly_cap,omitempty"`
	ExpiresAt  *int64    `json:"expires_at,omitempty"`
	Revoked    *bool     `json:"revoked,omitempty"`
}

// Expired reports whether the grant has passed its expiry (0 = never).
func (g Grant) Expired(now time.Time) bool {
	return g.ExpiresAt != 0 && now.Unix() >= g.ExpiresAt
}

// GrantPrice returns the price the grant bills at: 0/0 for a free or self grant,
// else its custom (PriceIn, PriceOut). A negative stored price is clamped to 0 here -
// the billing chokepoint every settle path reads - so even a legacy/corrupt negative
// row can never yield a negative cost (which Finalize would turn into a minted credit).
// The HTTP create/edit paths reject a negative price outright; this is defense in depth.
func (g Grant) GrantPrice() (in, out float64) {
	if g.Free || g.Self {
		return 0, 0
	}
	in, out = g.PriceIn, g.PriceOut
	if in < 0 {
		in = 0
	}
	if out < 0 {
		out = 0
	}
	return in, out
}

// GrantAdd is one account's copy of a priced grant: the caller added the grant's secret to
// their account once, and after that their own login and account keys get the grant's price
// on the owner's stations (features/grants/priced_grants.feature). It holds no secret.
type GrantAdd struct {
	Grant   Grant
	Gen     int64 // the grant's generation when it was added
	AddedAt int64
	// Covers is what the add reply showed the caller: station id -> the models priced there.
	// The copy gives its price on these pairs only, so a station or model the owner brings on
	// air later is never charged at a price the caller was not shown (P17, P18). Adding the
	// grant again refreshes it.
	Covers map[string][]string
}

// GrantScope is a reprice's new station and model lists. A nil field keeps the grant's own;
// an empty list means every station or model.
type GrantScope struct {
	Nodes, Models *[]string
}

// Current reports whether the copy was added under the grant's present secret. A copy from
// before a reprice is not: it gives no price until the caller adds the new secret.
func (a GrantAdd) Current() bool { return a.Gen == a.Grant.Gen }

// ErrGrantAddLimit: the wallet already holds the most grants an account may add.
var ErrGrantAddLimit = errors.New("store: account grant limit")

// applyPatch returns g with the non-nil patch fields applied.
func (g Grant) applyPatch(p GrantPatch) Grant {
	if p.Label != nil {
		g.Label = *p.Label
	}
	if p.Nodes != nil {
		g.Nodes = *p.Nodes
	}
	if p.Models != nil {
		g.Models = *p.Models
	}
	if p.Free != nil {
		g.Free = *p.Free
	}
	if p.PriceIn != nil {
		g.PriceIn = *p.PriceIn
	}
	if p.PriceOut != nil {
		g.PriceOut = *p.PriceOut
	}
	if p.RPM != nil {
		g.RPM = *p.RPM
	}
	if p.Burst != nil {
		g.Burst = *p.Burst
	}
	if p.DailyCap != nil {
		g.DailyCap = *p.DailyCap
	}
	if p.MonthlyCap != nil {
		g.MonthlyCap = *p.MonthlyCap
	}
	if p.ExpiresAt != nil {
		g.ExpiresAt = *p.ExpiresAt
	}
	if p.Revoked != nil {
		g.Revoked = *p.Revoked
	}
	return g
}

// dayKey / monthKey are the UTC-window keys for the grant usage rollup.
func dayKey(t time.Time) string   { return t.UTC().Format("2006-01-02") }
func monthKey(t time.Time) string { return t.UTC().Format("2006-01") }

// --- Mem grant storage ---------------------------------------------------
//
// A second small map set on Mem, mirroring owners/nodeAcct. Guarded by its own
// mutex so grant ops never contend with the wallet/ledger lock.

type grantStore struct {
	mu       sync.Mutex
	grants   map[string]Grant               // id -> grant
	bySecret map[string]string              // secret_hash -> id
	dayUsage map[string]int64               // "id|YYYY-MM-DD" -> tokens
	monUsage map[string]int64               // "id|YYYY-MM" -> tokens
	adds     map[string]map[string]GrantAdd // wallet -> grant id -> the copy (Grant unset)
}

func newGrantStore() *grantStore {
	return &grantStore{
		grants: map[string]Grant{}, bySecret: map[string]string{},
		dayUsage: map[string]int64{}, monUsage: map[string]int64{},
		adds: map[string]map[string]GrantAdd{},
	}
}

func (m *Mem) CreateGrant(g Grant) error {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	if g.CreatedAt == 0 {
		g.CreatedAt = time.Now().Unix()
	}
	m.gs.grants[g.ID] = g
	m.gs.bySecret[g.SecretHash] = g.ID
	return nil
}

func (m *Mem) GrantBySecretHash(hash string) (Grant, bool, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	id, ok := m.gs.bySecret[hash]
	if !ok {
		return Grant{}, false, nil
	}
	g, ok := m.gs.grants[id]
	return g, ok, nil
}

func (m *Mem) GrantsByOwner(owner string) ([]Grant, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	var out []Grant
	for _, g := range m.gs.grants {
		if g.Owner == owner {
			out = append(out, g)
		}
	}
	return out, nil
}

func (m *Mem) SetGrantRevoked(id, owner string, revoked bool) (bool, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	g, ok := m.gs.grants[id]
	if !ok || g.Owner != owner { // owner-scoped: never touch another owner's grant
		return false, nil
	}
	g.Revoked = revoked
	m.gs.grants[id] = g
	return true, nil
}

func (m *Mem) UpdateGrant(id, owner string, patch GrantPatch) (Grant, bool, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	g, ok := m.gs.grants[id]
	if !ok || g.Owner != owner {
		return Grant{}, false, nil
	}
	g = g.applyPatch(patch)
	m.gs.grants[id] = g
	return g, true, nil
}

func (m *Mem) GrantUsageOf(id string, now time.Time) (GrantUsage, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	return GrantUsage{
		DayTokens:   m.gs.dayUsage[id+"|"+dayKey(now)],
		MonthTokens: m.gs.monUsage[id+"|"+monthKey(now)],
	}, nil
}

func (m *Mem) AddGrantUsage(id string, tokens int64, now time.Time) error {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	m.gs.dayUsage[id+"|"+dayKey(now)] += tokens
	m.gs.monUsage[id+"|"+monthKey(now)] += tokens
	return nil
}

func (m *Mem) RepriceGrant(id, owner, newHash string, in, out float64, scope GrantScope) (Grant, bool, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	g, ok := m.gs.grants[id]
	if !ok || g.Owner != owner || g.Revoked { // a revoked grant is never given a new secret
		return Grant{}, false, nil
	}
	delete(m.gs.bySecret, g.SecretHash)
	g.SecretHash, g.PriceIn, g.PriceOut, g.Free, g.Gen = newHash, in, out, in == 0 && out == 0, g.Gen+1
	if scope.Nodes != nil {
		g.Nodes = append([]string(nil), *scope.Nodes...)
	}
	if scope.Models != nil {
		g.Models = append([]string(nil), *scope.Models...)
	}
	m.gs.grants[id] = g
	m.gs.bySecret[newHash] = id
	return g, true, nil
}

func (m *Mem) AddGrantToWallet(wallet, grantID string, gen int64, covers map[string][]string, max int, now time.Time) error {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	held := m.gs.adds[wallet]
	cp := make(map[string][]string, len(covers))
	for k, v := range covers {
		cp[k] = append([]string(nil), v...)
	}
	covers = cp
	if cur, ok := held[grantID]; ok {
		cur.Gen, cur.Covers = gen, covers
		held[grantID] = cur
		return nil
	}
	if len(held) >= max {
		return ErrGrantAddLimit
	}
	if held == nil {
		held = map[string]GrantAdd{}
		m.gs.adds[wallet] = held
	}
	held[grantID] = GrantAdd{Gen: gen, AddedAt: now.Unix(), Covers: covers}
	return nil
}

func (m *Mem) RemoveGrantFromWallet(wallet, grantID string) (bool, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	_, ok := m.gs.adds[wallet][grantID]
	delete(m.gs.adds[wallet], grantID)
	return ok, nil
}

func (m *Mem) GrantAddsOfWallet(wallet string) ([]GrantAdd, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	var out []GrantAdd
	for id, a := range m.gs.adds[wallet] {
		if g, ok := m.gs.grants[id]; ok {
			a.Grant = g
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AddedAt != out[j].AddedAt {
			return out[i].AddedAt < out[j].AddedAt
		}
		return out[i].Grant.ID < out[j].Grant.ID
	})
	return out, nil
}

func (m *Mem) GrantAddCounts(owner string) (map[string]int, error) {
	m.gs.mu.Lock()
	defer m.gs.mu.Unlock()
	out := map[string]int{}
	for _, held := range m.gs.adds {
		for id, a := range held {
			if g, ok := m.gs.grants[id]; ok && g.Owner == owner && a.Gen == g.Gen {
				out[id]++
			}
		}
	}
	return out, nil
}
