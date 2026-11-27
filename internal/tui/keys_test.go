package tui

// keys_test.go: the corners of the KEYS panel that features/tui/account_keys.feature does not
// walk scenario by scenario - a missing hook, a broker that fails mid-change, the odd inputs
// to the limit parser, and the row that has to be cut.

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
	"rogerai.fm/roger/v7/internal/client"
)

func keysModel(h Hooks) model {
	var tm tea.Model = NewWithHooks("http://broker.local", "tester", nil, h)
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m := asModel(tm)
	m.loggedIn = true
	return m
}

// allKeyHooks is a broker that holds one key and one grant and fails nothing.
func allKeyHooks() Hooks {
	return Hooks{
		KeysList: func(string) ([]client.AccountKey, error) {
			return []client.AccountKey{{ID: "key_aaaa1", Name: "ci", Reset: "weekly", LimitUSD: 5}}, nil
		},
		KeysMint: func(string, map[string]any) (client.AccountKey, string, error) {
			return client.AccountKey{}, "rog-key_FAKEONLY", nil
		},
		KeysSet:    func(string, string, map[string]any) (client.AccountKey, error) { return client.AccountKey{}, nil },
		KeysRevoke: func(string, string) error { return nil },
		GrantList: func(string) ([]GrantRow, error) {
			return []GrantRow{{Name: "friends", Price: "free", Status: "active"}}, nil
		},
		GrantRevoke: func(string, string) error { return nil },
	}
}

func TestParseKeyLimit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		limit  float64
		window string
		bad    string
	}{
		{"nothing stated", nil, 0, "", keysLimitHelp},
		{"a plain amount", []string{"5"}, 5, "", ""},
		{"a dollar sign", []string{"$5.50"}, 5.5, "", ""},
		{"none", []string{"none"}, 0, "", ""},
		{"none with a window", []string{"none", "daily"}, 0, "daily", ""},
		{"zero", []string{"0"}, 0, "", client.ErrKeyLimitZero.Error()},
		{"a word", []string{"five"}, 0, "", keysLimitHelp},
		{"not a number", []string{"NaN"}, 0, "", keysLimitHelp},
		{"infinity", []string{"Inf"}, 0, "", keysLimitHelp},
		{"negative", []string{"-0.01"}, 0, "", "a limit cannot be negative"},
		{"each window", []string{"1", "monthly"}, 1, "monthly", ""},
		{"an unknown window", []string{"1", "hourly"}, 0, "", "the window is daily, weekly, monthly or none"},
		{"none is a window", []string{"1", "none"}, 1, "none", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			limit, window, bad := parseKeyLimit(tc.args)
			require.Equal(t, tc.bad, bad)
			require.Equal(t, tc.limit, limit)
			require.Equal(t, tc.window, window)
		})
	}
}

func TestKeyLimitTextAndCleanName(t *testing.T) {
	require.Equal(t, "unlimited", client.AccountKey{}.LimitText())
	require.Equal(t, "$2.00 total", client.AccountKey{LimitUSD: 2}.LimitText(), "a key with no window stated counts over its life")
	require.Equal(t, "$2.00 total", client.AccountKey{LimitUSD: 2, Reset: "none"}.LimitText())
	require.Equal(t, "$2.00 daily", client.AccountKey{LimitUSD: 2, Reset: "daily"}.LimitText())

	require.Equal(t, "(no name)", cleanName(""))
	require.Equal(t, "(no name)", cleanName(" \x1b\r\n"))
	require.Equal(t, "a[0mb", cleanName("a\x1b[0mb"))
}

func TestPanelRowFallsBackToACutLine(t *testing.T) {
	wide := stripANSI(panelRow(80, 6, "ci", []string{"key_aaaa1", "active"}, []string{"active"}))
	require.Contains(t, wide, "key_aaaa1", "the first set is used when it fits")
	mid := stripANSI(panelRow(20, 6, "ci", []string{"key_aaaa1aaaaaaaaaaaaa", "active"}, []string{"active"}))
	require.NotContains(t, mid, "key_aaaa1")
	require.Contains(t, mid, "active")
	tiny := stripANSI(panelRow(12, 6, "ci", []string{"a-long-column"}))
	require.LessOrEqual(t, len([]rune(tiny)), 12, "even the last set is cut to the width")
}

