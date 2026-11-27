package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"

	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/netbind"
)

// channel.go - the console's ENDPOINT PLATE (features/console/endpoint_plate.feature):
// the BASE URL / API KEY / MODEL plate `roger use` and the TUI end on, opened from the browser.
//
// It serves client.ProxyHandlerLive, the handler those two already serve, so the endpoint's
// own behaviour is not written again here. This file is only opening, showing, re-pointing
// and closing it.
//
// THE KEY LEAVES IN TWO REPLIES AND NOWHERE ELSE: the open that minted it and an explicit
// POST for it. Never the status, the state snapshot or the event stream.

// channelAddr is where the scan for a free port starts, the same 4141 `roger use` starts at.
const channelAddr = "127.0.0.1:4141"

// channelBodyCap bounds an open request. It carries a model id and a price.
const channelBodyCap = 4 << 10

// channelListen binds the endpoint. A variable so a test can make the bind fail.
var channelListen = netbind.ListenFree

// consoleChannel is the console's own open channel, if any. Its own, not the TUI's, for the
// reason the console keeps its own agent session: the TUI's lives in the TUI.
type consoleChannel struct {
	mu      sync.Mutex
	hs      *http.Server
	ln      net.Listener
	holder  *client.ProxyOptionsHolder
	base    string
	key     string
	model   string
	station string
	priced  bool
	// closes counts every close asked for, so an open that was checked before a logout can
	// tell the logout happened and check again.
	closes int
}

type channelReply struct {
	OK            bool    `json:"ok"`
	Open          bool    `json:"open"`
	Message       string  `json:"message,omitempty"`
	BaseURL       string  `json:"base_url,omitempty"`
	Model         string  `json:"model,omitempty"`
	Station       string  `json:"station,omitempty"`
	Key           string  `json:"key,omitempty"`
	KeyHint       string  `json:"key_hint,omitempty"`
	LoginNeeded   bool    `json:"login_needed,omitempty"`
	OverLimit     bool    `json:"over_limit,omitempty"`
	ConfirmNeeded bool    `json:"confirm_needed,omitempty"`
	PriceOut      float64 `json:"price_out,omitempty"`
}

// channelQuote is what a channel on a band would cost and whether it would open.
type channelQuote struct {
	Model         string  `json:"model"`
	OnAir         bool    `json:"on_air"`
	Free          bool    `json:"free"`
	PriceIn       float64 `json:"price_in"`
	PriceOut      float64 `json:"price_out"`
	RangeMin      float64 `json:"range_min"`
	RangeMax      float64 `json:"range_max"`
	Stations      int     `json:"stations"`
	CapOut        float64 `json:"cap_out"`
	OverLimit     bool    `json:"over_limit"`
	ConfirmNeeded bool    `json:"confirm_needed"`
	station       string
	minTPS        float64
}

func writeChannel(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// capFor is the operator's out-price cap and speed floor for model: their own, or the consumer
// default cap when they set none, as `roger use` applies it.
func (s *Server) capFor(model string) (capOut, minTPS float64) {
	capOut = client.ConsumerDefaultMaxOut
	if s.opts.ReadLimits != nil {
		l := s.opts.ReadLimits()[model]
		minTPS = l.MinTPS
		if l.MaxOut > 0 {
			capOut = l.MaxOut
		}
	}
	return capOut, minTPS
}

// status is the plate without the key. Call with c.mu held.
func (c *consoleChannel) status() channelReply {
	if c.hs == nil {
		return channelReply{OK: true}
	}
	return channelReply{OK: true, Open: true, BaseURL: c.base, Model: c.model, Station: c.station, KeyHint: c.key[len(c.key)-4:]}
}

// quoteChannel prices model from the live offers and the operator's own cap for it.
func (s *Server) quoteChannel(model string) (channelQuote, error) {
	q := channelQuote{Model: model}
	q.CapOut, q.minTPS = s.capFor(model)
	offers, err := client.Discover(s.opts.Broker)
	if err != nil {
		return q, err
	}
	br, ok := client.BandRangeOf(offers, model)
	if !ok {
		return q, nil
	}
	q.OnAir, q.PriceIn, q.PriceOut, q.RangeMin, q.RangeMax, q.Stations, q.station = true, br.CheapIn, br.Min, br.Min, br.Max, br.Stations, br.CheapNode
	q.Free = br.Min == 0
	for _, o := range offers {
		if o.NodeID == br.CheapNode && o.Model == model && o.FreeNow {
			q.Free = true
		}
	}
	q.OverLimit = br.Min > q.CapOut
	q.ConfirmNeeded = br.Min > client.ConsumerConfirmThreshold
	return q, nil
}

// channelUnavailable answers the two cases where no channel can exist on this console.
func (s *Server) channelUnavailable(w http.ResponseWriter) bool {
	if s.opts.LANDial != nil {
		writeChannel(w, http.StatusOK, channelReply{Message: "the endpoint plate is not available in LAN mode yet - use `roger --lan use <model>` in a terminal"})
		return true
	}
	if s.opts.Broker == "" {
		writeChannel(w, http.StatusServiceUnavailable, channelReply{Message: "no broker configured - a channel needs one to reach a band"})
		return true
	}
	return false
}

// handleChannel reads the plate (GET): open or not, and where. Never the key.
func (s *Server) handleChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.channel.mu.Lock()
	st := s.channel.status()
	s.channel.mu.Unlock()
	writeChannel(w, http.StatusOK, st)
}

