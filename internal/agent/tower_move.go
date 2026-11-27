package agent

// tower_move.go: the node's half of a move (features/tower/station_mobility.feature). Core asks
// a node to attach again by putting a nudge on its cancel long-poll; the nudge names no Tower,
// so the worst a forged one can do is cause one extra attach to Core over its pinned address.
// The node finishes what it is serving first, acts at most once a minute, and ignores a nudge
// it could not act on before it expired.

import (
	"context"
	"errors"
	randv2 "math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// errMoved ends a tenancy because Core asked this node to attach again.
var errMoved = errors.New("roger core asked this station to attach again")

// nudgeEvery is the most often a node acts on a nudge.
const nudgeEvery = time.Minute

// nudgeExpiry is how long a node waits to be idle before it gives up on a nudge: the broker
// keeps a nudge for the same two minutes.
var nudgeExpiry = 2 * time.Minute

// nudge is one re-attach request as the node received it.
type nudge struct {
	reason string
	at     time.Time
}

// fresh reports whether the nudge can still be acted on: one older than nudgeExpiry is stale,
// whatever happened to it in between.
func (n nudge) fresh(now time.Time) bool { return now.Sub(n.at) <= nudgeExpiry }

var (
	nudges sync.Map // node id -> chan nudge
	gates  sync.Map // node id -> *moveGate
)

// nudgeChan is the one-slot channel the cancel poll hands a node's nudges to and its relay
// loop listens on. One slot: a nudge waiting already says everything a second one would.
func nudgeChan(nodeID string) chan nudge {
	ch, _ := nudges.LoadOrStore(nodeID, make(chan nudge, 1))
	return ch.(chan nudge)
}

// gateFor is the node's move gate. It belongs to the node, not to one tenancy, so the
// one-a-minute pace holds across re-attaches.
func gateFor(nodeID string) *moveGate {
	g, _ := gates.LoadOrStore(nodeID, &moveGate{})
	return g.(*moveGate)
}

// deliverNudge hands a nudge to the relay loop without ever blocking the cancel poll.
func deliverNudge(nodeID, reason string) {
	select {
	case nudgeChan(nodeID) <- nudge{reason: reason, at: time.Now()}:
	default:
	}
}

// moveGate counts the relay work in flight and paces nudges.
type moveGate struct {
	inflight atomic.Int64 // relay work in flight on the current tenancy
	mu       sync.Mutex
	last     time.Time
	now      func() time.Time
}

func (g *moveGate) begin() { g.inflight.Add(1) }
func (g *moveGate) end()   { g.inflight.Add(-1) }

func (g *moveGate) clock() time.Time {
	if g.now != nil {
		return g.now()
	}
	return time.Now()
}

// allow reports whether a nudge may be acted on now, and if so starts the minute.
func (g *moveGate) allow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.clock()
	if !g.last.IsZero() && now.Sub(g.last) < nudgeEvery {
		return false
	}
	g.last = now
	return true
}

// waitIdle waits for the relay work in flight to finish, up to expiry.
func (g *moveGate) waitIdle(ctx context.Context, expiry time.Duration) bool {
	deadline := time.Now().Add(expiry)
	for g.inflight.Load() > 0 {
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(20 * time.Millisecond):
		}
	}
	return true
}

// movedError ends a tenancy for a move and carries the move's trigger.
type movedError struct{ reason string }

func (e movedError) Error() string { return errMoved.Error() + " (" + e.reason + ")" }
func (e movedError) Unwrap() error { return errMoved }

var moveWhy = map[string]string{
	"better_tower":          "a better relay was found for it",
	"tower_failing":         "its relay was failing",
	"stage_cap":             "its relay was full for its trust stage",
	"tower_gone":            "its relay was gone",
	"left_network":          "it is no longer on its relay's local network",
	"unreachable":           "it could not reach its relay",
	"no work since a nudge": "Roger Core asked it to move and no work came for five minutes",
}

