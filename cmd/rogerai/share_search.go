package main

// share_search.go is `roger share --search <engine>`, which puts a search engine on the market
// as a search station (features/search/share_cli.feature), and `roger search check`, which runs
// the search/1 contract checks against a native service without going on air
// (features/search/native_provider.feature). One flag picks the engine; roger checks the engine
// works before anything goes on air.

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"

	"rogerai.fm/roger/v7/internal/agent"
	"rogerai.fm/roger/v7/internal/client"
	"rogerai.fm/roger/v7/internal/protocol"
	"rogerai.fm/roger/v7/internal/searchengine"
)

// braveEndpoint is where a Brave share sends its queries (a var so a test can point it at a
// local engine; production never assigns it).
var braveEndpoint = searchengine.BraveEndpoint

// searchShareUsage is `roger share --search <engine> --help`.
const searchShareUsage = `roger share --search <engine> - put a search engine on the market (docs/search-providers.md)

  roger share --search brave --price-per-1k 5          Brave, with your own key
  roger share --search searxng --upstream <url>        a SearXNG you run
  roger share --search native --upstream <url>         your own index (test it: roger search check)

  --model <id>          the station id (default web-search; native: search)
  --price-per-1k <$>    dollars per 1,000 queries (0 = free; brave must set one)
  --scope "<line>"      what your index covers, up to 120 characters
  --upstream-key <k>    the engine's key (brave: else BRAVE_SEARCH_API_KEY)
  --private             a private band: never listed, reachable only with its code
  --accept-engine-terms accept the engine's terms in a script

A search station needs a signed-in owner: run ` + "`roger login`" + ` first.
`

// searchCheckUsage is `roger search check --help`.
const searchCheckUsage = `roger search check --upstream <url> [--upstream-key <key>]

Runs the search/1 contract checks against your own search service, without going on air.
Run it until every check passes, then share it with roger share --search native.
`

// bandOf reads a session's private band (a seam: a test session has no broker behind it).
var bandOf = func(s *agent.Session) (id, code, display string) { return s.Band() }

// searchOnAirWait is how long a search share waits to see the broker accept it.
var searchOnAirWait = 3 * time.Second

// searchCheckQuery is the one fixed query a share sends its engine before going on air (C5).
const searchCheckQuery = "weather"

// searchShareIO is everything a search share reads from or writes to the person running it.
type searchShareIO struct {
	getenv      func(string) string
	interactive bool
	in          *bufio.Reader
	readSecret  func() (string, error)
	out         io.Writer
}

func defaultSearchShareIO() searchShareIO {
	tty := isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
	return searchShareIO{
		getenv:      os.Getenv,
		interactive: tty,
		in:          bufio.NewReader(os.Stdin),
		readSecret: func() (string, error) {
			b, err := term.ReadPassword(os.Stdin.Fd())
			fmt.Println()
			return string(b), err
		},
		out: os.Stdout,
	}
}

// wantsSearchShare reports whether a `roger share` command line is a search share.
func wantsSearchShare(args []string) bool {
	for _, a := range args {
		if a == "--search" || a == "-search" || strings.HasPrefix(a, "--search=") || strings.HasPrefix(a, "-search=") {
			return true
		}
	}
	return false
}

// acceptedSearchTerms reports whether the provider already answered yes for engine's terms.
func (c config) acceptedSearchTerms(engine string) bool {
	return slices.Contains(c.SearchTerms, engine)
}

