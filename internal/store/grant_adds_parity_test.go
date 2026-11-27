package store

// grant_adds_parity_test.go: an account's added grants (features/grants/priced_grants.feature)
// on BOTH backends. A caller adds a priced grant to their account once; a reprice gives the
// grant a new secret and a new generation, which ends every copy added before it.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGrantAddsParity(t *testing.T) {
	for name, s := range parityStores(t) {
		t.Run(name, func(t *testing.T) {
			sfx := time.Now().Format("150405.000000")
			owner, other := "own-"+sfx, "oth-"+sfx
			a, b := "u_adds_a_"+sfx, "u_adds_b_"+sfx
			g1 := Grant{ID: "grant_a1_" + sfx, SecretHash: "h1-" + sfx, Owner: owner, Label: "friends", PriceOut: 0.05,
				Models: []string{"m1"}, DailyCap: 9, CreatedAt: 1}
			g2 := Grant{ID: "grant_a2_" + sfx, SecretHash: "h2-" + sfx, Owner: owner, Label: "club", PriceOut: 0.03, CreatedAt: 2}
			require.NoError(t, s.CreateGrant(g1))
			require.NoError(t, s.CreateGrant(g2))
			now := time.Unix(1_000, 0)

			adds, err := s.GrantAddsOfWallet(a)
			require.NoError(t, err)
			require.Empty(t, adds)

			require.NoError(t, s.AddGrantToWallet(a, g1.ID, 0, map[string][]string{"n1": {"m1"}}, 2, now))
			require.NoError(t, s.AddGrantToWallet(a, g1.ID, 0, map[string][]string{"n1": {"m1"}, "n2": {"m1"}}, 2, now.Add(time.Hour)), "adding the same grant twice keeps one copy")
			require.NoError(t, s.AddGrantToWallet(a, g2.ID, 0, nil, 2, now.Add(time.Minute)))
			require.NoError(t, s.AddGrantToWallet(b, g1.ID, 0, map[string][]string{"n1": {"m1"}}, 2, now))

			adds, err = s.GrantAddsOfWallet(a)
			require.NoError(t, err)
			require.Len(t, adds, 2)
			require.Equal(t, g1.ID, adds[0].Grant.ID, "oldest first")
			require.Equal(t, int64(1_000), adds[0].AddedAt, "the first add's time is kept")
			require.Equal(t, []string{"m1"}, adds[0].Grant.Models)
			require.Equal(t, map[string][]string{"n1": {"m1"}, "n2": {"m1"}}, adds[0].Covers, "adding again refreshes what the copy covers (P17, P18)")
			require.True(t, adds[0].Current())
			require.Equal(t, g2.ID, adds[1].Grant.ID)
			require.Empty(t, adds[1].Covers, "a copy added while the owner had no station covers none")

			// The limit counts what the wallet holds. A grant it already holds is never refused.
			g3 := Grant{ID: "grant_a3_" + sfx, SecretHash: "h3-" + sfx, Owner: owner, Label: "third", PriceOut: 1, CreatedAt: 3}
			require.NoError(t, s.CreateGrant(g3))
			require.ErrorIs(t, s.AddGrantToWallet(a, g3.ID, 0, nil, 2, now), ErrGrantAddLimit)
			require.NoError(t, s.AddGrantToWallet(a, g2.ID, 0, nil, 2, now))
			adds, _ = s.GrantAddsOfWallet(a)
			require.Len(t, adds, 2)

			counts, err := s.GrantAddCounts(owner)
			require.NoError(t, err)
			require.Equal(t, 2, counts[g1.ID])
			require.Equal(t, 1, counts[g2.ID])
			require.Zero(t, counts[g3.ID])
			counts, err = s.GrantAddCounts(other)
			require.NoError(t, err)
			require.Empty(t, counts)

			// Reprice: owner-scoped, a new hash, a new generation, everything else kept.
			_, ok, err := s.RepriceGrant(g1.ID, other, "h1b-"+sfx, 0, 0.08, GrantScope{})
			require.NoError(t, err)
			require.False(t, ok, "another owner cannot reprice it")
			_, ok, err = s.RepriceGrant("grant_missing_"+sfx, owner, "hx-"+sfx, 0, 0.08, GrantScope{})
			require.NoError(t, err)
			require.False(t, ok)
			require.NoError(t, s.AddGrantUsage(g1.ID, 7, now))
			wider := []string{"n1", "n9"}
			got, ok, err := s.RepriceGrant(g1.ID, owner, "h1b-"+sfx, 0.01, 0.08, GrantScope{Nodes: &wider})
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, int64(1), got.Gen)
			require.Equal(t, 0.08, got.PriceOut)
			require.Equal(t, 0.01, got.PriceIn)
			require.False(t, got.Free)
			require.Equal(t, "friends", got.Label)
			require.Equal(t, int64(9), got.DailyCap)
			require.Equal(t, []string{"m1"}, got.Models, "a scope the reprice does not name is kept")
			require.Equal(t, wider, got.Nodes, "a reprice may widen the stations (P17)")
			_, found, _ := s.GrantBySecretHash("h1-" + sfx)
			require.False(t, found, "the old secret is gone")
			byNew, found, _ := s.GrantBySecretHash("h1b-" + sfx)
			require.True(t, found)
			require.Equal(t, int64(1), byNew.Gen)
			u, _ := s.GrantUsageOf(g1.ID, now)
			require.Equal(t, int64(7), u.DayTokens, "usage survives a reprice")

			adds, _ = s.GrantAddsOfWallet(a)
			require.Len(t, adds, 2)
			require.False(t, adds[0].Current(), "a copy added before the reprice is no longer current")
			counts, _ = s.GrantAddCounts(owner)
			require.Zero(t, counts[g1.ID], "nobody holds the new secret yet")

			// Adding the new secret brings the copy up to date, even at the limit.
			require.NoError(t, s.AddGrantToWallet(a, g1.ID, 1, map[string][]string{"n1": {"m1"}, "n9": {"m1"}}, 2, now.Add(2*time.Hour)))
			adds, _ = s.GrantAddsOfWallet(a)
			require.Len(t, adds, 2)
			for _, ad := range adds {
				require.True(t, ad.Current(), ad.Grant.ID)
			}
			counts, _ = s.GrantAddCounts(owner)
			require.Equal(t, 1, counts[g1.ID])

			// A reprice to zero is a free grant.
			every := []string{}
			got, ok, err = s.RepriceGrant(g2.ID, owner, "h2b-"+sfx, 0, 0, GrantScope{Models: &every})
			require.NoError(t, err)
			require.True(t, ok)
			require.True(t, got.Free)
			require.Empty(t, got.Models, "an empty list is every model")

			// A revoked grant is never given a new secret (2026-11-24 audit).
			ok, err = s.SetGrantRevoked(g3.ID, owner, true)
			require.NoError(t, err)
			require.True(t, ok)
			_, ok, err = s.RepriceGrant(g3.ID, owner, "h3b-"+sfx, 0, 2, GrantScope{})
			require.NoError(t, err)
			require.False(t, ok, "a revoked grant cannot be repriced")
			_, found, _ = s.GrantBySecretHash("h3b-" + sfx)
			require.False(t, found)

			removed, err := s.RemoveGrantFromWallet(a, g1.ID)
			require.NoError(t, err)
			require.True(t, removed)
			removed, err = s.RemoveGrantFromWallet(a, g1.ID)
			require.NoError(t, err)
			require.False(t, removed)
			adds, _ = s.GrantAddsOfWallet(a)
			require.Len(t, adds, 1)
			adds, _ = s.GrantAddsOfWallet(b)
			require.Len(t, adds, 1, "another wallet's copy is untouched")
		})
	}
}
