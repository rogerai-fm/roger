package tui

// account_keys_bdd_test.go - the godog harness for features/tui/account_keys.feature (account
// keys and grants in the TUI). The steps drive the REAL model through its two command
// dispatchers (run under the band list, runSession in the channel), its key handler and its
// View. The only stand-in is the broker, behind the Hooks seam the host wires to the client:
// it records every call, so "the broker receives ..." reads what was really asked.

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

const keysBDDBroker = "http://broker.local"

// grantRefusal is what the client reports when the broker answers 403 on /grants.
const grantRefusal = "grants require a linked operator account - sign in (GitHub or Apple) first"

type keySetCall struct {
	id     string
	fields map[string]any
}

type grantCreateCall struct {
	name string
	free bool
}

type keysBDD struct {
	t    *testing.T
	m    model
	clip *bytes.Buffer

	// the broker stand-in
	keys         []client.AccountKey
	grants       []GrantRow
	unreachable  bool
	grantsRefuse bool
	grantListErr error
	minted       int

	// what the broker was asked
	calls        int
	mints        []map[string]any
	sets         []keySetCall
	revokes      []string
	grantCreates []grantCreateCall
	grantRevokes []string

	secrets     []string // every secret the broker handed out, in order
	inChannel   bool
	transcript0 int
	noHooks     bool
	lateID      string
}

func (s *keysBDD) reset() {
	*s = keysBDD{t: s.t, clip: &bytes.Buffer{}}
	clipboardOut = s.clip
	SetLANDial(nil)
}

func (s *keysBDD) hooks() Hooks {
	if s.noHooks {
		return Hooks{}
	}
	down := func() error {
		s.calls++
		if s.unreachable {
			return fmt.Errorf("%w: dial tcp: connection refused", client.ErrBrokerUnreachable)
		}
		return nil
	}
	return Hooks{
		KeysList: func(string) ([]client.AccountKey, error) {
			if err := down(); err != nil {
				return nil, err
			}
			return append([]client.AccountKey(nil), s.keys...), nil
		},
		KeysMint: func(_ string, fields map[string]any) (client.AccountKey, string, error) {
			if err := down(); err != nil {
				return client.AccountKey{}, "", err
			}
			s.mints = append(s.mints, fields)
			if len(s.keys) >= 32 {
				return client.AccountKey{}, "", errors.New("at most 32 keys per account - delete one first")
			}
			s.minted++
			secret := fmt.Sprintf("rog-key_FAKE%02d%s", s.minted, strings.Repeat("0a1b2c3d", 7)+"9f3c")
			s.secrets = append(s.secrets, secret)
			k := client.AccountKey{ID: fmt.Sprintf("key_new%d", s.minted), Hint: "..." + secret[len(secret)-4:], Reset: "none"}
			k.Name, _ = fields["name"].(string)
			k.LimitUSD, _ = fields["limit_usd"].(float64)
			if r, ok := fields["reset"].(string); ok {
				k.Reset = r
			}
			s.keys = append([]client.AccountKey{k}, s.keys...)
			return k, secret, nil
		},
		KeysSet: func(_, id string, fields map[string]any) (client.AccountKey, error) {
			if err := down(); err != nil {
				return client.AccountKey{}, err
			}
			s.sets = append(s.sets, keySetCall{id, fields})
			for i := range s.keys {
				if s.keys[i].ID != id {
					continue
				}
				if v, ok := fields["limit_usd"].(float64); ok {
					s.keys[i].LimitUSD = v
				}
				if v, ok := fields["reset"].(string); ok {
					s.keys[i].Reset = v
				}
				if v, ok := fields["disabled"].(bool); ok {
					s.keys[i].Disabled = v
				}
				return s.keys[i], nil
			}
			return client.AccountKey{}, errors.New("no such key")
		},
		KeysRevoke: func(_, id string) error {
			if err := down(); err != nil {
				return err
			}
			s.revokes = append(s.revokes, id)
			for i := range s.keys {
				if s.keys[i].ID == id {
					s.keys = append(s.keys[:i:i], s.keys[i+1:]...)
					break
				}
			}
			return nil
		},
		GrantCreate: func(_, name string, free bool) (string, error) {
			if err := down(); err != nil {
				return "", err
			}
			s.grantCreates = append(s.grantCreates, grantCreateCall{name, free})
			if s.grantsRefuse {
				return "", errors.New(grantRefusal)
			}
			secret := "rog-grant_FAKE" + strings.Repeat("7e6d5c4b", 7)
			s.secrets = append(s.secrets, secret)
			s.grants = append(s.grants, GrantRow{Name: name, Price: "free", Status: "active"})
			return secret, nil
		},
		GrantList: func(string) ([]GrantRow, error) {
			if err := down(); err != nil {
				return nil, err
			}
			if s.grantsRefuse {
				return nil, errors.New(grantRefusal)
			}
			if s.grantListErr != nil {
				return nil, s.grantListErr
			}
			return append([]GrantRow(nil), s.grants...), nil
		},
		GrantRevoke: func(_, name string) error {
			if err := down(); err != nil {
				return err
			}
			s.grantRevokes = append(s.grantRevokes, name)
			for i := range s.grants {
				if s.grants[i].Name == name {
					s.grants[i].Status = "revoked"
				}
			}
			return nil
		},
	}
}

// --- driving the model -------------------------------------------------------------------

// pump runs a command and feeds what it returns back into the model, the way the program
// loop does, until nothing more comes back.
func (s *keysBDD) pump(cmd tea.Cmd, depth int) {
	if cmd == nil || depth > 20 {
		return
	}
	msg := cmd()
	switch v := msg.(type) {
	case nil:
		return
	case tea.BatchMsg:
		for _, c := range v {
			s.pump(c, depth+1)
		}
		return
	}
	if strings.Contains(strings.ToLower(fmt.Sprintf("%T", msg)), "tick") || strings.Contains(fmt.Sprintf("%T", msg), "Blink") {
		return // a clock beat, not an answer
	}
	tm, next := s.m.Update(msg)
	s.m = asModel(tm)
	s.pump(next, depth+1)
}

// dispatch hands a typed line to the dispatcher that owns it where the operator is, and
// returns the command without running it.
func (s *keysBDD) dispatch(line string) tea.Cmd {
	var tm tea.Model
	var cmd tea.Cmd
	if s.inChannel {
		tm, cmd = s.m.runSession(line)
	} else {
		tm, cmd = s.m.run(line)
	}
	s.m = asModel(tm)
	return cmd
}

