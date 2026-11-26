package towercli

// Contract: features/tower/tower_operator_guidance.feature.
//
// ONE SOURCE OF WORDS FOR WHERE A TOWER STANDS. `serve` and `status` used to describe the same
// state in two different ways, and neither said the two things an operator most needs: who has
// to act (RogerAI approves every joined Tower, again after a suspension or a lapse) and what to
// do in the meantime (nothing but keep it running, in every state RogerAI can lift). The first
// outside operator had to write in to ask, and the house Tower restart-looped for 17 hours on
// 2026-11-24 because nothing said a suspension waits on RogerAI rather than on a restart.
//
// Every line fits in 80 columns: a next step breaks at each "; " onto continuation lines, and
// any longer sentence wraps under a hanging indent.

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// termWidth is the width every guidance line fits in.
const termWidth = 80

// approvalLine is said by init and register for a joined Tower.
const approvalLine = "RogerAI reviews and approves every joined Tower before it carries public traffic, " +
	"and again after a suspension or a lapse."

// standaloneLine is said by init for a standalone Tower.
const standaloneLine = "A standalone Tower serves its own network and needs no approval from RogerAI."

// towerPage is where an operator learns more and asks RogerAI anything, so no mailbox is written
// into the binary (the spec's rule).
const towerPage = "https://rogerai.fm/tower"

// stateGuide is what a state means and what the operator does next.
type stateGuide struct {
	label, meaning, next string
	// also is one more line some states carry: the plain fact behind the state.
	also string
}

// guideFor describes state ("lease_lapsed" for a lease that ran out) with the reason code Core
// gave for a suspension and the id to quote. ok is false for a state this binary does not know.
func guideFor(state, reason, towerID string) (stateGuide, bool) {
	quote := quoteID(towerID)
	switch state {
	case "quarantine":
		return stateGuide{label: "QUARANTINE", meaning: "waiting for RogerAI to approve this Tower",
			next: "keep this running; approval switches it on by itself; questions: " + towerPage + ", " + quote,
			also: "Nothing is broken. RogerAI approves each Tower and this terminal will say when."}, true
	case "active":
		return stateGuide{label: "ACTIVE", meaning: "approved and carrying traffic", next: "nothing to do",
			also: "Stations can attach now. Each job prints here as it settles."}, true
	case "draining":
		return stateGuide{label: "DRAINING", meaning: "finishing current work, taking no new work", next: "nothing to do"}, true
	case "suspended":
		meaning := "stopped by RogerAI"
		if w := reasonWords(reason); w != "" {
			meaning += ": " + w
		}
		return stateGuide{label: "SUSPENDED", meaning: meaning,
			next: "keep this running; RogerAI must review and approve it again; " + quote,
			also: "RogerAI has suspended this Tower, pending review."}, true
	case "lease_lapsed":
		return stateGuide{label: "LAPSED", meaning: "its approval lapsed and cannot be renewed from here",
			next: "keep this running; RogerAI must re-admit and approve it again; " + quote,
			also: "RogerAI says this Tower's approval has lapsed."}, true
	case "revoked":
		return stateGuide{label: "REVOKED", meaning: "permanently retired",
			next: "enroll a new Tower if you want to run one again",
			also: "RogerAI has revoked this Tower permanently."}, true
	}
	return stateGuide{}, false
}

// reasonWords turns the reason code Core gives for a suspension into plain words. Core sends a
// code from a fixed list; anything else, including any text that is not a code, is said as the
// one fixed phrase, so nothing an administrator typed can ever reach the operator through here.
func reasonWords(reason string) string {
	switch reason {
	case "":
		return ""
	case "reputation":
		return "its results failed RogerAI's checks far more often than the rest of the network's"
	case "owner_banned":
		return "the account that owns it is banned"
	case "cert_revoked":
		return "its certificate was revoked"
	}
	return "suspended by a RogerAI administrator"
}

// writeGuide prints a state's lines, wrapped to termWidth. prefix indents them under a heading.
func writeGuide(out io.Writer, prefix string, g stateGuide) {
	const stateHead, nextHead = "state:     ", "next step: "
	hang := prefix + strings.Repeat(" ", len(nextHead))
	for _, line := range wrap(prefix+stateHead, g.label+" - "+g.meaning, hang) {
		fmt.Fprintln(out, line)
	}
	steps := strings.Split(g.next, "; ")
	for i, step := range steps {
		if i < len(steps)-1 {
			step += ";"
		}
		head := hang
		if i == 0 {
			head = prefix + nextHead
		}
		for _, line := range wrap(head, step, hang) {
			fmt.Fprintln(out, line)
		}
	}
}

// wrap lays text out after head, breaking at spaces so no line passes termWidth, with each
// continuation line starting with hang.
func wrap(head, text, hang string) []string {
	var lines []string
	cur := head
	fresh := true // nothing after the head yet
	for _, word := range strings.Fields(text) {
		switch {
		case fresh:
			cur += word
			fresh = false
		case len([]rune(cur))+1+len([]rune(word)) <= termWidth:
			cur += " " + word
		default:
			lines = append(lines, cur)
			cur = hang + word
		}
	}
	return append(lines, cur)
}

// wrapPlain wraps a plain sentence for the terminal, continuation lines indented four spaces.
func wrapPlain(s string) string { return strings.Join(wrap("", s, "    "), "\n") }

// sinceWords is "in quarantine for 3 days": how long a Tower has been in its state.
func sinceWords(state string, since, now time.Time) string {
	word := map[string]string{"quarantine": "in quarantine", "lease_lapsed": "lapsed"}[state]
	if word == "" {
		word = state
	}
	return word + " for " + spanWords(now.Sub(since))
}

// spanWords is a duration the way a person says it: whole days, else hours, else minutes.
func spanWords(d time.Duration) string {
	plural := func(n int, unit string) string {
		if n == 1 {
			return "1 " + unit
		}
		return fmt.Sprintf("%d %ss", n, unit)
	}
	switch {
	case d >= 24*time.Hour:
		return plural(int(d/(24*time.Hour)), "day")
	case d >= time.Hour:
		return plural(int(d/time.Hour), "hour")
	default:
		return plural(int(d/time.Minute), "minute")
	}
}

// quoteID is the "quote the tower id" clause every next step that involves RogerAI ends with.
func quoteID(towerID string) string {
	if towerID == "" {
		return "quote the tower id"
	}
	return "quote the tower id " + towerID
}

// refusalWords is a refusal said the way every state is said: the plain fact, then the same
// state and next-step lines `status` prints, all wrapped.
func refusalWords(g stateGuide) string {
	var b strings.Builder
	if g.also != "" {
		fmt.Fprintln(&b, wrapPlain(g.also))
	}
	writeGuide(&b, "", g)
	return strings.TrimRight(b.String(), "\n")
}
