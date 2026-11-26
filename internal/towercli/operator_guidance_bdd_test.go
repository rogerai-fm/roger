package towercli

// Contract: features/tower/tower_operator_guidance.feature.
//
// The operator half of tower_operator_guidance.feature: `roger tower init`, `register`, `help`,
// `serve` and `status` tell the operator that RogerAI approves every joined Tower, who must act
// in each state, and what to do next, in the same words in every command. The admin half (the
// pending list mark and the suspension email) runs in cmd/rogerai-broker. The public-docs
// scenario reads the site's source here, because the words are the contract.
//
// serve is driven for real (runLink against the package's Core stub): a state Core reports on
// the session is announced, and a state Core refuses with is said by the refusal path.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
	"rogerai.fm/roger/v7/internal/tower"
	"rogerai.fm/roger/v7/internal/towercore/link"
	"rogerai.fm/roger/v7/internal/towerjoin"
)

// approvalSentence is the one sentence every joined-Tower setup command says.
const approvalSentence = "RogerAI reviews and approves every joined Tower before it carries public traffic"

type guidanceState struct {
	t         *testing.T
	initOut   string
	regOut    string
	helpOut   string
	state     string
	reason    string
	since     time.Time
	serveOut  string
	statusOut string
	restored  bool
	dir       string
	towerID   string
	core      *coreStub
	regDir    string
}

func (s *guidanceState) initThenRegister() error {
	s.regDir = filepath.Join(s.t.TempDir(), "j")
	out, err := runCLI(s.t, "init", "--dir", s.regDir, "--mode", "joined")
	if err != nil {
		return err
	}
	s.initOut = out
	prev := registerFn
	registerFn = func(st *tower.State, _ towerjoin.Account) error {
		return os.WriteFile(filepath.Join(st.Dir(), "admission.json"), []byte(`{"tower_id":"tw-9"}`), 0o600)
	}
	defer func() { registerFn = prev }()
	out, err = runCLI(s.t, "register", "--dir", s.regDir)
	s.regOut = out
	return err
}

func (s *guidanceState) eachSaysApproval() error {
	for name, out := range map[string]string{"init": wrappedJoin(s.initOut), "register": wrappedJoin(s.regOut)} {
		if !strings.Contains(out, approvalSentence) {
			return fmt.Errorf("%s does not say %q:\n%s", name, approvalSentence, out)
		}
	}
	return nil
}

func (s *guidanceState) registerSaysWaiting() error {
	for _, want := range []string{"now waiting for that approval", "tw-9", "roger tower status"} {
		if !strings.Contains(wrappedJoin(s.regOut), want) {
			return fmt.Errorf("register does not say %q:\n%s", want, s.regOut)
		}
	}
	return nil
}

func (s *guidanceState) initStandalone() error {
	out, err := runCLI(s.t, "init", "--dir", filepath.Join(s.t.TempDir(), "s"), "--mode", "standalone")
	s.initOut = out
	return err
}

func (s *guidanceState) standaloneNeedsNoApproval() error {
	const want = "A standalone Tower serves its own network and needs no approval from RogerAI"
	if !strings.Contains(wrappedJoin(s.initOut), want) {
		return fmt.Errorf("standalone init does not say %q:\n%s", want, s.initOut)
	}
	return nil
}

func (s *guidanceState) runsHelp() error {
	out, err := runCLI(s.t, "help")
	s.helpOut = out
	return err
}

func (s *guidanceState) helpSaysApproval() error {
	for _, line := range strings.Split(s.helpOut, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "register") {
			if strings.Contains(line, "RogerAI's approval") {
				return nil
			}
			return fmt.Errorf("the register line does not say it needs RogerAI's approval: %q", line)
		}
	}
	return fmt.Errorf("help has no register line:\n%s", s.helpOut)
}

// --- the state table -----------------------------------------------------------------------

func (s *guidanceState) joinedInState(state string) error {
	s.state = state
	if s.since.IsZero() {
		s.since = time.Now().Add(-2 * time.Hour)
	}
	return nil
}

