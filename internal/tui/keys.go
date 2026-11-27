package tui

// keys.go: the KEYS panel - account keys (rog-key_...) and grants (rog-grant_...) in one
// place, with `/keys` and the `/grant` verbs that feed it (features/tui/account_keys.feature).
//
// The two kinds, as the panel says them: an account key is for you or your own bot, works on
// any public station and pays from your wallet. A grant is for someone else, works on your
// stations only and is free by default.
//
// A NEW SECRET lives only in this panel's state. It is copied to the clipboard and drawn
// once at the bottom of the panel (the rows a short terminal keeps), never on the status
// line and never in a transcript, and it is dropped when the panel closes or any other
// /keys or /grant command runs. Nothing here logs or stores it.
//
// A command that will ask a question or show a secret opens the panel when it is typed, so
// the operator is already there when the answer lands. An answer for a panel that was
// closed in the meantime is not shown: a question is never asked over other work, where a
// stray key could answer it.

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"rogerai.fm/roger/v7/internal/client"
)

const (
	keysLoginFirst  = "log in first - /login (keys are per account)"
	keysNoLimitNote = client.KeyUnlimitedNote
	keysNameW       = 24 // the widest a name is drawn before it is cut with an ellipsis
)

// keysPanel is the KEYS panel's state (model.keys).
type keysPanel struct {
	prev     mode // where esc returns to
	keys     []client.AccountKey
	keysErr  string
	grants   []GrantRow
	grantErr string
	added    []client.AddedGrant // grants this account added (grant_adds.go)
	addedErr string

	secret     string // a new secret, shown once; "" when none is showing
	secretKind string // "key" or "grant"
	noLimit    bool   // the showing key was minted with no limit
	minting    string // the name of a key mint or grant create the broker has not answered yet
	confirm    *keysConfirm

	adding bool            // the hidden grant-secret input is open (/grant add)
	addIn  textinput.Model // that input; its value is a secret and is never drawn

	scroll int // the first list row drawn, when the lists are taller than the terminal

	visit int // counts the times the panel closed: an answer for an older visit is not shown
	loads int // counts list reads: only the newest one lands
}

// keysLimitHelp is the client's refusal of an unreadable limit, with the TUI's word for none.
var keysLimitHelp = client.ErrKeyLimit.Error() + `, or "none" for unlimited`

// keysConfirm is a revoke waiting for its y/N.
type keysConfirm struct {
	grant    bool // a grant you made: revoke it for everyone
	added    bool // a grant you added: remove it from this account only
	id, name string
}

type (
	keysListMsg struct {
		load     int
		keys     []client.AccountKey
		keysErr  string
		grants   []GrantRow
		grantErr string
		added    []client.AddedGrant
		addedErr string
	}
	// keyMintedMsg is the broker's answer to a key mint or a grant create.
	keyMintedMsg struct {
		visit      int
		kind, name string // "key" or "grant", and what it was called
		secret     string
		noLimit    bool
		err        error
	}
	keysConfirmMsg struct {
		visit int
		keysConfirm
	}
	keysDoneMsg string // a change the broker accepted, as the line to show
)

// clearSecret drops a showing secret and any pending confirm (ruling K3).
func (p *keysPanel) clearSecret() {
	p.secret, p.secretKind, p.noLimit, p.confirm = "", "", false, nil
	p.adding, p.addIn = false, textinput.Model{} // a half-typed grant secret goes too
}

// keysErrText is a broker error in the TUI's words: the client names the CLI's login
// command, and from here the way in is /login. Whatever the broker said is drawn clean.
func keysErrText(err error) string {
	if errors.Is(err, client.ErrKeysLogin) {
		return keysLoginFirst
	}
	return cleanText(err.Error())
}

// enterKeys shows the panel, remembering where the operator was.
func (m *model) enterKeys() {
	if m.mode != modeKeys {
		m.keys.prev, m.mode, m.keys.scroll = m.mode, modeKeys, 0
	}
}

// closeKeys leaves the panel. A showing secret goes with it, and so does the right of any
// answer still on its way to be shown here.
func (m *model) closeKeys() {
	m.keys.clearSecret()
	m.keys.visit++
	m.mode = m.keys.prev
}

