package agent

// Contract: features/tower/station_mobility_idempotency.feature (ruled 2026-11-24: a flat 30 s
// retry after "no data plane"; every other refusal keeps the doubling backoff).

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAfterNoDataPlaneTheRetryIsFlatNotDoubling(t *testing.T) {
	gone := errors.New(`attach refused (503): {"error":{"message":"the tower this node is attached to has no data plane right now - try again shortly"}}`)
	for consecutive := 0; consecutive < 8; consecutive++ {
		d := ReattachDelayFor(gone, consecutive)
		require.GreaterOrEqual(t, d, 25*time.Second, "try %d", consecutive)
		require.LessOrEqual(t, d, 35*time.Second, "try %d", consecutive)
	}
	other := errors.New(`attach refused (503): {"error":{"message":"no tower can host this node right now - try again shortly"}}`)
	require.Greater(t, ReattachDelayFor(other, 6), 35*time.Second, "other refusals keep the doubling backoff")
	require.Greater(t, ReattachDelayFor(nil, 6), 35*time.Second)
}

// While Core answers "no data plane", a first attach keeps asking on the flat retry and never gives up.
func TestAFirstAttachNeverGivesUpWhileCoreAnswersNoDataPlane(t *testing.T) {
	gone := errors.New(`attach refused (503): {"error":{"message":"the tower this node is attached to has no data plane right now - try again shortly"}}`)
	require.False(t, firstAttachGivesUp(gone, firstAttachAttempts+10))
	other := errors.New("attach refused (503): no tower can host this node right now")
	require.True(t, firstAttachGivesUp(other, firstAttachAttempts))
	require.True(t, firstAttachGivesUp(ErrPrivateShareNeverRelays, 1))
}
