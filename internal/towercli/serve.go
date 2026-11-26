package towercli

// serve.go holds the joined relay link open.
//
// Before this, `roger-tower serve` told the operator the link had not shipped. Core's routes
// existed and were exercised only by tests speaking HTTP directly - a protocol with one
// participant. This is the other one.
//
// WHAT SERVING MEANS TODAY, said plainly here so the command can say it plainly too: the
// Tower registers, opens a session, pushes a signed inventory of whatever its Stations have
// signed, heartbeats, and drains cleanly on shutdown. It does NOT carry customer traffic:
// dispatch is not built.
//
// The inventory it pushes is typically EMPTY now: self-attached `roger share --tower` nodes
// register their offers directly with Core at attach, and the legacy leaf-offer files (from
// the retired roger-station binary) have no producer. An inventory of zero leaves is the
// honest "I am here"; the offers directory remains read for byte-for-byte relay of any
// legacy files an operator still carries, and Core excludes what nothing can serve.

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	randv2 "math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"rogerai.fm/roger/v7/internal/localplane/serve"
	"rogerai.fm/roger/v7/internal/netpriv"
	"rogerai.fm/roger/v7/internal/tower"
	"rogerai.fm/roger/v7/internal/towercore/link"
	"rogerai.fm/roger/v7/internal/towerjoin"
)

