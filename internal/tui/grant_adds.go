package tui

// grant_adds.go: grants this account ADDED - someone else's priced grant, added by its secret -
// in the KEYS panel, with `/grant add`, `/grant added` and `/grant remove`
// (features/grants/grant_surfaces.feature, rulings S3, S7, S9, S10, S11).
//
// THE SECRET. `/grant add` with no secret opens a hidden input in the panel; what is typed
// there is never drawn, logged or written to the transcript, and the input is emptied the
// moment it is sent or cancelled. A secret typed inline (`/grant add rog-grant_...`) works
// too, and the command history keeps only "grant add" (scrubSecretArgs). The broker's reply
// carries no secret, so nothing after the add can hold one.

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"rogerai.fm/roger/v7/internal/client"
)

const (
	grantAddLoginFirst = "log in first: /login" // ruling S11
	grantPricedElse    = "priced grants are made with `roger grant create` or the console - /grant create makes free grants only"
)

// grantAddedMsg is the broker's answer to an add.
type grantAddedMsg struct {
	g   client.AddedGrant
	err error
}

// grantAddErrText is an add, list or remove error in the TUI's words: logged out names
// /login, and the limit refusal names /grant remove instead of the CLI command (S10).
func grantAddErrText(err error) string {
	switch {
	case errors.Is(err, client.ErrNotSignedIn):
		return grantAddLoginFirst
	case errors.Is(err, client.ErrGrantAddLimit):
		msg := strings.ReplaceAll(cleanText(err.Error()), "roger grant remove", "/grant remove")
		if !strings.Contains(msg, "/grant remove") {
			msg += " - /grant remove <name> drops one"
		}
		return msg
	}
	return cleanText(err.Error())
}

// grantAdd is `/grant add [secret]`. With no secret it opens the hidden input.
func (m model) grantAdd(args []string) (tea.Model, tea.Cmd) {
	if m.hooks.GrantAdd == nil {
		m.status = stDim.Render("grants unavailable in this build - run `roger grant add`")
		return m, nil
	}
	if !m.loggedInState() {
		m.status = stDim.Render(grantAddLoginFirst)
		return m, nil
	}
	m.enterKeys()
	if len(args) == 0 {
		in := textinput.New()
		in.EchoMode, in.EchoCharacter = textinput.EchoPassword, '•'
		in.Prompt = "  > "
		in.Focus()
		m.keys.adding, m.keys.addIn = true, in
		m.status = stDim.Render("paste the grant secret, then enter - esc cancels")
		return m, nil
	}
	return m.grantAddSend(strings.Join(args, " "))
}

// grantAddSend sends a secret to the broker. The status line never repeats it.
func (m model) grantAddSend(secret string) (tea.Model, tea.Cmd) {
	if strings.TrimSpace(secret) == "" {
		m.status = stEmber.Render("! paste the grant secret")
		return m, nil
	}
	m.status = stDim.Render("adding grant…")
	add, broker := m.hooks.GrantAdd, m.broker
	return m, func() tea.Msg {
		g, err := add(broker, secret)
		return grantAddedMsg{g: g, err: err}
	}
}

// onGrantAddKey owns every key while the hidden input is open: enter sends, esc cancels, and
// the rest are typing. Either way out empties the input.
func (m model) onGrantAddKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.keys.adding, m.keys.addIn = false, textinput.Model{}
		m.status = stDim.Render("no grant added")
		return m, nil
	case "enter":
		secret := m.keys.addIn.Value()
		m.keys.adding, m.keys.addIn = false, textinput.Model{}
		return m.grantAddSend(secret)
	}
	var c tea.Cmd
	m.keys.addIn, c = m.keys.addIn.Update(k)
	return m, c
}

// onGrantAdded lands an add: the grant goes on top of the panel with its stations.
func (m model) onGrantAdded(msg grantAddedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = stEmber.Render("! " + grantAddErrText(msg.err))
		return m, nil
	}
	// Shown at once with what the add returned (stations, models); the list read that
	// follows replaces it with the broker's own rows.
	g := msg.g
	if g.State == "" {
		g.State = "active"
	}
	rows := []client.AddedGrant{g}
	for _, o := range m.keys.added {
		if o.ID != g.ID {
			rows = append(rows, o)
		}
	}
	m.keys.added = rows
	m.status = stLive.Render("added " + cleanName(g.Name) + " - your requests to its stations are now charged at its price, from your own wallet")
	return m, m.reloadKeys()
}

