package dispatch

// Contract: features/tower/station_mobility_idempotency.feature (R5).
//
// OpenUntil is what the mobility idle gate reads: a Station's attempts that have not settled and
// can still settle. Both stores, so the in-memory one the broker tests use cannot drift from SQL.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParityOpenUntilCountsOnlyAttemptsThatCanStillSettle(t *testing.T) {
	eachStore(t, func(t *testing.T, s Store) {
		now := time.Unix(1_700_000_000, 0).UTC()
		put := func(id, station string, deadline time.Time) {
			require.NoError(t, s.Put(Record{AttemptID: id, JobID: "j-" + id, TowerID: "tw-1", StationID: station,
				Model: "m", Modality: "text", RequestDigest: "d", Nonce: "n-" + id, Deadline: deadline,
				Grant: []byte("g"), AssertionKey: []byte("k"), State: StateIssued}))
		}
		put("a-open", "st-1", now.Add(5*time.Minute))
		put("a-later", "st-1", now.Add(9*time.Minute))
		put("a-past", "st-1", now.Add(-time.Minute))
		put("a-settled", "st-1", now.Add(20*time.Minute))
		put("a-other", "st-2", now.Add(30*time.Minute))
		_, err := s.ClaimByID("a-settled", "tw-1", now)
		require.NoError(t, err)
		_, err = s.Settle("a-settled", now)
		require.NoError(t, err)

		latest, n, err := s.OpenUntil("st-1", now)
		require.NoError(t, err)
		require.Equal(t, 2, n, "the past and the settled attempts do not count")
		require.True(t, latest.Equal(now.Add(9*time.Minute)), "latest %s", latest)

		_, n, err = s.OpenUntil("st-none", now)
		require.NoError(t, err)
		require.Zero(t, n)
	})
}
