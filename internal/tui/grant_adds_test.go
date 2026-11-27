package tui

// grant_adds_test.go - the branches of grant_adds.go the grant_surfaces scenarios do not
// reach: builds without the hooks, logged out, usage, cancel, broker errors, and how added
// grants are drawn. It reuses the BDD harness's broker stand-in (the Hooks seam).

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/client"
)

func newGrantAddsRig(t *testing.T) *grantSurfacesBDD {
	t.Helper()
	s := &grantSurfacesBDD{}
	s.reset()
	require.NoError(t, s.oliveMadeFriends())
	require.NoError(t, s.frankLoggedIn())
	return s
}

func TestGrantAddErrText(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"logged out", client.ErrNotSignedIn, grantAddLoginFirst},
		{"limit names the TUI verb", gsLimitErr{}, "an account can hold 10 grants - remove one first (`/grant remove <name>`)"},
		{"limit with other words gains the verb", fmt.Errorf("%w: full", client.ErrGrantAddLimit), "grant add limit: full - /grant remove <name> drops one"},
		{"other text as given, cleaned", errors.New("grant key invalid\x1b[2J or revoked"), "grant key invalid[2J or revoked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, grantAddErrText(tc.err))
		})
	}
}

func TestAddedPriceText(t *testing.T) {
	for _, tc := range []struct {
		in, out float64
		want    string
	}{
		{0, 0, "free"},
		{0, 0.05, "$0.05/1M out"},
		{0.01, 0.05, "$0.01 in $0.05 out /1M"},
	} {
		require.Equal(t, tc.want, addedPriceText(tc.in, tc.out))
	}
}

func TestAddedRows(t *testing.T) {
	rows := strings.Join(addedRows(120, 10, []client.AddedGrant{
		{Name: "live", PriceOut: 0.2, State: "active", AddedAt: 1795478400,
			Stations: []client.AddedGrantStation{{ID: "olive-box", Models: []string{"qwen-small"}, PublicOut: 0.1, AbovePublic: true}}},
		{Name: "empty", PriceOut: 0.05, State: "active"},
		{Name: "gone", PriceOut: 0.05, State: "replaced", Stations: []client.AddedGrantStation{{ID: "hidden-box"}}},
	}), "\n")
	rows = stripANSI(rows)
	require.Contains(t, rows, "added 2026-11-24")
	require.Contains(t, rows, "olive-box  public $0.1/1M out  qwen-small  - the grant price is above the public price")
	require.Contains(t, rows, "covers no station yet")
	require.Contains(t, rows, "replaced - ask the owner for the new secret")
	require.NotContains(t, rows, "hidden-box", "a grant that stopped applying lists no stations")
}

func TestGrantAddsWithoutHooks(t *testing.T) {
	for _, line := range []string{"/grant add", "/grant added", "/grant remove friends"} {
		s := newGrantAddsRig(t)
		s.m.hooks.GrantAdd, s.m.hooks.GrantAddedList, s.m.hooks.GrantRemove = nil, nil, nil
		s.typeLine(line)
		require.Contains(t, stripANSI(s.m.status), "grants unavailable in this build", line)
		require.Zero(t, s.calls, line)
	}
}

func TestGrantAddsLoggedOut(t *testing.T) {
	for _, line := range []string{"/grant added", "/grant remove friends"} {
		s := newGrantAddsRig(t)
		require.NoError(t, s.frankLoggedOut())
		s.typeLine(line)
		require.Contains(t, stripANSI(s.m.status), grantAddLoginFirst, line)
		require.Zero(t, s.calls, line)
	}
}

func TestGrantAddHiddenInputCancelAndEmpty(t *testing.T) {
	s := newGrantAddsRig(t)
	s.typeLine("/grant add")
	require.True(t, s.m.keys.adding)
	require.Contains(t, s.screen(), "enter add")
	s.send(keyMsg("rog-grant_half"))
	require.NotContains(t, s.screen(), "rog-grant_half", "the input is hidden")
	s.send(tea.KeyMsg{Type: tea.KeyEsc})
	require.False(t, s.m.keys.adding)
	require.Empty(t, s.m.keys.addIn.Value())
	require.Contains(t, stripANSI(s.m.status), "no grant added")
	require.Equal(t, modeKeys, s.m.mode, "esc leaves the input, not the panel")

	s.typeLine("/grant add")
	s.send(tea.KeyMsg{Type: tea.KeyEnter})
	require.Contains(t, stripANSI(s.m.status), "paste the grant secret")
	require.Empty(t, s.added)
}

func TestGrantAddBrokerRefusal(t *testing.T) {
	s := newGrantAddsRig(t)
	s.typeLine("/grant add rog-grant_UNKNOWN")
	require.Contains(t, stripANSI(s.m.status), "! grant key invalid or revoked")
	require.Empty(t, s.added)
}

func TestGrantAddedListError(t *testing.T) {
	s := newGrantAddsRig(t)
	s.signedOut = true // the client refuses before sending, as when the sign-in went stale
	s.typeLine("/grant added")
	require.Contains(t, s.screen(), "could not read the grants you added: "+grantAddLoginFirst)
}