// handleChannelQuote prices a band before it is opened (GET ?model=). It opens nothing.
func (s *Server) handleChannelQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.channelUnavailable(w) {
		return
	}
	q, err := s.quoteChannel(strings.TrimSpace(r.URL.Query().Get("model")))
	if err != nil {
		writeChannel(w, http.StatusBadGateway, channelReply{Message: err.Error()})
		return
	}
	writeChannel(w, http.StatusOK, q)
}

// handleChannelOpen opens the endpoint on a band, or re-points the open one. Every refusal
// comes before the bind, so a refused open never leaves a listener or a key behind and a
// refused re-point leaves the open channel as it was.
func (s *Server) handleChannelOpen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model           string   `json:"model"`
		ConfirmPriceOut *float64 `json:"confirm_price_out"`
	}
	if code, msg := strictBody(w, r, &req, "malformed request"); code != 0 {
		writeChannel(w, code, channelReply{Message: msg})
		return
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		writeChannel(w, http.StatusBadRequest, channelReply{Message: "pick a model first"})
		return
	}
	if s.channelUnavailable(w) {
		return
	}
	c := &s.channel
	for {
		c.mu.Lock()
		closes := c.closes
		c.mu.Unlock()
		// The checks talk to the broker, so they run outside the lock.
		q, refusal, code := s.admitChannel(model, req.ConfirmPriceOut)
		if refusal != nil {
			writeChannel(w, code, refusal)
			return
		}
		headerRouting := client.NegotiateRouting(s.opts.Broker)
		c.mu.Lock()
		if c.closes != closes {
			// A close or a logout landed while this was being checked: check again.
			c.mu.Unlock()
			continue
		}
		defer c.mu.Unlock()
		opts := client.ProxyOptions{
			Broker: s.opts.Broker, User: s.opts.User, Model: model,
			MaxPriceOut: q.CapOut, MinTPS: q.minTPS, HeaderRouting: headerRouting,
			ReasoningFallbackOff: client.RawReasoningEnv(),
			// A turn through the plate is a session on this machine's Edge, as a chat turn is.
			OnReceipt: s.recordConsoleSession,
		}
		if c.hs == nil {
			ln, err := channelListen(channelAddr)
			if err != nil {
				writeChannel(w, http.StatusOK, channelReply{Message: "the endpoint could not bind: " + err.Error()})
				return
			}
			c.key = client.NewSessionKey()
			opts.SessionKey = c.key
			c.base = "http://" + ln.Addr().String() + "/v1"
			c.holder = client.NewProxyOptionsHolder(opts)
			c.ln, c.hs = ln, &http.Server{Handler: client.ProxyHandlerLive(c.holder)}
			go func(hs *http.Server, ln net.Listener) { _ = hs.Serve(ln) }(c.hs, ln)
		} else {
			// The SAME endpoint and key, re-pointed: a tool already configured keeps working.
			c.holder.SetBand(opts)
		}
		// Priced by its LISTED price: a free-now window ends, and the channel outlives it.
		c.model, c.station, c.priced = model, q.station, q.PriceOut > 0
		out := c.status()
		out.Key = c.key
		writeChannel(w, http.StatusOK, out)
		return
	}
}