func (s *keysBDD) runs(line string) error {
	s.pump(s.dispatch(line), 0)
	return nil
}

func (s *keysBDD) press(key string) {
	k := keyMsg(key)
	switch key {
	case "esc":
		k = tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		k = tea.KeyMsg{Type: tea.KeyEnter}
	}
	tm, cmd := s.m.Update(k)
	s.m = asModel(tm)
	s.pump(cmd, 0)
}

func (s *keysBDD) screen() string { return stripANSI(s.m.View()) }
func (s *keysBDD) status() string { return stripANSI(s.m.status) }

// flat is the screen with every line break and run of spaces removed, so wrapped text and
// a secret broken across lines can still be found whole.
func (s *keysBDD) flat() string { return strings.Join(strings.Fields(s.screen()), "") }

func squash(text string) string { return strings.Join(strings.Fields(text), "") }

// section is the text of one headed part of the panel: from its header to the next header
// or the end of the panel body.
func (s *keysBDD) section(name string) (string, error) {
	scr := s.screen()
	i := strings.Index(scr, name)
	if i < 0 {
		return "", fmt.Errorf("no %q section on screen:\n%s", name, scr)
	}
	rest := scr[i+len(name):]
	for _, other := range []string{"ACCOUNT KEYS", "GRANTS", "new key", "new grant", "revoke "} {
		if other == name {
			continue
		}
		if j := strings.Index(rest, "\n  "+other); j >= 0 {
			rest = rest[:j]
		}
	}
	return rest, nil
}

// lineFor is the one row of the panel that starts with this name.
func (s *keysBDD) lineFor(name string) (string, error) {
	for _, ln := range strings.Split(s.screen(), "\n") {
		if f := strings.Fields(ln); len(f) > 1 && f[0] == name {
			return ln, nil
		}
	}
	return "", fmt.Errorf("no row for %q on screen:\n%s", name, s.screen())
}

func (s *keysBDD) clipboard() string {
	re := regexp.MustCompile("\x1b\\]52;c;([A-Za-z0-9+/=]*)\a")
	all := re.FindAllStringSubmatch(s.clip.String(), -1)
	if len(all) == 0 {
		return ""
	}
	raw, _ := base64.StdEncoding.DecodeString(all[len(all)-1][1])
	return string(raw)
}

func (s *keysBDD) lastSecret() (string, error) {
	if len(s.secrets) == 0 {
		return "", errors.New("the broker handed out no secret")
	}
	return s.secrets[len(s.secrets)-1], nil
}

func has(hay, want string) error {
	if !strings.Contains(hay, want) {
		return fmt.Errorf("want %q in:\n%s", want, hay)
	}
	return nil
}

// --- Given -------------------------------------------------------------------------------

func (s *keysBDD) running() error {
	m := NewWithHooks(keysBDDBroker, "tester", nil, s.hooks())
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	s.m = asModel(tm)
	return nil
}

func (s *keysBDD) loggedIn() error  { s.m.loggedIn = true; return nil }
func (s *keysBDD) loggedOut() error { s.m.loggedIn, s.m.ghLogin = false, ""; return nil }

func (s *keysBDD) withoutHooks() error {
	s.noHooks = true
	s.m.hooks = Hooks{}
	return nil
}

func (s *keysBDD) lanMode() error {
	SetLANDial(func() client.LANDial { return client.LANDial{} })
	return nil
}

func (s *keysBDD) width60() error {
	tm, _ := s.m.Update(tea.WindowSizeMsg{Width: 60, Height: 60})
	s.m = asModel(tm)
	return nil
}

func (s *keysBDD) inTheChannel() error {
	s.m.connected = &offer{NodeID: "nyx", Model: "qwen3-4b", Online: true}
	s.m.endpoint, s.m.apikey = "http://127.0.0.1:4141/v1", "sess-0123456789abcdef"
	s.m.mode = modeChat
	s.m.transcript = []string{"hello", "world"}
	s.inChannel, s.transcript0 = true, len(s.m.transcript)
	return nil
}

func (s *keysBDD) channelOpen() error {
	s.m.connected = &offer{NodeID: "nyx", Model: "qwen3-4b", Online: true}
	s.m.endpoint, s.m.apikey = "http://127.0.0.1:4141/v1", "sess-0123456789abcdef"
	return nil
}

func (s *keysBDD) addKey(k client.AccountKey) {
	if k.Hint == "" {
		k.Hint = "...9f3c"
	}
	if k.Reset == "" {
		k.LimitUSD, k.Reset = 5, "weekly"
	}
	s.keys = append(s.keys, k)
}

func (s *keysBDD) keysTable(tbl *godog.Table) error {
	for _, row := range tbl.Rows[1:] {
		c := func(i int) string { return row.Cells[i].Value }
		limit, _ := strconv.ParseFloat(c(3), 64)
		used, _ := strconv.ParseFloat(c(5), 64)
		k := client.AccountKey{ID: c(0), Name: c(1), Hint: c(2), LimitUSD: limit, Reset: c(4), Disabled: c(6) == "disabled", Expired: c(6) == "expired"}
		k.Usage, k.UsageDaily, k.UsageWeekly, k.UsageMonthly = used, used, used, used
		s.keys = append(s.keys, k)
	}
	return nil
}

func (s *keysBDD) grantsTable(tbl *godog.Table) error {
	for _, row := range tbl.Rows[1:] {
		s.grants = append(s.grants, GrantRow{Name: row.Cells[0].Value, Price: row.Cells[1].Value, Status: row.Cells[2].Value})
	}
	return nil
}

func (s *keysBDD) nothingYet() error { s.keys, s.grants = nil, nil; return nil }

func (s *keysBDD) keyWithLimit(limit float64, reset string) error {
	s.keys = append(s.keys, client.AccountKey{ID: "key_aaaa1", Name: "ci", Hint: "...9f3c", LimitUSD: limit, Reset: reset})
	return nil
}

func (s *keysBDD) keyThatSpent(reset string) error {
	s.keys = append(s.keys, client.AccountKey{ID: "key_aaaa1", Name: "ci", Hint: "...9f3c", LimitUSD: 50, Reset: reset,
		UsageDaily: 1, UsageWeekly: 2, UsageMonthly: 3, Usage: 4})
	return nil
}