// ineligibleWhy words the reasons a relay may stop taking work (station_mobility_idempotency.feature
// R4); Core sends them as "ineligible:<reason>".
var ineligibleWhy = map[string]string{
	"lease_expired":     "its relay's approval lapsed",
	"suspended":         "its relay was suspended",
	"draining":          "its relay is shutting down",
	"failing_admission": "its relay is waiting to be approved again",
	"banned":            "its relay's owner may no longer run one",
}

// movedNotice is what the operator sees in `roger share` once a moved Station has attached
// again: where it was, where it is, and why, in plain words. Tower ids only, never the
// operators behind them.
func movedNotice(from, to, reason string) string {
	why, ok := moveWhy[reason]
	if r, isInel := strings.CutPrefix(reason, "ineligible:"); isInel {
		why, ok = ineligibleWhy[r]
		if !ok {
			why, ok = "its relay stopped taking work", true
		}
	}
	if !ok {
		why = "Roger Core asked it to attach again"
	}
	if from == to {
		return "tower: this Station attached again and stayed on relay " + to + " (" + why + ")"
	}
	return "tower: this Station moved from relay " + from + " to relay " + to + " (" + why + ")"
}

// gatedExec counts each served job on the gate, so a nudge waits for it.
type gatedExec struct {
	inner sealedExec
	gate  *moveGate
}

func (g gatedExec) Serve(ctx context.Context, grant, envelope []byte) ([]byte, []byte, string) {
	g.gate.begin()
	defer g.gate.end()
	return g.inner.Serve(ctx, grant, envelope)
}

// reattachWait waits out a re-attach backoff, and a fresh nudge cuts it short
// (station_mobility_idempotency.feature R3). Core nudges a node it is moving every 30 seconds and
// undoes the move after three unanswered nudges; a node sitting out a backoff of several minutes
// used to hear none of them, so its move was undone and it waited anyway. The one-a-minute pace
// still holds. false means ctx ended.
func reattachWait(ctx context.Context, d time.Duration, nodeID string) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	var nudges chan nudge
	if nodeID != "" {
		nudges = nudgeChan(nodeID)
	}
	for {
		select {
		case <-t.C:
			return true
		case <-ctx.Done():
			return false
		case n := <-nudges:
			if n.fresh(time.Now()) && gateFor(nodeID).allow() {
				return true
			}
		}
	}
}

// nudgeActionable reports whether a nudge may end a tenancy that began at started: it must be
// fresh, and it must have arrived after that attach. A nudge buffered before the attach was
// answered by the attach itself, so acting on it would only re-attach once more for nothing.
func nudgeActionable(n nudge, started, now time.Time) bool {
	return n.fresh(now) && !n.at.Before(started)
}

// ReattachDelayFor is how long a node waits before asking Core again after a failed re-attach.
//
// After "no data plane" (its Tower is not linked) the wait is a flat 25 to 35 seconds, not the
// doubling backoff (station_mobility_idempotency.feature, ruled 2026-11-24): Core declares a Tower
// gone after a fixed window and then moves its Stations as they ask, so a node that had backed off
// to many minutes would sit dark long after Core was ready to place it. Every other refusal keeps
// the doubling backoff.
func ReattachDelayFor(err error, consecutive int) time.Duration {
	if err != nil && strings.Contains(err.Error(), noDataPlane) {
		return 25*time.Second + time.Duration(randv2.Int64N(int64(10*time.Second)))
	}
	return reattachDelay(consecutive)
}

// noDataPlane is the words Core answers a re-attach with when the Station's Tower is not linked.
const noDataPlane = "has no data plane right now"

// firstAttachGivesUp reports whether a node's FIRST attach stops trying after tries refusals. While
// Core answers "no data plane" it never does: that answer means Core has this node's Station and is
// waiting for its Tower to come back or be declared gone, and the node retries on the flat 30 s.
func firstAttachGivesUp(err error, tries int) bool {
	if errors.Is(err, ErrPrivateShareNeverRelays) {
		return true
	}
	if err != nil && strings.Contains(err.Error(), noDataPlane) {
		return false
	}
	return tries >= firstAttachAttempts
}
