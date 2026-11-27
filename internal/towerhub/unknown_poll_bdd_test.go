package towerhub

// Contract: features/tower/station_mobility_idempotency.feature.
//
// The hub half of "A node is never stranded": a poll from a Station the hub does not know yet asks
// Core (through OnUnknownStation, which in `roger tower serve` re-fetches Core's registrations)
// before answering 401, and a flood of unknown ids cannot make it ask more than once a second.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"rogerai.fm/roger/v7/internal/bddrun"
)

type hupState struct {
	t      *testing.T
	s      *Server
	srv    *httptest.Server
	id     *testNode
	asked  atomic.Int64
	err    error
	bodies []string
}

func (h *hupState) justAttached() error {
	h.id = newTestNode(h.t)
	h.s = NewServer(New(), stubCheck, ServerOptions{TowerID: testTowerID, EpochKey: testHubKey,
		SubmitTTL: time.Second, PollTTL: 50 * time.Millisecond})
	// Core knows the Station; the hub learns it only when it asks.
	h.s.OnUnknownStation = func(station string) {
		h.asked.Add(1)
		if station == "st-new" {
			h.s.RegisterNode("st-new", h.id.auth())
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc(PathPoll, h.s.Poll)
	h.srv = httptest.NewServer(mux)
	h.t.Cleanup(h.srv.Close)
	return nil
}

func (h *hupState) pollsBeforeRefresh() error {
	c := h.id.client(h.srv.URL, 0)
	_, _, h.err = c.PollJob(context.Background(), "st-new")
	return nil
}

func (h *hupState) askedCore() error {
	if h.asked.Load() < 1 {
		return fmt.Errorf("the hub answered without asking Core about the unknown Station")
	}
	return nil
}

func (h *hupState) servedOnThatPoll() error {
	if h.err != nil {
		return fmt.Errorf("the poll was refused: %v", h.err)
	}
	return nil
}

func (h *hupState) flood() error {
	if err := h.justAttached(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := h.id.client(h.srv.URL, 0)
			_, _, err := c.PollJob(context.Background(), fmt.Sprintf("st-random-%d", i))
			mu.Lock()
			if err != nil {
				h.bodies = append(h.bodies, err.Error())
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return nil
}

func (h *hupState) debounced() error {
	if n := h.asked.Load(); n > 2 {
		return fmt.Errorf("200 unknown polls in under a second made the hub ask Core %d times", n)
	}
	return nil
}

func (h *hupState) same401Words() error {
	if len(h.bodies) != 200 {
		return fmt.Errorf("%d of 200 unknown polls were refused", len(h.bodies))
	}
	for _, b := range h.bodies {
		if !strings.Contains(b, "401") || !strings.Contains(b, "no node is registered for this Station on this tower") {
			return fmt.Errorf("an unknown Station got other words: %s", b)
		}
	}
	return nil
}

func TestUnknownStationPollBDD(t *testing.T) {
	h := &hupState{t: t}
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*h = hupState{t: t}
				return ctx, nil
			})
			sc.Step(`^a signed-in operator whose node is registered and co-signed by their account$`, func() error { return nil })
			sc.Step(`^two live joined Towers, A and B, that may take that node's Station$`, func() error { return nil })
			sc.Step(`^the node's Station S is attached to Tower A at epoch 4$`, func() error { return nil })
			sc.Step(`^S just attached to Tower B$`, h.justAttached)
			sc.Step(`^S's node polls Tower B's hub before B's next 30-second refresh$`, h.pollsBeforeRefresh)
			sc.Step(`^the hub asks Core about the unknown Station before answering 401$`, h.askedCore)
			sc.Step(`^a Station Core confirms is served on that poll$`, h.servedOnThatPoll)
			sc.Step(`^a caller polls Tower B's hub with many unknown Station ids$`, h.flood)
			sc.Step(`^the hub asks Core no more often than its existing debounce allows$`, h.debounced)
			sc.Step(`^a Station Core does not confirm gets the same 401 words as today$`, h.same401Words)
		},
		Options: &godog.Options{Format: "pretty", Strict: true, TestingT: t,
			Paths: bddrun.Scenarios("../../features/tower/station_mobility_idempotency.feature", bddrun.Titled(false,
				"The hub learns a new Station on a poll, not only on a submit",
				"Polls naming random Station ids cannot make the hub flood Core",
			))},
	}
	if bddrun.Run(suite) != 0 {
		t.Fatal("station_mobility_idempotency.feature (unknown poll, hub) failed")
	}
}