// openKeys shows the panel and reads both lists. Logged out it asks the broker nothing.
func (m model) openKeys() (tea.Model, tea.Cmd) {
	if LANMode() {
		m.status = stDim.Render("keys need the public network - not available in LAN mode")
		return m, nil
	}
	m.enterKeys()
	if !m.loggedInState() {
		return m, nil
	}
	return m, m.loadKeys()
}

// loadKeys reads the account keys and the grants. One list failing leaves the other shown.
// Reads can overlap, so each is numbered and only the newest lands.
func (m *model) loadKeys() tea.Cmd {
	m.keys.loads++
	broker, h, load := m.broker, m.hooks, m.keys.loads
	return func() tea.Msg {
		out := keysListMsg{load: load}
		var err error
		switch {
		case h.KeysList == nil:
			out.keysErr = "keys unavailable in this build - run `roger keys`"
		default:
			if out.keys, err = h.KeysList(broker); err != nil {
				out.keysErr = "could not read keys: " + keysErrText(err)
			}
		}
		switch {
		case h.GrantList == nil:
			out.grantErr = "grants unavailable in this build - run `roger grant list`"
		default:
			if out.grants, err = h.GrantList(broker); err != nil {
				out.grantErr = "could not read grants: " + cleanText(err.Error())
			}
		}
		if h.GrantAddedList != nil {
			if out.added, _, err = h.GrantAddedList(broker); err != nil {
				out.addedErr = "could not read the grants you added: " + grantAddErrText(err)
			}
		}
		return out
	}
}

// doKeys is `/keys`: list | mint | limit | on | off | revoke (rm and delete are the same).
func (m model) doKeys(args []string) (tea.Model, tea.Cmd) {
	m.keys.clearSecret()
	h := m.hooks
	if LANMode() {
		return m.openKeys() // says why, and asks nothing
	}
	if h.KeysList == nil || h.KeysMint == nil || h.KeysSet == nil || h.KeysRevoke == nil {
		m.status = stDim.Render("keys unavailable in this build - run `roger keys`")
		return m, nil
	}
	if len(args) == 0 || args[0] == "list" || args[0] == "ls" {
		return m.openKeys()
	}
	if !m.loggedInState() {
		m.status = stDim.Render(keysLoginFirst)
		return m, nil
	}
	verb, rest := args[0], args[1:]
	switch verb {
	case "mint", "new":
		return m.keysMint(rest)
	case "limit":
		if len(rest) == 0 || len(rest) > 3 {
			m.status = stDim.Render("usage: /keys limit <key> <limit> [daily|weekly|monthly]")
			return m, nil
		}
		limit, window, bad := parseKeyLimit(rest[1:])
		if bad != "" {
			m.status = stEmber.Render(bad)
			return m, nil
		}
		fields := map[string]any{"limit_usd": limit}
		if window != "" {
			fields["reset"] = window
		}
		m.enterKeys()
		return m, m.keysChange(rest[0], fields)
	case "on", "off":
		if len(rest) != 1 {
			m.status = stDim.Render("usage: /keys " + verb + " <key>")
			return m, nil
		}
		m.enterKeys()
		return m, m.keysChange(rest[0], map[string]any{"disabled": verb == "off"})
	case "rm", "revoke", "delete":
		if len(rest) != 1 {
			m.status = stDim.Render("usage: /keys revoke <key>")
			return m, nil
		}
		m.enterKeys()
		broker, arg, visit := m.broker, rest[0], m.keys.visit
		return m, func() tea.Msg {
			k, err := findKey(h.KeysList, broker, arg)
			if err != nil {
				return flowErrMsg(keysErrText(err))
			}
			return keysConfirmMsg{visit, keysConfirm{id: k.ID, name: k.Name}}
		}
	}
	m.status = stDim.Render("/keys: list, mint, limit, on, off, revoke")
	return m, nil
}

// parseKeyLimit reads `<limit> [window]`: the limit as `roger keys --limit` reads it
// (client.ParseKeyLimit), then daily, weekly, monthly or none. bad is the line to show when it
// cannot be read.
func parseKeyLimit(args []string) (limit float64, window, bad string) {
	if len(args) == 0 {
		return 0, "", keysLimitHelp
	}
	limit, err := client.ParseKeyLimit(args[0])
	switch {
	case errors.Is(err, client.ErrKeyLimit):
		return 0, "", keysLimitHelp
	case err != nil:
		return 0, "", err.Error()
	}
	if len(args) > 1 {
		if window = args[1]; !client.KeyResetOK(window) {
			return 0, "", "the window is daily, weekly, monthly or none"
		}
	}
	return limit, window, ""
}