// serveJoined runs the link until the process is interrupted. It supplies the two things the
// loop cannot invent for itself - a real signal and a real clock - and then gets out of the
// way; everything that can go wrong is in runLink, where a test can reach it.
func serveJoined(st *tower.State, out io.Writer, relayPublic string, hub hubOptions) error {
	hub.Advertised = relayPublic
	// THE MODE CHECK COMES FIRST, BEFORE ANY LISTENER AND BEFORE ANY DIAL.
	//
	// runLink refuses a standalone Tower, and did so correctly - but it runs AFTER the hub, and
	// the hub's very first act is to fetch Roger Core's grant key and then this Tower's node
	// list from Core. So `roger-tower serve --hub` on a standalone data directory made two
	// public-network calls before being told it should not have. features/tower/modes.feature
	// says a standalone Tower "performs no RogerAI DNS lookup or network connection", and that
	// was true only of the flag combination nobody had tried.
	//
	// Signed hub polls make the hub's dependence on Core heavier still - the assertion keys it
	// verifies node polls against arrive on that same fetch - so the refusal moves ahead of it
	// rather than the fetch being made conditional. runLink keeps its own check: it is called
	// directly, and a guarantee this size should hold at both doors.
	if st.Mode != tower.ModeJoined {
		return terminalError{msg: errStandaloneCannotServeJoined.Error(), cause: errStandaloneCannotServeJoined}
	}
	// THE HUB'S CERTIFICATE IS RESOLVED HERE, BEFORE EITHER THE LISTENER OR THE LINK, because
	// both halves of the change depend on it and they must not be able to disagree. The
	// listener presents these bytes; the link advertises their fingerprint to Core, which hands
	// it to every node and consumer routed here. Resolving it in one place, once, is what makes
	// "the pin Core published is the certificate this hub presents" true by construction rather
	// than by two code paths happening to read the same file.
	//
	// AFTER THE MODE CHECK, deliberately. A standalone Tower must not so much as mint a key on
	// the strength of a joined-mode flag; it is refused above with nothing written.
	relay := link.RelayPlane{Endpoint: relayPublic}
	if hub.Addr != "" && hub.tlsWanted() {
		mat, terr := hubTLS(st.Dir(), st.TowerID, hub.TLSCert, hub.TLSKey)
		if terr != nil {
			return terr
		}
		hub.cert = &mat.Cert
		relay.TLSSPKI = mat.Pin
		fmt.Fprintf(out, "hub: TLS certificate pin %s - Roger Core publishes this to every node "+
			"and consumer it routes here, and they accept no other certificate\n", mat.Pin)
	}
	if relay.TLSSPKI == "" && relay.Endpoint != "" {
		// SAID AT THE TOWER, NOT ONLY AT THE NODE. The node has always printed a plaintext
		// notice, but the operator who could fix it never saw it: they run this process, not
		// somebody else's `roger share`. One line, at the moment the tower decides to advertise
		// a plaintext data plane.
		fmt.Fprint(out, "NOTE: this tower advertises a PLAINTEXT hub. The sealed job and its "+
			"answer stay private either way, but every poll puts a Station's long-term "+
			"assertion public key - its payment identity - on the wire in the clear, and "+
			"nothing authenticates this hub's answers to a node. Pass --hub-tls to close both.\n")
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	// Adapt the signal channel to the loop's plain one, so runLink has no opinion about where
	// "stop" came from and a test does not have to send its own process a signal to exercise
	// the shutdown path. A test that did would be racing the Go runtime for the delivery.
	stopped := make(chan struct{})
	var once sync.Once
	windDown := func() { once.Do(func() { close(stopped) }) }
	go func() {
		<-stop
		windDown()
	}()

	// THE HUB - the data plane (Option C, Topology 2): the tower-hosted job queue consumers
	// submit sealed work
	// to and self-attached roger share nodes poll. Started before the link so a consumer's
	// very first submit after we advertise has somewhere to land; fails fast if Core's grant
	// key cannot be fetched.
	hub.session = &sessionBox{}
	if hub.Addr != "" {
		waitForHub, herr := startHub(st, hub, out, stopped)
		if errors.Is(herr, errStoppedWaiting) {
			windDown()
			fmt.Fprintln(out, "\nstopping")
			return nil
		}
		if herr != nil {
			windDown()
			return herr
		}
		defer waitForHub()
	}
	// RENEWAL, alongside the link and the relay. Without it the certificate and the lease
	// both lapse in a day and the Tower is finished - re-enrollment through quarantine, for
	// an operator who did nothing wrong. It runs independently of the link for the same
	// reason the relay does: a control-plane blip must not also cost the credential.
	go towerjoin.KeepRenewed(st, out, stopped, realTicker)
	err := runLinkWith(st, out, stopped, realTicker, relay, hub.session)
	// THE LINK RETURNING WINDS EVERYTHING DOWN, error or not. Without this, a serve whose
	// link failed at startup - not registered, wrong mode - HUNG forever: the deferred
	// relay-wait was waiting on a stop signal only ctrl-c could send, over a loop whose
	// error had already been decided. Found by a test that timed out instead of failing.
	windDown()
	return err
}

// Contract: features/tower/tower_canary_offline_station.feature.
//
// startHub starts the hub, waiting out any failure to fetch Core's grant key - unreachable, a 5xx,
// a 429, a proxy's 404 mid-deploy, a body that is not a key: the endpoint is public, so none of
// them is a fault here, and a Tower that boots during a Core redeploy must come up when Core
// does, not exit and stay down. Only a fault on this machine is terminal - an address it cannot bind, TLS asked for
// with no certificate, an identity or admission it cannot read - and a bind failure says so.
func startHub(st *tower.State, hub hubOptions, out io.Writer, stop <-chan struct{}) (func(), error) {
	said := false
	for attempt := 0; ; attempt++ {
		waitForHub, err := runHubInBackground(st, hub, out, stop)
		switch {
		case err == nil:
			return waitForHub, nil
		case errors.Is(err, errGrantKey):
			if !said {
				said = true
				fmt.Fprintln(out, refusalWords(stateGuide{label: "WAITING",
					meaning: "could not reach RogerAI for its grant key, which the hub checks work with (" +
						strings.TrimPrefix(err.Error(), errGrantKey.Error()+": ") + ")",
					next: "keep this running; it asks again every 30 s, backing off to 15 min; " +
						"the hub starts by itself once RogerAI answers"}))
			}
			select {
			case <-stop:
				return nil, errStoppedWaiting
			case <-serveWait(linkBackoff(attempt)):
			}
		case strings.Contains(err.Error(), "cannot listen on"):
			return nil, terminalError{msg: refusalWords(stateGuide{label: "STOPPED",
				meaning: fmt.Sprintf("the hub cannot bind %s; another process is probably using that address", hub.Addr),
				next:    "free that address or pass a different --hub; then start again"}) + "\n  " + err.Error(), cause: err}
		default:
			return nil, terminalError{msg: fmt.Sprintf("the hub could not start: %v", err), cause: err}
		}
	}
}

// errStandaloneCannotServeJoined is the one refusal both doors give, so the two cannot drift
// into saying different things about the same rule.
var errStandaloneCannotServeJoined = errors.New(
	"this Tower is standalone: `roger tower serve --dir DIR` serves its consumer plane on its own " +
		"network (loopback by default). The link and the hub are for a JOINED Tower - initialize a " +
		"new data directory with --mode joined to join the public network")

func realTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// runLink is the link loop proper: open, push, heartbeat, re-open on refusal, drain on exit.
func runLink(st *tower.State, out io.Writer, stop <-chan struct{}, ticker func(time.Duration) (<-chan time.Time, func()), relay link.RelayPlane) error {
	return runLinkWith(st, out, stop, ticker, relay, nil)
}

// runLinkWith is runLink publishing each session id to box, which the pulls are bound to.
//
// Contract: features/tower/tower_canary_offline_station.feature.
//
// A REFUSAL IS EITHER WORTH WAITING OUT OR IT IS NOT. Every refusal used to end the process, and
// a service manager restarted it every few seconds for as long as the refusal lasted - the house
// Tower spent 17 hours doing that while suspended. Now a refusal Core may lift (suspended, a
// lapsed lease, Core unreachable or failing) keeps the process up, says so once per state, and
// asks again on the node re-attach backoff, so a restore takes effect on the next retry with no
// restart. A refusal that cannot lift by waiting (revoked, a revoked certificate, a Tower Core
// does not recognise, the wrong mode) exits with exitTerminal, which the packaged unit tells its
// service manager not to restart.
func runLinkWith(st *tower.State, out io.Writer, stop <-chan struct{}, ticker func(time.Duration) (<-chan time.Time, func()), relay link.RelayPlane, box *sessionBox) error {
	if st.Mode != tower.ModeJoined {
		return terminalError{msg: errStandaloneCannotServeJoined.Error(), cause: errStandaloneCannotServeJoined}
	}

	// The head we last had accepted. Quoting it on connect is what lets Core say "resume"
	// instead of demanding everything - when Core is in step with us.
	head := towerjoin.Head{}
	var revision int64
	var sess towerjoin.Session
	sessOpen, pushRefused, linkedSaid := false, false, false
	said := map[string]bool{}
	lastState := ""
	// The id the operator quotes to RogerAI, said in every next step.
	towerID := ""
	if adm, ok := towerjoin.LoadAdmission(st.Dir()); ok {
		towerID = adm.TowerID
	}

	// connect opens a session and pushes the inventory when Core needs it (always on the first
	// connect: we only ever have everything). One step for the retry, because a refusal of the
	// push usually means the session it was pushed on is already gone.
	connect := func(first bool) error {
		// A session whose push was refused is closed before another is opened, or Core would
		// hold one per retry until each aged out. ONLY that one: a session ended by a refused
		// heartbeat may already be replaced on another instance, and a close is what drops a
		// Tower's routable fleet, so after a heartbeat refusal the Tower just re-opens.
		if sessOpen && pushRefused {
			_ = sess.Close(st)
			sessOpen = false
		}
		s, err := towerjoin.OpenSession(st, head, relay)
		if err != nil {
			return err
		}
		sess, sessOpen, pushRefused = s, true, false
		if first && !linkedSaid {
			linkedSaid = true
			fmt.Fprintf(out, "linked to RogerAI as %s (session %s)\n", sess.TowerID, sess.SessionID)
		}
		box.set(sess.SessionID)
		announceStateFor(out, &lastState, sess.State, towerID)
		if first || sess.NeedFullInventory {
			revision, head, err = pushInventory(st, out, revision, head)
			// Only the FIRST push's session is closed before a retry: that session never carried
			// anything, so closing it can drop nothing that is serving.
			pushRefused = err != nil && first
		}
		return err
	}
	if err := untilLinked(out, stop, said, towerID, func() error { return connect(true) }); err != nil {
		if errors.Is(err, errStoppedWaiting) {
			fmt.Fprintln(out, "\nstopping")
			return nil
		}
		return err
	}

	// DRAIN ON THE WAY OUT, on every exit path. Leaving without it means Core keeps offering
	// this Tower's Stations for a full freshness window after it has gone.
	defer func() {
		if !sessOpen {
			return // closed already, before a re-open that did not land
		}
		if cerr := sess.Close(st); cerr != nil {
			fmt.Fprintf(out, "warning: could not drain cleanly: %v\n", cerr)
			return
		}
		fmt.Fprintln(out, "drained: RogerAI has dropped this Tower's inventory")
	}()

	beat := sess.Heartbeat
	if beat <= 0 {
		beat = 60 * time.Second
	}
	beats, stopBeats := ticker(beat)
	defer stopBeats()

	// AND THE INVENTORY REFRESH, which is not optional. A pushed revision EXPIRES, and once
	// it does Core has nothing routable for this Tower - while the heartbeats keep
	// succeeding and everything keeps looking healthy. Without this the loop pushed once and
	// went dark half an hour later, silently, in production only.
	refreshes, stopRefresh := ticker(inventoryRefresh)
	defer stopRefresh()

	fmt.Fprintf(out, "holding the link (heartbeat every %s, inventory refresh every %s) - "+
		"ctrl-c to drain and exit\n", beat, inventoryRefresh)

	for {
		select {
		case <-stop:
			fmt.Fprintln(out, "\nstopping")
			return nil
		case <-refreshes:
			// A refusal here is NOT fatal: the inventory we already pushed is good until it
			// expires, so there is time for the next refresh to succeed. Tearing the link
			// down over one bad push would turn a blip into an outage.
			next, nextHead, rerr := pushInventory(st, out, revision, head)
			if rerr != nil {
				fmt.Fprintf(out, "could not refresh the inventory (%v) - will retry\n", rerr)
				continue
			}
			revision, head = next, nextHead
		case <-beats:
			if state, err := sess.SendHeartbeat(st); err == nil {
				announceStateFor(out, &lastState, state, towerID)
				continue
			} else if errors.Is(err, towerjoin.ErrUnreachable) {
				// Transport, not refusal: the freshness window is several heartbeats wide, so
				// one lost frame costs nothing and reconnecting immediately would be worse.
				fmt.Fprintf(out, "heartbeat did not reach RogerAI (%v) - will retry\n", err)
				continue
			}
			// Refused: the session is gone (Core restarted, or our lease lapsed). Re-open and
			// find out which, rather than heartbeating into nothing.
			fmt.Fprintln(out, "the session was refused - re-opening")
			if err := untilLinked(out, stop, said, towerID, func() error { return connect(false) }); err != nil {
				if errors.Is(err, errStoppedWaiting) {
					fmt.Fprintln(out, "\nstopping")
					return nil
				}
				return err
			}
		}
	}
}

// exitTerminal is the exit status of a `roger tower serve` that cannot recover by waiting. The
// packaged unit (packaging/tower/roger-tower.service) lists it in RestartPreventExitStatus, so a
// revoked or unrecognised Tower is not restarted in a loop. 78 is EX_CONFIG: what is wrong is the
// Tower's standing or its configuration, and the operator has to act.
const exitTerminal = 78

// terminalError ends serve with exitTerminal. It wraps its cause, so errors.Is still finds the
// refusal underneath (Run maps it to the status).
type terminalError struct {
	msg   string
	cause error
}

func (e terminalError) Error() string { return e.msg }
func (e terminalError) Unwrap() error { return e.cause }

// terminal is the error a terminal refusal ends serve with.
func terminal(msg string) error { return terminalError{msg: msg} }

// terminalBecause is terminal with the refusal that caused it kept, and Core's own sentence
// said after the operator's plain words.
func terminalBecause(words string, cause error) error {
	return terminalError{msg: words + "\n  Roger Core said: " + cause.Error(), cause: cause}
}

// errStoppedWaiting is ctrl-c arriving while serve waits out a refusal: a clean stop.
var errStoppedWaiting = errors.New("stopped while waiting")

// serveWait is the wait between asks after a recoverable refusal. A variable so a test can watch
// the waits without spending them.
var serveWait = func(d time.Duration) <-chan time.Time { return time.After(d) }

// linkBackoff is the node re-attach backoff (internal/agent reattachDelay): 30 s doubling to a
// 15-minute cap, with jitter over the upper half so a fleet refused together does not return
// together.
func linkBackoff(attempt int) time.Duration {
	const base, cap = 30 * time.Second, 15 * time.Minute
	d := base
	for i := 0; i < attempt && d < cap; i++ {
		d *= 2
	}
	if d > cap {
		d = cap
	}
	half := d / 2
	return half + time.Duration(randv2.Int64N(int64(half))+1)
}

// untilLinked runs step until it succeeds, waiting out recoverable refusals on linkBackoff and
// saying each state once. A terminal refusal comes back as a terminal error; anything it cannot
// classify comes back as it is.
//
// A refusal that names no state is waited out like a recoverable one, and becomes terminal only
// once it has lasted statelessRefusalLimit: Core sends it for a Tower it does not recognise, but
// an older Core sends it for every refusal, and a single answer is not enough to give up on.
//
// A success clears said, so a state that comes back after the link was restored is told again.
func untilLinked(out io.Writer, stop <-chan struct{}, said map[string]bool, towerID string, step func() error) error {
	var statelessSince time.Time
	for attempt := 0; ; attempt++ {
		err := step()
		if err == nil {
			clear(said)
			return nil
		}
		kind, words := classifyRefusal(err, towerID)
		switch kind {
		case refusalTerminal:
			return terminalBecause(words, err)
		case refusalOther:
			return err
		case refusalStateless:
			if statelessSince.IsZero() {
				statelessSince = serveNow()
			}
			if serveNow().Sub(statelessSince) >= statelessRefusalLimit {
				return terminalBecause(refusalWords(stateGuide{label: "REFUSED", meaning: "RogerAI has refused this Tower without saying why for 24 hours, so this Tower stops asking",
					next: "check --dir and `roger tower status`; if RogerAI no longer knows this Tower, enroll again with `roger tower register`; " + quoteID(towerID)}), err)
			}
		default:
			statelessSince = time.Time{}
		}
		if !said[words] {
			said[words] = true
			fmt.Fprintln(out, words)
		}
		select {
		case <-stop:
			return errStoppedWaiting
		case <-serveWait(linkBackoff(attempt)):
		}
	}
}

const (
	refusalOther = iota
	refusalRecoverable
	refusalTerminal
	refusalStateless
)

// serveNow is the clock serve measures a lasting refusal by. A variable so a test can run a day.
var serveNow = time.Now

// statelessRefusalLimit is how long serve keeps asking through a refusal that names no state.
const statelessRefusalLimit = 24 * time.Hour

// classifyRefusal sorts an error from the link into recoverable, terminal or neither, with the
// plain words the operator is told: for a state, the same lines `status` prints
// (tower_operator_guidance.feature), so the two commands never describe one state two ways.
func classifyRefusal(err error, towerID string) (int, string) {
	waiting := wrapPlain("This Tower stays up and asks again (every 30 s, backing off to 15 min), " +
		"and carries work again on its own once that changes.")
	if errors.Is(err, towerjoin.ErrUnreachable) {
		return refusalRecoverable, refusalWords(stateGuide{label: "UNREACHABLE", meaning: "could not reach RogerAI",
			next: "keep this running; check this machine's network if it lasts"}) + "\n" + waiting
	}
	var rf *towerjoin.Refusal
	if !errors.As(err, &rf) {
		return refusalOther, ""
	}
	guided := func(state string) string {
		g, _ := guideFor(state, rf.Reason, towerID)
		return refusalWords(g)
	}
	switch {
	case rf.State == "state_suspended":
		return refusalRecoverable, guided("suspended") + "\n" + waiting
	case rf.State == "lease_lapsed":
		return refusalRecoverable, guided("lease_lapsed") + "\n" + waiting
	case rf.State == "state_revoked":
		return refusalTerminal, guided("revoked")
	case rf.State == "cert_revoked":
		return refusalTerminal, refusalWords(stateGuide{label: "REVOKED", meaning: "RogerAI has revoked this Tower's certificate, so it will not be let back on",
			next: "enroll a new Tower with `roger tower register` if you want to run one again"})
	case rf.State == "state_expired":
		return refusalTerminal, refusalWords(stateGuide{label: "EXPIRED", meaning: "this Tower's enrollment expired before RogerAI approved it",
			next: "enroll again with `roger tower register`"})
	case rf.State != "":
		return refusalRecoverable, refusalWords(stateGuide{label: strings.ToUpper(strings.TrimPrefix(rf.State, "state_")),
			meaning: "RogerAI is not letting this Tower hold a link right now",
			next:    "keep this running; questions: " + towerPage + ", " + quoteID(towerID)}) + "\n" + waiting
	case rf.Status == http.StatusForbidden:
		return refusalStateless, refusalWords(stateGuide{label: "REFUSED", meaning: "RogerAI refused this Tower without saying why",
			next: "check --dir: RogerAI may not know this Tower id, or this directory's key may not be the one it enrolled with; " + quoteID(towerID)}) +
			"\n" + wrapPlain("This Tower keeps asking (every 30 s, backing off to 15 min) and stops if the refusal lasts 24 hours.")
	case rf.Status == http.StatusTooManyRequests:
		return refusalRecoverable, refusalWords(stateGuide{label: "SLOWED", meaning: "RogerAI asked this Tower to slow down",
			next: "nothing to do"}) + "\n" + waiting
	case rf.Status >= 500:
		return refusalRecoverable, refusalWords(stateGuide{label: "UNAVAILABLE", meaning: fmt.Sprintf("RogerAI is having trouble right now (error %d)", rf.Status),
			next: "keep this running"}) + "\n" + waiting
	}
	return refusalOther, ""
}

// inventoryRefresh is how often the fleet is re-pushed.
//
// DERIVED from the lifetime, deliberately: a hardcoded interval beside it is how the two
// drift apart when one changes, and the failure that produces is invisible - every heartbeat
// still succeeds while Core quietly has nothing routable. A third leaves room for one push
// to fail and the next to still land inside the window.
const inventoryRefresh = towerjoin.InventoryLifetime / 3

// pushInventory sends the current fleet and returns the new chain position.
func pushInventory(st *tower.State, out io.Writer, revision int64, head towerjoin.Head) (int64, towerjoin.Head, error) {
	leaves, err := localOffers(st, out)
	if err != nil {
		return revision, head, err
	}
	next := revision + 1
	prev := head.Hash
	if prev == "" {
		prev = "genesis"
	}
	res, err := towerjoin.PushFullInventory(st, next, prev, leaves)
	switch {
	case errors.Is(err, towerjoin.ErrNeedFullInventory):
		// We only ever send full snapshots today, so being asked for one again means our
		// chain position is not Core's. Start from one rather than guessing.
		res, err = towerjoin.PushFullInventory(st, 1, "genesis", leaves)
		if err != nil {
			return revision, head, err
		}
		next = 1
	case err != nil:
		return revision, head, err
	}

	if len(leaves) == 0 {
		fmt.Fprintf(out, "inventory revision %d accepted: no Stations attached yet, so this "+
			"Tower is on the network and carrying nothing\n", res.Revision)
	} else {
		fmt.Fprintf(out, "inventory revision %d accepted: %d of %d Station offer(s) eligible\n",
			res.Revision, res.Routable, len(leaves))
	}
	// The exclusions are the answer to "why is my Station idle", and they are the only place
	// an operator can get it.
	for _, ex := range res.Excluded {
		fmt.Fprintf(out, "  · %s is not eligible: %s\n", ex.StationID, ex.Reason)
	}
	return next, towerjoin.Head{Revision: res.Revision, Hash: res.Hash}, nil
}

// offersDir is where a Tower looks for legacy Station-signed offer files. The producer (the
// roger-station binary) is retired; the directory is still read so an operator's existing
// files surface as explicit Core-side exclusions rather than silently vanishing.
const offersDir = "offers"

// localOffers reads the Station-signed offers this Tower is relaying.
//
// RELAYED VERBATIM. The bytes are passed through untouched and are never decoded and
// re-encoded, because a Station signs its own offers with an assertion key this Tower does
// not hold and must never hold. Re-encoding one would invalidate the signature at best and,
// at worst, quietly change what the Station said. That is also why this reads FILES rather
// than building offers from the Tower's own configuration: there is no configuration a Tower
// could hold that would let it produce a leaf Core accepts, and that is by design.
//
// A file that is not JSON is REPORTED AND SKIPPED rather than fatal. One bad file should not
// take a whole fleet off the network - but a silent skip is how an operator ends up staring
// at a Station that never appears, so it is named on the way past. Core applies its own
// nineteen-row rejection table to every leaf that does get through and reports what it
// excluded, which is the answer to "why is my Station idle".
func localOffers(st *tower.State, out io.Writer) ([]json.RawMessage, error) {
	dir := filepath.Join(st.Dir(), offersDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		// Not an error, and not silent either: a Tower with no offers directory is the
		// ordinary state of one whose Stations have not been set up yet.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read the offers directory %s: %w", dir, err)
	}
	var leaves []json.RawMessage
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			fmt.Fprintf(out, "warning: skipping %s: %v\n", e.Name(), rerr)
			continue
		}
		if !json.Valid(raw) {
			fmt.Fprintf(out, "warning: skipping %s: it is not valid JSON\n", e.Name())
			continue
		}
		leaves = append(leaves, json.RawMessage(raw))
	}
	return leaves, nil
}

