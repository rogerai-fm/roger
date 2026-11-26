package towercli

// Contract: features/tower/tower_operator_guidance.feature.
//
// Review fixes for the operator guidance (2026-11-24): a reason Core sends is never echoed
// unless it is one this binary knows; a suspended or revoked Tower whose lease also ran out
// reads as suspended or revoked in both serve and status; serve and status agree on the state
// line as well as the next step, the suspension reason included; every guidance line fits in
// 80 columns with a real-length tower id.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/towercore/link"
	"rogerai.fm/roger/v7/internal/towerjoin"
)

// refuseWithReason answers like the broker's towerCallerOr for a request verified as the
// Tower's own: the state, and for a suspension its allow-listed reason.
func refuseWithReason(w http.ResponseWriter, state, reason string) {
	w.WriteHeader(http.StatusForbidden)
	meta := map[string]any{"tower_state": state}
	if reason != "" {
		meta["tower_reason"] = reason
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{
		"code": "tower_refused", "message": "this link requires a registered Tower's own signed request",
		"metadata": meta}})
}

// realID is a tower id of real length: "tw-" and 24 hex digits.
const realID = "tw-3081d0565b48999cdb34fbce"

// wrappedJoin undoes the 80-column wrap: a continuation line (indented 4 or more) joins the line
// before it, so a test reads a wrapped sentence as one.
func wrappedJoin(s string) string { return regexp.MustCompile(`\n {4,}`).ReplaceAllString(s, " ") }

// guideLine is the text after label on its (joined) line.
func guideLine(out, label string) string {
	m := regexp.MustCompile(`(?m)` + regexp.QuoteMeta(label) + `\s*(.*)$`).FindStringSubmatch(wrappedJoin(out))
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func TestAnUnknownReasonIsNeverEchoed(t *testing.T) {
	for _, reason := range []string{"administrator", "canary_fail=5 ev-123 see ticket 42", "some_new_code"} {
		g, ok := guideFor("suspended", reason, realID)
		require.True(t, ok)
		require.Contains(t, g.meaning, "suspended by a RogerAI administrator", reason)
		require.NotContains(t, g.meaning, "canary_fail")
		require.NotContains(t, g.meaning, "ticket")
		require.NotContains(t, g.meaning, "some_new_code")
	}
	g, _ := guideFor("suspended", "reputation", realID)
	require.Contains(t, g.meaning, "its results failed RogerAI's checks far more often than the rest of the network's")
	g, _ = guideFor("suspended", "owner_banned", realID)
	require.Contains(t, g.meaning, "the account that owns it is banned")
}

// serveAndStatus drives the real serve loop and the real status command against one Core stub,
// for a Tower Core refuses with refusal (and reason) and reports as statusState (with the lease
// expired when lapsed is set).
func serveAndStatus(t *testing.T, refusal, reason, statusState string, lapsed bool) (serveOut, statusOut string) {
	t.Helper()
	s := &guidanceState{t: t, state: statusState, reason: reason, since: time.Now().Add(-time.Hour)}
	core := newCoreStub(t)
	st := servingTower(t)
	core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool {
		refuseWithReason(w, refusal, reason)
		return true
	}
	prevWait := serveWait
	serveWait = func(time.Duration) <-chan time.Time { return make(chan time.Time) }
	t.Cleanup(func() { serveWait = prevWait })
	out := &syncBuffer{}
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- runLink(st, out, stop, manualTicker(make(chan time.Time)), link.RelayPlane{}) }()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case err := <-done:
			serveOut = out.String()
			if err != nil {
				serveOut += "\n" + err.Error()
			}
			goto status
		default:
		}
		if strings.Contains(out.String(), "next step:") || time.Now().After(deadline) {
			close(stop)
			<-done
			serveOut = out.String()
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
status:
	lease := time.Now().Add(time.Hour).Unix()
	if lapsed {
		lease = time.Now().Add(-time.Hour).Unix()
	}
	entry := map[string]any{"tower_id": "tw-1", "state": statusState, "lease_expires": lease,
		"link_live": false, "carries_traffic": true, "state_since": s.since.Unix()}
	if reason != "" {
		entry["state_reason"] = reason
	}
	raw, _ := json.Marshal(map[string]any{"towers": []any{entry}})
	dir := joinedRegisteredTower(t)
	core.reply["/tower/status"] = func(w http.ResponseWriter, _ int) bool {
		_, _ = w.Write(raw)
		return true
	}
	o, err := runCLI(t, "status", "--dir", dir)
	require.NoError(t, err)
	return serveOut, o
}

