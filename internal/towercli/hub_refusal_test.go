package towercli

// Contract: features/tower/tower_canary_offline_station.feature ("prints the state in plain words,
// with what the operator can do, once per state") and tower_operator_guidance.feature.
//
// Found live 2026-11-26 on the house Tower while it was suspended: the hub's node refresh printed
// "Roger Core refused the inventory: listing hub nodes requires the Tower's own signed request"
// every 30 seconds. The request WAS the Tower's own; Core refused it because the Tower is
// suspended, and said so in error.metadata.tower_state. The link loop already tells the operator
// the state; the refresh says once that it is paused, and nothing misleading.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/towerjoin"
)

func TestAHubRefreshRefusedForTheTowersStateSaysSoOncePerState(t *testing.T) {
	var out bytes.Buffer
	var last string
	suspended := &towerjoin.Refusal{Status: 403, State: "state_suspended"}
	for i := 0; i < 5; i++ {
		hubRefreshFailed(&out, suspended, &last)
	}
	require.Equal(t, 1, strings.Count(out.String(), "hub:"), "said more than once per state:\n%s", out.String())
	require.Contains(t, out.String(), "suspended")
	require.Contains(t, out.String(), "resumes by itself")
	require.NotContains(t, out.String(), "requires the Tower's own signed request")

	// A different state is said once more.
	hubRefreshFailed(&out, &towerjoin.Refusal{Status: 403, State: "lease_lapsed"}, &last)
	require.Equal(t, 2, strings.Count(out.String(), "hub:"), "a new state was not said:\n%s", out.String())
}

func TestAHubRefreshThatFailsForAnyOtherReasonStillSaysWhy(t *testing.T) {
	var out bytes.Buffer
	last := "state_suspended" // a state was said before; this failure is not about it
	hubRefreshFailed(&out, errors.New("dial tcp: connection refused"), &last)
	require.Contains(t, out.String(), "could not refresh node registrations: dial tcp: connection refused")
	require.Empty(t, last, "after a non-state failure a later refusal is said again")

	// An unverified refusal (no state: Core could not tell it was this Tower) keeps Core's words.
	out.Reset()
	hubRefreshFailed(&out, &towerjoin.Refusal{Status: 403}, &last)
	require.Contains(t, out.String(), "could not refresh node registrations")
}

// Pre-push audit (2026-11-26): a TERMINAL state never resumes, so the hub never says it will.
// serve prints the terminal words and exits (classifyRefusal); the hub adds nothing.
func TestAHubRefreshRefusedForATerminalStateNeverSaysItResumes(t *testing.T) {
	for _, state := range []string{"state_revoked", "cert_revoked", "state_expired"} {
		var out bytes.Buffer
		var last string
		hubRefreshFailed(&out, &towerjoin.Refusal{Status: 403, State: state}, &last)
		require.NotContains(t, out.String(), "resumes by itself", state)
		require.NotContains(t, out.String(), "requires the Tower's own signed request", state)
	}
}

// The paused line uses plain words, not Core's state tokens.
func TestAHubRefreshPausedLineUsesPlainWords(t *testing.T) {
	var out bytes.Buffer
	var last string
	hubRefreshFailed(&out, &towerjoin.Refusal{Status: 403, State: "lease_lapsed"}, &last)
	require.NotContains(t, out.String(), "lease_lapsed")
	require.Contains(t, out.String(), "approval lapsed")
}