// statusReply is what Core's /tower/status says for the Tower in s.state.
func (s *guidanceState) statusReply() string {
	state, lease := s.state, time.Now().Add(time.Hour).Unix()
	if state == "lease_lapsed" {
		state, lease = "active", time.Now().Add(-time.Hour).Unix()
	}
	entry := map[string]any{
		"tower_id": "tw-1", "state": state, "may_take_work": state == "active",
		"link_live": state != "revoked", "lease_expires": lease, "carries_traffic": true,
		"state_since": s.since.Unix(),
	}
	if s.reason != "" {
		entry["state_reason"] = s.reason
	}
	raw, _ := json.Marshal(map[string]any{"towers": []any{entry}})
	return string(raw)
}

func (s *guidanceState) runStatus() error {
	s.core = newCoreStub(s.t)
	s.dir = joinedRegisteredTower(s.t)
	body := s.statusReply()
	s.core.reply["/tower/status"] = func(w http.ResponseWriter, _ int) bool {
		_, _ = w.Write([]byte(body))
		return true
	}
	out, err := runCLI(s.t, "status", "--dir", s.dir)
	s.statusOut = out
	return err
}

// runServe drives the real link loop: a state Core reports on the session is announced; a
// state Core refuses with is said by the refusal path.
func (s *guidanceState) runServe() error {
	core := newCoreStub(s.t)
	st := servingTower(s.t)
	refusal := map[string]string{"suspended": "state_suspended", "lease_lapsed": "lease_lapsed",
		"revoked": "state_revoked"}[s.state]
	state := s.state
	core.reply["/tower/session"] = func(w http.ResponseWriter, _ int) bool {
		if refusal != "" {
			refuse(w, refusal)
			return true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"session_id": "sess-1", "heartbeat_seconds": 30, "freshness_seconds": 180, "state": state,
		})
		return true
	}
	prevWait := serveWait
	serveWait = func(time.Duration) <-chan time.Time { return make(chan time.Time) }
	defer func() { serveWait = prevWait }()

	out := &syncBuffer{}
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- runLink(st, out, stop, manualTicker(make(chan time.Time)), link.RelayPlane{})
	}()
	deadline := time.After(10 * time.Second)
	for {
		select {
		case err := <-done:
			s.serveOut = out.String()
			if err != nil {
				s.serveOut += "\n" + err.Error()
			}
			return nil
		case <-deadline:
			close(stop)
			return fmt.Errorf("serve said nothing about %s:\n%s", s.state, out.String())
		case <-time.After(10 * time.Millisecond):
		}
		if strings.Contains(out.String(), "next step:") {
			close(stop)
			<-done
			s.serveOut = out.String()
			return nil
		}
	}
}

func (s *guidanceState) runsServeOrStatus() error {
	if err := s.runServe(); err != nil {
		return err
	}
	return s.runStatus()
}

func (s *guidanceState) stateLineSays(meaning string) error {
	if s.state == "suspended" {
		meaning = "stopped by RogerAI"
	}
	for name, out := range map[string]string{"serve": wrappedJoin(s.serveOut), "status": wrappedJoin(s.statusOut)} {
		if !strings.Contains(out, meaning) {
			return fmt.Errorf("%s does not say %q for %s:\n%s", name, meaning, s.state, out)
		}
	}
	return nil
}

func (s *guidanceState) nextStepSays(next string) error {
	for name, out := range map[string]string{"serve": wrappedJoin(s.serveOut), "status": wrappedJoin(s.statusOut)} {
		if !strings.Contains(out, next) {
			return fmt.Errorf("%s does not give the next step %q for %s:\n%s", name, next, s.state, out)
		}
	}
	return nil
}

// sameWords: the state line and the next-step line each command prints carry the same text
// after their labels, read with the 80-column wrap undone.
func (s *guidanceState) sameWords() error {
	for _, label := range []string{"state:", "next step:"} {
		a, b := guideLine(s.serveOut, label), guideLine(s.statusOut, label)
		if a == "" || a != b {
			return fmt.Errorf("serve and status disagree on %q:\n serve:  %q\n status: %q", label, a, b)
		}
	}
	return nil
}

func (s *guidanceState) suspendedWithReason(reason string) error {
	s.state, s.reason = "suspended", reason
	return s.runStatus()
}

func (s *guidanceState) reasonInPlainWords() error {
	if !strings.Contains(wrappedJoin(s.statusOut), "its results failed RogerAI's checks far more often than the rest of the network's") {
		return fmt.Errorf("the reputation reason is not in plain words:\n%s", s.statusOut)
	}
	return nil
}