// cmdServe is the `roger-tower serve` entry point.
// serveFlagModes names the serve flags that belong to only one mode, and whether each takes a
// value (so a refusal can repeat what the operator typed).
var serveFlagModes = map[string]struct {
	mode  tower.Mode
	value bool
}{
	"bind": {tower.ModeStandalone, true}, "allow-public": {tower.ModeStandalone, false},
	"advertise": {tower.ModeStandalone, false}, "label": {tower.ModeStandalone, true},
	"check": {tower.ModeStandalone, false},
	"hub":   {tower.ModeJoined, true}, "relay-public": {tower.ModeJoined, true},
	"hub-tls": {tower.ModeJoined, false}, "hub-tls-cert": {tower.ModeJoined, true},
	"hub-tls-key": {tower.ModeJoined, true}, "hub-legacy-bearer": {tower.ModeJoined, false},
}

// cmdServe is `roger tower serve`: the data directory's recorded mode decides what is served.
// A standalone directory runs the Core-free consumer plane (internal/localplane/serve); a joined
// one runs the link and the hub. A flag of the other mode is refused (exit 2) before anything is
// opened, naming the flag as typed and the directory's mode.
func cmdServe(args []string, out io.Writer) error {
	for _, a := range args {
		if a == "-h" || a == "--help" || a == "-help" {
			fmt.Fprint(out, serveHelp)
			return errHelpShown
		}
	}
	dir := flagValue(args, "dir")
	if dir == "" {
		// With only --config, the directory is the config's identity.dir (the packaged unit
		// passes both; a unit written for the old name passed only --config).
		if cfgPath := flagValue(args, "config"); cfgPath != "" {
			if c, err := tower.LoadConfig(cfgPath); err == nil && c.Identity.Dir != "" {
				dir = c.Identity.Dir
				args = append(append([]string(nil), args...), "--dir", dir)
			}
		}
	}
	if dir == "" {
		return exitError{2, "--dir is required"}
	}
	// A flag neither mode takes is a usage mistake, reported before the directory is touched.
	for _, a := range args {
		if !strings.HasPrefix(a, "-") || a == "-" || a == "--" {
			continue
		}
		name, _, _ := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if _, ok := serveFlagModes[name]; !ok && name != "dir" && name != "config" && name != "h" && name != "help" {
			return exitError{2, fmt.Sprintf("flag provided but not defined: -%s\nusage: %s", name, commandUsage("serve"))}
		}
	}
	st, err := tower.Open(dir)
	if err != nil {
		// No mode to dispatch on. A joined flag means the operator meant the joined serve,
		// whose flag checks come before the directory (a flag mistake is reported as one).
		for _, a := range args {
			name, _, _ := strings.Cut(strings.TrimLeft(a, "-"), "=")
			if f, ok := serveFlagModes[name]; ok && strings.HasPrefix(a, "-") && f.mode == tower.ModeJoined {
				return cmdServeJoined(args, out)
			}
		}
		return err
	}
	for i, a := range args {
		if !strings.HasPrefix(a, "-") || a == "-" || a == "--" {
			continue
		}
		name, _, hasVal := strings.Cut(strings.TrimLeft(a, "-"), "=")
		f, ok := serveFlagModes[name]
		if !ok || f.mode == st.Mode {
			continue
		}
		typed := a
		if f.value && !hasVal && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			typed = a + " " + args[i+1]
		}
		return exitError{2, fmt.Sprintf("%s is a %s flag; this directory is %s", typed, f.mode, st.Mode)}
	}
	if st.Mode == tower.ModeStandalone {
		return serveStandalone(args, out)
	}
	return cmdServeJoined(args, out)
}

