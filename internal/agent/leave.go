package agent

// Contract: features/tower/tower_canary_offline_station.feature.
//
// leave.go is the stopping half of a Station's tenancy: when the operator stops `roger share`,
// every attachment this process holds is taken off the air at once with a signed leave, instead
// of staying routable until its node has been silent for nodeTTL. A leave that does not land is
// harmless: Core's liveness predicate takes the Station out of rotation once the node is silent.

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"rogerai.fm/roger/v7/internal/protocol"
)

type leaveTarget struct {
	broker string
	priv   ed25519.PrivateKey
	at     TowerAttachment
}

// held is the attachment each successful attach in this process obtained, by Station id. A re-attach
// overwrites its Station's entry, so the latest Tower and epoch are the ones left.
var held = struct {
	sync.Mutex
	m map[string]leaveTarget
}{m: map[string]leaveTarget{}}

func holdAttachment(broker string, priv ed25519.PrivateKey, at TowerAttachment) {
	if at.StationID == "" || at.Epoch <= 0 {
		return
	}
	t := leaveTarget{broker: broker, priv: priv, at: at}
	// Hold or leave is decided under the lock StopAttaching and LeaveAll take, so an attach
	// either lands before the snapshot (and LeaveAll leaves it) or sees the share leaving.
	held.Lock()
	if !leaving.Load() {
		held.m[at.StationID] = t
		held.Unlock()
		return
	}
	held.Unlock()
	// It was in flight when the share began leaving, so LeaveAll did not see it: leave it now.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = sendLeave(ctx, t)
}

// LeaveAll sends a leave for every attachment this process holds, signed by the node key it
// attached with, and returns how many Core accepted. It gives up when ctx ends.
func LeaveAll(ctx context.Context) int {
	// Each held attachment is tried once: after its leave it is no longer this process's to
	// leave, whatever Core answered.
	held.Lock()
	targets := make([]leaveTarget, 0, len(held.m))
	for id, t := range held.m {
		targets = append(targets, t)
		delete(held.m, id)
	}
	held.Unlock()
	var wg sync.WaitGroup
	var mu sync.Mutex
	n := 0
	for _, t := range targets {
		wg.Add(1)
		go func(t leaveTarget) {
			defer wg.Done()
			if sendLeave(ctx, t) == nil {
				mu.Lock()
				n++
				mu.Unlock()
			}
		}(t)
	}
	wg.Wait()
	return n
}

func sendLeave(ctx context.Context, t leaveTarget) error {
	body, err := json.Marshal(map[string]any{"station_id": t.at.StationID, "tower_id": t.at.TowerID, "epoch": t.at.Epoch})
	if err != nil {
		return err
	}
	const path = "/tower/edge/leave"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.broker+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	pub, ts, sig := protocol.SignRequest(t.priv, http.MethodPost, path, body)
	req.Header.Set(protocol.HeaderPubkey, pub)
	req.Header.Set(protocol.HeaderTS, fmt.Sprintf("%d", ts))
	req.Header.Set(protocol.HeaderSig, sig)
	resp, err := (&http.Client{CheckRedirect: protocol.NoDowngradeRedirect}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("leave refused (%d)", resp.StatusCode)
	}
	return nil
}

// ErrLeaving refuses an attach once this process has begun taking its Stations off the air.
var ErrLeaving = errors.New("this share is stopping: its Stations are leaving, so it does not attach again")

// leaving is set once this process has begun taking its Stations off the air. From then on no
// attach is sent, and one that was already in flight when it was set is left as soon as it lands.
var leaving atomic.Bool

// StopAttaching stops every attach in this process, to be called BEFORE LeaveAll so that no
// ServeTower loop re-attaches behind the leaves. resume undoes it; only a test calls it - a
// stopping share exits.
func StopAttaching() (resume func()) {
	held.Lock()
	leaving.Store(true)
	held.Unlock()
	return func() { leaving.Store(false) }
}
