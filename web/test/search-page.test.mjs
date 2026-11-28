// /search.html, the search directory (2026-11-26): the voices page's sibling over the broker's public
// /search/stations. The roster's behaviour is test/unit/roster.test.js; this holds the page around it.
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { DIST, WEB } from "./_html.mjs";

const dist = (p) => readFileSync(path.join(DIST, p), "utf8");
const main = (p) => dist(p).replace(/<!--[\s\S]*?-->/g, "").match(/<main\b[\s\S]*<\/main>/)[0];
const words = (html) => html.replace(/<(script|style)\b[\s\S]*?<\/\1>/g, "").replace(/<[^>]+>/g, " ").replace(/\s+/g, " ");

test("search: the page prerenders the tuner, an honest loading table and the request, as one island", () => {
  const html = dist("search.html");
  assert.match(html, /<div data-island="search">/);
  assert.match(html, /<body data-onair="idle"/, "the body carries the rail offset, as on voices");
  assert.match(dist("voices.html"), /<body data-onair="idle"/);
  assert.match(html, /<script type="module" src="js\/islands\/search\.js[^"]*">/);
  const m = main("search.html");
  for (const id of ["rosterSearch", "rosterSort", "rosterPick", "rosterFree", "rosterList", "rosterCmd"]) assert.match(m, new RegExp(`id="${id}"`), id);
  assert.equal((m.match(/class="voice-skel"/g) || []).length, 4, "four skeleton rows until the first read");
  assert.doesNotMatch(m, /class="voice-row"(?! voice-row--head)[^>]*>\s*<span class="voice-cell/, "no station row is built in: rows come from the broker only");
  assert.match(words(m), /curl https:\/\/broker\.rogerai\.fm\/v1\/search -H "Authorization: Bearer \$ROGER_KEY" -d '\{"query": "red shoes", "count": 5\}'/);
});

test("search: the copy claims nothing the record does not hold", () => {
  const t = words(main("search.html"));
  assert.doesNotMatch(t, /\bBrave\b|SearXNG/i, "the static copy names no engine; the engine column is the broker's data");
  assert.doesNotMatch(t, /house-or|roger-search/, "no station is written into the page");
  assert.doesNotMatch(t, /\d[\d,.]* (queries|searches|users|stations) (a|per) (day|month)|revenue|earning/i, "no volume, user or revenue figure");
  assert.doesNotMatch(t.replace(/FIG\. \d+ - /g, ""), /—| - /, "no em dash and no spaced hyphen in new copy (the figure label is the site's own form)");
  assert.match(t, /real data only/);
});

test("search: it is reachable from the nav, the footer, its siblings, the sitemap and llms.txt", () => {
  for (const p of ["index.html", "models.html", "voices.html", "search.html"]) {
    assert.match(dist(p), /<a href="\/search\.html"[^>]*><b>Search<\/b>/, `${p}: the Models panel lists Search`);
    assert.match(dist(p), /<footer[\s\S]*<a href="\/search\.html">Search<\/a>/, `${p}: the footer lists Search`);
  }
  assert.match(dist("search.html"), /<a href="\/search\.html" aria-current="page"/, "the panel marks the page you are on");
  assert.match(main("voices.html"), /class="dir-sibling"[\s\S]*?href="\/search\.html"/);
  assert.match(main("search.html"), /class="dir-sibling"[\s\S]*?href="\/voices\.html"[\s\S]*?href="\/manual\.html#sSearch"/);
  assert.match(dist("sitemap.xml"), /https:\/\/rogerai\.fm\/search\.html/);
  assert.match(dist("llms.txt"), /search\.html/);
  assert.match(readFileSync(path.join(WEB, "app/routes/manual/+page.svelte"), "utf8"), /id="sSearch"/, "the manual section the page links to exists");
});

test("rosters: voices and search are one component, and the old page script is gone", () => {
  for (const p of ["voices", "search"]) assert.match(readFileSync(path.join(WEB, `app/islands/${p}.js`), "utf8"), new RegExp(`island\\("${p}", Roster, \\{ kind: "${p}" \\}\\)`));
  assert.throws(() => readFileSync(path.join(WEB, "src/js/voices.js")), "src/js/voices.js is removed");
  assert.match(dist("voices.html"), /<div data-island="voices">/);
});
