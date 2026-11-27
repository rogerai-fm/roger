package store

// Contract: features/tower/station_mobility_idempotency.feature ("Every check above reads shared
// state where the approved spec requires it"): a banned owner's Tower is ineligible since the
// ban's recorded time, so the time must be the first ban's, kept, and gone after an unban.

import (
	"testing"
	"time"
)

func TestOwnerBannedSinceIsTheFirstBansTimeParity(t *testing.T) {
	for name, db := range parityStores(t) {
		t.Run(name, func(t *testing.T) {
			if _, found, err := db.OwnerBannedSince("acct-none"); err != nil || found {
				t.Fatalf("an unbanned account reads banned: %v %v", found, err)
			}
			before := time.Now().Add(-time.Minute)
			if err := db.BanOwner("acct-b", "first", "{}"); err != nil {
				t.Fatal(err)
			}
			first, found, err := db.OwnerBannedSince("acct-b")
			if err != nil || !found || first.Before(before) {
				t.Fatalf("ban time %v found=%v err=%v", first, found, err)
			}
			time.Sleep(20 * time.Millisecond)
			if err := db.BanOwner("acct-b", "second", "{}"); err != nil {
				t.Fatal(err)
			}
			again, _, _ := db.OwnerBannedSince("acct-b")
			if !again.Equal(first) {
				t.Fatalf("a second ban moved the time from %v to %v; the first ban is kept", first, again)
			}
			if _, err := db.ForgiveOwner("acct-b"); err != nil {
				t.Fatal(err)
			}
			if _, found, _ := db.OwnerBannedSince("acct-b"); found {
				t.Fatal("an unbanned account still has a ban time")
			}
		})
	}
}