// shareSearch is `roger share --search <engine> ...`.
func shareSearch(cfg config, args []string, sio searchShareIO) error {
	if cfg.lan.LAN {
		return errors.New("search stations serve the market: --lan does not apply to --search")
	}
	fs := flag.NewFlagSet("share", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	engine := fs.String("search", "", "put a search engine on the market: brave | searxng | native")
	upstream := fs.String("upstream", "", "the engine's address (searxng, native; brave defaults to Brave's API)")
	upKey := fs.String("upstream-key", "", "the engine's key (brave: else BRAVE_SEARCH_API_KEY)")
	model := fs.String("model", "", "the offer id (default web-search; native: search)")
	scope := fs.String("scope", "", "one line saying what the index covers")
	per1K := fs.Float64("price-per-1k", 0, "dollars per 1,000 queries (0 = free)")
	accept := fs.Bool("accept-engine-terms", false, "accept that you answer for the engine's terms (for scripts)")
	node := fs.String("node", "", "station callsign; persisted")
	broker := fs.String("broker", cfg.Broker, "broker URL")
	private := fs.Bool("private", false, "go on air as a private band")
	// flags that mean something else on a search share: refused with the reason below
	fs.String("modality", "", "")
	fs.String("curated", "", "")
	fs.Float64("price-in", 0, "")
	fs.Float64("price-out", 0, "")
	if err := fs.Parse(args); errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(sio.out, searchShareUsage)
		return nil
	} else if err != nil {
		return err
	}
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	kind := strings.TrimSpace(*engine)
	switch {
	case !protocol.ValidEngine(kind):
		return fmt.Errorf("unknown search engine %q: use brave, searxng or native", kind)
	case set["modality"]:
		return errors.New("--search already sets the modality; drop --modality")
	case set["curated"]:
		return errors.New("a search station cannot be curated; the engine label already says what it is")
	case set["price-in"] || set["price-out"]:
		return errors.New("search is priced with --price-per-1k (dollars per 1,000 queries)")
	}
	up := strings.TrimSpace(*upstream)
	switch {
	case up == "" && kind == protocol.EngineSearXNG:
		return errors.New("--search searxng needs --upstream, the address of your SearXNG")
	case up == "" && kind == protocol.EngineNative:
		return errors.New("--search native needs --upstream, the address of your search service")
	case up == "":
		up = braveEndpoint
	}
	if u, err := url.Parse(up); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("--upstream must be an http or https address")
	}
	price := *per1K
	if math.IsNaN(price) || math.IsInf(price, 0) || price < 0 {
		return errors.New("--price-per-1k must be a number of dollars at or above 0")
	}
	if kind == protocol.EngineBrave && !set["price-per-1k"] {
		return errors.New("Brave charges you per query, so set --price-per-1k (0 is allowed if you mean it)")
	}
	mdl := strings.TrimSpace(*model)
	if mdl == "" {
		mdl = "web-search"
		if kind == protocol.EngineNative {
			mdl = "search"
		}
	}

	// Every search station has a signed-in owner, free or paid (owner_required.feature OR2):
	// say so before any terms question or check query.
	if client.LinkedLogin() == "" {
		return errors.New("a search station needs a signed-in owner: run `roger login` first")
	}

	// Engine terms (C6): brave and searxng serve someone else's results, so the provider
	// answers for those terms. Asked once per engine; a script passes --accept-engine-terms.
	changed := false
	if kind != protocol.EngineNative && !cfg.acceptedSearchTerms(kind) {
		if !*accept {
			if !sio.interactive {
				return fmt.Errorf("first %s share: you answer for %s's terms for serving results to others - pass --accept-engine-terms to accept in a script", kind, kind)
			}
			fmt.Fprintf(sio.out, "you are responsible for following %s's terms for serving results to others.\ncontinue? [y/N] ", kind)
			line, _ := sio.in.ReadString('\n')
			if a := strings.ToLower(strings.TrimSpace(line)); a != "y" && a != "yes" {
				return fmt.Errorf("not sharing: %s terms not accepted", kind)
			}
		}
		cfg.SearchTerms = append(cfg.SearchTerms, kind)
		changed = true
	}

	// The engine key (C7): stays in the local config only, and is never printed back.
	key := strings.TrimSpace(*upKey)
	if kind == protocol.EngineBrave {
		if key == "" {
			key = strings.TrimSpace(sio.getenv("BRAVE_SEARCH_API_KEY"))
		}
		if key == "" {
			key = cfg.SearchKeys[kind]
		}
		if key == "" {
			if !sio.interactive {
				return errors.New("Brave needs a key: pass --upstream-key or set BRAVE_SEARCH_API_KEY")
			}
			fmt.Fprint(sio.out, "Brave API key (input hidden): ")
			k, err := sio.readSecret()
			if err != nil || strings.TrimSpace(k) == "" {
				return errors.New("no Brave key given")
			}
			key = strings.TrimSpace(k)
			if cfg.SearchKeys == nil {
				cfg.SearchKeys = map[string]string{}
			}
			cfg.SearchKeys[kind] = key
			changed = true
		}
	}
	if changed {
		if err := saveConfig(cfg); err != nil {
			return err
		}
	}
	keyNote := "no key"
	if key != "" {
		keyNote = "key set"
	}
	fmt.Fprintf(sio.out, "search engine: %s (%s)\n", kind, keyNote)

	// One check query before anything goes on air (C5).
	eng := searchengine.Engine{Kind: kind, Upstream: up, Key: key}
	n, err := checkSearchShare(context.Background(), eng)
	if err != nil {
		fmt.Fprintf(sio.out, "check failed: %v\n", err)
		return fmt.Errorf("search station not on air: %w", err)
	}
	if n == 0 {
		fmt.Fprintf(sio.out, "check passed: 0 results for %q (fine for a focused index)\n", searchCheckQuery)
	} else {
		fmt.Fprintf(sio.out, "check passed: %d results\n", n)
	}

	station := agentSlugStation(*node)
	if station != "" {
		saveStation(station)
	} else {
		station = loadOrCreateStation()
	}
	nodeID := agent.ShareNodeID(station, mdl, 0)
	acfg := agent.Config{
		Broker: *broker, Upstream: up, UpstreamKey: key,
		NodeID: nodeID, Station: station, Region: "home", Model: mdl,
		Modality: protocol.ModalitySearch, SearchEngine: kind, Scope: strings.TrimSpace(*scope),
		PriceIn: price * 1000, Parallel: 4, Private: *private,
	}
	release, err := acquireOnAirLock(nodeID, station, mdl)
	if err != nil {
		return err
	}
	defer release()
	sess, err := agentStart(acfg)
	if err != nil {
		return err
	}
	waitOnAir(sess, searchOnAirWait)
	mode := "free"
	if price > 0 {
		mode = fmt.Sprintf("$%s per 1,000 queries", trimAmt(price))
	}
	fmt.Fprintf(sio.out, "on air: search station %s as @%s (%s, %s)\n", mdl, station, kind, mode)
	if *private {
		_, code, display := bandOf(sess)
		printFrequencyCode(sio.out, mdl, code, display, "POST /v1/search with the header X-Roger-Freq set to the code above")
	}
	shareBlock()
	return nil
}