func (s *keysBDD) keyThatIs(condition string) error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci",
		Disabled: strings.Contains(condition, "disabled"), Expired: strings.Contains(condition, "expired")})
	return nil
}

func (s *keysBDD) keyNamed(name string) error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: name})
	return nil
}

func (s *keysBDD) keyIDNamed(id, name string) error {
	s.addKey(client.AccountKey{ID: id, Name: name})
	return nil
}

func (s *keysBDD) twoKeysNamed(a, b, name string) error {
	s.addKey(client.AccountKey{ID: a, Name: name})
	s.addKey(client.AccountKey{ID: b, Name: name})
	return nil
}

func (s *keysBDD) keyWithLongName() error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: strings.Repeat("n", 128)})
	return nil
}

func (s *keysBDD) keyWithEscapeName() error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci\x1b[2J\x1b[31mred\x07\r"})
	return nil
}

func (s *keysBDD) keysAndGrants() error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci"})
	s.addKey(client.AccountKey{ID: "key_bbbb2", Name: "deploy"})
	s.grants = append(s.grants, GrantRow{Name: "my-bots", Price: "free", Status: "active"})
	return nil
}

func (s *keysBDD) thirtyTwoKeys() error {
	for i := 0; i < 32; i++ {
		s.addKey(client.AccountKey{ID: fmt.Sprintf("key_full%02d", i), Name: fmt.Sprintf("k%02d", i)})
	}
	return nil
}

func (s *keysBDD) grantNamed(name string) error {
	s.grants = append(s.grants, GrantRow{Name: name, Price: "free", Status: "active"})
	return nil
}

func (s *keysBDD) brokerUnreachable() error   { s.unreachable = true; return nil }
func (s *keysBDD) brokerSlow() error          { return nil } // the "twice" step holds the answer back
func (s *keysBDD) brokerRefusesGrants() error { s.grantsRefuse = true; return nil }
func (s *keysBDD) grantListFails() error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: "ci"})
	s.grantListErr = errors.New("broker returned status 500")
	return nil
}

func (s *keysBDD) mintedAndShowing() error {
	if err := s.runs("/keys mint ci 5 weekly"); err != nil {
		return err
	}
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	return has(s.flat(), sec)
}

func (s *keysBDD) panelOpen() error { return s.runs("/keys") }

func (s *keysBDD) lateKey(name string) error {
	s.lateID = "key_late9"
	s.addKey(client.AccountKey{ID: s.lateID, Name: name})
	return nil
}

func (s *keysBDD) askedToConfirmKey(name string) error {
	s.addKey(client.AccountKey{ID: "key_aaaa1", Name: name})
	if err := s.runs("/keys rm " + name); err != nil {
		return err
	}
	return s.asksToConfirmKey(name)
}

func (s *keysBDD) confirmedRevokingKey(name string) error {
	if err := s.askedToConfirmKey(name); err != nil {
		return err
	}
	s.press("y")
	return nil
}

func (s *keysBDD) askedToConfirmGrant(name string) error {
	s.grants = append(s.grants, GrantRow{Name: name, Price: "free", Status: "active"})
	if err := s.runs("/grant revoke " + name); err != nil {
		return err
	}
	return s.asksToConfirmGrant(name)
}

// --- When --------------------------------------------------------------------------------

func (s *keysBDD) runsThenEsc(line string) error {
	if err := s.runs(line); err != nil {
		return err
	}
	s.press("esc")
	return nil
}

func (s *keysBDD) closesAndRuns(line string) error {
	s.press("esc")
	return s.runs(line)
}

// runsTwice sends the same line twice before the first answer comes back, then lets the
// broker be asked. The answer itself is held, so the screen is as it was while waiting.
func (s *keysBDD) runsTwice(line string) error { return s.runsBoth(line, line) }

// runsUnanswered types a line and holds its answer back.
func (s *keysBDD) runsUnanswered(line string) error {
	s.dispatch(line)
	return nil
}

// runsClosesAnswered types a line, closes the panel, and only then lets the answer land.
func (s *keysBDD) runsClosesAnswered(line string) error {
	cmd := s.dispatch(line)
	s.press("esc")
	s.pump(cmd, 0)
	return nil
}

// manyKeysShortTerminal gives the account n keys (key_0001 named k01, ...) on a terminal of
// the given height.
func (s *keysBDD) manyKeysShortTerminal(n, rows int) error {
	for i := 1; i <= n; i++ {
		s.keys = append(s.keys, client.AccountKey{ID: fmt.Sprintf("key_%04d", i), Name: fmt.Sprintf("k%02d", i), LimitUSD: 5, Reset: "weekly"})
	}
	tm, _ := s.m.Update(tea.WindowSizeMsg{Width: 100, Height: rows})
	s.m = asModel(tm)
	return nil
}

func (s *keysBDD) keyOnScreen(n int, not string) error {
	on := strings.Contains(s.screen(), fmt.Sprintf("key_%04d", n))
	if on == (not != "") {
		return fmt.Errorf("key %d on screen is %v:\n%s", n, on, s.screen())
	}
	return nil
}

func (s *keysBDD) panelDoesNotSay(text string) error {
	if strings.Contains(s.screen(), text) {
		return fmt.Errorf("the panel says %q:\n%s", text, s.screen())
	}
	return nil
}

func (s *keysBDD) saysRowsBelow() error {
	if !regexp.MustCompile(`\d+ more below`).MatchString(s.screen()) {
		return fmt.Errorf("the panel does not say how many rows are below:\n%s", s.screen())
	}
	return nil
}

func (s *keysBDD) screenFitsHeight() error {
	if rows := lineRows(strings.TrimRight(s.m.View(), "\n")); rows > s.m.height {
		return fmt.Errorf("the screen is %d rows on a terminal of %d", rows, s.m.height)
	}
	return nil
}

func (s *keysBDD) pressesEach(keys string) error {
	for _, k := range strings.Split(keys, ",") {
		s.press(k)
	}
	return nil
}

func (s *keysBDD) hintOnScreen(_, text string) error { return has(s.screen(), text) }

func (s *keysBDD) panelNotShown() error {
	if s.m.mode == modeKeys {
		return fmt.Errorf("the KEYS panel is still open:\n%s", s.screen())
	}
	return nil
}

func (s *keysBDD) clipboardNoSecret() error {
	if c := s.clipboard(); strings.Contains(c, "rog-key_") || strings.Contains(c, "rog-grant_") {
		return errors.New("a secret is on the clipboard")
	}
	return nil
}