func TestKeysExportTrimsATrailingSlash(t *testing.T) {
	require.Equal(t, "export OPENAI_BASE_URL=https://b.example/v1\nexport OPENAI_API_KEY=rog-key_FAKEONLY",
		client.KeyExportLines("https://b.example/", "rog-key_FAKEONLY"))
}

func TestKeysErrTextPointsAtTheTUILogin(t *testing.T) {
	got := keysErrText(fmt.Errorf("%w - run `roger login` (keys are per account)", client.ErrKeysLogin))
	require.Equal(t, keysLoginFirst, got)
}

func TestLoadKeysWithMissingHooksSaysSoPerSection(t *testing.T) {
	km := keysModel(Hooks{})
	msg := km.loadKeys()().(keysListMsg)
	require.Contains(t, msg.keysErr, "roger keys")
	require.Contains(t, msg.grantErr, "roger grant list")

	h := allKeyHooks()
	h.KeysList = func(string) ([]client.AccountKey, error) {
		return nil, fmt.Errorf("%w - run `roger login` (keys are per account)", client.ErrKeysLogin)
	}
	m := keysModel(h)
	tm, cmd := m.run("keys")
	tm, _ = tm.Update(cmd())
	out := stripANSI(asModel(tm).keysView(100, 0))
	require.Contains(t, out, "could not read keys: "+keysLoginFirst)
	require.Contains(t, out, "friends", "the grant list still shows")
}

func TestKeysVerbsThatNeedOneKey(t *testing.T) {
	for _, tc := range []struct{ line, want string }{
		{"keys on", "usage: /keys on <key>"},
		{"keys off a b", "usage: /keys off <key>"},
		{"keys rm a b", "usage: /keys revoke <key>"},
		{"keys limit a 1 daily extra", "usage: /keys limit <key> <limit> [daily|weekly|monthly]"},
		{"grant revoke", "usage: /grant revoke <name>"},
		{"grant rm a b", "usage: /grant revoke <name>"},
	} {
		t.Run(tc.line, func(t *testing.T) {
			tm, cmd := keysModel(allKeyHooks()).run(tc.line)
			require.Nil(t, cmd)
			require.Contains(t, stripANSI(asModel(tm).status), tc.want)
		})
	}
}

func TestKeyIsAnAliasForKeys(t *testing.T) {
	tm, cmd := keysModel(allKeyHooks()).run("key")
	require.NotNil(t, cmd)
	require.Equal(t, modeKeys, asModel(tm).mode)
}

func TestLoggedOutCannotChangeAKey(t *testing.T) {
	m := keysModel(allKeyHooks())
	m.loggedIn = false
	for _, line := range []string{"keys off ci", "keys rm ci", "keys limit ci 5"} {
		tm, cmd := m.run(line)
		require.Nil(t, cmd, line)
		require.Contains(t, stripANSI(asModel(tm).status), keysLoginFirst, line)
	}
}

func TestBrokerFailuresComeBackAsOneLine(t *testing.T) {
	boom := errors.New("broker returned status 500")
	for _, tc := range []struct {
		name string
		edit func(*Hooks)
		line string
		key  string // a key to press once the first answer has landed
	}{
		{"the list fails before a change", func(h *Hooks) { h.KeysList = func(string) ([]client.AccountKey, error) { return nil, boom } }, "keys off ci", ""},
		{"the change fails", func(h *Hooks) {
			h.KeysSet = func(string, string, map[string]any) (client.AccountKey, error) { return client.AccountKey{}, boom }
		}, "keys limit ci 9", ""},
		{"the list fails before a revoke", func(h *Hooks) { h.KeysList = func(string) ([]client.AccountKey, error) { return nil, boom } }, "keys rm ci", ""},
		{"the revoke fails", func(h *Hooks) { h.KeysRevoke = func(string, string) error { return boom } }, "keys rm ci", "y"},
		{"the grant list fails before a revoke", func(h *Hooks) { h.GrantList = func(string) ([]GrantRow, error) { return nil, boom } }, "grant revoke friends", ""},
		{"the grant revoke fails", func(h *Hooks) { h.GrantRevoke = func(string, string) error { return boom } }, "grant revoke friends", "y"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := allKeyHooks()
			tc.edit(&h)
			tm, cmd := keysModel(h).run(tc.line)
			require.NotNil(t, cmd)
			msg := cmd()
			if tc.key != "" {
				tm, _ = tm.Update(msg)
				tm, cmd = tm.Update(keyMsg(tc.key))
				require.NotNil(t, cmd)
				msg = cmd()
			}
			require.Equal(t, flowErrMsg("broker returned status 500"), msg)
			tm, _ = tm.Update(msg)
			require.Contains(t, stripANSI(asModel(tm).status), "broker returned status 500")
		})
	}
}