// grantAddedOpen is `/grant added`: the panel, with the added list read fresh.
func (m model) grantAddedOpen() (tea.Model, tea.Cmd) {
	if m.hooks.GrantAddedList == nil {
		m.status = stDim.Render("grants unavailable in this build - run `roger grant list`")
		return m, nil
	}
	if !m.loggedInState() {
		m.status = stDim.Render(grantAddLoginFirst)
		return m, nil
	}
	return m.openKeys()
}

// grantRemoveAsk is `/grant remove <name|id>`: it finds the one added grant meant, then asks.
func (m model) grantRemoveAsk(args []string) (tea.Model, tea.Cmd) {
	h := m.hooks
	if h.GrantAddedList == nil || h.GrantRemove == nil {
		m.status = stDim.Render("grants unavailable in this build - run `roger grant remove`")
		return m, nil
	}
	if !m.loggedInState() {
		m.status = stDim.Render(grantAddLoginFirst)
		return m, nil
	}
	if len(args) != 1 {
		m.status = stDim.Render("usage: /grant remove <name|id>")
		return m, nil
	}
	m.enterKeys()
	broker, ref, visit := m.broker, args[0], m.keys.visit
	return m, func() tea.Msg {
		added, _, err := h.GrantAddedList(broker)
		if err != nil {
			return flowErrMsg(grantAddErrText(err))
		}
		var hits []client.AddedGrant
		for _, g := range added {
			if g.ID == ref {
				hits = []client.AddedGrant{g}
				break
			}
			if g.Name == ref {
				hits = append(hits, g)
			}
		}
		switch len(hits) {
		case 0:
			return flowErrMsg("no grant named " + cleanText(ref) + " in your added grants - /grant added lists them")
		case 1:
			return keysConfirmMsg{visit, keysConfirm{added: true, id: hits[0].ID, name: hits[0].Name}}
		}
		ids := make([]string, len(hits))
		for i, g := range hits {
			ids[i] = cleanText(g.ID)
		}
		return flowErrMsg(strconv.Itoa(len(hits)) + " added grants are named " + cleanText(ref) + " - /grant remove <id>: " + strings.Join(ids, ", "))
	}
}

// grantMoney is a per-1M-token price as the panel writes it.
func grantMoney(v float64) string { return "$" + strconv.FormatFloat(v, 'f', -1, 64) }

// addedPriceText is a grant's price per 1M tokens: output, and input when it has one.
func addedPriceText(in, out float64) string {
	switch {
	case in == 0 && out == 0:
		return "free"
	case in == 0:
		return grantMoney(out) + "/1M out"
	}
	return grantMoney(in) + " in " + grantMoney(out) + " out /1M"
}

// addedRows draws the "grants you added" section's grant rows: one per grant, then, for a
// live one, a row per station with its public price, and the models.
func addedRows(w, nameW int, gs []client.AddedGrant) []string {
	var out []string
	for _, g := range gs {
		price, state := addedPriceText(g.PriceIn, g.PriceOut), g.StateWords()
		full := []string{price, state}
		if g.AddedAt > 0 {
			full = append(full, "added "+time.Unix(g.AddedAt, 0).UTC().Format("2006-01-02"))
		}
		out = append(out, panelRow(w, nameW, g.Name, full, []string{price, state}))
		if state != "active" {
			continue
		}
		if len(g.Stations) == 0 {
			out = append(out, "      "+stDim.Render("covers no station yet - add it again later, when one is on air"))
		}
		for _, st := range g.Stations {
			ln := cleanText(st.ID) + "  public " + addedPriceText(st.PublicIn, st.PublicOut) + "  " + cleanText(strings.Join(st.Models, ", "))
			if st.AbovePublic {
				ln += "  - the grant price is above the public price"
			}
			out = append(out, "      "+stDim.Render(truncVisible(ln, max(8, w-6))))
		}
	}
	return out
}