func (s *keysBDD) revokedGrantNamed(name string) error {
	s.grants = append(s.grants, GrantRow{Name: name, Price: "free", Status: "revoked"})
	return nil
}

// oneAndNoOther checks that exactly one secret was asked for, of the kind named first.
func (s *keysBDD) oneAndNoOther(what string) error {
	mints, grants := 1, 0
	if strings.HasPrefix(what, "one grant create") {
		mints, grants = 0, 1
	}
	if len(s.mints) != mints || len(s.grantCreates) != grants {
		return fmt.Errorf("the broker received %d mints and %d grant creates", len(s.mints), len(s.grantCreates))
	}
	return nil
}

func (s *keysBDD) runsBoth(a, b string) error {
	first := s.dispatch(a)
	second := s.dispatch(b)
	// Both were typed before either answer. The broker does its work; the answers are still
	// on their way, so the screen is as the operator sees it while waiting.
	var ask func(tea.Cmd)
	ask = func(c tea.Cmd) {
		if c == nil {
			return
		}
		if b, ok := c().(tea.BatchMsg); ok {
			for _, sub := range b {
				ask(sub)
			}
		}
	}
	ask(first)
	ask(second)
	return nil
}

func (s *keysBDD) presses(key string) error { s.press(key); return nil }

func (s *keysBDD) opensHelp() error { return nil }

// types opens the command prompt and types the rest, key by key.
func (s *keysBDD) types(text string) error {
	for _, r := range text {
		tm, _ := s.m.Update(keyMsg(string(r)))
		s.m = asModel(tm)
	}
	return nil
}

// --- Then: the panel ---------------------------------------------------------------------

func (s *keysBDD) panelShown() error {
	scr := s.screen()
	for _, want := range []string{"KEYS", "ACCOUNT KEYS", "GRANTS"} {
		if err := has(scr, want); err != nil {
			return err
		}
	}
	return nil
}

func (s *keysBDD) sectionOneLinePer(name, kind string) error {
	if err := s.panelShown(); err != nil {
		return err
	}
	if kind == "key" {
		for _, k := range s.keys {
			if _, err := s.lineFor(k.Name); err != nil {
				return err
			}
		}
		return nil
	}
	for _, g := range s.grants {
		if _, err := s.lineFor(g.Name); err != nil {
			return err
		}
	}
	return nil
}

func (s *keysBDD) lineShows5(name, a, b, c, d, e string) error {
	ln, err := s.lineFor(name)
	if err != nil {
		return err
	}
	for _, want := range []string{a, b, c, d, e} {
		if err := has(ln, want); err != nil {
			return err
		}
	}
	return nil
}

func (s *keysBDD) lineShows2(name, a, b string) error {
	ln, err := s.lineFor(name)
	if err != nil {
		return err
	}
	if err := has(ln, a); err != nil {
		return err
	}
	return has(ln, b)
}

func (s *keysBDD) sectionSays(name, text string) error {
	sec, err := s.section(name)
	if err != nil {
		return err
	}
	if !strings.Contains(squash(sec), squash(text)) {
		return fmt.Errorf("%s section lacks %q:\n%s", name, text, sec)
	}
	return nil
}

func (s *keysBDD) thatKeyReads(text string) error {
	ln, err := s.lineFor("ci")
	if err != nil {
		return err
	}
	return has(ln, text)
}

func (s *keysBDD) thatKeyUsed(text string) error { return s.thatKeyReads(text + " used") }

func (s *keysBDD) thatKeyState(state string) error {
	ln, err := s.lineFor("ci")
	if err != nil {
		return err
	}
	if f := strings.Fields(ln); f[len(f)-1] != state {
		return fmt.Errorf("state is %q, want %q, in %q", f[len(f)-1], state, ln)
	}
	return nil
}

func (s *keysBDD) backInChannel() error {
	if s.m.mode != modeChat {
		return fmt.Errorf("mode is %v, want the channel", s.m.mode)
	}
	return nil
}

func (s *keysBDD) transcriptUnchanged() error {
	if len(s.m.transcript) != s.transcript0 {
		return fmt.Errorf("transcript grew from %d to %d lines", s.transcript0, len(s.m.transcript))
	}
	return nil
}

func (s *keysBDD) lineShowsNameLimitState(name string) error {
	ln, err := s.lineFor(name)
	if err != nil {
		return err
	}
	if err := has(ln, "$5.00 weekly"); err != nil {
		return err
	}
	return has(ln, "active")
}

func (s *keysBDD) noLineTooWide() error {
	// The frame clips every line to the terminal as a last resort, so the panel itself is
	// checked too: nothing it draws may need that clip.
	for _, ln := range strings.Split(s.screen()+"\n"+stripANSI(s.m.keysView(s.m.effWidth(), 0)), "\n") {
		if n := len([]rune(ln)); n > s.m.width {
			return fmt.Errorf("line is %d cells in a %d cell terminal: %q", n, s.m.width, ln)
		}
	}
	return nil
}

func (s *keysBDD) longNameFits() error {
	ln, err := s.lineFor(strings.Fields(s.longLine())[0])
	if err != nil {
		return err
	}
	if n := len([]rune(ln)); n > s.m.width {
		return fmt.Errorf("line is %d cells wide in a %d cell terminal", n, s.m.width)
	}
	return s.noLineTooWide()
}

// longLine is the row of the key whose name was cut.
func (s *keysBDD) longLine() string {
	for _, ln := range strings.Split(s.screen(), "\n") {
		if strings.Contains(ln, "nnnnnnnn") {
			return ln
		}
	}
	return "missing"
}

func (s *keysBDD) nameEndsWith(tail string) error {
	name := strings.Fields(s.longLine())[0]
	if !strings.HasSuffix(name, tail) || strings.Count(name, "n") >= 128 {
		return fmt.Errorf("name %q is not cut with %q", name, tail)
	}
	return nil
}

func (s *keysBDD) nameWithoutControls() error {
	scr := s.m.View()
	for _, bad := range []string{"\x1b[2J", "\x07", "\r"} {
		if strings.Contains(scr, bad) {
			return fmt.Errorf("the screen carries the control sequence %q", bad)
		}
	}
	return has(s.screen(), "ci[2J[31mred")
}

