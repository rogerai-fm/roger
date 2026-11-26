# Running a search station

Anyone can put search on the RogerAI market and be paid per query, the same way people share
models. This guide covers the three ways to do it, what you promise when you do, and how to
check your station before it goes on the air.

Callers' side of the same API: [search-api.md](search-api.md).

## Before you start

- **roger 7.2.0 or later.** Older copies do not have `roger share --search` or
  `roger search check`. Install or upgrade with:

  ```
  curl -fsSL https://rogerai.fm/install.sh | sh
  ```

  or `brew upgrade roger` if you installed it with Homebrew. `roger --version` shows the version.
- **A signed-in owner.** Every search station, free or paid, needs `roger login` first.
- **To be paid,** finish payout onboarding once with `roger payout onboard` (identity check
  through Stripe). `roger payout status` shows what you have earned and when it pays out.

`roger share --search <engine> --help` lists every flag.

## Three routes

| route | what answers | you need |
|---|---|---|
| Brave | Brave's Search API, with your own key | a Brave API key and its plan |
| SearXNG | a SearXNG instance you run, which asks Google, Bing, DuckDuckGo and others | SearXNG with JSON output turned on |
| native | your own index, served over the search/1 contract below | a service you built |

Every station is labeled with its route everywhere it is listed (`engine: brave`, `searxng`
or `native`), so callers can tell a resold engine from someone's own index.

### Brave

```
roger share --search brave --price-per-1k 6
```

roger reads the key from `--upstream-key`, then `BRAVE_SEARCH_API_KEY`, then asks for it with
input hidden. The key stays in your local roger config and is never sent to RogerAI. Brave
charges you per query, so roger makes you set a price; `--price-per-1k 0` is allowed if you
mean it.

### SearXNG

```
roger share --search searxng --upstream http://127.0.0.1:8888
```

SearXNG must allow JSON output. In its `settings.yml`, add `json` to `search.formats`. SearXNG
has no country filter, so your station is listed as ignoring `country`.

### Native: your own index

```
roger search check --upstream http://127.0.0.1:9000
roger share --search native --upstream http://127.0.0.1:9000 --scope "Go module docs"
```

A native station is any service that answers the search/1 contract. It can sit in front of
Meilisearch, Typesense, OpenSearch or Elasticsearch, a YaCy peer, or a crawler you wrote.
`--scope` is one line saying what your index covers, so callers can tell a general web index
from a focused one. It is cut to 120 characters, and the content screen can remove it: if it
does, your station still goes on air without a scope line.

`roger search check` is the first step for a native provider: run it against your service
until every check passes, then share it.

Before going on air, `roger share --search` sends your engine one fixed test query,
"weather". A focused index (a shop's products, a parts catalog) may have nothing for it: a
well-formed reply with an empty `results` array passes, and roger says
`check passed: 0 results for "weather" (fine for a focused index)`. A reply whose results all
lack an http or https link fails.

## The search/1 contract

roger sends:

```
POST /v1/search
Content-Type: application/json
X-Search-Contract: search/1
Authorization: Bearer <key>        (only when you set --upstream-key)

{"query": "tides", "count": 5, "freshness": "week", "country": "US"}
```

`freshness` (`day`, `week`, `month`, `year`) and `country` (two letters) are optional. Your
service answers within 10 s, with a reply under 512 KiB:

```json
{"results": [
  {"title": "Tides", "url": "https://example.com/tides",
   "snippet": "How the moon moves the sea", "published": "2026-11-01"}
]}
```

Any non-2xx status is a failure. Its body is never shown to callers.

`roger search check` runs these checks, by these names:

- check `endpoint`: POST /v1/search answers HTTP 200
- check `reply is JSON`: the reply is a JSON object
- check `results array`: the object has a `results` array
- check `respects count`: never more results than `count`
- check `http(s) urls only`: every `url` is an absolute http or https URL
- check `every result has a title`: no result has an empty title
- check `published is RFC 3339`: `published`, when present, is an RFC 3339 date or date-time
- check `answers within 10 s`: the reply arrives within 10 s
- check `reply under 512 KiB`: the reply is at most 512 KiB
- check `accepts freshness`: a query with `freshness` succeeds
- check `accepts Unicode`: a query with non-ASCII text succeeds

It also warns, without failing, when your service seems to ignore `freshness` or `country`.
It sends only fixed test queries.

A short example service, enough to pass the checks:

```python
from http.server import BaseHTTPRequestHandler, HTTPServer
import json

DOCS = [("Tides", "https://example.com/tides", "How the moon moves the sea")]

class Search(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/v1/search":
            self.send_response(404); self.end_headers(); return
        req = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        words = req["query"].lower().split()
        hits = [d for d in DOCS if any(w in (d[0] + " " + d[2]).lower() for w in words)]
        out = {"results": [{"title": t, "url": u, "snippet": s} for t, u, s in hits[: req["count"]]]}
        body = json.dumps(out).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body)

HTTPServer(("127.0.0.1", 9000), Search).serve_forever()
```

## What you promise

- **Do not log, keep or sell queries.** Callers are told their words reach your engine and
  nowhere else.
- **Do not try to identify callers.** You receive a pseudonym that is different for every
  station, and nothing else about who asked.
- **Brave and SearXNG providers answer for those engines' terms.** You serve results from
  someone else's engine to other people. Whether that engine's terms allow it is your
  responsibility. roger asks you once per engine; scripts pass `--accept-engine-terms`.

The broker also checks your replies. It rebuilds every reply from cleaned, size-limited fields,
drops anything that is not an http or https link, and screens the results before they reach a
caller. A station that returns child sexual abuse material receives a strike.

## What it can take to build an index

A focused index is realistic for one person or a small team: your docs, a niche corpus, a
dataset you curate. A general web index is a large job: a crawler that respects `robots.txt`,
storage, ranking, keeping results fresh and fighting spam. If you want general web search
without building it, run SearXNG.

## Price and payouts

You set your price in dollars per 1,000 queries with `--price-per-1k`. The ceiling is $50 per
1,000. A search with no station named only reaches stations at or under $10 per 1,000 unless
the caller raises its cap. Searches settle on the same ledger as everything else: the standard
platform fee, then the standard payout hold, minimum and schedule: a 30-day hold with 10% held to day 90, a $25 minimum, paid
monthly (`roger payout onboard` once, then `roger payout status`).
Every search station needs a signed-in owner, free or paid: run `roger login` before `roger share --search`.

Add `--private` to keep the station off every public listing. roger prints a frequency code
once, on your screen only; it is never shown again (if you lose it, `roger bands new-code` gives
the band a new one). Only requests that carry it in the `X-Roger-Freq` header reach the
station, and a paid private station still charges the caller's wallet. The free plan allows
one private band per account; a second private share is refused until you release the first.

A failed search, a reply the broker cannot use, or results the content screen withholds pay
nothing. A well-formed reply with zero results is paid.

Designed by RogerAI in Orange County, California.
