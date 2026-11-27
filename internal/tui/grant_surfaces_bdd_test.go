package tui

// grant_surfaces_bdd_test.go - the godog harness for the TUI Rule of
// features/grants/grant_surfaces.feature: /grant add, /grant added and /grant remove, and that
// a grant secret never reaches the transcript, the status line, the input history or a saved
// file (rulings S3, S7, S10, S11). The steps drive the REAL model: lines are typed into the
// command palette and sent with enter, keys go through Update, and the screen is View. The
// only stand-in is the broker, behind the Hooks seam the host wires to the client; it counts
// every call, so "no request reached the broker" reads what was really asked.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cucumber/godog"
	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/client"
)

// gsLimitErr is the broker's limit refusal as the client reports it: errors.Is matches
// client.ErrGrantAddLimit and the text is the broker's own.
type gsLimitErr struct{}

func (gsLimitErr) Error() string {
	return "an account can hold 10 grants - remove one first (`roger grant remove <name>`)"
}
func (gsLimitErr) Is(target error) bool { return target == client.ErrGrantAddLimit }

type grantSurfacesBDD struct {
	m    model
	clip *bytes.Buffer

	// the broker stand-in
	bySecret  map[string]client.AddedGrant
	added     []client.AddedGrant
	signedOut bool

	calls       int
	createCalls int
	statuses    []string // every status line the operator saw
}

const gsFriendsSecret = "rog-grant_FRIENDS0a1b2c3d4e5f60718293a4b5c6d7e8f9"

func (s *grantSurfacesBDD) reset() {
	*s = grantSurfacesBDD{clip: &bytes.Buffer{}, bySecret: map[string]client.AddedGrant{}}
	clipboardOut = s.clip
	SetLANDial(nil)
}

func (s *grantSurfacesBDD) hooks() Hooks {
	return Hooks{
		GrantList: func(string) ([]GrantRow, error) { s.calls++; return nil, nil },
		GrantCreate: func(string, string, bool) (string, error) {
			s.calls++
			s.createCalls++
			return "rog-grant_FREE00000000000000000000000000000000", nil
		},
		KeysList:    func(string) ([]client.AccountKey, error) { s.calls++; return nil, nil },
		KeysMint:    func(string, map[string]any) (client.AccountKey, string, error) { return client.AccountKey{}, "", nil },
		KeysSet:     func(string, string, map[string]any) (client.AccountKey, error) { return client.AccountKey{}, nil },
		KeysRevoke:  func(string, string) error { return nil },
		GrantRevoke: func(string, string) error { return nil },
		GrantAdd: func(_, secret string) (client.AddedGrant, error) {
			s.calls++
			if s.signedOut {
				return client.AddedGrant{}, client.ErrNotSignedIn
			}
			g, ok := s.bySecret[strings.TrimSpace(secret)]
			switch {
			case !ok:
				return client.AddedGrant{}, errors.New("grant key invalid or revoked")
			case len(s.added) >= 10:
				return client.AddedGrant{}, gsLimitErr{}
			}
			g.State, g.AddedAt = "active", 1795478400 // 2026-11-24
			s.added = append(s.added, g)
			return g, nil
		},
		GrantAddedList: func(string) ([]client.AddedGrant, int, error) {
			s.calls++
			if s.signedOut {
				return nil, 0, client.ErrNotSignedIn
			}
			return append([]client.AddedGrant(nil), s.added...), 10, nil
		},
		GrantRemove: func(_, ref string) (client.AddedGrant, error) {
			s.calls++
			if s.signedOut {
				return client.AddedGrant{}, client.ErrNotSignedIn
			}
			for i, g := range s.added {
				if g.ID == ref || g.Name == ref {
					s.added = append(s.added[:i:i], s.added[i+1:]...)
					return g, nil
				}
			}
			return client.AddedGrant{}, fmt.Errorf("no grant named %q in your added grants", ref)
		},
	}
}