func TestGrantRevokeWithoutItsHooks(t *testing.T) {
	h := allKeyHooks()
	h.GrantRevoke = nil
	tm, cmd := keysModel(h).run("grant revoke friends")
	require.Nil(t, cmd)
	require.Contains(t, stripANSI(asModel(tm).status), "grants unavailable in this build - run `roger grant revoke`")
}

func TestKeysPanelKeys(t *testing.T) {
	open := func() tea.Model {
		tm, cmd := keysModel(allKeyHooks()).run("keys")
		tm, _ = tm.Update(cmd())
		return tm
	}
	t.Run("a stray key does nothing", func(t *testing.T) {
		tm, cmd := open().Update(keyMsg("x"))
		require.Nil(t, cmd)
		require.Equal(t, modeKeys, asModel(tm).mode)
	})
	t.Run("esc closes back to the band list", func(t *testing.T) {
		tm, _ := open().Update(tea.KeyMsg{Type: tea.KeyEsc})
		require.Equal(t, modeBrowse, asModel(tm).mode)
	})
	for _, k := range []tea.KeyMsg{{Type: tea.KeyEnter}, keyMsg("q"), keyMsg("y")} {
		t.Run(k.String()+" does not close it: only esc and a command do", func(t *testing.T) {
			tm, cmd := open().Update(k)
			require.Nil(t, cmd)
			require.Equal(t, modeKeys, asModel(tm).mode)
		})
	}
	for _, lead := range []string{"/", ":"} {
		t.Run(lead+" closes it and starts a command, dropping a shown secret", func(t *testing.T) {
			m := asModel(open())
			m.keys.secret, m.keys.secretKind = "rog-key_TESTONLY", "key"
			tm, _ := m.Update(keyMsg(lead))
			got := asModel(tm)
			require.Equal(t, modeCommand, got.mode)
			require.Equal(t, lead, got.cmdLead)
			require.Empty(t, got.keys.secret)
			require.NotContains(t, stripANSI(got.View()), "rog-key_TESTONLY")
		})
	}
	t.Run("declining a revoke says the key was kept", func(t *testing.T) {
		tm := open()
		tm, cmd := asModel(tm).run("keys rm ci")
		tm, _ = tm.Update(cmd())
		require.Empty(t, asModel(tm).status, "the question replaces whatever the status said")
		tm, cmd = tm.Update(keyMsg("n"))
		require.Nil(t, cmd)
		require.Contains(t, stripANSI(asModel(tm).status), "kept ci")
		require.Nil(t, asModel(tm).keys.confirm)
	})
	t.Run("the footer teaches the panel's keys", func(t *testing.T) {
		require.Contains(t, stripANSI(open().View()), "esc close")
		tm, cmd := asModel(open()).run("keys rm ci")
		tm, _ = tm.Update(cmd())
		require.Contains(t, stripANSI(tm.View()), "y revoke")
	})
}

func TestGrantCommandsInTheChannelOpenThePanel(t *testing.T) {
	m := keysModel(allKeyHooks())
	m.mode = modeChat
	lines := len(m.transcript)
	tm, cmd := m.runSession("/grant")
	require.NotNil(t, cmd)
	got := asModel(tm)
	require.Equal(t, modeKeys, got.mode)
	require.Equal(t, modeChat, got.keys.prev)
	require.Len(t, got.transcript, lines)
	require.True(t, strings.Contains(stripANSI(got.View()), "GRANTS"))
}

// msgOf runs cmd, and every command a batch holds, and returns the one message of type T.
func msgOf[T tea.Msg](t *testing.T, cmd tea.Cmd) T {
	t.Helper()
	var found []T
	var walk func(tea.Cmd)
	walk = func(c tea.Cmd) {
		if c == nil {
			return
		}
		switch msg := c().(type) {
		case tea.BatchMsg:
			for _, sub := range msg {
				walk(sub)
			}
		case T:
			found = append(found, msg)
		}
	}
	walk(cmd)
	require.Len(t, found, 1)
	return found[0]
}