// checkSearchShare sends the engine the one fixed check query and explains a failure in the
// provider's terms.
func checkSearchShare(ctx context.Context, e searchengine.Engine) (int, error) {
	rs, sent, err := e.Check(ctx, searchengine.Request{Query: searchCheckQuery, Count: 3})
	switch {
	case errors.Is(err, searchengine.ErrUnreachable):
		return 0, errors.New("could not reach your search engine at the --upstream address")
	case errors.Is(err, searchengine.ErrTimeout):
		return 0, errors.New("the engine did not answer within 10 seconds")
	case errors.Is(err, searchengine.ErrJSONOff):
		return 0, errors.New("SearXNG's JSON output is off: add json to search.formats in settings.yml")
	case errors.Is(err, searchengine.ErrKeyRefused) && e.Kind == protocol.EngineBrave:
		return 0, errors.New("Brave refused the key")
	case errors.Is(err, searchengine.ErrUnreadable):
		return 0, errors.New("the engine's reply does not match the search contract")
	case err != nil:
		return 0, err
	case len(rs) == 0 && sent > 0:
		// Results came back and none was usable (no http(s) URL). A well-formed EMPTY reply
		// passes: a focused index has nothing for "weather" (founder ruling 2026-11-25).
		return 0, errors.New("the engine returned no usable results for the check query")
	}
	return len(rs), nil
}

// cmdSearch is `roger search check --upstream <url> [--upstream-key K]`: the search/1 contract
// checks against a native service, without going on air.
func cmdSearch(cfg config, args []string, out io.Writer) error {
	if len(args) == 0 || args[0] != "check" {
		return errors.New("usage: roger search check --upstream <url> [--upstream-key <key>]")
	}
	fs := flag.NewFlagSet("search check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	upstream := fs.String("upstream", "", "the search service's address")
	upKey := fs.String("upstream-key", "", "the service's bearer key, if it needs one")
	if err := fs.Parse(args[1:]); errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(out, searchCheckUsage)
		return nil
	} else if err != nil {
		return err
	}
	if u, err := url.Parse(*upstream); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("--upstream must be an http or https address")
	}
	keyNote := "no key"
	if *upKey != "" {
		keyNote = "key set"
	}
	fmt.Fprintf(out, "checking %s against %s (%s)\n", *upstream, searchengine.ContractVersion, keyNote)
	rs := searchengine.CheckContract(context.Background(), searchengine.Engine{Kind: protocol.EngineNative, Upstream: *upstream, Key: *upKey})
	for _, r := range rs {
		line := fmt.Sprintf("  %-4s  %s", r.Status, r.Name)
		if r.Detail != "" {
			line += ": " + r.Detail
		}
		fmt.Fprintln(out, line)
	}
	if searchengine.Failed(rs) {
		return errors.New("the service does not meet the search contract yet")
	}
	fmt.Fprintln(out, "the service meets the search contract: share it with `roger share --search native --upstream "+*upstream+"`")
	return nil
}