// --- driving the model -------------------------------------------------------------------

func (s *grantSurfacesBDD) pump(cmd tea.Cmd, depth int) {
	s.statuses = append(s.statuses, stripANSI(s.m.status))
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
		return
	}
	tm, next := s.m.Update(msg)
	s.m = asModel(tm)
	s.pump(next, depth+1)
}

func (s *grantSurfacesBDD) send(k tea.KeyMsg) {
	tm, cmd := s.m.Update(k)
	s.m = asModel(tm)
	s.pump(cmd, 0)
}

// typeLine types a line into the command palette and presses enter, as the operator does.
func (s *grantSurfacesBDD) typeLine(line string) {
	line = strings.ReplaceAll(line, "<friends secret>", gsFriendsSecret)
	if s.m.mode == modeKeys {
		s.send(tea.KeyMsg{Type: tea.KeyEsc})
	}
	s.send(keyMsg("/"))
	if s.m.mode != modeCommand {
		s.m.mode = modeCommand
	}
	s.m.cmd.SetValue(strings.TrimPrefix(line, "/"))
	s.send(tea.KeyMsg{Type: tea.KeyEnter})
}

func (s *grantSurfacesBDD) screen() string { return stripANSI(s.m.View()) }

// --- Given -------------------------------------------------------------------------------

func (s *grantSurfacesBDD) oliveMadeFriends() error {
	s.bySecret[gsFriendsSecret] = client.AddedGrant{
		ID: "g_friends01", Name: "friends", Price: "$0/$0.05", PriceOut: 0.05,
		Models:   []string{"qwen-small"},
		Stations: []client.AddedGrantStation{{ID: "olive-box", Models: []string{"qwen-small"}, PublicOut: 0.10}},
	}
	return nil
}

func (s *grantSurfacesBDD) frankLoggedIn() error {
	m := NewWithHooks(keysBDDBroker, "tester", nil, s.hooks())
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	s.m = asModel(tm)
	s.m.loggedIn, s.m.ghLogin = true, "frank"
	return nil
}

func (s *grantSurfacesBDD) frankLoggedOut() error {
	s.m.loggedIn, s.m.ghLogin, s.signedOut = false, "", true
	return nil
}

func (s *grantSurfacesBDD) frankAddedFriends() error {
	g := s.bySecret[gsFriendsSecret]
	g.State, g.AddedAt = "active", 1795478400
	s.added = append(s.added, g)
	return nil
}

func (s *grantSurfacesBDD) frankAddedInTUI() error {
	s.typeLine("/grant add <friends secret>")
	return s.grantIsAdded()
}

func (s *grantSurfacesBDD) frankAddedTen() error {
	for i := range 10 {
		s.added = append(s.added, client.AddedGrant{ID: fmt.Sprintf("g_other%02d", i), Name: fmt.Sprintf("other-%d", i), State: "active"})
	}
	return nil
}

// --- When --------------------------------------------------------------------------------

func (s *grantSurfacesBDD) types(_, line string) error { s.typeLine(line); return nil }

func (s *grantSurfacesBDD) pastesAndEnter() error {
	s.send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(gsFriendsSecret), Paste: true})
	s.send(tea.KeyMsg{Type: tea.KeyEnter})
	return nil
}

func (s *grantSurfacesBDD) typesAndAnswers(line, answer string) error {
	s.typeLine(line)
	if s.m.keys.confirm == nil {
		return fmt.Errorf("no question was asked after %q:\n%s", line, s.screen())
	}
	if !strings.Contains(s.screen(), "y/N") {
		return fmt.Errorf("the question does not default to no:\n%s", s.screen())
	}
	if answer == "yes" {
		s.send(keyMsg("y"))
	} else {
		s.send(keyMsg("n"))
	}
	return nil
}

func (s *grantSurfacesBDD) addsInTUI() error { s.typeLine("/grant add <friends secret>"); return nil }

