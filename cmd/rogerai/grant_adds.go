package main

// grant_adds.go: `roger grant add`, `remove` and `reprice`, and the "grants you added" half of
// `roger grant list` (features/grants/grant_surfaces.feature, rulings S1 to S6, S9 to S11).
// The data calls are in internal/client/grant_adds.go; this file reads flags and the secret,
// asks before a reprice, and prints.

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"rogerai.fm/roger/v7/internal/client"
)

// grantIn is where `grant add` reads a secret and `grant reprice` reads its answer; a test
// replaces it. A terminal gets a hidden prompt (S3).
var grantIn io.Reader = os.Stdin

func cmdGrantAdd(cfg config, args []string) error {
	secret := ""
	if len(args) > 0 {
		secret = args[0]
	} else {
		s, err := readGrantSecret()
		if err != nil {
			return err
		}
		secret = s
	}
	g, err := client.GrantAdd(cfg.Broker, secret)
	if err != nil {
		return err
	}
	printAddedGrant(os.Stdout, g)
	return nil
}

// readGrantSecret reads the secret without echoing it on a terminal, or one line from a pipe.
func readGrantSecret() (string, error) {
	if f, ok := grantIn.(*os.File); ok && term.IsTerminal(f.Fd()) {
		fmt.Fprint(os.Stderr, "paste the grant secret (it is not shown): ")
		b, err := term.ReadPassword(f.Fd())
		fmt.Fprintln(os.Stderr)
		return strings.TrimSpace(string(b)), err
	}
	line, err := bufio.NewReader(grantIn).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// grantPriceWords is a grant price as a person reads it.
func grantPriceWords(in, out float64) string {
	switch {
	case in == 0 && out == 0:
		return "free"
	case in == 0:
		return fmt.Sprintf("$%s per 1M output tokens", fmtPrice(out))
	case out == 0:
		return fmt.Sprintf("$%s per 1M input tokens", fmtPrice(in))
	}
	return fmt.Sprintf("$%s per 1M input and $%s per 1M output tokens", fmtPrice(in), fmtPrice(out))
}

func fmtPrice(v float64) string {
	s := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", v), "0"), ".")
	if !strings.Contains(s, ".") {
		s += ".00"
	} else if len(s)-strings.Index(s, ".") == 2 {
		s += "0"
	}
	return s
}

// printAddedGrant shows what the broker showed for an add (S4): price, stations with their
// public price, a line where the grant is dearer, and the models. Never the secret.
func printAddedGrant(w io.Writer, g client.AddedGrant) {
	fmt.Fprintf(w, "added %q: %s\n", g.Name, grantPriceWords(g.PriceIn, g.PriceOut))
	if len(g.Stations) == 0 {
		fmt.Fprintln(w, "it covers no station yet: the owner has none on air. Add it again later to pick them up.")
		return
	}
	fmt.Fprintf(w, "  %-20s %-14s %s\n", "STATION", "PUBLIC PRICE", "MODELS")
	for _, s := range g.Stations {
		fmt.Fprintf(w, "  %-20s %-14s %s\n", s.ID, "$"+fmtPrice(s.PublicOut), strings.Join(s.Models, ", "))
	}
	for _, s := range g.Stations {
		if s.AbovePublic {
			fmt.Fprintf(w, "note: on %s the grant price is above the public price.\n", s.ID)
		}
	}
	fmt.Fprintln(w, "your requests to these stations are now charged at this price, from your own wallet.")
}

// printAddedList is the "grants you added" section of `roger grant list` (S1, S9).
func printAddedList(w io.Writer, broker string) error {
	fmt.Fprintln(w, "grants you added")
	added, _, err := client.GrantAddedList(broker)
	if err != nil {
		return err
	}
	if len(added) == 0 {
		fmt.Fprintln(w, "  none - `roger grant add` adds a grant someone gave you")
		return nil
	}
	fmt.Fprintf(w, "  %-16s %-34s %-12s %s\n", "NAME", "PRICE", "ADDED", "STATE")
	for _, g := range added {
		fmt.Fprintf(w, "  %-16s %-34s %-12s %s\n", g.Name, grantPriceWords(g.PriceIn, g.PriceOut),
			time.Unix(g.AddedAt, 0).UTC().Format("2006-01-02"), g.StateWords())
	}
	return nil
}

func cmdGrantList(cfg config) error {
	if !signedInHere() {
		return client.ErrNotSignedIn
	}
	fmt.Println("grants you made")
	if err := client.GrantList(cfg.Broker); err != nil {
		fmt.Printf("  %v\n", err)
	}
	fmt.Println()
	return printAddedList(os.Stdout, cfg.Broker)
}

// signedInHere is client.SignedIn, as a bool.
func signedInHere() bool { _, ok := client.SignedIn(); return ok }

func cmdGrantRemove(cfg config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: roger grant remove <name|id>")
	}
	g, err := client.GrantRemove(cfg.Broker, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("removed %q from your account. Your requests to its stations are charged the public price again.\n", g.Name)
	return nil
}

// cmdGrantRevoke is `roger grant revoke`: a name that is only a grant you added is refused,
// naming `remove` (S2).
func cmdGrantRevoke(cfg config, name string) error {
	err := client.GrantRevoke(cfg.Broker, name)
	if err != nil && strings.HasPrefix(err.Error(), "no grant named") && client.GrantAddedOnlyByMe(cfg.Broker, name) {
		return fmt.Errorf("%q is a grant you added, not one you made - `roger grant remove %s` drops it from your account", name, name)
	}
	return err
}

// grantEditFlags reads `<name|id>` and the price and scope flags shared by `grant set` and
// `grant reprice`. A list flag given empty means every station or model: it is sent as [],
// never null (null keeps the grant's own).
func grantEditFlags(verb, usage string, args []string) (string, client.RepriceOpts, bool, error) {
	var o client.RepriceOpts
	fs := flag.NewFlagSet("grant "+verb, flag.ContinueOnError)
	priceOut := fs.String("price-out", "", "$/1M output tokens, or \"free\"")
	priceIn := fs.String("price-in", "", "$/1M input tokens")
	nodes := fs.String("nodes", "", "the stations it covers (comma-separated)")
	models := fs.String("models", "", "the models it covers (comma-separated)")
	yes := fs.Bool("yes", false, "do not ask first")
	fs.Usage = func() { fmt.Print(usage) }
	ref, rest := "", args
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		ref, rest = rest[0], rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return "", o, false, err
	}
	if ref == "" && fs.NArg() > 0 {
		ref = fs.Arg(0)
	}
	if ref == "" {
		fs.Usage()
		return "", o, false, fmt.Errorf("usage: %s", strings.SplitN(usage, "\n", 2)[0])
	}
	for _, p := range []struct {
		raw string
		dst **float64
	}{{*priceOut, &o.PriceOut}, {*priceIn, &o.PriceIn}} {
		if p.raw == "" {
			continue
		}
		v, err := client.ParseGrantPrice(p.raw)
		if err != nil {
			return "", o, false, err
		}
		*p.dst = &v
	}
	if flagPassed(fs, "nodes") {
		l := append([]string{}, splitCSV(*nodes)...)
		o.Nodes = &l
	}
	if flagPassed(fs, "models") {
		l := append([]string{}, splitCSV(*models)...)
		o.Models = &l
	}
	return ref, o, *yes, nil
}