// keysMint is `/keys mint <name> <limit> [window]`. A limit must be stated (ruling K1).
func (m model) keysMint(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 || len(args) > 3 {
		m.status = stDim.Render("usage: /keys mint <name> <limit> [daily|weekly|monthly]")
		return m, nil
	}
	limit, window, bad := parseKeyLimit(args[1:])
	if bad != "" {
		m.status = stEmber.Render(bad)
		return m, nil
	}
	if window == "" {
		window = "none"
	}
	mint := m.hooks.KeysMint
	return m.making("key", args[0], "minting "+cleanText(args[0])+"…", limit == 0, func(broker string) (string, error) {
		_, secret, err := mint(broker, map[string]any{"name": args[0], "limit_usd": limit, "reset": window})
		return secret, err
	})
}

// grantCreate is `/grant create [name]`: a free grant, its secret handled like a key's.
func (m model) grantCreate(args []string) (tea.Model, tea.Cmd) {
	create := m.hooks.GrantCreate
	if create == nil {
		m.status = stDim.Render("grants unavailable in this build - run `roger grant create`")
		return m, nil
	}
	// Free only (ruling K6, S7). A flag is never taken for a name.
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--price"):
			m.status = stDim.Render(grantPricedElse)
			return m, nil
		case strings.HasPrefix(a, "-"):
			m.status = stDim.Render("usage: /grant create [name] - free grants only")
			return m, nil
		}
	}
	name := "my-bots"
	if len(args) >= 1 {
		name = args[0]
	}
	return m.making("grant", name, "creating free grant "+cleanText(name)+"…", false, func(broker string) (string, error) {
		secret, err := create(broker, name, true)
		if err != nil {
			err = fmt.Errorf("grant create failed: %w · for your own bot, use /keys mint", err)
		}
		return secret, err
	})
}

// making starts the one thing that ends in a new secret. Only one runs at a time: two
// answers would each claim the screen and the clipboard, and one secret would be lost. The
// panel opens now, so the operator is there when the secret lands.
func (m model) making(kind, name, status string, noLimit bool, mk func(broker string) (string, error)) (tea.Model, tea.Cmd) {
	if m.keys.minting != "" {
		m.status = stDim.Render("still minting " + cleanText(m.keys.minting))
		return m, nil
	}
	m.enterKeys()
	m.keys.minting = name
	m.status = stDim.Render(status)
	broker, visit := m.broker, m.keys.visit
	return m, tea.Batch(m.loadKeys(), func() tea.Msg {
		secret, err := mk(broker)
		return keyMintedMsg{visit: visit, kind: kind, name: name, secret: secret, noLimit: noLimit, err: err}
	})
}

// findKey reads the key list fresh and matches arg against it as the CLI does
// (client.FindKey): an id first, then a name (ruling K5).
func findKey(list func(string) ([]client.AccountKey, error), broker, arg string) (client.AccountKey, error) {
	keys, err := list(broker)
	if err != nil {
		return client.AccountKey{}, err
	}
	k, err := client.FindKey(keys, arg)
	if errors.Is(err, client.ErrKeyNotFound) {
		err = fmt.Errorf("%w - /keys lists them", err)
	}
	return k, err
}

// keysChange sends only the named fields to the one key arg names.
func (m model) keysChange(arg string, fields map[string]any) tea.Cmd {
	broker, h := m.broker, m.hooks
	return func() tea.Msg {
		k, err := findKey(h.KeysList, broker, arg)
		if err == nil {
			_, err = h.KeysSet(broker, k.ID, fields)
		}
		if err != nil {
			return flowErrMsg(keysErrText(err))
		}
		done := "updated " + cleanName(k.Name)
		if limit, ok := fields["limit_usd"]; ok && limit == 0.0 {
			done += " - " + keysNoLimitNote // ruling K13
		}
		return keysDoneMsg(done)
	}
}