// admitChannel is every reason a channel on model may not open now: off air, a paid band
// with no login, a price above the operator's cap, an expensive band not typed back.
func (s *Server) admitChannel(model string, typed *float64) (channelQuote, *channelReply, int) {
	q, err := s.quoteChannel(model)
	switch {
	case err != nil:
		return q, &channelReply{Message: err.Error()}, http.StatusBadGateway
	case !q.OnAir:
		// A private band is not in the open market's list, so it ends here too: the console
		// reaches the open market only (chat.go).
		return q, &channelReply{Message: fmt.Sprintf("no station on air for %q right now", model)}, http.StatusOK
	case !q.Free && !s.loggedIn():
		return q, &channelReply{LoginNeeded: true, Message: "this band is paid - log in on ACCOUNT to use your wallet"}, http.StatusOK
	case q.OverLimit:
		return q, &channelReply{OverLimit: true, PriceOut: q.PriceOut,
			Message: fmt.Sprintf("cheapest on air is %.2f $/1M out, above your max of %.2f - raise it in SETTINGS", q.PriceOut, q.CapOut)}, http.StatusOK
	case q.ConfirmNeeded && (typed == nil || cents(*typed) != cents(q.PriceOut)):
		return q, &channelReply{ConfirmNeeded: true, PriceOut: q.PriceOut,
			Message: fmt.Sprintf("this band is %.2f $/1M out - type that price to open it", q.PriceOut)}, http.StatusOK
	}
	return q, nil, http.StatusOK
}

// cents compares a typed price with a quoted one at the precision the page shows.
func cents(usd float64) int64 { return int64(math.Round(usd * 100)) }

// loggedIn asks the broker, as ACCOUNT does, and falls back to what this run already knows.
func (s *Server) loggedIn() bool {
	if bal, err := client.FetchBalance(s.opts.Broker, s.opts.User); err == nil {
		return bal.LoggedIn
	}
	return s.ctrl.LoggedIn()
}

// handleChannelKey hands the plate WITH its key to the page that asked for it (POST only).
func (s *Server) handleChannelKey(w http.ResponseWriter, r *http.Request) {
	s.channel.mu.Lock()
	out := s.channel.status()
	out.Key = s.channel.key
	s.channel.mu.Unlock()
	if !out.Open {
		writeChannel(w, http.StatusConflict, channelReply{Message: "no channel open"})
		return
	}
	writeChannel(w, http.StatusOK, out)
}

// handleChannelClose ends the endpoint and its key. Closing a closed channel is not an error.
func (s *Server) handleChannelClose(w http.ResponseWriter, r *http.Request) {
	s.closeChannel(false)
	writeChannel(w, http.StatusOK, channelReply{OK: true})
}

// closeChannel stops listening at once and lets a turn in flight finish and be receipted.
// pricedOnly closes only a channel on a paid band (a logout: a free band needs no account).
func (s *Server) closeChannel(pricedOnly bool) {
	c := &s.channel
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closes++
	if c.hs == nil || (pricedOnly && !c.priced) {
		return
	}
	// The listener closes now, so the port refuses at once, and a request not yet started
	// finds no band. Shutdown then lets a turn in flight finish and be receipted.
	_ = c.ln.Close()
	c.holder.Disconnect()
	go func(hs *http.Server) { _ = hs.Shutdown(context.Background()) }(c.hs)
	c.hs, c.ln, c.holder, c.key, c.base, c.model, c.station, c.priced = nil, nil, nil, "", "", "", "", false
}

// recapChannel re-reads the operator's cap for model onto the open channel, so a cap changed
// in SETTINGS binds the next request rather than the next channel.
func (s *Server) recapChannel(model string) {
	c := &s.channel
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hs == nil || c.model != model {
		return
	}
	opts := c.holder.Get()
	opts.MaxPriceOut, opts.MinTPS = s.capFor(model)
	c.holder.SetBand(opts)
}