const grantRepriceUsage = `roger grant reprice <name|id> --price-out <P> [--price-in <P>] [--nodes a,b] [--models m] [--yes]

A reprice issues a NEW secret and ends the old one. Everyone who added the grant is on the
public price until they add the new secret, which shows them the new price first.
A price cut does not need this: ` + "`roger grant set`" + ` keeps the same secret.
`

const grantSetUsage = `roger grant set <name|id> [--price-out <P>] [--price-in <P>] [--nodes a,b] [--models m]

Edits a grant you made and keeps its secret. A price cut and fewer stations or models apply
at once. A higher price, or more stations or models on a priced grant, needs a new secret:
use ` + "`roger grant reprice`" + `.
`

func cmdGrantReprice(cfg config, args []string) error {
	ref, o, yes, err := grantEditFlags("reprice", grantRepriceUsage, args)
	if err != nil {
		return err
	}
	if o.PriceIn == nil && o.PriceOut == nil {
		return errors.New(client.MsgRepriceNoPrice)
	}
	if !signedInHere() {
		return client.ErrNotSignedIn
	}
	id, name, curIn, curOut, err := client.GrantCurrentPrice(cfg.Broker, ref)
	if err != nil {
		return err
	}
	newIn, newOut := curIn, curOut
	if o.PriceIn != nil {
		newIn = *o.PriceIn
	}
	if o.PriceOut != nil {
		newOut = *o.PriceOut
	}
	if o.Nodes == nil && o.Models == nil && newIn <= curIn && newOut <= curOut && (newIn < curIn || newOut < curOut) {
		return fmt.Errorf("%v - %s", client.ErrGrantCut, grantCutHow(name, newIn, newOut))
	}
	if !yes && !askYes(fmt.Sprintf("reprice %q to %s? The old secret stops working, and everyone who added it goes back to the public price until they add the new one. [y/N] ", name, grantPriceWords(newIn, newOut))) {
		fmt.Println("nothing changed.")
		return nil
	}
	res, err := client.GrantReprice(cfg.Broker, id, o)
	if err != nil {
		return err
	}
	fmt.Printf("repriced %q to %s. The new secret, shown once:\n\n  %s\n\n", res.Name, grantPriceWords(newIn, newOut), res.Secret)
	fmt.Println("the old secret has stopped working.")
	switch res.Dropped {
	case 0:
		fmt.Println("no account had added it.")
	case 1:
		fmt.Println("1 account that added it is on the public price until they add the new one.")
	default:
		fmt.Printf("%d accounts that added it are on the public price until they add the new one.\n", res.Dropped)
	}
	return nil
}