// keysRevoke carries out a confirmed revoke.
func (m model) keysRevoke(c keysConfirm) tea.Cmd {
	broker, h := m.broker, m.hooks
	return func() tea.Msg {
		done, err := "revoked "+cleanName(c.name), error(nil)
		switch {
		case c.added:
			done = "removed " + cleanName(c.name) + " from your account - its stations are back at their public price"
			if _, err = h.GrantRemove(broker, c.id); err != nil {
				return flowErrMsg(grantAddErrText(err))
			}
		case c.grant:
			done, err = "revoked grant "+cleanName(c.name), h.GrantRevoke(broker, c.name)
		default:
			err = h.KeysRevoke(broker, c.id)
		}
		if err != nil {
			return flowErrMsg(keysErrText(err))
		}
		return keysDoneMsg(done)
	}
}

// grantRevokeAsk is `/grant revoke <name>`: it checks a live grant has that name, then asks.
func (m model) grantRevokeAsk(args []string) (tea.Model, tea.Cmd) {
	h := m.hooks
	if h.GrantList == nil || h.GrantRevoke == nil {
		m.status = stDim.Render("grants unavailable in this build - run `roger grant revoke`")
		return m, nil
	}
	if len(args) != 1 {
		m.status = stDim.Render("usage: /grant revoke <name>")
		return m, nil
	}
	m.enterKeys()
	broker, name, visit := m.broker, args[0], m.keys.visit
	return m, func() tea.Msg {
		rows, err := h.GrantList(broker)
		if err != nil {
			return flowErrMsg(keysErrText(err))
		}
		revoked := false
		for _, g := range rows {
			if g.Name == name && g.Status != "revoked" {
				return keysConfirmMsg{visit, keysConfirm{grant: true, name: name}}
			}
			revoked = revoked || g.Name == name
		}
		if revoked {
			return flowErrMsg("grant " + cleanText(name) + " is already revoked")
		}
		return flowErrMsg("no grant named " + cleanText(name) + " - /grant lists them")
	}
}

// onKeysMsg lands the panel's own messages.
func (m model) onKeysMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case keysListMsg:
		if msg.load == m.keys.loads {
			m.keys.keys, m.keys.keysErr, m.keys.grants, m.keys.grantErr = msg.keys, msg.keysErr, msg.grants, msg.grantErr
			m.keys.added, m.keys.addedErr = msg.added, msg.addedErr
		}
		return m, nil
	case grantAddedMsg:
		return m.onGrantAdded(msg)
	case keyMintedMsg:
		return m.onMinted(msg)
	case keysConfirmMsg:
		// A question nobody is there to read is not asked: a key pressed for something else
		// must never answer it.
		if !m.keysWaiting(msg.visit) {
			return m, nil
		}
		m.keys.confirm = &msg.keysConfirm
		m.status = "" // the question is the only thing to read now
		return m, m.loadKeys()
	}
	m.status = stLive.Render(string(msg.(keysDoneMsg)))
	return m, m.reloadKeys()
}

// keysWaiting reports whether the panel is still open on the visit an answer was asked from.
func (m model) keysWaiting(visit int) bool { return m.mode == modeKeys && visit == m.keys.visit }

// reloadKeys reads the lists again if the panel is there to show them.
func (m *model) reloadKeys() tea.Cmd {
	if m.mode != modeKeys {
		return nil
	}
	return m.loadKeys()
}

// onMinted lands a key mint or a grant create. The secret goes in the panel and its export
// block on the clipboard; the status line says only that something was copied. If the
// operator closed the panel while it was on its way, the secret is dropped (ruling K3) and
// the status line says what to do about the key or grant that now exists.
func (m model) onMinted(msg keyMintedMsg) (tea.Model, tea.Cmd) {
	m.keys.minting = ""
	name := cleanName(msg.name)
	undo := "/keys revoke " + name
	if msg.kind == "grant" {
		undo = "/grant revoke " + name
	}
	switch {
	case msg.err != nil:
		m.status = stEmber.Render("! " + keysErrText(msg.err))
		return m, nil
	case msg.secret == "":
		m.status = stEmber.Render("! the broker made " + name + " but returned no secret - " + undo + ", then make another")
	case !m.keysWaiting(msg.visit):
		m.status = stEmber.Render("! " + name + " was made while the panel was closed, so its secret was dropped - " + undo + ", then make another")
	default:
		m.keys.secret, m.keys.secretKind, m.keys.noLimit = msg.secret, msg.kind, msg.noLimit
		m.status = copiedToast("the new " + msg.kind + ", as export lines")
		return m, tea.Batch(clipboardWrite(client.KeyExportLines(m.broker, msg.secret)), m.loadKeys())
	}
	return m, m.reloadKeys()
}

