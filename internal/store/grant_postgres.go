package store

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// Postgres grant storage (GRANT-KEYS-DESIGN). Mirrors the additive style of the
// owners / earning_lots methods: JSONB for the node/model allow-lists, an indexed
// secret_hash for the hot auth lookup, and a small grant_usage rollup for caps.

// nullStr maps an empty string to a SQL NULL (so an untagged receipt's grant_id
// stays NULL rather than ""), else the value itself.
func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// jsonStrSlice marshals a string slice to JSONB; nil becomes "[]" so the column
// default holds and round-trips cleanly.
func jsonStrSlice(s []string) []byte {
	if s == nil {
		s = []string{}
	}
	b, _ := json.Marshal(s)
	return b
}

func (p *Postgres) CreateGrant(g Grant) error {
	if g.CreatedAt == 0 {
		g.CreatedAt = time.Now().Unix()
	}
	_, err := p.db.Exec(`INSERT INTO rogerai.grants
		(id,secret_hash,owner,label,nodes,models,free,price_in,price_out,rpm,burst,daily_cap,monthly_cap,self,expires_at,revoked,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		g.ID, g.SecretHash, g.Owner, g.Label, jsonStrSlice(g.Nodes), jsonStrSlice(g.Models),
		g.Free, g.PriceIn, g.PriceOut, g.RPM, g.Burst, g.DailyCap, g.MonthlyCap, g.Self,
		g.ExpiresAt, g.Revoked, g.CreatedAt)
	return err
}

// scanGrant maps one grants row into a Grant.
func (p *Postgres) scanGrant(row interface{ Scan(...any) error }) (Grant, error) {
	var g Grant
	var nodes, models []byte
	err := row.Scan(&g.ID, &g.SecretHash, &g.Owner, &g.Label, &nodes, &models,
		&g.Free, &g.PriceIn, &g.PriceOut, &g.RPM, &g.Burst, &g.DailyCap, &g.MonthlyCap,
		&g.Self, &g.ExpiresAt, &g.Revoked, &g.CreatedAt, &g.Gen)
	if err != nil {
		return Grant{}, err
	}
	_ = json.Unmarshal(nodes, &g.Nodes)
	_ = json.Unmarshal(models, &g.Models)
	return g, nil
}

const grantCols = `id,secret_hash,owner,label,nodes,models,free,price_in,price_out,rpm,burst,daily_cap,monthly_cap,self,expires_at,revoked,created_at,gen`

func (p *Postgres) GrantBySecretHash(hash string) (Grant, bool, error) {
	g, err := p.scanGrant(p.db.QueryRow(`SELECT `+grantCols+` FROM rogerai.grants WHERE secret_hash=$1`, hash))
	if err == sql.ErrNoRows {
		return Grant{}, false, nil
	}
	if err != nil {
		return Grant{}, false, err
	}
	return g, true, nil
}

func (p *Postgres) GrantsByOwner(owner string) ([]Grant, error) {
	rows, err := p.db.Query(`SELECT `+grantCols+` FROM rogerai.grants WHERE owner=$1 ORDER BY created_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Grant
	for rows.Next() {
		g, err := p.scanGrant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (p *Postgres) SetGrantRevoked(id, owner string, revoked bool) (bool, error) {
	res, err := p.db.Exec(`UPDATE rogerai.grants SET revoked=$3 WHERE id=$1 AND owner=$2`, id, owner, revoked)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (p *Postgres) UpdateGrant(id, owner string, patch GrantPatch) (Grant, bool, error) {
	// Read-modify-write inside a transaction, owner-scoped: an owner can only ever
	// edit their own grants (the WHERE owner clause is the gate).
	tx, err := p.db.Begin()
	if err != nil {
		return Grant{}, false, err
	}
	defer tx.Rollback()
	g, err := p.scanGrant(tx.QueryRow(`SELECT `+grantCols+` FROM rogerai.grants WHERE id=$1 AND owner=$2 FOR UPDATE`, id, owner))
	if err == sql.ErrNoRows {
		return Grant{}, false, nil
	}
	if err != nil {
		return Grant{}, false, err
	}
	g = g.applyPatch(patch)
	if _, err := tx.Exec(`UPDATE rogerai.grants SET
		label=$3,nodes=$4,models=$5,free=$6,price_in=$7,price_out=$8,rpm=$9,burst=$10,
		daily_cap=$11,monthly_cap=$12,expires_at=$13,revoked=$14 WHERE id=$1 AND owner=$2`,
		id, owner, g.Label, jsonStrSlice(g.Nodes), jsonStrSlice(g.Models), g.Free,
		g.PriceIn, g.PriceOut, g.RPM, g.Burst, g.DailyCap, g.MonthlyCap, g.ExpiresAt, g.Revoked); err != nil {
		return Grant{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Grant{}, false, err
	}
	return g, true, nil
}

func (p *Postgres) GrantUsageOf(id string, now time.Time) (GrantUsage, error) {
	var u GrantUsage
	row := p.db.QueryRow(`SELECT
		COALESCE((SELECT tokens FROM rogerai.grant_usage WHERE grant_id=$1 AND bucket=$2),0),
		COALESCE((SELECT tokens FROM rogerai.grant_usage WHERE grant_id=$1 AND bucket=$3),0)`,
		id, dayKey(now), monthKey(now))
	if err := row.Scan(&u.DayTokens, &u.MonthTokens); err != nil {
		return GrantUsage{}, err
	}
	return u, nil
}

func (p *Postgres) AddGrantUsage(id string, tokens int64, now time.Time) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, win := range []string{dayKey(now), monthKey(now)} {
		if _, err := tx.Exec(`INSERT INTO rogerai.grant_usage(grant_id,bucket,tokens) VALUES($1,$2,$3)
			ON CONFLICT (grant_id,bucket) DO UPDATE SET tokens=rogerai.grant_usage.tokens+$3`, id, win, tokens); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (p *Postgres) RepriceGrant(id, owner, newHash string, in, out float64, scope GrantScope) (Grant, bool, error) {
	list := func(l *[]string) any { // NULL keeps the stored list
		if l == nil {
			return nil
		}
		b, _ := json.Marshal(append([]string{}, *l...))
		return string(b)
	}
	g, err := p.scanGrant(p.db.QueryRow(`UPDATE rogerai.grants SET secret_hash=$3,price_in=$4,price_out=$5,free=$6,gen=gen+1,
		nodes=COALESCE($7::jsonb,nodes),models=COALESCE($8::jsonb,models)
		WHERE id=$1 AND owner=$2 AND NOT revoked RETURNING `+grantCols, id, owner, newHash, in, out, in == 0 && out == 0, list(scope.Nodes), list(scope.Models)))
	if err == sql.ErrNoRows {
		return Grant{}, false, nil
	}
	if err != nil {
		return Grant{}, false, err
	}
	return g, true, nil
}

// AddGrantToWallet inserts only while the wallet holds fewer than max grants, or already holds
// this one. Two adds of DIFFERENT grants racing at the limit can both see room under READ
// COMMITTED, so the limit can be passed by the racing adds; it bounds clutter, not money.
func (p *Postgres) AddGrantToWallet(wallet, grantID string, gen int64, covers map[string][]string, max int, now time.Time) error {
	if covers == nil {
		covers = map[string][]string{}
	}
	nj, _ := json.Marshal(covers)
	res, err := p.db.Exec(`INSERT INTO rogerai.grant_adds(wallet,grant_id,gen,added_at,covers)
		SELECT $1,$2,$3,$4,$6::jsonb WHERE (SELECT count(*) FROM rogerai.grant_adds WHERE wallet=$1) < $5
			OR EXISTS (SELECT 1 FROM rogerai.grant_adds WHERE wallet=$1 AND grant_id=$2)
		ON CONFLICT (wallet,grant_id) DO UPDATE SET gen=EXCLUDED.gen, covers=EXCLUDED.covers`, wallet, grantID, gen, now.Unix(), max, string(nj))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrGrantAddLimit
	}
	return nil
}

func (p *Postgres) RemoveGrantFromWallet(wallet, grantID string) (bool, error) {
	res, err := p.db.Exec(`DELETE FROM rogerai.grant_adds WHERE wallet=$1 AND grant_id=$2`, wallet, grantID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (p *Postgres) GrantAddsOfWallet(wallet string) ([]GrantAdd, error) {
	rows, err := p.db.Query(`SELECT a.gen,a.added_at,a.covers,g.`+strings.ReplaceAll(grantCols, ",", ",g.")+`
		FROM rogerai.grant_adds a JOIN rogerai.grants g ON g.id=a.grant_id
		WHERE a.wallet=$1 ORDER BY a.added_at, a.grant_id`, wallet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GrantAdd
	for rows.Next() {
		var a GrantAdd
		var covers []byte
		g, err := p.scanGrant(prefixScan{rows, []any{&a.Gen, &a.AddedAt, &covers}})
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(covers, &a.Covers)
		a.Grant = g
		out = append(out, a)
	}
	return out, rows.Err()
}

// prefixScan scans a row whose first columns are not part of the grant.
type prefixScan struct {
	row interface{ Scan(...any) error }
	pre []any
}

func (s prefixScan) Scan(dest ...any) error { return s.row.Scan(append(s.pre, dest...)...) }

func (p *Postgres) GrantAddCounts(owner string) (map[string]int, error) {
	rows, err := p.db.Query(`SELECT a.grant_id, count(*) FROM rogerai.grant_adds a
		JOIN rogerai.grants g ON g.id=a.grant_id AND g.gen=a.gen WHERE g.owner=$1 GROUP BY a.grant_id`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, rows.Err()
}