func TestServeAndStatusAgreeIncludingExpiredLeases(t *testing.T) {
	cases := []struct {
		name, refusal, reason, status, label string
		lapsed                               bool
	}{
		{"suspended", "state_suspended", "", "suspended", "SUSPENDED", false},
		{"suspended with reason", "state_suspended", "reputation", "suspended", "SUSPENDED", false},
		{"suspended, lease expired", "state_suspended", "", "suspended", "SUSPENDED", true},
		{"revoked, lease expired", "state_revoked", "", "revoked", "REVOKED", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			serve, status := serveAndStatus(t, c.refusal, c.reason, c.status, c.lapsed)
			for _, label := range []string{"state:", "next step:"} {
				a, b := guideLine(serve, label), guideLine(status, label)
				require.NotEmpty(t, a, "serve has no %q line:\n%s", label, serve)
				require.Equal(t, a, b, "serve and status disagree on %q\nserve:\n%s\nstatus:\n%s", label, serve, status)
			}
			require.True(t, strings.HasPrefix(guideLine(status, "state:"), c.label), status)
			require.NotContains(t, status, "LAPSED", "a %s Tower must not read as lapsed", c.status)
		})
	}
}

// Every guidance line fits in 80 columns, with a real-length tower id, in every state, for
// every reason, and for every refusal serve can meet.
func TestGuidanceLinesFitIn80Columns(t *testing.T) {
	var all bytes.Buffer
	for _, state := range []string{"quarantine", "active", "draining", "suspended", "lease_lapsed", "revoked"} {
		for _, reason := range []string{"", "reputation", "owner_banned", "administrator"} {
			g, ok := guideFor(state, reason, realID)
			require.True(t, ok)
			writeGuide(&all, "  ", g)
			fmt.Fprintln(&all, refusalWords(g))
			last := ""
			announceStateFor(&all, &last, state, realID)
		}
	}
	for _, err := range []error{
		towerjoin.ErrUnreachable,
		&towerjoin.Refusal{Status: 403, State: "state_suspended", Reason: "reputation"},
		&towerjoin.Refusal{Status: 403, State: "lease_lapsed"},
		&towerjoin.Refusal{Status: 403, State: "state_revoked"},
		&towerjoin.Refusal{Status: 403, State: "cert_revoked"},
		&towerjoin.Refusal{Status: 403, State: "state_expired"},
		&towerjoin.Refusal{Status: 403, State: "state_draining"},
		&towerjoin.Refusal{Status: 403},
		&towerjoin.Refusal{Status: 429},
		&towerjoin.Refusal{Status: 503},
	} {
		_, words := classifyRefusal(err, realID)
		fmt.Fprintln(&all, words)
	}
	fmt.Fprintln(&all, wrapPlain(approvalLine))
	fmt.Fprintln(&all, wrapPlain(standaloneLine))
	for _, line := range strings.Split(all.String(), "\n") {
		require.LessOrEqual(t, len([]rune(line)), 80, "line over 80 columns: %q", line)
	}
	out, err := runCLI(t, "help")
	require.NoError(t, err)
	require.Contains(t, out, "register  submit a joined Tower for RogerAI's approval (after login): --dir DIR")
}

func TestTheNewWords(t *testing.T) {
	g, _ := guideFor("quarantine", "", realID)
	require.Equal(t, "Nothing is broken. RogerAI approves each Tower and this terminal will say when.", g.also)
	g, _ = guideFor("active", "", realID)
	require.Equal(t, "Stations can attach now. Each job prints here as it settles.", g.also)
	g, _ = guideFor("lease_lapsed", "", realID)
	require.Equal(t, "LAPSED", g.label)
	require.Contains(t, wrappedJoin(refusalWords(g)), "approval has lapsed")
	_, words := classifyRefusal(&towerjoin.Refusal{Status: 503}, realID)
	require.Contains(t, wrappedJoin(words), "RogerAI is having trouble right now (error 503)")
	// The fallback state line has the same layout as every known state.
	var b bytes.Buffer
	last := ""
	announceStateFor(&b, &last, "some-future-state", realID)
	require.Equal(t, "state:     some-future-state\n", b.String())
}

func TestStatusSaysRogerAISays(t *testing.T) {
	_, status := serveAndStatus(t, "state_suspended", "", "suspended", false)
	require.Contains(t, status, "RogerAI says:")
	require.NotContains(t, status, "Roger Core says")
}