func (s *grantSurfacesBDD) runsSave() error {
	tm, cmd := s.m.runSession("/save")
	s.m = asModel(tm)
	s.pump(cmd, 0)
	return nil
}

// --- Then --------------------------------------------------------------------------------

func (s *grantSurfacesBDD) hiddenInputAsks() error {
	if s.m.mode != modeKeys || !s.m.keys.adding {
		return fmt.Errorf("no secret input is open (mode %v):\n%s", s.m.mode, s.screen())
	}
	if s.m.keys.addIn.EchoMode != textinput.EchoPassword {
		return errors.New("the secret input echoes what is typed")
	}
	return has(s.screen(), "grant secret")
}

func (s *grantSurfacesBDD) panelShowsFriends() error {
	if s.m.mode != modeKeys {
		return fmt.Errorf("the KEYS panel is not open (mode %v)", s.m.mode)
	}
	scr := s.screen()
	for _, want := range []string{"friends", "$0.05", "olive-box", "qwen-small"} {
		if err := has(scr, want); err != nil {
			return err
		}
	}
	return nil
}

func (s *grantSurfacesBDD) secretNowhere() error {
	sec := gsFriendsSecret
	if strings.Contains(s.m.transcriptText(), sec) {
		return errors.New("the transcript holds the secret")
	}
	for _, st := range s.statuses {
		if strings.Contains(st, sec) {
			return fmt.Errorf("a status line held the secret: %q", st)
		}
	}
	if strings.Contains(squash(s.screen()), sec) {
		return errors.New("the screen shows the secret")
	}
	if strings.Contains(strings.Join(tuiLog.snapshot(), "\n"), sec) {
		return errors.New("the log holds the secret")
	}
	return nil
}

func (s *grantSurfacesBDD) grantIsAdded() error {
	for _, g := range s.added {
		if g.Name == "friends" {
			return nil
		}
	}
	return fmt.Errorf("friends was not added; status %q", stripANSI(s.m.status))
}

func (s *grantSurfacesBDD) upDoesNotRecall() error {
	if s.m.mode == modeKeys {
		s.send(tea.KeyMsg{Type: tea.KeyEsc})
	}
	s.send(keyMsg("/"))
	if s.m.mode != modeCommand {
		return fmt.Errorf("the command input did not open (mode %v)", s.m.mode)
	}
	for range 5 {
		s.send(tea.KeyMsg{Type: tea.KeyUp})
		if strings.Contains(s.m.cmd.Value(), gsFriendsSecret) {
			return errors.New("up brought the secret back into the input")
		}
	}
	return nil
}

func (s *grantSurfacesBDD) noFileHoldsSecret() error {
	if strings.Contains(s.m.transcriptText(), gsFriendsSecret) {
		return errors.New("the transcript a save would write holds the secret")
	}
	return filepath.WalkDir(os.Getenv("HOME"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if raw, _ := os.ReadFile(p); bytes.Contains(raw, []byte(gsFriendsSecret)) {
			return fmt.Errorf("%s holds the secret", p)
		}
		return nil
	})
}

func (s *grantSurfacesBDD) addedSection(name, price, state string) error {
	scr := s.screen()
	i := strings.Index(strings.ToLower(scr), "grants you added")
	if s.m.mode != modeKeys || i < 0 {
		return fmt.Errorf("no \"grants you added\" section in the KEYS panel:\n%s", scr)
	}
	for _, ln := range strings.Split(scr[i:], "\n") {
		if f := strings.Fields(ln); len(f) > 0 && f[0] == name && strings.Contains(ln, price) && strings.Contains(ln, state) {
			return nil
		}
	}
	return fmt.Errorf("no row for %q with %s and %q under grants you added:\n%s", name, price, state, scr[i:])
}

func (s *grantSurfacesBDD) stillHas(name string) error {
	for _, g := range s.added {
		if g.Name == name {
			return nil
		}
	}
	return fmt.Errorf("frank no longer has %q", name)
}