func (s *guidanceState) noInternalCounters() error {
	for _, bad := range []string{"canary_fail", "canary_pass", "audit_mismatch", "evidence", "reputation"} {
		if strings.Contains(s.statusOut, bad) {
			return fmt.Errorf("status shows %q:\n%s", bad, s.statusOut)
		}
	}
	return nil
}

func (s *guidanceState) restoredToQuarantine() error {
	s.state, s.restored = "quarantine", true
	return s.runServe()
}

func (s *guidanceState) quarantineNotFault() error {
	if !strings.Contains(wrappedJoin(s.serveOut), "waiting for RogerAI to approve") {
		return fmt.Errorf("serve does not say waiting for RogerAI to approve:\n%s", s.serveOut)
	}
	for _, bad := range []string{"error", "failed", "fault:"} {
		if strings.Contains(strings.ToLower(s.serveOut), bad) {
			return fmt.Errorf("serve reads as a fault (%q):\n%s", bad, s.serveOut)
		}
	}
	return nil
}

func (s *guidanceState) inQuarantineFor(days int) error {
	s.state, s.since = "quarantine", time.Now().Add(-time.Duration(days)*24*time.Hour-time.Minute)
	return nil
}

func (s *guidanceState) runsStatus() error { return s.runStatus() }

func (s *guidanceState) statusSays(want string) error {
	if !strings.Contains(wrappedJoin(s.statusOut), want) {
		return fmt.Errorf("status does not say %q:\n%s", want, s.statusOut)
	}
	return nil
}

// The public-docs scenario ("The Tower page and the manual describe approval") is checked once,
// by web/test/tower-approval.test.mjs, which CI runs with the rest of the web tests.

func TestTowerOperatorGuidanceBDD(t *testing.T) {
	s := &guidanceState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*s = guidanceState{t: t}
				return ctx, nil
			})
			sc.Step("^the operator runs `roger tower init` for a joined Tower, and then `roger tower register`$", s.initThenRegister)
			sc.Step(`^each says that RogerAI reviews and approves every joined Tower before it carries public traffic$`, s.eachSaysApproval)
			sc.Step("^register says the Tower is now waiting for that approval, names the tower id to quote, and how to check: `roger tower status`$", s.registerSaysWaiting)
			sc.Step(`^the operator initializes a standalone Tower$`, s.initStandalone)
			sc.Step(`^it says a standalone Tower serves its own network and needs no approval from RogerAI$`, s.standaloneNeedsNoApproval)
			sc.Step("^the operator runs `roger tower help`$", s.runsHelp)
			sc.Step(`^the joined-Tower lines say approval by RogerAI is required, per Tower$`, s.helpSaysApproval)

			sc.Step(`^a joined Tower in state "([^"]*)"$`, s.joinedInState)
			sc.Step("^the operator runs `roger tower serve` or `roger tower status`$", s.runsServeOrStatus)
			sc.Step(`^the state line says "([^"]*)"$`, s.stateLineSays)
			sc.Step(`^the next-step line says "([^"]*)"$`, s.nextStepSays)
			sc.Step(`^both commands use the same words$`, s.sameWords)
			sc.Step(`^Core suspended the Tower with reason "([^"]*)"$`, s.suspendedWithReason)
			sc.Step(`^the suspended line names that reason in plain words$`, s.reasonInPlainWords)
			sc.Step(`^it does not show internal counters or evidence ids$`, s.noInternalCounters)
			sc.Step(`^an administrator moved a suspended Tower back to quarantine$`, s.restoredToQuarantine)
			sc.Step(`^serve announces quarantine with the "waiting for RogerAI to approve" words, not as a fault$`, s.quarantineNotFault)
			sc.Step(`^a Tower has been in quarantine for (\d+) days$`, s.inQuarantineFor)
			sc.Step(`^the operator runs status$`, s.runsStatus)
			sc.Step(`^it says "([^"]*)"$`, s.statusSays)

		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/tower_operator_guidance.feature",
				bddrun.Titled(true,
					"The Tower page and the manual describe approval",
					"Approving a restored Tower is the same one action as approving a new one",
					"A suspension tells the admin why, so re-approval is an informed decision"))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("tower_operator_guidance.feature (operator half) failed")
	}
}