func (s *keysBDD) noSecretPrefixes(a, b string) error {
	if err := s.sectionOneLinePer("ACCOUNT KEYS", "key"); err != nil {
		return err // the check means nothing unless the list is really on screen
	}
	scr := s.screen()
	if strings.Contains(scr, a) || strings.Contains(scr, b) {
		return fmt.Errorf("a secret prefix is on screen:\n%s", scr)
	}
	return nil
}

func (s *keysBDD) panelShownWithNewKey() error {
	if err := s.panelShown(); err != nil {
		return err
	}
	_, err := s.lineFor("ci")
	return err
}

func (s *keysBDD) showsSecretUnder(label string) error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	scr := s.screen()
	i := strings.Index(scr, label)
	if i < 0 {
		return fmt.Errorf("no %q on screen:\n%s", label, scr)
	}
	if !strings.Contains(squash(scr[i:]), sec) {
		return fmt.Errorf("the secret is not under %q:\n%s", label, scr)
	}
	if n := strings.Count(squash(scr), sec); n != 1 {
		return fmt.Errorf("the secret is on screen %d times, want once", n)
	}
	return nil
}

func (s *keysBDD) panelSays(text string) error {
	if !strings.Contains(squash(s.screen()), squash(text)) {
		return fmt.Errorf("panel lacks %q:\n%s", text, s.screen())
	}
	return nil
}

func (s *keysBDD) everyCharOnScreen() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	return has(s.flat(), sec)
}

func (s *keysBDD) noSecretShown() error {
	scr := s.flat()
	if strings.Contains(scr, "rog-key_") || strings.Contains(scr, "rog-grant_") {
		return fmt.Errorf("a secret is on screen:\n%s", s.screen())
	}
	return nil
}

func (s *keysBDD) secretNotOnScreen() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	if strings.Contains(s.flat(), sec) || strings.Contains(s.status(), sec) {
		return fmt.Errorf("the secret is still on screen:\n%s", s.screen())
	}
	return nil
}

func (s *keysBDD) firstSecretNotOnScreen() error {
	if len(s.secrets) == 0 {
		return errors.New("no secret was handed out")
	}
	if strings.Contains(s.flat(), s.secrets[0]) {
		return fmt.Errorf("the first secret is still on screen:\n%s", s.screen())
	}
	return nil
}

func (s *keysBDD) newKeyListedWithHint() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	ln, err := s.lineFor("ci")
	if err != nil {
		return err
	}
	return has(ln, "..."+sec[len(sec)-4:])
}

func (s *keysBDD) panelShowsKeyAsReturned() error {
	if err := s.panelShown(); err != nil {
		return err
	}
	ln, err := s.lineFor("ci")
	if err != nil {
		return err
	}
	k := s.keys[0]
	want := "unlimited"
	if k.LimitUSD > 0 {
		want = fmt.Sprintf("$%.2f", k.LimitUSD)
		if k.Reset != "none" {
			want += " " + k.Reset
		}
	}
	if err := has(ln, want); err != nil {
		return err
	}
	return has(ln, k.State())
}

func (s *keysBDD) panelNoLongerLists(name string) error {
	if err := s.panelShown(); err != nil {
		return err
	}
	if ln, err := s.lineFor(name); err == nil {
		return fmt.Errorf("%q is still listed: %q", name, ln)
	}
	return nil
}

func (s *keysBDD) panelAsks(text string) error { return s.panelSays(text) }

func (s *keysBDD) asksToConfirmKey(name string) error {
	return s.panelSays("revoke " + name + " (key_aaaa1)? anything using it stops working. y/N")
}

func (s *keysBDD) asksToConfirmGrant(name string) error {
	return s.panelSays("revoke grant " + name + "? anyone using it loses access. y/N")
}

func (s *keysBDD) panelShowsGrantAs(name, status string) error {
	ln, err := s.lineFor(name)
	if err != nil {
		return err
	}
	return has(ln, status)
}

func (s *keysBDD) sectionLists(name, key string) error {
	sec, err := s.section(name)
	if err != nil {
		return err
	}
	return has(sec, key)
}

func (s *keysBDD) sectionListsTheKeys(name string) error { return s.sectionLists(name, "ci") }

func (s *keysBDD) grantsCarryReason(name string) error {
	return s.sectionSays(name, grantRefusal)
}

func (s *keysBDD) grantsCouldNotBeRead(name string) error {
	if err := s.sectionSays(name, "could not read grants"); err != nil {
		return err
	}
	return s.sectionSays(name, "broker returned status 500")
}

// --- Then: the broker --------------------------------------------------------------------

func (s *keysBDD) oneMint() (map[string]any, error) {
	if len(s.mints) != 1 {
		return nil, fmt.Errorf("the broker received %d mints, want 1", len(s.mints))
	}
	return s.mints[0], nil
}

func fieldsAre(got map[string]any, want map[string]any) error {
	if len(got) != len(want) {
		return fmt.Errorf("fields are %v, want exactly %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			return fmt.Errorf("field %s is %v (%T), want %v, in %v", k, got[k], got[k], v, got)
		}
	}
	return nil
}

func (s *keysBDD) mintNameLimitReset(name string, limit float64, reset string) error {
	f, err := s.oneMint()
	if err != nil {
		return err
	}
	return fieldsAre(f, map[string]any{"name": name, "limit_usd": limit, "reset": reset})
}

func (s *keysBDD) mintNameLimit(name string, limit float64) error {
	f, err := s.oneMint()
	if err != nil {
		return err
	}
	if f["name"] != name || f["limit_usd"] != limit {
		return fmt.Errorf("mint was %v, want name %q and limit_usd %v", f, name, limit)
	}
	return nil
}

func (s *keysBDD) mintLimit(limit float64) error {
	f, err := s.oneMint()
	if err != nil {
		return err
	}
	if f["limit_usd"] != limit {
		return fmt.Errorf("mint limit_usd is %v, want %v", f["limit_usd"], limit)
	}
	return nil
}

func (s *keysBDD) noMint() error {
	if len(s.mints) != 0 {
		return fmt.Errorf("the broker received a mint: %v", s.mints)
	}
	return nil
}

func (s *keysBDD) exactlyOneMint() error { _, err := s.oneMint(); return err }

func (s *keysBDD) oneSet(id string) (map[string]any, error) {
	if len(s.sets) != 1 {
		return nil, fmt.Errorf("the broker received %d changes, want 1", len(s.sets))
	}
	if s.sets[0].id != id {
		return nil, fmt.Errorf("the change went to %q, want %q", s.sets[0].id, id)
	}
	return s.sets[0].fields, nil
}