func (s *grantSurfacesBDD) noLongerHas(name string) error {
	if s.stillHas(name) == nil {
		return fmt.Errorf("frank still has %q", name)
	}
	return nil
}

func (s *grantSurfacesBDD) refusalNames(want string) error {
	st := stripANSI(s.m.status)
	if err := has(st, want); err != nil {
		return err
	}
	if strings.Contains(st, "roger grant remove") {
		return fmt.Errorf("the TUI refusal names the CLI command: %q", st)
	}
	return nil
}

func (s *grantSurfacesBDD) saysLoginNothingSent(want string) error {
	if err := has(stripANSI(s.m.status), want); err != nil {
		return err
	}
	if s.calls != 0 {
		return fmt.Errorf("%d requests reached the broker", s.calls)
	}
	return nil
}

func (s *grantSurfacesBDD) pricedElsewhere() error {
	st := stripANSI(s.m.status)
	for _, want := range []string{"priced grants", "`roger grant create`", "console"} {
		if err := has(st, want); err != nil {
			return err
		}
	}
	if s.createCalls != 0 {
		return errors.New("a grant was created")
	}
	return nil
}

func TestGrantSurfacesTUIBDD(t *testing.T) {
	st := &grantSurfacesBDD{}
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
			// Background
			sc.Step(`^olive made a grant "friends" at \$0\.05 per 1M output tokens on "olive-box"$`, st.oliveMadeFriends)
			sc.Step(`^frank is logged in$`, st.frankLoggedIn)
			// Given
			sc.Step(`^frank is logged out$`, st.frankLoggedOut)
			sc.Step(`^frank added "friends" in the TUI$`, st.frankAddedInTUI)
			sc.Step(`^frank added "friends"$`, st.frankAddedFriends)
			sc.Step(`^frank has added 10 grants$`, st.frankAddedTen)
			// When
			sc.Step(`^(frank|olive) types "([^"]*)"(?: in the TUI)?$`, st.types)
			sc.Step(`^frank pastes the "friends" secret and presses enter$`, st.pastesAndEnter)
			sc.Step(`^frank types "([^"]*)" and answers (yes|no)$`, st.typesAndAnswers)
			sc.Step(`^frank adds the "friends" secret in the TUI$`, st.addsInTUI)
			sc.Step(`^frank runs /save$`, st.runsSave)
			// Then
			sc.Step(`^a hidden input asks for the grant secret$`, st.hiddenInputAsks)
			sc.Step(`^the KEYS panel shows "friends" at \$0\.05 with its stations and models$`, st.panelShowsFriends)
			sc.Step(`^the transcript and status line never contain the secret$`, st.secretNowhere)
			sc.Step(`^the grant is added$`, st.grantIsAdded)
			sc.Step(`^pressing up in the input does not bring the secret back$`, st.upDoesNotRecall)
			sc.Step(`^no saved file contains the secret$`, st.noFileHoldsSecret)
			sc.Step(`^the KEYS panel has a "grants you added" section with "([^"]*)", (\$[0-9.]+) and "([^"]*)"$`, st.addedSection)
			sc.Step(`^frank still has "([^"]*)"$`, st.stillHas)
			sc.Step(`^frank no longer has "([^"]*)"$`, st.noLongerHas)
			sc.Step(`^the refusal names "([^"]*)"$`, st.refusalNames)
			sc.Step(`^the TUI says "([^"]*)" and no request reached the broker$`, st.saysLoginNothingSent)
			sc.Step("^the TUI says priced grants are made with `roger grant create` or the console$", st.pricedElsewhere)
		},
		Options: &godog.Options{
			Format: "progress", TestingT: t, Strict: true,
			Paths: bddrun.ScenariosInRule("../../features/grants/grant_surfaces.feature", "The TUI"),
		},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("grant surface TUI scenarios failed (see godog output above)")
	}
}