// onKeysKey owns every key while the panel is open: y answers a revoke, esc closes, and the
// arrow and page keys scroll lists too tall for the terminal. A "/" or ":" closes it too and
// goes on to start the command it begins.
func (m model) onKeysKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.keys.adding {
		return m.onGrantAddKey(k)
	}
	// The scroll keys move the lists and nothing else: they never answer a revoke question.
	if page, top := m.keysPage(); top > 0 || m.keys.scroll > 0 {
		step, ok := map[string]int{"up": -1, "down": 1, "pgup": -page, "pgdown": page, "home": -top, "end": top}[k.String()]
		if ok {
			m.keys.scroll = max(0, min(top, m.keys.scroll+step))
			return m, nil
		}
	}
	if c := m.keys.confirm; c != nil {
		m.keys.confirm = nil
		if s := k.String(); s != "y" && s != "Y" {
			m.status = stDim.Render("kept " + cleanName(c.name))
			return m, nil
		}
		return m, m.keysRevoke(*c)
	}
	switch k.String() {
	case "esc":
		m.closeKeys()
	case "/", ":":
		m.closeKeys()
		return m.Update(k)
	}
	return m, nil
}

// cleanText drops control and invisible formatting characters (escape sequences, bidi
// overrides), so nothing the broker returns can draw on the terminal.
func cleanText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}

// cleanName is cleanText for a name, which may come out empty.
func cleanName(s string) string {
	if s = cleanText(s); strings.TrimSpace(s) != "" {
		return s
	}
	return "(no name)"
}

// panelRow draws one row: the name, cut to nameW, then the first set of columns that fits
// the width. The last set is cut to the width if even it does not fit.
func panelRow(w, nameW int, name string, colSets ...[]string) string {
	name = truncVisibleTail(cleanName(name), nameW)
	head := "  " + stKey.Render(name) + strings.Repeat(" ", nameW-lipgloss.Width(name)) + "  "
	var ln string
	for _, cols := range colSets {
		for i := range cols {
			cols[i] = cleanText(cols[i])
		}
		ln = head + stDim.Render(strings.Join(cols, "  "))
		if lipgloss.Width(ln) <= w {
			return ln
		}
	}
	return truncVisible(ln, w)
}

// keysView draws the panel in the rows the chrome leaves it (0: all of it): both lists, then
// at the bottom a pending revoke question or a new secret. The bottom is always drawn whole,
// whatever the login state reads, so a secret that exists is never left off the screen. Lists
// too tall for the rows left are drawn as a window that the scroll keys move, with a line
// above and below saying how much is hidden.
func (m model) keysView(w, rows int) string {
	lists, tail := m.keysParts(w)
	page, top := keysWindow(len(lists), len(tail), rows)
	if top > 0 {
		at := min(m.keys.scroll, top)
		above, below := "", ""
		if at > 0 {
			above = "  " + stDim.Render(fmt.Sprintf("↑ %d more above", at))
		}
		if left := top - at; left > 0 {
			below = "  " + stDim.Render(fmt.Sprintf("↓ %d more below", left))
		}
		lists = append(append([]string{above}, lists[at:at+page]...), below)
	}
	return strings.Join(append(lists, tail...), "\n")
}

// keysWindow sizes the list window: page is how many list rows are drawn at once and top is
// the furthest the window scrolls (0 when everything fits, or no height is known).
func keysWindow(lists, tail, rows int) (page, top int) {
	if rows <= 0 || lists+tail <= rows {
		return lists, 0
	}
	page = max(1, rows-tail-2) // two rows say what is hidden above and below
	return page, max(0, lists-page)
}

// keysPage is keysWindow for the panel as it would be drawn now, for the scroll keys.
func (m model) keysPage() (page, top int) {
	w := m.effWidth()
	out, _, _, _ := m.layout(w)
	lists, tail := m.keysParts(w)
	return keysWindow(len(lists), len(tail), m.fitRows(out))
}