// changeWithOnly reads the outline's field list: `limit_usd 10`, `limit_usd 10 and reset
// "monthly"`, `disabled true`.
func (s *keysBDD) changeWithOnly(id, spec string) error {
	got, err := s.oneSet(id)
	if err != nil {
		return err
	}
	want := map[string]any{}
	for _, part := range strings.Split(spec, " and ") {
		kv := strings.Fields(part)
		switch kv[0] {
		case "limit_usd":
			want[kv[0]], _ = strconv.ParseFloat(kv[1], 64)
		case "reset":
			want[kv[0]] = strings.Trim(kv[1], `"`)
		case "disabled":
			want[kv[0]] = kv[1] == "true"
		default:
			return fmt.Errorf("unknown field in %q", spec)
		}
	}
	return fieldsAre(got, want)
}

func (s *keysBDD) changeTo(id string) error { _, err := s.oneSet(id); return err }

func (s *keysBDD) changeToThatKey() error { return s.changeTo(s.lateID) }

func (s *keysBDD) noChange() error {
	if len(s.sets) != 0 {
		return fmt.Errorf("the broker received a change: %v", s.sets)
	}
	return nil
}

func (s *keysBDD) noRevoke() error {
	if len(s.revokes)+len(s.grantRevokes) != 0 {
		return fmt.Errorf("the broker received a revoke: %v %v", s.revokes, s.grantRevokes)
	}
	return nil
}

func (s *keysBDD) brokerOutcome(outcome string) error {
	if outcome == "receives no revoke" {
		return s.noRevoke()
	}
	id := strings.TrimPrefix(outcome, "receives a revoke of ")
	if len(s.revokes) != 1 || s.revokes[0] != id {
		return fmt.Errorf("revokes are %v, want one of %q", s.revokes, id)
	}
	return nil
}

func (s *keysBDD) grantCreate(name string) error {
	if len(s.grantCreates) != 1 || s.grantCreates[0] != (grantCreateCall{name, true}) {
		return fmt.Errorf("grant creates are %v, want one free grant named %q", s.grantCreates, name)
	}
	return nil
}

func (s *keysBDD) grantRevoked(name string) error {
	if len(s.grantRevokes) != 1 || s.grantRevokes[0] != name {
		return fmt.Errorf("grant revokes are %v, want one of %q", s.grantRevokes, name)
	}
	return nil
}

func (s *keysBDD) brokerNotAsked() error {
	if s.calls != 0 {
		return fmt.Errorf("the broker was asked %d times", s.calls)
	}
	return nil
}

// --- Then: status line, clipboard, transcript, log, help ----------------------------------

func (s *keysBDD) statusSays(text string) error { return has(s.status(), text) }

func (s *keysBDD) statusNotSay(text string) error {
	if strings.Contains(s.status(), text) {
		return fmt.Errorf("status says %q: %q", text, s.status())
	}
	return nil
}

func (s *keysBDD) statusNoSecret() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	return s.statusNotSay(sec)
}

func (s *keysBDD) statusUnreachable() error {
	return has(s.status(), client.ErrBrokerUnreachable.Error())
}

func (s *keysBDD) statusCarriesReason() error { return has(s.status(), grantRefusal) }

func (s *keysBDD) exportOnClipboard() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	want := "export OPENAI_BASE_URL=" + keysBDDBroker + "/v1\nexport OPENAI_API_KEY=" + sec
	if got := s.clipboard(); got != want {
		return fmt.Errorf("clipboard is %q, want the export block", got)
	}
	return nil
}

func (s *keysBDD) clipboardHolds(doc *godog.DocString) error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	want := strings.NewReplacer("<broker>", keysBDDBroker, "<secret>", sec).Replace(doc.Content)
	if got := s.clipboard(); got != want {
		return fmt.Errorf("clipboard is %q, want %q", got, want)
	}
	return nil
}

func (s *keysBDD) transcriptNoSecret() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	if strings.Contains(s.m.transcriptText(), sec) {
		return errors.New("the transcript holds the secret")
	}
	return nil
}

func (s *keysBDD) copyAllNoSecret() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	s.clip.Reset()
	tm, cmd := s.m.runSession("/copy all")
	s.m = asModel(tm)
	s.pump(cmd, 0)
	got := s.clipboard()
	if got == "" {
		return errors.New("/copy all copied nothing, so the check proves nothing")
	}
	if strings.Contains(got, sec) {
		return errors.New("/copy all put the secret on the clipboard")
	}
	return nil
}

// saveNoSecret: /save in the channel keeps the transcript in memory and writes no file today
// (update.go). So this runs /save and then checks both things a save could ever write from:
// the transcript, and every file under the config home the tests run in.
func (s *keysBDD) saveNoSecret() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	tm, cmd := s.m.runSession("/save")
	s.m = asModel(tm)
	s.pump(cmd, 0)
	if strings.Contains(s.m.transcriptText(), sec) {
		return errors.New("the transcript a save would write holds the secret")
	}
	return filepath.WalkDir(os.Getenv("HOME"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if raw, _ := os.ReadFile(p); bytes.Contains(raw, []byte(sec)) {
			return fmt.Errorf("%s holds the secret", p)
		}
		return nil
	})
}

func (s *keysBDD) logNoSecret() error {
	sec, err := s.lastSecret()
	if err != nil {
		return err
	}
	if s.m.mode != modeLog {
		return fmt.Errorf("mode is %v, want the log", s.m.mode)
	}
	if strings.Contains(s.flat(), sec) || strings.Contains(strings.Join(tuiLog.snapshot(), "\n"), sec) {
		return errors.New("the log holds the secret")
	}
	return nil
}

func (s *keysBDD) helpRow(cmd, text string) error {
	for _, ln := range strings.Split(stripANSI(s.m.helpView()), "\n") {
		if f := strings.Fields(ln); len(f) > 0 && f[0] == cmd && strings.Contains(ln, text) {
			return nil
		}
	}
	return fmt.Errorf("no help row for %q saying %q:\n%s", cmd, text, stripANSI(s.m.helpView()))
}

func (s *keysBDD) helpNotContain(text string) error {
	if strings.Contains(stripANSI(s.m.helpView()), text) {
		return fmt.Errorf("the help contains %q", text)
	}
	return nil
}