// serveHelp is `roger tower serve --help`, including what the standalone guarantee rests on.
const serveHelp = `usage: roger tower serve --dir DIR [--config FILE] [flags]

Serves the data directory as its recorded mode says.

  standalone  the consumer plane on your own network:
                --bind HOST:PORT (default 127.0.0.1:8787)  --allow-public  --advertise
                --label NAME  --check (resolve and report, without listening)
  joined      the link to Roger Core and the sealed hub:
                --hub :8444  --relay-public HOST:PORT  --hub-tls [--hub-tls-cert F --hub-tls-key F]
                --hub-legacy-bearer=false

A flag of the other mode is refused (exit 2). With only --config, the directory is the
config's identity.dir.

Standalone: the standalone serve package imports no Core code, and standalone mode refuses
every non-local connection (the process's default HTTP client, transport and resolver), and it
refuses to start if a Core client was built.
`

// flagValue is the value of --name (or --name=v, or -name) in args, read before parsing so the
// dispatch can choose the mode first.
func flagValue(args []string, name string) string {
	for i, a := range args {
		n, val, hasVal := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if !strings.HasPrefix(a, "-") || n != name {
			continue
		}
		if hasVal {
			return val
		}
		if i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// serveStandalone is the standalone consumer plane (a seam: a test stands in for the blocking
// serve).
var serveStandalone = serve.Run

func cmdServeJoined(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	dir, cfg := dirAndConfig(fs)
	relayPublic := fs.String("relay-public", "", "the PUBLIC host:port consumers reach this tower's data plane (the hub) at, advertised to Roger Core")
	hubAddr := fs.String("hub", "", "address to serve the data-plane HUB on, e.g. :8444 - where consumers submit sealed work and this tower's self-attached nodes poll")
	hubTLSOn := fs.Bool("hub-tls", false, "serve the hub over TLS. With no --hub-tls-cert, this tower mints and keeps a self-signed certificate and Roger Core publishes its fingerprint to every node and consumer it routes here, so no publicly-trusted certificate and no domain name are needed")
	hubCert := fs.String("hub-tls-cert", "", "TLS certificate (PEM) for the hub listener - implies --hub-tls, and with --hub-tls-key the hub serves https")
	hubKey := fs.String("hub-tls-key", "", "TLS private key (PEM) for the hub listener")
	// THE TRANSITION SWITCH, and it is a real switch. A node released before signed hub polls
	// authenticates with a bearer token, and this tower keeps accepting one for a release so
	// that provider keeps earning while they update. An operator who knows every node on their
	// tower has updated can end that early; one release from now the flag and the code behind
	// it both go. It had been documented as "default true" while being settable from nowhere
	// but a test, which is a claim an operator cannot act on.
	legacyBearer := fs.Bool("hub-legacy-bearer", true, "accept the pre-signature bearer token from nodes that have not updated yet (a station that has signed to this hub always refuses its own token from then on); -hub-legacy-bearer=false requires a signature from every node")
	if err := parseFlags(fs, args, out); err != nil {
		return err
	}
	// The hub's payload is sealed end-to-end, but the grant metadata and every node's
	// long-term assertion PUBLIC KEY ride the transport in the clear - so TLS here is real
	// protection, not ceremony (audit M1). It is no longer the polling token: a current node
	// signs each request and transmits nothing reusable. What TLS buys now is that an observer
	// cannot tie a station's payment identity to an address. Half a key pair is a mistake, not
	// a mode.
	if (*hubCert == "") != (*hubKey == "") {
		return fmt.Errorf("--hub-tls-cert and --hub-tls-key must be given together")
	}
	if (*hubCert != "" || *hubTLSOn) && *hubAddr == "" {
		return fmt.Errorf("--hub-tls without --hub: there is no hub listener to protect")
	}
	// Checked BEFORE the data directory is touched when there is no config to fill the gap:
	// a flag mistake should be reported as a flag mistake, not as whatever the directory
	// happens to complain about first.
	if *relayPublic != "" {
		resolved, note, aerr := resolveAdvertised(*relayPublic)
		if aerr != nil {
			return aerr
		}
		*relayPublic = resolved
		if note != "" {
			fmt.Fprintln(out, note)
		}
	}
	if *cfg == "" && *relayPublic != "" && *hubAddr == "" {
		return fmt.Errorf("--relay-public advertises a data plane, but no --hub is serving one")
	}
	// serve takes --config for the same reason the state commands do, and it matters MORE
	// here: an operator whose `attach` wrote to the database while `serve` read local disk
	// would be relaying an inventory that does not describe their fleet.
	st, release, err := openDirWith(*dir, *cfg)
	if err != nil {
		return err
	}
	defer release()
	// THE CONFIG IS NOT DECORATION. A data plane declared in the file and then ignored
	// because the operator did not also pass flags is the exact failure an audit found
	// across the rest of the schema; flags win when both are given, because a flag is the
	// more deliberate of the two.
	if *cfg != "" {
		c, cerr := loadConfig(*cfg)
		if cerr != nil {
			return cerr
		}
		for _, u := range c.Unenforced() {
			fmt.Fprintf(out, "IGNORED: %s\n", u)
		}
		if *relayPublic == "" && c.Relay != nil {
			*relayPublic = c.Relay.Public
		}
		if c.Hub != nil {
			if *hubAddr == "" {
				*hubAddr = c.Hub.Address
			}
			if *hubCert == "" && *hubKey == "" {
				*hubCert, *hubKey = c.Hub.TLSCert, c.Hub.TLSKey
			}
			// The file can only turn TLS ON. A config saying tls:false while the operator typed
			// --hub-tls would be the file quietly downgrading the more deliberate of the two,
			// and the thing it would be downgrading is the channel's confidentiality.
			if c.Hub.TLS {
				*hubTLSOn = true
			}
			// The config can only turn the tolerance OFF, and only when the flag was left at
			// its default. A file that said "true" while the operator typed
			// -hub-legacy-bearer=false would be the config quietly overriding the more
			// deliberate of the two, which is the rule this whole block exists to keep.
			if c.Hub.AllowLegacyBearer != nil && !fsSet(fs, "hub-legacy-bearer") {
				*legacyBearer = *c.Hub.AllowLegacyBearer
			}
		}
	}
	// The address Core hands to self-attaching nodes is often not the listen address - ":8444"
	// is not dialable by anyone. With no --relay-public the hub is advertised at this machine's
	// own address, because work now reaches the Tower by pull over its link and a hub on a home
	// network serves the Stations on that network (private_hub_placement.feature).
	if *relayPublic != "" && *hubAddr == "" {
		return fmt.Errorf("--relay-public advertises a data plane, but no --hub is serving one")
	}
	if *hubAddr != "" && *relayPublic == "" {
		resolved, note, aerr := resolveAdvertised(*hubAddr)
		if aerr != nil {
			return fmt.Errorf("the hub could not be advertised at this machine's own address (%v) - pass --relay-public", aerr)
		}
		*relayPublic = resolved
		if note != "" {
			fmt.Fprintln(out, note)
		}
		// Who that address can serve, said the same way as for an address the operator typed.
		if host, _, herr := net.SplitHostPort(resolved); herr == nil {
			if ip := net.ParseIP(host); ip != nil {
				if tier := classifyAdvertised(host, []net.IP{ip}, false); tier != "" && !strings.Contains(note, tier) {
					fmt.Fprintln(out, tier)
				}
			}
		}
		if advertisedPrivate(resolved) && !*hubTLSOn && *hubCert == "" {
			// A private hub is handed to Stations only with a certificate pin, so the hub this
			// command chose the address for gets one. An operator who typed the address keeps
			// their own choice and is told what it costs, below.
			*hubTLSOn = true
			fmt.Fprintln(out, "hub: TLS turned on - a hub at a private address is handed to Stations only with a certificate pin")
		}
	}
	if advertisedPrivate(*relayPublic) && !*hubTLSOn && *hubCert == "" {
		fmt.Fprintln(out, "NOTE: a private hub needs --hub-tls. Roger Core places no Station on a private "+
			"address it cannot hand out with a certificate pin.")
	}
	if !*legacyBearer {
		fmt.Fprint(out, "hub: pre-signature bearer tokens are REFUSED on this tower - a node "+
			"older than signed hub polls will not be served here.\n")
	}
	return serveJoined(st, out, *relayPublic, hubOptions{
		Addr: *hubAddr, TLS: *hubTLSOn, TLSCert: *hubCert, TLSKey: *hubKey,
		AllowLegacyBearer: *legacyBearer,
	})
}

// fsSet reports whether a flag was actually typed, as opposed to sitting at its default. It is
// what lets configuration lower a default without ever overriding an explicit flag - the rule
// the rest of cmdServe follows by checking for an empty string, which a bool cannot do.
func fsSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// announceState prints the admission state when it CHANGES, in the operator's language, with
// no tower id to quote.
func announceState(out io.Writer, last *string, state string) { announceStateFor(out, last, state, "") }

// announceStateFor prints the admission state when it CHANGES, in the words `status` uses for
// it (tower_operator_guidance.feature): who has to act and what to do next. Quarantine is the
// state every Tower is admitted into, so it reads as the waiting room it is, not as a fault.
// A state this binary does not recognise is shown verbatim rather than hidden, because an old
// CLI meeting a new Core should say something true.
func announceStateFor(out io.Writer, last *string, state, towerID string) {
	if state == "" || state == *last {
		return
	}
	*last = state
	g, ok := guideFor(state, "", towerID)
	if !ok {
		fmt.Fprintf(out, "state:     %s\n", state)
		return
	}
	writeGuide(out, "", g)
	if g.also != "" {
		fmt.Fprintln(out, strings.Join(wrap("  ", g.also, "    "), "\n"))
	}
}

// resolveAdvertised turns the operator's --relay-public into the address Core will hand
// to nodes and consumers, and says anything worth saying about it.
//
// An empty host (":8444") means "this machine": it resolves to the machine's own
// outbound address and PRINTS the choice, because advertising the literal ":8444" made
// every node dial itself - silent nonsense. A loopback host is accepted and named for
// what it is: a same-machine test rig the public network cannot reach. Neither is an
// error; both are the operator's business - said out loud.
func resolveAdvertised(endpoint string) (addr, note string, err error) {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("--relay-public must be a dialable host:port, got %q", endpoint)
	}
	parsed := net.ParseIP(host)
	// An empty host (":8444") or an UNSPECIFIED one ("0.0.0.0", "::") is a BIND wildcard,
	// not a reachable address - the thing you pass to --hub to listen on every interface,
	// mistaken for the thing you advertise. You cannot dial 0.0.0.0. Both mean "this
	// machine", so resolve to the machine's own outbound address and say what happened.
	if host == "" || (parsed != nil && parsed.IsUnspecified()) {
		ip, derr := outboundIPFn()
		if derr != nil {
			return "", "", fmt.Errorf("--relay-public %q is a bind wildcard, not a reachable address, "+
				"and this machine's own address could not be determined (%v) - pass the address explicitly", endpoint, derr)
		}
		resolved := net.JoinHostPort(ip, port)
		why := "had no host"
		if host != "" {
			why = fmt.Sprintf("was %s, a bind wildcard nothing can dial", host)
		}
		return resolved, fmt.Sprintf("relay-public %s: advertising this machine's address, %s", why, resolved), nil
	}
	if parsed != nil {
		return endpoint, classifyAdvertised(host, []net.IP{parsed}, false), nil
	}
	// A NAME - "roggentoo", "hub.example.net". Resolve it here, on the operator's own
	// machine, and say what it points at: a LAN name is a first-class home-lab tier, and
	// the operator deserves to know which tier they just advertised - and that the name
	// must resolve on every device that will dial it, which an /etc/hosts entry on this
	// box alone does not give them.
	rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ips, rerr := lookupIPFn(rctx, "ip", host)
	if rerr != nil || len(ips) == 0 {
		return "", "", fmt.Errorf("--relay-public names %q, which does not resolve on this machine (%v). "+
			"Every device dialing the hub must be able to resolve it - use the IP address if unsure", host, rerr)
	}
	return endpoint, classifyAdvertised(host, ips, true), nil
}

// lookupIPFn is the advert's resolver, a seam so tests classify names without DNS.
var lookupIPFn = net.DefaultResolver.LookupIP

// classifyAdvertised names the tier an advert lands in, in the operator's language:
// loopback is a same-machine test rig, a private address is the home-lab LAN tier, and a
// public one says nothing because nothing needs saying.
func classifyAdvertised(host string, ips []net.IP, named bool) string {
	loop, private, linkLocal := false, false, false
	shown := ips[0].String()
	for _, ip := range ips {
		// One rule (internal/netpriv), the same classes `roger lan add` uses: link-local is
		// local but not routable, so it is named as local and said to be unusable.
		switch netpriv.Classify(ip) {
		case netpriv.Loopback:
			loop = true
		case netpriv.LinkLocal:
			linkLocal = true
		case netpriv.Private, netpriv.ULA, netpriv.CGNAT:
			private = true
		}
	}
	switch {
	case loop:
		return "relay-public is loopback: only THIS machine can reach the hub, so Roger Core places " +
			"no Station here. Advertise this machine's LAN address instead."
	case linkLocal:
		return fmt.Sprintf("relay-public %s is a LOCAL network address, but a link-local one: Roger "+
			"Core hands no Station a link-local address. Advertise this machine's LAN address instead.", host)
	case private && named:
		return fmt.Sprintf("relay-public %q resolves to %s - a LOCAL network address. Stations on "+
			"this network will be placed here; Stations elsewhere need a public address for now. "+
			"Note: every device dialing the hub must resolve %q itself - if the name lives only "+
			"in this machine's hosts file, advertise %s instead.", host, shown, host, shown)
	case private:
		return fmt.Sprintf("relay-public %s is a LOCAL network address. Stations on this network "+
			"will be placed here; Stations elsewhere need a public address for now.", host)
	}
	return ""
}

// advertisedPrivate reports an advertised hub at a private address (RFC 1918, ULA, CGNAT) - the
// kind Roger Core hands out only with a certificate pin. Names are judged by Core, not here.
func advertisedPrivate(endpoint string) bool {
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	switch netpriv.Classify(ip) {
	case netpriv.Private, netpriv.ULA, netpriv.CGNAT:
		return true
	}
	return false
}

// outboundIPFn is outboundIP behind a seam, so the wildcard rules are tested on any machine,
// including one with no route out at all.
var outboundIPFn = outboundIP

// outboundIP is the address this machine uses to reach the world: a UDP "dial" that
// sends nothing and asks the kernel which source address it would pick.
func outboundIP() (string, error) {
	conn, err := net.Dial("udp", "203.0.113.1:9")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	la, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || la.IP.IsUnspecified() {
		return "", fmt.Errorf("no outbound address")
	}
	return la.IP.String(), nil
}