// keysParts builds the panel's lines: the lists, which may scroll, and the tail under them
// (a revoke question, a new secret), which never does.
func (m model) keysParts(w int) (lists, tail []string) {
	var b strings.Builder
	prose := func(style lipgloss.Style, text string) {
		for _, ln := range wordWrap(text, w-4) {
			b.WriteString("  " + style.Render(ln) + "\n")
		}
	}
	b.WriteString("  " + stBrand.Render("KEYS") + stDim.Render("   ") + stKey.Render("esc") + stDim.Render(" close") + "\n\n")
	p := m.keys
	if !m.loggedInState() {
		prose(stDim, "log in to manage keys - /login")
	} else {
		nameW := 4
		for _, k := range p.keys {
			nameW = max(nameW, lipgloss.Width(cleanName(k.Name)))
		}
		for _, g := range p.grants {
			nameW = max(nameW, lipgloss.Width(cleanName(g.Name)))
		}
		for _, g := range p.added {
			nameW = max(nameW, lipgloss.Width(cleanName(g.Name)))
		}
		nameW = min(nameW, min(keysNameW, max(4, w-8)))

		b.WriteString("  " + stBrand.Render("ACCOUNT KEYS") + "\n")
		prose(stDim, "for you or your own bot - any public station - pays from your wallet")
		switch {
		case p.keysErr != "":
			prose(stEmber, p.keysErr)
		case len(p.keys) == 0:
			prose(stDim, "none yet - /keys mint <name> <limit> [daily|weekly|monthly]")
		}
		for _, k := range p.keys {
			limit, used := k.LimitText(), k.UsedText()
			b.WriteString(panelRow(w, nameW, k.Name,
				[]string{k.ID, k.Hint, limit, used, k.State()},
				[]string{limit, used, k.State()},
				[]string{limit, k.State()}) + "\n")
		}

		b.WriteString("\n  " + stBrand.Render("GRANTS") + "\n")
		prose(stDim, "for someone else - your stations only - free by default")
		switch {
		case p.grantErr != "":
			prose(stEmber, p.grantErr)
		case len(p.grants) == 0:
			prose(stDim, "none yet - /grant create <name>")
		}
		for _, g := range p.grants {
			b.WriteString(panelRow(w, nameW, g.Name, []string{g.Price, g.Status}) + "\n")
		}

		if m.hooks.GrantAddedList != nil || len(p.added) > 0 {
			b.WriteString("\n  " + stBrand.Render("GRANTS YOU ADDED") + "\n")
			prose(stDim, "someone else's grant - their stations - you pay their price")
			switch {
			case p.addedErr != "":
				prose(stEmber, p.addedErr)
			case len(p.added) == 0:
				prose(stDim, "none yet - /grant add, then paste the secret")
			}
			for _, ln := range addedRows(w, nameW, p.added) {
				b.WriteString(ln + "\n")
			}
		}
	}

	lists = strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	b.Reset()
	if c := p.confirm; c != nil {
		b.WriteString("\n")
		if c.added {
			prose(stEmber, "remove "+cleanName(c.name)+" from your account? its stations go back to their public price for you. y/N")
		} else if c.grant {
			prose(stEmber, "revoke grant "+cleanName(c.name)+"? anyone using it loses access. y/N")
		} else {
			prose(stEmber, "revoke "+cleanName(c.name)+" ("+cleanText(c.id)+")? anything using it stops working. y/N")
		}
	}
	if p.adding {
		b.WriteString("\n")
		prose(stLive, "paste the grant secret - it is not shown - enter adds it, esc cancels")
		b.WriteString(p.addIn.View() + "\n")
	}
	if p.secret != "" {
		b.WriteString("\n")
		prose(stLive, "new "+p.secretKind+" - store it now, it is not shown again")
		for r, step := []rune(p.secret), max(8, w-6); len(r) > 0; r = r[min(step, len(r)):] {
			b.WriteString("    " + stKey.Render(string(r[:min(step, len(r))])) + "\n")
		}
		if p.noLimit {
			prose(stEmber, keysNoLimitNote)
		}
	}
	if b.Len() > 0 {
		tail = strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	}
	return lists, tail
}