func (s *keysBDD) suggestionsInclude(name string) error {
	if s.m.mode != modeCommand || s.m.cmd.Value() == "" {
		return fmt.Errorf("the command prompt is not open with text (mode %v, %q)", s.m.mode, s.m.cmd.Value())
	}
	if err := has(s.screen(), name); err != nil {
		return err
	}
	for _, c := range paletteMatch(s.m.cmd.Value()) {
		if c.name == name {
			return nil
		}
	}
	return fmt.Errorf("typing /ke does not suggest %q", name)
}

func (s *keysBDD) plateNotContain(a, b string) error {
	out := stripANSI(s.m.endpointBlock(100))
	if err := has(out, "127.0.0.1:4141"); err != nil {
		return err
	}
	if strings.Contains(out, a) || strings.Contains(out, b) {
		return fmt.Errorf("the endpoint plate points at account keys:\n%s", out)
	}
	return nil
}

func TestAccountKeysBDD(t *testing.T) {
	st := &keysBDD{t: t}
	t.Cleanup(func() {
		clipboardOut = io.Discard
		SetLANDial(nil)
	})
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				st.reset()
				return ctx, nil
			})
			// Given
			sc.Step(`^the TUI is running against a broker$`, st.running)
			sc.Step(`^the account is logged in$`, st.loggedIn)
			sc.Step(`^the account is logged out$`, st.loggedOut)
			sc.Step(`^the TUI was started without the account key hooks$`, st.withoutHooks)
			sc.Step(`^the TUI is in LAN mode$`, st.lanMode)
			sc.Step(`^the terminal is 60 columns wide$`, st.width60)
			sc.Step(`^the operator is in the channel$`, st.inTheChannel)
			sc.Step(`^a channel is open$`, st.channelOpen)
			sc.Step(`^the account has the account keys$`, st.keysTable)
			sc.Step(`^the account has the grants$`, st.grantsTable)
			sc.Step(`^the account has no account keys and no grants$`, st.nothingYet)
			sc.Step(`^the account has an account key with limit ([\d.]+) and reset "([^"]*)"$`, st.keyWithLimit)
			sc.Step(`^the account has an account key with reset "([^"]*)" that spent 1 today, 2 this week, 3 this month and 4 in all$`, st.keyThatSpent)
			sc.Step(`^the account has an account key that is (disabled and expired|disabled|expired|neither)$`, st.keyThatIs)
			sc.Step(`^the account has an account key named "([^"]*)"$`, st.keyNamed)
			sc.Step(`^the account has an account key "([^"]*)" named "([^"]*)"$`, st.keyIDNamed)
			sc.Step(`^the account has account keys "([^"]*)" and "([^"]*)" both named "([^"]*)"$`, st.twoKeysNamed)
			sc.Step(`^the account has an account key whose name is 128 characters long$`, st.keyWithLongName)
			sc.Step(`^the broker returns an account key whose name holds an escape sequence$`, st.keyWithEscapeName)
			sc.Step(`^the account has account keys and grants$`, st.keysAndGrants)
			sc.Step(`^the account already has 32 account keys$`, st.thirtyTwoKeys)
			sc.Step(`^the account has a grant named "([^"]*)"$`, st.grantNamed)
			sc.Step(`^the broker is unreachable$`, st.brokerUnreachable)
			sc.Step(`^the broker is slow to answer$`, st.brokerSlow)
			sc.Step(`^the broker refuses grants for this account$`, st.brokerRefusesGrants)
			sc.Step(`^the broker answers the key list and fails the grant list$`, st.grantListFails)
			sc.Step(`^the operator minted a key and the secret is showing$`, st.mintedAndShowing)
			sc.Step(`^the operator has the KEYS panel open$`, st.panelOpen)
			sc.Step(`^a key named "([^"]*)" is minted on the site afterwards$`, st.lateKey)
			sc.Step(`^the operator was asked to confirm revoking "([^"]*)"$`, st.askedToConfirmKey)
			sc.Step(`^the operator confirmed revoking "([^"]*)"$`, st.confirmedRevokingKey)
			sc.Step(`^the operator was asked to confirm revoking grant "([^"]*)"$`, st.askedToConfirmGrant)
			// When
			sc.Step(`^the operator runs "([^"]*)" and then presses esc$`, st.runsThenEsc)
			sc.Step(`^the operator runs "([^"]*)" twice before the first answer$`, st.runsTwice)
			sc.Step(`^the operator closes the panel and runs "([^"]*)"$`, st.closesAndRuns)
			sc.Step(`^the operator runs "([^"]*)"$`, st.runs)
			sc.Step(`^the operator presses "([^"]*)"$`, st.pressesEach)
			sc.Step(`^the operator opens the help$`, st.opensHelp)
			sc.Step(`^the operator types "/ke"$`, func() error { return st.types("/ke") })
			// Then: the panel
			sc.Step(`^the KEYS panel is shown$`, st.panelShown)
			sc.Step(`^the panel has a section "([^"]*)" with one line per (key|grant)$`, st.sectionOneLinePer)
			sc.Step(`^the line for "([^"]*)" shows "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)"$`, st.lineShows5)
			sc.Step(`^the line for "([^"]*)" shows "([^"]*)" and "([^"]*)"$`, st.lineShows2)
			sc.Step(`^the line for "([^"]*)" shows its name, its limit and its state$`, st.lineShowsNameLimitState)
			sc.Step(`^the "([^"]*)" section says "([^"]*)"$`, st.sectionSays)
			sc.Step(`^the "([^"]*)" section lists "([^"]*)"$`, st.sectionLists)
			sc.Step(`^the "([^"]*)" section lists the keys$`, st.sectionListsTheKeys)
			sc.Step(`^the "([^"]*)" section carries the broker's reason$`, st.grantsCarryReason)
			sc.Step(`^the "([^"]*)" section says the grants could not be read$`, st.grantsCouldNotBeRead)
			sc.Step(`^that key's limit reads "([^"]*)"$`, st.thatKeyReads)
			sc.Step(`^that key's used amount reads "([^"]*)"$`, st.thatKeyUsed)
			sc.Step(`^that key's state reads "([^"]*)"$`, st.thatKeyState)
			sc.Step(`^the operator is back in the channel$`, st.backInChannel)
			sc.Step(`^the transcript has no new lines$`, st.transcriptUnchanged)
			sc.Step(`^no line is wider than the terminal$`, st.noLineTooWide)
			sc.Step(`^that key's line fits the panel width$`, st.longNameFits)
			sc.Step(`^the name ends with "([^"]*)"$`, st.nameEndsWith)
			sc.Step(`^the panel shows the name with the control characters removed$`, st.nameWithoutControls)
			sc.Step(`^nothing on screen starts with "([^"]*)" or "([^"]*)"$`, st.noSecretPrefixes)
			sc.Step(`^the KEYS panel is shown with the new key in the list$`, st.panelShownWithNewKey)
			sc.Step(`^the (?:KEYS )?panel shows the secret once under "([^"]*)"$`, st.showsSecretUnder)
			sc.Step(`^the panel says "([^"]*)"$`, st.panelSays)
			sc.Step(`^the KEYS panel says "([^"]*)"$`, st.panelSays)
			sc.Step(`^every character of the secret is on screen$`, st.everyCharOnScreen)
			sc.Step(`^no secret is shown$`, st.noSecretShown)
			sc.Step(`^the secret is not on screen$`, st.secretNotOnScreen)
			sc.Step(`^the first secret is not on screen$`, st.firstSecretNotOnScreen)
			sc.Step(`^the new key is in the list with its hint$`, st.newKeyListedWithHint)
			sc.Step(`^the KEYS panel shows the key as the broker returned it$`, st.panelShowsKeyAsReturned)
			sc.Step(`^the KEYS panel no longer lists "([^"]*)"$`, st.panelNoLongerLists)
			sc.Step(`^the panel asks "([^"]*)"$`, st.panelAsks)
			sc.Step(`^the panel asks to confirm revoking "([^"]*)"$`, st.asksToConfirmKey)
			sc.Step(`^the panel asks to confirm revoking grant "([^"]*)"$`, st.asksToConfirmGrant)
			sc.Step(`^the KEYS panel shows "([^"]*)" as "([^"]*)"$`, st.panelShowsGrantAs)
			// Then: the broker
			sc.Step(`^the broker receives a mint with name "([^"]*)", limit_usd ([\d.]+) and reset "([^"]*)"$`, st.mintNameLimitReset)
			sc.Step(`^the broker receives a mint with name "([^"]*)" and limit_usd ([\d.]+)$`, st.mintNameLimit)
			sc.Step(`^the broker receives a mint with limit_usd ([\d.]+)$`, st.mintLimit)
			sc.Step(`^the broker receives no mint$`, st.noMint)
			sc.Step(`^the broker receives one mint$`, st.exactlyOneMint)
			sc.Step(`^the broker receives a change to "([^"]*)" with only (.+)$`, st.changeWithOnly)
			sc.Step(`^the broker receives a change to "([^"]*)"$`, st.changeTo)
			sc.Step(`^the broker receives a change to that key$`, st.changeToThatKey)
			sc.Step(`^the broker receives no change$`, st.noChange)
			sc.Step(`^the broker (receives a revoke of \S+|receives no revoke)$`, st.brokerOutcome)
			sc.Step(`^the broker receives a grant create with name "([^"]*)" and free true$`, st.grantCreate)
			sc.Step(`^the broker receives a revoke of grant "([^"]*)"$`, st.grantRevoked)
			sc.Step(`^the broker is not asked$`, st.brokerNotAsked)
			sc.Step(`^the broker is asked for no secrets$`, st.brokerNotAsked)
			// Then: status line, clipboard, transcript, log, help
			sc.Step(`^the status line says "(.*)"$`, st.statusSays)
			sc.Step(`^the status line does not say "([^"]*)"$`, st.statusNotSay)
			sc.Step(`^the status line does not contain the secret$`, st.statusNoSecret)
			sc.Step(`^the status line says the broker could not be reached$`, st.statusUnreachable)
			sc.Step(`^the status line carries the broker's reason$`, st.statusCarriesReason)
			sc.Step(`^the export block is on the clipboard$`, st.exportOnClipboard)
			sc.Step(`^the clipboard holds exactly these lines, with the broker's address and the new secret$`, st.clipboardHolds)
			sc.Step(`^the transcript does not contain the secret$`, st.transcriptNoSecret)
			sc.Step(`^"/copy all" does not put the secret on the clipboard$`, st.copyAllNoSecret)
			sc.Step(`^"/save" does not put the secret anywhere$`, st.saveNoSecret)
			sc.Step(`^the operator runs "([^"]*)" and the answer has not come back$`, st.runsUnanswered)
			sc.Step(`^the operator runs "([^"]*)", closes the panel, and the answer comes back$`, st.runsClosesAnswered)
			sc.Step(`^the operator runs "([^"]*)" and then "([^"]*)" before the first answer$`, st.runsBoth)
			sc.Step(`^the KEYS panel is not shown$`, st.panelNotShown)
			sc.Step(`^the account has (\d+) account keys and the terminal is (\d+) rows tall$`, st.manyKeysShortTerminal)
			sc.Step(`^key (\d+) is (not )?on screen$`, st.keyOnScreen)
			sc.Step(`^the panel does not say "([^"]*)"$`, st.panelDoesNotSay)
			sc.Step(`^the panel says how many rows are below$`, st.saysRowsBelow)
			sc.Step(`^the screen is no taller than the terminal$`, st.screenFitsHeight)
			sc.Step(`^the "([^"]*)" hint "([^"]*)" is on screen$`, st.hintOnScreen)
			sc.Step(`^the clipboard does not hold a secret$`, st.clipboardNoSecret)
			sc.Step(`^the account has a revoked grant named "([^"]*)"$`, st.revokedGrantNamed)
			sc.Step(`^the broker receives (one mint and no grant create|one grant create and no mint)$`, st.oneAndNoOther)
			sc.Step(`^the log does not contain the secret$`, st.logNoSecret)
			sc.Step(`^the help has a row for "([^"]*)" that says "([^"]*)"$`, st.helpRow)
			sc.Step(`^the help does not contain "([^"]*)"$`, st.helpNotContain)
			sc.Step(`^the command suggestions include "([^"]*)"$`, st.suggestionsInclude)
			sc.Step(`^the endpoint plate does not contain "([^"]*)" or "([^"]*)"$`, st.plateNotContain)
		},
		Options: &godog.Options{Format: "progress", Paths: []string{"../../features/tui/account_keys.feature"}, TestingT: t, Strict: true},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("account key scenarios failed (see godog output above)")
	}
}
