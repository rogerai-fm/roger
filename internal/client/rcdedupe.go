package client

import "sync"

// rcDedupeSize is how many recent inbound ids a host remembers. A redelivery comes back within
// a poll or two (rc_inbox_delivery.feature D2), so a few hundred is far more than enough.
const rcDedupeSize = 256

// RCDedupe remembers the ids of recent remote-control inbounds, so a turn the broker hands out
// again after a lost ack is run once (rc_inbox_delivery.feature D3).
type RCDedupe struct {
	mu   sync.Mutex
	seen map[string]bool
	ring []string
	next int
}

// NewRCDedupe returns an empty dedupe set.
func NewRCDedupe() *RCDedupe {
	return &RCDedupe{seen: map[string]bool{}, ring: make([]string, rcDedupeSize)}
}

// First reports whether id is new, and remembers it. An empty id (an older broker) is always new.
func (d *RCDedupe) First(id string) bool {
	if id == "" {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seen[id] {
		return false
	}
	if old := d.ring[d.next]; old != "" {
		delete(d.seen, old)
	}
	d.ring[d.next] = id
	d.next = (d.next + 1) % len(d.ring)
	d.seen[id] = true
	return true
}
