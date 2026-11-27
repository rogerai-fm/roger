package agent

// Contract: features/tower/station_mobility_idempotency.feature.
//
// relayWatch is what one tenancy knows about how its relay is answering, for the "never stranded"
// rule: a hub that says "no node is registered" (a grace while a new attachment propagates, then
// a hub failure like any other), the operator's warning and recovery lines, and empty polls after
// a nudge the node could not act on.

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// noNodeGrace is how long a new attachment's hub may not know it yet: twice the hub's
	// 30-second refresh of Core's registrations.
	noNodeGrace = 60 * time.Second
	// emptyAfterNudge is how long a node that heard a nudge it could not act on waits through
	// empty polls before asking Core again.
	emptyAfterNudge = 5 * time.Minute
)

type relayWatch struct {
	mu      sync.Mutex
	towerID string
	begun   time.Time
	now     func() time.Time
	notify  func(error)

	unknownShown bool
	nudgeHeard   time.Time
}

func newRelayWatch(towerID string, begun time.Time, notify func(error)) *relayWatch {
	return &relayWatch{towerID: towerID, begun: begun, now: time.Now, notify: notify}
}

// hubNoNode reports a hub 401 that says it has no node registered for this Station: the hub has
// not learned the attachment (yet), as opposed to refusing this node's identity.
func hubNoNode(err error) bool {
	return hubRefusedIdentity(err) && strings.Contains(err.Error(), "no node is registered")
}

// onNoNode handles a "no node is registered" 401 and reports whether it counts toward the standing
// window. Inside the grace it is quiet: the hub refreshes Core's registrations every 30 seconds and
// a new attachment is simply not there yet. After it, it is a hub failure like a refused dial (the
// node asks Core again after the standing window), and the operator is told once what is true.
func (w *relayWatch) onNoNode(err error) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.now().Sub(w.begun) < noNodeGrace {
		return false
	}
	if !w.unknownShown {
		w.unknownShown = true
		w.notify(fmt.Errorf("relay: Tower %s does not know this node yet; it keeps polling and will ask Roger Core "+
			"again if that lasts (%w)", w.towerID, err))
	}
	return true
}

// onPoll observes a poll the hub answered (job says whether it carried work) and reports whether
// the node should ask Core again: after five minutes of nothing but empty polls since a nudge it
// heard and could not act on. A poll that answers after the "does not know this node" line says
// once that the node is serving again.
func (w *relayWatch) onPoll(job bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.unknownShown {
		w.unknownShown = false
		w.notify(fmt.Errorf("relay: serving through Tower %s again", w.towerID))
	}
	if job {
		w.nudgeHeard = time.Time{}
		return false
	}
	return !w.nudgeHeard.IsZero() && w.now().Sub(w.nudgeHeard) >= emptyAfterNudge
}

// onNudgeHeard records a nudge the node heard but could not act on (the one-a-minute pace, or work
// that did not finish in time). The first one starts the clock.
func (w *relayWatch) onNudgeHeard() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.nudgeHeard.IsZero() {
		w.nudgeHeard = w.now()
	}
}
