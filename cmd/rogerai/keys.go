package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"rogerai.fm/roger/v7/internal/client"
)

// keysSecretWarning follows a freshly minted secret, the only time it is ever shown.
const keysSecretWarning = "store it now - it is not shown again"

// cmdKeys is the account-keys verb group: list | mint | set | rm, over the same signed
// /account/keys endpoints the web account page uses (features/auth/key_guardrails.feature).
func cmdKeys(cfg config, args []string) error {
	return runKeys(cfg.Broker, args, os.Stdout, os.Stderr, isatty.IsTerminal(os.Stdout.Fd()))
}

func runKeys(broker string, args []string, stdout, stderr io.Writer, tty bool) error {
	if len(args) == 0 {
		keysUsage(stdout)
		return nil
	}
	switch args[0] {
	case "list", "ls":
		keys, err := client.KeysList(broker)
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			fmt.Fprintln(stdout, "no keys yet - `roger keys mint --name ci` mints one.")
			return nil
		}
		client.WriteKeysTable(stdout, keys)
		return nil
	case "mint", "new":
		fields, err := keysFields("mint", args[1:], false)
		if err != nil {
			return err
		}
		k, secret, err := client.KeysMint(broker, fields)
		if err != nil {
			return err
		}
		defer keysUnlimitedNote(stderr, k)
		if !tty {
			// A pipe gets the secret alone (machine-readable); the warning goes to the person.
			fmt.Fprintln(stdout, secret)
			fmt.Fprintf(stderr, "minted %s - %s\n", k.ID, keysSecretWarning)
			return nil
		}
		fmt.Fprintf(stdout, "minted %s\n\n  %s\n\n%s\n", k.ID, secret, keysSecretWarning)
		return nil
	case "set":
		if len(args) < 2 || strings.HasPrefix(args[1], "-") {
			return fmt.Errorf("usage: roger keys set <key> [--name N] [--limit $] [--reset R] [--models a,b] [--nodes a,b] [--expires T] [--disable|--enable]")
		}
		fields, err := keysFields("set", args[2:], true)
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return fmt.Errorf("nothing to change - pass at least one flag (see `roger keys`)")
		}
		id, _, err := keysResolve(broker, args[1])
		if err != nil {
			return err
		}
		k, err := client.KeysSet(broker, id, fields)
		if err != nil {
			return err
		}
		client.WriteKeysTable(stdout, []client.AccountKey{k})
		if _, limited := fields["limit_usd"]; limited {
			keysUnlimitedNote(stderr, k)
		}
		return nil
	case "revoke", "rm", "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: roger keys revoke <key>")
		}
		id, label, err := keysResolve(broker, args[1])
		if err != nil {
			return err
		}
		if err := client.KeysRevoke(broker, id); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "revoked %s\n", label)
		return nil
	case "-h", "--help", "help":
		keysUsage(stdout)
		return nil
	}
	return fmt.Errorf("unknown keys command %q (try list|mint|set|revoke)", args[0])
}

// keysResolve turns what was typed into a key id. An id is used as it is. Anything else is a
// name, looked up in the account's own list as the TUI does (client.FindKey), so nothing typed
// becomes part of a broker path. label is how to name the key back to the person.
func keysResolve(broker, arg string) (id, label string, err error) {
	if client.KeyIDOK(arg) {
		return arg, arg, nil
	}
	keys, err := client.KeysList(broker)
	if err != nil {
		return "", "", err
	}
	k, err := client.FindKey(keys, arg)
	if errors.Is(err, client.ErrKeyNotFound) {
		err = fmt.Errorf("%w - `roger keys list` shows them", err)
	}
	return k.ID, fmt.Sprintf("%s (%s)", k.Name, k.ID), err
}

// keysUnlimitedNote tells the person, never the pipe, that a key has no limit.
func keysUnlimitedNote(stderr io.Writer, k client.AccountKey) {
	if k.LimitUSD <= 0 {
		fmt.Fprintln(stderr, client.KeyUnlimitedNote)
	}
}

// keysFields parses the key flags into the JSON fields the broker takes, naming only the
// flags actually passed (a PATCH changes nothing it does not name).
func keysFields(verb string, args []string, patch bool) (map[string]any, error) {
	fs := flag.NewFlagSet("keys "+verb, flag.ExitOnError)
	name := fs.String("name", "", "label shown in the key list")
	limit := fs.String("limit", "", "spend limit in $ per window (none = unlimited)")
	reset := fs.String("reset", "", "limit window: daily, weekly, monthly or none")
	models := fs.String("models", "", "allow only these models (comma-separated; empty = any)")
	nodes := fs.String("nodes", "", "allow only these stations (comma-separated; empty = any)")
	expires := fs.String("expires", "", "expiry: 30d, 12h, 2026-12-31 or never")
	disable := fs.Bool("disable", false, "turn the key off (set only)")
	enable := fs.Bool("enable", false, "turn the key back on (set only)")
	_ = fs.Parse(args)
	out := map[string]any{}
	var bad error
	fs.Visit(func(f *flag.Flag) {
		if bad != nil {
			return
		}
		switch f.Name {
		case "name":
			out["name"] = *name
		case "limit":
			v, err := client.ParseKeyLimit(*limit)
			if err != nil {
				bad = fmt.Errorf("--limit %q: %w", *limit, err)
				return
			}
			out["limit_usd"] = v
		case "reset":
			if !client.KeyResetOK(*reset) {
				bad = fmt.Errorf("--reset %q: daily, weekly, monthly or none", *reset)
				return
			}
			out["reset"] = *reset
		case "models":
			out["allowed_models"] = nonNilCSV(*models)
		case "nodes":
			out["allowed_nodes"] = nonNilCSV(*nodes)
		case "expires":
			if strings.EqualFold(strings.TrimSpace(*expires), "never") {
				out["expires_at"] = nil
				return
			}
			t, err := parseExpires(*expires)
			if err != nil {
				bad = err
				return
			}
			out["expires_at"] = time.Unix(t, 0).UTC().Format(time.RFC3339)
		case "disable", "enable":
			if !patch {
				bad = fmt.Errorf("--%s applies to `roger keys set`", f.Name)
				return
			}
			out["disabled"] = *disable && !*enable
		}
	})
	if *disable && *enable {
		return nil, fmt.Errorf("--disable and --enable together")
	}
	return out, bad
}

// nonNilCSV is splitCSV that sends an empty list (clear the allow-list) rather than null.
func nonNilCSV(s string) []string {
	if v := splitCSV(s); v != nil {
		return v
	}
	return []string{}
}

func keysUsage(w io.Writer) {
	fmt.Fprint(w, `roger keys - account API keys (rog-key_...) for scripts, CI and other tools

  roger keys list                                    your keys, their limits and use
  roger keys mint --name ci --limit 5 --reset weekly mint a key (the secret is shown ONCE)
  roger keys set <key> --limit 10 --models a,b       change a key
  roger keys set <key> --disable                     turn a key off (--enable to turn it back on)
  roger keys revoke <key>                            revoke a key

A <key> is its id or its name.

  --name <label>  --limit <$|none>  --reset daily|weekly|monthly|none
  --models a,b  --nodes a,b  --expires 30d|2026-12-31|never

A limit is in dollars. Use --limit none for an unlimited key; 0 is refused, because it
would mean unlimited and not no spending. --limit is optional: leaving it out also mints an
unlimited key, and says so.

A key pays from your account. Use one with `+"`roger use <model> --key rog-key_...`"+`, or as the
OpenAI API key for any tool pointed at the broker. Keys need `+"`roger login`"+`.
`)
}
