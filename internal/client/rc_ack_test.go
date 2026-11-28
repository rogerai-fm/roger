package client

// The host side of features/multinode/rc_inbox_delivery.feature (D1, D3): the real RCBridge poll
// loop says it acks, acks each inbound on its next poll (and keeps sending the ack until a poll
// gets an answer), and runs an id it has already seen only once.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rogerai.fm/roger/v7/internal/protocol"
)

func TestRCBridgeAcksAndRunsARedeliveryOnce(t *testing.T) {
	answers := []protocol.RCInbound{
		{ID: "x1", Kind: protocol.RCInTurn, Text: "a"},
		{ID: "x1", Kind: protocol.RCInTurn, Text: "a"}, // a redelivery: its ack was lost
		{ID: "x2", Kind: protocol.RCInTurn, Text: "b"},
	}
	var mu sync.Mutex
	var polls []string // "<header>|<ack>" per poll
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		n := len(polls)
		polls = append(polls, r.Header.Get("X-Roger-RC-Ack")+"|"+r.URL.Query().Get("ack"))
		mu.Unlock()
		if n == 1 {
			// The second poll "never reaches the broker": the transport fails, so the host must
			// send the same ack again on the poll after.
			hj, ok := w.(http.Hijacker)
			require.True(t, ok)
			c, _, _ := hj.Hijack()
			_ = c.Close()
			return
		}
		i := n
		if n > 1 {
			i = n - 1
		}
		if i < len(answers) {
			_ = json.NewEncoder(w).Encode(answers[i])
			return
		}
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	rb := NewRCBridge(srv.URL, "s1", "tok")
	rb.Run()
	defer rb.Stop()
	var got []string
	for len(got) < 2 {
		select {
		case in := <-rb.Inbound():
			got = append(got, in.Text)
		case <-time.After(10 * time.Second):
			t.Fatalf("the host ran %v, want a then b", got)
		}
	}
	require.Equal(t, []string{"a", "b"}, got, "a redelivered id runs once")
	select {
	case in := <-rb.Inbound():
		t.Fatalf("an extra inbound reached the host: %+v", in)
	case <-time.After(200 * time.Millisecond):
	}
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return len(polls) >= 5 }, 5*time.Second, 10*time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"1|", "1|x1", "1|x1", "1|x1", "1|x2"}, polls[:5],
		fmt.Sprintf("every poll says it acks; each acks the last inbound until a poll is answered: %v", polls))
}

func TestRCDedupeIsBounded(t *testing.T) {
	d := NewRCDedupe()
	require.True(t, d.First("a"))
	require.False(t, d.First("a"))
	require.True(t, d.First(""), "an inbound with no id (an older broker) is always run")
	require.True(t, d.First(""))
	for i := 0; i < rcDedupeSize; i++ {
		d.First(fmt.Sprintf("n%d", i))
	}
	require.True(t, d.First("a"), "the oldest id is forgotten once the set is full")
}