// keysDrive runs cmd and everything it leads to against m, like the event loop, and returns the
// model. hold names message types to keep back; they are returned undelivered, in order.
func keysDrive(t *testing.T, m model, cmd tea.Cmd, hold func(tea.Msg) bool) (model, []tea.Msg) {
	t.Helper()
	var held []tea.Msg
	queue := []tea.Cmd{cmd}
	for steps := 0; len(queue) > 0; steps++ {
		require.Less(t, steps, 200, "the commands never settled")
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		msg := c()
		if b, ok := msg.(tea.BatchMsg); ok {
			queue = append(queue, b...)
			continue
		}
		if msg == nil {
			continue
		}
		if hold != nil && hold(msg) {
			held = append(held, msg)
			continue
		}
		tm, next := m.Update(msg)
		m = asModel(tm)
		queue = append(queue, next)
	}
	return m, held
}

func isMinted(msg tea.Msg) bool  { _, ok := msg.(keyMintedMsg); return ok }
func isConfirm(msg tea.Msg) bool { _, ok := msg.(keysConfirmMsg); return ok }

func land(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	tm, cmd := m.Update(msg)
	got, _ := keysDrive(t, asModel(tm), cmd, nil)
	return got
}

const lateSecret = "rog-key_TESTONLYlate"

func lateHooks() Hooks {
	h := allKeyHooks()
	h.KeysMint = func(string, map[string]any) (client.AccountKey, string, error) {
		return client.AccountKey{ID: "key_new1"}, lateSecret, nil
	}
	h.GrantCreate = func(string, string, bool) (string, error) { return "rog-grant_TESTONLYlate", nil }
	return h
}

// A mint or a revoke opens the panel when it is typed. If the operator closes it before the
// answer, the answer is not drawn over whatever they are doing now.
func TestKeysAnswersForAClosedPanelAreNotShown(t *testing.T) {
	t.Run("the panel opens when the mint is typed", func(t *testing.T) {
		tm, _ := keysModel(lateHooks()).run("keys mint ci 5")
		require.Equal(t, modeKeys, asModel(tm).mode)
		require.Contains(t, stripANSI(asModel(tm).status), "minting ci")
	})
	t.Run("a secret that lands after esc is dropped, and the status says what to do", func(t *testing.T) {
		tm, cmd := keysModel(lateHooks()).run("keys mint ci 5")
		m, held := keysDrive(t, asModel(tm), cmd, isMinted)
		require.Len(t, held, 1)
		tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = land(t, asModel(tm), held[0])
		require.Equal(t, modeBrowse, m.mode, "the panel does not reopen itself")
		require.Empty(t, m.keys.secret)
		require.Empty(t, m.keys.minting, "the next mint is not blocked")
		require.NotContains(t, stripANSI(m.View()), lateSecret)
		require.Contains(t, stripANSI(m.status), "ci was made while the panel was closed, so its secret was dropped - /keys revoke ci")
	})
	t.Run("closing and reopening does not bring it back either", func(t *testing.T) {
		tm, cmd := keysModel(lateHooks()).run("keys mint ci 5")
		m, held := keysDrive(t, asModel(tm), cmd, isMinted)
		tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		tm, cmd = asModel(tm).run("keys")
		m, _ = keysDrive(t, asModel(tm), cmd, nil)
		m = land(t, m, held[0])
		require.Equal(t, modeKeys, m.mode)
		require.Empty(t, m.keys.secret)
		require.NotContains(t, stripANSI(m.View()), lateSecret)
	})
	t.Run("a late grant secret is dropped the same way", func(t *testing.T) {
		tm, cmd := keysModel(lateHooks()).run("grant create friends")
		m, held := keysDrive(t, asModel(tm), cmd, isMinted)
		tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = land(t, asModel(tm), held[0])
		require.NotContains(t, stripANSI(m.View()), "rog-grant_TESTONLYlate")
		require.Contains(t, stripANSI(m.status), "/grant revoke friends")
	})
	t.Run("a revoke question that lands after esc is not asked, so no key can answer it", func(t *testing.T) {
		revoked := 0
		h := lateHooks()
		h.KeysRevoke = func(string, string) error { revoked++; return nil }
		tm, cmd := keysModel(h).run("keys rm ci")
		m, held := keysDrive(t, asModel(tm), cmd, isConfirm)
		require.Len(t, held, 1)
		tm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = land(t, asModel(tm), held[0])
		require.Nil(t, m.keys.confirm)
		require.Equal(t, modeBrowse, m.mode)
		tm, cmd = m.Update(keyMsg("y"))
		keysDrive(t, asModel(tm), cmd, nil)
		require.Zero(t, revoked)
	})
	t.Run("a y typed before the question lands does not answer it", func(t *testing.T) {
		revoked := 0
		h := lateHooks()
		h.KeysRevoke = func(string, string) error { revoked++; return nil }
		tm, cmd := keysModel(h).run("keys rm ci")
		m, held := keysDrive(t, asModel(tm), cmd, isConfirm)
		tm, early := m.Update(keyMsg("y"))
		m, _ = keysDrive(t, asModel(tm), early, nil)
		m = land(t, m, held[0])
		require.NotNil(t, m.keys.confirm, "the question is still waiting for its own answer")
		require.Zero(t, revoked)
	})
}