func TestGrantAddedEmptySection(t *testing.T) {
	s := newGrantAddsRig(t)
	s.typeLine("/grant added")
	require.Contains(t, s.screen(), "GRANTS YOU ADDED")
	require.Contains(t, s.screen(), "none yet - /grant add")
}

func TestGrantRemoveBranches(t *testing.T) {
	t.Run("usage", func(t *testing.T) {
		s := newGrantAddsRig(t)
		s.typeLine("/grant remove")
		require.Contains(t, stripANSI(s.m.status), "usage: /grant remove <name|id>")
	})
	t.Run("not added", func(t *testing.T) {
		s := newGrantAddsRig(t)
		s.typeLine("/grant remove nothing-here")
		require.Contains(t, stripANSI(s.m.status), "no grant named nothing-here in your added grants")
		require.Nil(t, s.m.keys.confirm)
	})
	t.Run("two with one name lists the ids", func(t *testing.T) {
		s := newGrantAddsRig(t)
		s.added = []client.AddedGrant{{ID: "g_a", Name: "friends"}, {ID: "g_b", Name: "friends"}}
		s.typeLine("/grant remove friends")
		require.Contains(t, stripANSI(s.m.status), "2 added grants are named friends - /grant remove <id>: g_a, g_b")
		require.Nil(t, s.m.keys.confirm)
	})
	t.Run("by id", func(t *testing.T) {
		s := newGrantAddsRig(t)
		s.added = []client.AddedGrant{{ID: "g_a", Name: "friends"}, {ID: "g_b", Name: "friends"}}
		s.typeLine("/grant remove g_b")
		require.NotNil(t, s.m.keys.confirm)
		require.Contains(t, s.footer(), "y remove")
		s.send(keyMsg("y"))
		require.Equal(t, []client.AddedGrant{{ID: "g_a", Name: "friends"}}, s.added)
		require.Contains(t, stripANSI(s.m.status), "removed friends from your account")
	})
	t.Run("list fails", func(t *testing.T) {
		s := newGrantAddsRig(t)
		s.signedOut = true
		s.typeLine("/grant remove friends")
		require.Contains(t, stripANSI(s.m.status), grantAddLoginFirst)
	})
	t.Run("remove fails", func(t *testing.T) {
		s := newGrantAddsRig(t)
		require.NoError(t, s.frankAddedFriends())
		s.typeLine("/grant remove friends")
		s.m.hooks.GrantRemove = func(string, string) (client.AddedGrant, error) { return client.AddedGrant{}, errors.New("store error") }
		s.send(keyMsg("y"))
		require.Contains(t, stripANSI(s.m.status), "store error")
	})
}

func (s *grantSurfacesBDD) footer() string { return stripANSI(s.m.footer(s.m.effWidth())) }

func TestGrantCreateRefusesFlags(t *testing.T) {
	s := newGrantAddsRig(t)
	s.typeLine("/grant create --name x")
	require.Contains(t, stripANSI(s.m.status), "usage: /grant create [name] - free grants only")
	require.Zero(t, s.createCalls)
}

func TestScrubSecretArgsGrantAdd(t *testing.T) {
	require.Equal(t, "grant add", scrubSecretArgs("grant add rog-grant_X"))
	require.Equal(t, "/grant ADD", scrubSecretArgs("/grant ADD rog-grant_X"))
	require.Equal(t, "grant create bots", scrubSecretArgs("grant create bots"))
}

// In the channel, a /grant add typed into the chat input is kept in the chat history without
// its secret, and nothing reaches the transcript.
func TestGrantAddFromChannelKeepsSecretOut(t *testing.T) {
	s := newGrantAddsRig(t)
	s.m.connected = &offer{NodeID: "nyx", Model: "qwen3-4b", Online: true}
	s.m.mode = modeChat
	s.m.chatIn.SetValue("/grant add " + gsFriendsSecret)
	s.send(tea.KeyMsg{Type: tea.KeyEnter})
	require.NoError(t, s.grantIsAdded())
	require.NotContains(t, strings.Join(s.m.chatHist.entries, "\n"), gsFriendsSecret)
	require.NoError(t, s.secretNowhere())
}

// An add lands on top at once, replacing an older row of the same grant and keeping the rest;
// the add reply has no state, so it reads active.
func TestOnGrantAddedMergesRows(t *testing.T) {
	s := newGrantAddsRig(t)
	s.m.keys.added = []client.AddedGrant{{ID: "g_friends01", Name: "friends", State: "replaced"}, {ID: "g_x", Name: "x", State: "active"}}
	tm, _ := s.m.onGrantAdded(grantAddedMsg{g: client.AddedGrant{ID: "g_friends01", Name: "friends", PriceOut: 0.05}})
	s.m = asModel(tm)
	require.Equal(t, []client.AddedGrant{{ID: "g_friends01", Name: "friends", PriceOut: 0.05, State: "active"}, {ID: "g_x", Name: "x", State: "active"}}, s.m.keys.added)
}