// askYes asks a yes/no question on stderr, default no.
func askYes(q string) bool {
	fmt.Fprint(os.Stderr, q)
	line, _ := bufio.NewReader(grantIn).ReadString('\n')
	a := strings.ToLower(strings.TrimSpace(line))
	return a == "y" || a == "yes"
}

// grantCutHow is how to make a price cut, which keeps the same secret.
func grantCutHow(name string, in, out float64) string {
	flags := "--price-out " + fmtPrice(out)
	if in > 0 {
		flags += " --price-in " + fmtPrice(in)
	}
	return "`roger grant set " + name + " " + flags + "` lowers it at once"
}

// cmdGrantSet is `roger grant set`: an edit that keeps the secret (ruling S12). The broker
// refuses a raise or a widening of a priced grant, naming reprice.
func cmdGrantSet(cfg config, args []string) error {
	ref, o, _, err := grantEditFlags("set", grantSetUsage, args)
	if err != nil {
		return err
	}
	g, err := client.GrantSet(cfg.Broker, ref, o)
	if err != nil {
		if strings.HasPrefix(err.Error(), "no grant named") && client.GrantAddedOnlyByMe(cfg.Broker, ref) {
			return fmt.Errorf("%q is a grant you added, not one you made - `roger grant remove %s` drops it from your account", ref, ref)
		}
		return err
	}
	fmt.Printf("updated %q, same secret: %s, models %s.\n", g.Name, grantPriceWords(g.PriceIn, g.PriceOut), scopeWords(g.Models))
	return nil
}

func scopeWords(l []string) string {
	if len(l) == 0 {
		return "any"
	}
	return strings.Join(l, ", ")
}