// One thing that ends in a secret at a time: a second would take the screen and the
// clipboard from the first.
func TestOnlyOneSecretIsOnItsWayAtATime(t *testing.T) {
	for _, tc := range []struct{ first, second string }{
		{"keys mint ci 5", "grant create friends"},
		{"grant create friends", "grant create friends"},
		{"grant create friends", "keys mint ci 5"},
	} {
		t.Run(tc.first+" then "+tc.second, func(t *testing.T) {
			made := 0
			h := lateHooks()
			h.KeysMint = func(string, map[string]any) (client.AccountKey, string, error) {
				made++
				return client.AccountKey{}, lateSecret, nil
			}
			h.GrantCreate = func(string, string, bool) (string, error) { made++; return "rog-grant_TESTONLYlate", nil }
			tm, first := keysModel(h).run(tc.first)
			tm, second := asModel(tm).run(tc.second)
			require.Nil(t, second, "the second is refused before the broker is asked")
			require.Contains(t, stripANSI(asModel(tm).status), "still minting")
			m, _ := keysDrive(t, asModel(tm), first, nil)
			require.Equal(t, 1, made)
			require.NotEmpty(t, m.keys.secret)
			tm, third := m.run(tc.second)
			require.NotNil(t, third, "once the first has landed the next may start")
			require.Empty(t, asModel(tm).keys.secret, "and the first secret left the screen (K3)")
		})
	}
	t.Run("a grant made after a no-limit key does not carry the key's warning", func(t *testing.T) {
		tm, cmd := keysModel(lateHooks()).run("keys mint ci none")
		m, _ := keysDrive(t, asModel(tm), cmd, nil)
		require.Contains(t, stripANSI(m.View()), keysNoLimitNote)
		tm, cmd = m.run("grant create friends")
		m, _ = keysDrive(t, asModel(tm), cmd, nil)
		require.Contains(t, stripANSI(m.View()), "new grant")
		require.NotContains(t, stripANSI(m.View()), keysNoLimitNote)
	})
	t.Run("a failed mint frees the next one", func(t *testing.T) {
		h := lateHooks()
		h.GrantCreate = func(string, string, bool) (string, error) { return "", errors.New("nope") }
		tm, cmd := keysModel(h).run("grant create friends")
		m, _ := keysDrive(t, asModel(tm), cmd, nil)
		require.Empty(t, m.keys.minting)
		require.Contains(t, stripANSI(m.status), "grant create failed: nope")
		require.Contains(t, stripANSI(m.status), "use /keys mint")
	})
	t.Run("a broker that returns no secret is not called a success", func(t *testing.T) {
		h := lateHooks()
		h.KeysMint = func(string, map[string]any) (client.AccountKey, string, error) { return client.AccountKey{}, "", nil }
		tm, cmd := keysModel(h).run("keys mint ci 5")
		m, _ := keysDrive(t, asModel(tm), cmd, nil)
		require.NotContains(t, stripANSI(m.status), "Copied")
		require.Contains(t, stripANSI(m.status), "returned no secret - /keys revoke ci")
		require.Empty(t, m.keys.secret)
	})
}

// List reads can overlap; an older one that lands last must not put a revoked key back.
func TestOnlyTheNewestKeyListLands(t *testing.T) {
	m := keysModel(allKeyHooks())
	m.enterKeys()
	older := m.loadKeys()
	newer := m.loadKeys()
	tm, _ := m.Update(keysListMsg{load: newer().(keysListMsg).load, keys: []client.AccountKey{{ID: "key_b", Name: "kept"}}})
	old := older().(keysListMsg)
	old.keys = []client.AccountKey{{ID: "key_a", Name: "revoked-one"}, {ID: "key_b", Name: "kept"}}
	tm, _ = tm.Update(old)
	screen := stripANSI(tm.View())
	require.Contains(t, screen, "kept")
	require.NotContains(t, screen, "revoked-one")
}

// A panel opened before the account was known to be logged in reads the lists once it is.
func TestKeysPanelReadsOnceLoginIsKnown(t *testing.T) {
	m := keysModel(allKeyHooks())
	m.loggedIn, m.ghLogin = false, ""
	tm, cmd := m.run("keys")
	require.Nil(t, cmd, "logged out, the broker is asked nothing")
	tm, cmd = tm.Update(balanceMsg{loggedIn: true})
	got := msgOf[keysListMsg](t, cmd)
	require.NotEmpty(t, got.keys)
	// And only on the change: a later balance beat does not read again.
	_, cmd = tm.Update(balanceMsg{loggedIn: true})
	if cmd != nil {
		_, isList := cmd().(keysListMsg)
		require.False(t, isList)
	}
}

// A secret that exists is drawn even if the TUI's own login state says logged out.
func TestASecretIsDrawnWhateverTheLoginStateReads(t *testing.T) {
	m := keysModel(lateHooks())
	tm, cmd := m.run("grant create friends")
	got, _ := keysDrive(t, asModel(tm), cmd, nil)
	got.loggedIn, got.ghLogin = false, ""
	screen := stripANSI(got.View())
	require.Contains(t, screen, "log in to manage keys")
	require.Contains(t, strings.Join(strings.Fields(screen), ""), "rog-grant_TESTONLYlate")
}

func TestGrantRevokeAsksOnlyForALiveGrant(t *testing.T) {
	h := allKeyHooks()
	h.GrantList = func(string) ([]GrantRow, error) {
		return []GrantRow{{Name: "old", Price: "free", Status: "revoked"}, {Name: "both", Price: "free", Status: "revoked"}, {Name: "both", Price: "free", Status: "active"}}, nil
	}
	for _, tc := range []struct{ name, want string }{
		{"old", "grant old is already revoked"},
		{"nosuch", "no grant named nosuch - /grant lists them"},
	} {
		tm, cmd := keysModel(h).run("grant revoke " + tc.name)
		m, _ := keysDrive(t, asModel(tm), cmd, nil)
		require.Nil(t, m.keys.confirm)
		require.Contains(t, stripANSI(m.status), tc.want)
	}
	tm, cmd := keysModel(h).run("grant revoke both")
	m, _ := keysDrive(t, asModel(tm), cmd, nil)
	require.NotNil(t, m.keys.confirm, "a name with a live grant is asked about")
}

// Nothing the broker returns can draw on the terminal: not a name, an id, a hint, a price, a
// status or an error.
func TestBrokerTextIsDrawnClean(t *testing.T) {
	const esc, bidi = "\x1b[2J", "\u202e"
	h := allKeyHooks()
	h.KeysList = func(string) ([]client.AccountKey, error) {
		return []client.AccountKey{{ID: "key_" + esc + "a", Name: "n" + bidi, Hint: "..." + esc, Reset: "none"}}, nil
	}
	h.GrantList = func(string) ([]GrantRow, error) {
		return []GrantRow{{Name: "g", Price: "free" + esc, Status: "active" + bidi}}, nil
	}
	tm, cmd := keysModel(h).run("keys")
	m, _ := keysDrive(t, asModel(tm), cmd, nil)
	tm, cmd = m.run("keys rm key_" + esc + "a")
	m, _ = keysDrive(t, asModel(tm), cmd, nil)
	require.NotNil(t, m.keys.confirm)
	for _, text := range []string{m.keysView(200, 0), m.status} {
		require.NotContains(t, text, "\x1b")
		require.NotContains(t, text, bidi)
	}
	h.KeysList = func(string) ([]client.AccountKey, error) { return nil, errors.New("boom" + esc + bidi) }
	tm, cmd = keysModel(h).run("keys off ci")
	m, _ = keysDrive(t, asModel(tm), cmd, nil)
	require.NotContains(t, m.status, "\x1b")
	require.NotContains(t, m.status, bidi)
	require.Contains(t, m.status, "boom")
}
